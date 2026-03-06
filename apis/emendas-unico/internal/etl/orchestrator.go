// Radar do Povo ETL - https://radardopovo.com
package etl

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/radardopovo/emendas-etl/internal/config"
	"github.com/radardopovo/emendas-etl/internal/db"
	"github.com/radardopovo/emendas-etl/internal/httpx"
	"github.com/radardopovo/emendas-etl/internal/zipx"
)

const datasetKey = "emendas_unico"

type Orchestrator struct {
	cfg    config.Config
	log    *slog.Logger
	db     *sql.DB
	client *httpx.Client
}

type TableStats struct {
	Read     int64
	Inserted int64
	Ignored  int64
	Duration time.Duration
}

type DatasetStats struct {
	Read     int64
	Inserted int64
	Ignored  int64
	Duration time.Duration
	Skipped  bool
}

func New(cfg config.Config, log *slog.Logger, database *sql.DB, client *httpx.Client) *Orchestrator {
	return &Orchestrator{
		cfg:    cfg,
		log:    log,
		db:     database,
		client: client,
	}
}

func (o *Orchestrator) Run(ctx context.Context) error {
	o.log.Info("import_started",
		slog.String("dataset", datasetKey),
		slog.String("mode", o.cfg.ModeLabel()),
		slog.Int("batch", o.cfg.BatchSize),
		slog.Int("chunk", o.cfg.ChunkSize),
		slog.Int("writers", o.cfg.WriterWorkers),
		slog.Bool("dry_run", o.cfg.DryRun),
	)

	startAll := time.Now()
	st, err := o.processDataset(ctx)
	if err != nil {
		return err
	}

	totalDuration := time.Since(startAll)
	rowsPerSec := 0.0
	rowsPerMin := 0.0
	if totalDuration > 0 {
		rowsPerSec = float64(st.Inserted) / totalDuration.Seconds()
		rowsPerMin = rowsPerSec * 60
	}

	o.log.Info("import_finished",
		slog.String("dataset", datasetKey),
		slog.Bool("skipped", st.Skipped),
		slog.Int64("total_linhas", st.Read),
		slog.Int64("total_inseridas", st.Inserted),
		slog.Int64("total_ignoradas", st.Ignored),
		slog.Duration("tempo_total", totalDuration),
		slog.Float64("velocidade_rows_s", rowsPerSec),
		slog.Float64("velocidade_rows_min", rowsPerMin),
	)
	return nil
}

func (o *Orchestrator) processDataset(ctx context.Context) (res DatasetStats, err error) {
	start := time.Now()
	if err := ensureDirs(o.dataDir(), o.zipsDir(), o.extractedDir()); err != nil {
		return res, err
	}

	var st *db.ImportState
	if o.db != nil && !o.cfg.DryRun {
		st, err = db.GetImportState(ctx, o.db, datasetKey)
		if err != nil {
			return res, err
		}
	}

	zipPath := o.zipPath()
	extractPath := o.extractPath()
	if st != nil && st.Status == "done" && !o.cfg.Force && !o.cfg.OnlyDownload {
		if localSHA, ok := fileSHA256IfExists(zipPath); ok && localSHA == st.ZipSHA256 {
			o.log.Info("dataset_skipped_cache",
				slog.String("dataset", datasetKey),
				slog.String("zip_sha256", localSHA),
			)
			res.Skipped = true
			return res, nil
		}
	}

	zipSHA := ""
	if st != nil && st.ZipSHA256 != "" {
		zipSHA = st.ZipSHA256
	}

	var files zipx.EmendasCSVFiles
	haveExtracted := false
	if !o.cfg.Force && st != nil && (st.Status == "extracted" || st.Status == "importing" || st.Status == "error") {
		detected, derr := detectExtractedCSVs(extractPath)
		if derr == nil {
			files = detected
			haveExtracted = true
			o.log.Info("dataset_use_extracted_cache", slog.String("dataset", datasetKey), slog.String("dir", extractPath))
		}
	}

	if o.cfg.OnlyImport {
		if localSHA, ok := fileSHA256IfExists(zipPath); ok {
			zipSHA = localSHA
		}
	} else if !haveExtracted {
		state := o.baseState(st, zipSHA)
		state.Status = "downloading"
		state.LastStep = "download"
		state.ErrorMsg = ""
		state.FinishedAt = nil
		if err := o.persistImportState(ctx, state); err != nil {
			return res, err
		}

		zipSHA, err = o.ensureDatasetZIP(ctx, zipPath, st)
		if err != nil {
			o.markDatasetError(ctx, state, "download", err)
			return res, err
		}
	}
	if zipSHA == "" {
		if localSHA, ok := fileSHA256IfExists(zipPath); ok {
			zipSHA = localSHA
		}
	}

	reimport := o.cfg.Force || (st != nil && st.Status == "done" && st.ZipSHA256 != "" && zipSHA != "" && st.ZipSHA256 != zipSHA)
	if reimport && o.db != nil && !o.cfg.DryRun {
		o.log.Info("dataset_reimport", slog.String("dataset", datasetKey), slog.Bool("force", o.cfg.Force))
		if err := db.PurgeDataset(ctx, o.db); err != nil {
			return res, err
		}
	}
	if reimport {
		st = nil
	}

	if !haveExtracted {
		files, err = o.ensureExtracted(zipPath, extractPath)
		if err != nil {
			state := o.baseState(st, zipSHA)
			o.markDatasetError(ctx, state, "extract", err)
			return res, err
		}
	}

	state := o.baseState(st, zipSHA)
	state.ZipSHA256 = zipSHA
	state.Status = "extracted"
	state.LastStep = "extract"
	state.ErrorMsg = ""
	state.FinishedAt = nil
	if err := o.persistImportState(ctx, state); err != nil {
		return res, err
	}

	if o.cfg.OnlyDownload {
		o.log.Info("dataset_download_only_done",
			slog.String("dataset", datasetKey),
			slog.String("zip_sha256", zipSHA),
		)
		return res, nil
	}

	startStep := o.resolveStartStep(st, reimport)
	steps := []struct {
		name string
		run  func(context.Context, string) (TableStats, error)
		file string
	}{
		{name: "emendas", run: o.importEmendas, file: files.Emendas},
		{name: "favorecido", run: o.importEmendasPorFavorecido, file: files.PorFavorecido},
		{name: "convenio", run: o.importEmendasConvenios, file: files.Convenios},
	}

	startIdx := stepIndex(startStep)
	if startIdx < 0 {
		startIdx = 0
	}

	for i := startIdx; i < len(steps); i++ {
		s := steps[i]
		state.Status = "importing"
		state.ErrorMsg = ""
		if err := o.persistImportState(ctx, state); err != nil {
			return res, err
		}

		stats, runErr := s.run(ctx, s.file)
		if runErr != nil {
			o.markDatasetError(ctx, state, s.name, runErr)
			return res, runErr
		}

		switch s.name {
		case "emendas":
			state.RowsEmendas = stats.Inserted
		case "favorecido":
			state.RowsFavorecido = stats.Inserted
		case "convenio":
			state.RowsConvenio = stats.Inserted
		}
		state.LastStep = s.name
		if err := o.persistImportState(ctx, state); err != nil {
			return res, err
		}

		res.Read += stats.Read
		res.Inserted += stats.Inserted
		res.Ignored += stats.Ignored
	}

	finishedAt := time.Now().UTC()
	state.Status = "done"
	state.LastStep = "convenio"
	state.ErrorMsg = ""
	state.FinishedAt = &finishedAt
	if err := o.persistImportState(ctx, state); err != nil {
		return res, err
	}

	res.Duration = time.Since(start)
	o.log.Info("dataset_done",
		slog.String("dataset", datasetKey),
		slog.Duration("duracao", res.Duration),
		slog.Int64("lidas", res.Read),
		slog.Int64("inseridas", res.Inserted),
		slog.Int64("ignoradas", res.Ignored),
	)
	return res, nil
}

func (o *Orchestrator) baseState(prev *db.ImportState, zipSHA string) db.ImportState {
	now := time.Now().UTC()
	if prev == nil {
		return db.ImportState{
			DatasetKey:     datasetKey,
			ZipSHA256:      zipSHA,
			Status:         "downloading",
			LastStep:       "download",
			RowsEmendas:    0,
			RowsFavorecido: 0,
			RowsConvenio:   0,
			StartedAt:      &now,
			FinishedAt:     nil,
			ErrorMsg:       "",
		}
	}
	st := *prev
	if st.StartedAt == nil {
		st.StartedAt = &now
	}
	if zipSHA != "" {
		st.ZipSHA256 = zipSHA
	}
	return st
}

func (o *Orchestrator) markDatasetError(ctx context.Context, st db.ImportState, step string, runErr error) {
	st.Status = "error"
	st.LastStep = step
	st.ErrorMsg = runErr.Error()
	_ = o.persistImportState(ctx, st)
}

func (o *Orchestrator) persistImportState(ctx context.Context, st db.ImportState) error {
	if o.db == nil || o.cfg.DryRun {
		return nil
	}
	return db.UpsertImportState(ctx, o.db, st)
}

func (o *Orchestrator) resolveStartStep(prev *db.ImportState, reimport bool) string {
	if reimport || prev == nil {
		return "emendas"
	}
	switch prev.Status {
	case "importing":
		if next, ok := nextStep(prev.LastStep); ok {
			return next
		}
		if prev.LastStep == "convenio" {
			return "convenio"
		}
		return "emendas"
	case "error":
		if prev.LastStep == "" {
			return "emendas"
		}
		return prev.LastStep
	case "extracted":
		return "emendas"
	default:
		return "emendas"
	}
}

func nextStep(step string) (string, bool) {
	switch step {
	case "emendas":
		return "favorecido", true
	case "favorecido":
		return "convenio", true
	default:
		return "", false
	}
}

func stepIndex(step string) int {
	switch step {
	case "emendas":
		return 0
	case "favorecido":
		return 1
	case "convenio":
		return 2
	default:
		return -1
	}
}

func (o *Orchestrator) dataDir() string {
	return o.cfg.DataDir
}

func (o *Orchestrator) zipsDir() string {
	return filepath.Join(o.dataDir(), "zips")
}

func (o *Orchestrator) extractedDir() string {
	return filepath.Join(o.dataDir(), "extracted")
}

func (o *Orchestrator) zipPath() string {
	return filepath.Join(o.zipsDir(), "emendas_unico.zip")
}

func (o *Orchestrator) extractPath() string {
	return filepath.Join(o.extractedDir(), "emendas_unico")
}

func ensureDirs(dirs ...string) error {
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func (o *Orchestrator) ensureExtracted(zipPath, extractPath string) (zipx.EmendasCSVFiles, error) {
	if o.cfg.OnlyImport {
		files, err := detectExtractedCSVs(extractPath)
		if err == nil {
			return files, nil
		}
	}
	if _, err := os.Stat(zipPath); err != nil {
		return zipx.EmendasCSVFiles{}, fmt.Errorf("ZIP nao encontrado em %s", zipPath)
	}
	files, err := extractCSVsFromZIP(zipPath, extractPath)
	if err != nil {
		return zipx.EmendasCSVFiles{}, err
	}
	o.log.Info("zip_extracted", slog.String("dataset", datasetKey), slog.String("dir", extractPath))
	return files, nil
}
