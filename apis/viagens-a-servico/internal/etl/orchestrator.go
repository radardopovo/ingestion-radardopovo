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

	"github.com/radardopovo/viagens-etl/internal/config"
	"github.com/radardopovo/viagens-etl/internal/db"
	"github.com/radardopovo/viagens-etl/internal/httpx"
	"github.com/radardopovo/viagens-etl/internal/zipx"
)

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

type YearStats struct {
	Year     int
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
		slog.Int("from", o.cfg.FromYear),
		slog.Int("to", o.cfg.ToYear),
		slog.String("mode", o.cfg.ModeLabel()),
		slog.Int("batch", o.cfg.BatchSize),
		slog.Int("chunk", o.cfg.ChunkSize),
		slog.Int("writers", o.cfg.WriterWorkers),
		slog.Bool("dry_run", o.cfg.DryRun),
	)

	startAll := time.Now()
	var totalRead int64
	var totalInserted int64
	var totalIgnored int64
	var processed int

	for _, year := range o.cfg.Years() {
		ys, err := o.processYear(ctx, year)
		if err != nil {
			return err
		}
		if ys.Skipped {
			continue
		}
		processed++
		totalRead += ys.Read
		totalInserted += ys.Inserted
		totalIgnored += ys.Ignored
	}

	totalDuration := time.Since(startAll)
	rowsPerSec := 0.0
	rowsPerMin := 0.0
	if totalDuration > 0 {
		rowsPerSec = float64(totalInserted) / totalDuration.Seconds()
		rowsPerMin = rowsPerSec * 60
	}

	o.log.Info("import_finished",
		slog.Int("anos_processados", processed),
		slog.Int64("total_linhas", totalRead),
		slog.Int64("total_inseridas", totalInserted),
		slog.Int64("total_ignoradas", totalIgnored),
		slog.Duration("tempo_total", totalDuration),
		slog.Float64("velocidade_rows_s", rowsPerSec),
		slog.Float64("velocidade_rows_min", rowsPerMin),
	)
	return nil
}

func (o *Orchestrator) processYear(ctx context.Context, year int) (res YearStats, err error) {
	start := time.Now()
	res.Year = year

	if err := ensureDirs(o.dataDir(), o.zipsDir(), o.extractedDir()); err != nil {
		return res, err
	}

	var st *db.ImportState
	if o.db != nil && !o.cfg.DryRun {
		st, err = db.GetImportState(ctx, o.db, year)
		if err != nil {
			return res, err
		}
	}

	zipPath := o.zipPath(year)
	extractPath := o.extractPath(year)
	if st != nil && st.Status == "done" && !o.cfg.Force && !o.cfg.OnlyDownload {
		if localSHA, ok := fileSHA256IfExists(zipPath); ok && localSHA == st.ZipSHA256 {
			o.log.Info("year_skipped_cache",
				slog.Int("ano", year),
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

	var files zipx.CSVFiles
	haveExtracted := false
	if !o.cfg.Force && st != nil && (st.Status == "extracted" || st.Status == "importing" || st.Status == "error") {
		detected, derr := detectExtractedCSVs(extractPath)
		if derr == nil {
			files = detected
			haveExtracted = true
			o.log.Info("year_use_extracted_cache", slog.Int("ano", year), slog.String("dir", extractPath))
		}
	}

	if o.cfg.OnlyImport {
		if localSHA, ok := fileSHA256IfExists(zipPath); ok {
			zipSHA = localSHA
		}
	} else if !haveExtracted {
		state := o.baseState(year, st, zipSHA)
		state.Status = "downloading"
		state.LastStep = "download"
		state.ErrorMsg = ""
		state.FinishedAt = nil
		if err := o.persistImportState(ctx, state); err != nil {
			return res, err
		}

		zipSHA, err = o.ensureYearZIP(ctx, year, zipPath, st)
		if err != nil {
			o.markYearError(ctx, state, "download", err)
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
		o.log.Info("year_reimport", slog.Int("ano", year), slog.Bool("force", o.cfg.Force))
		if err := db.PurgeYear(ctx, o.db, year); err != nil {
			return res, err
		}
	}
	if reimport {
		st = nil
	}

	if !haveExtracted {
		files, err = o.ensureExtracted(year, zipPath, extractPath)
		if err != nil {
			state := o.baseState(year, st, zipSHA)
			o.markYearError(ctx, state, "extract", err)
			return res, err
		}
	}

	state := o.baseState(year, st, zipSHA)
	state.ZipSHA256 = zipSHA
	state.Status = "extracted"
	state.LastStep = "extract"
	state.ErrorMsg = ""
	state.FinishedAt = nil
	if err := o.persistImportState(ctx, state); err != nil {
		return res, err
	}

	if o.cfg.OnlyDownload {
		o.log.Info("year_download_only_done",
			slog.Int("ano", year),
			slog.String("zip_sha256", zipSHA),
		)
		return res, nil
	}

	startStep := o.resolveStartStep(st, reimport)
	steps := []struct {
		name string
		run  func(context.Context, int, string) (TableStats, error)
		file string
	}{
		{name: "viagem", run: o.importViagens, file: files.Viagem},
		{name: "trecho", run: o.importTrechos, file: files.Trecho},
		{name: "passagem", run: o.importPassagens, file: files.Passagem},
		{name: "pagamento", run: o.importPagamentos, file: files.Pagamento},
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

		stats, runErr := s.run(ctx, year, s.file)
		if runErr != nil {
			o.markYearError(ctx, state, s.name, runErr)
			return res, runErr
		}

		switch s.name {
		case "viagem":
			state.RowsViagem = stats.Inserted
		case "trecho":
			state.RowsTrecho = stats.Inserted
		case "passagem":
			state.RowsPassagem = stats.Inserted
		case "pagamento":
			state.RowsPagamento = stats.Inserted
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
	state.LastStep = "pagamento"
	state.ErrorMsg = ""
	state.FinishedAt = &finishedAt
	if err := o.persistImportState(ctx, state); err != nil {
		return res, err
	}

	res.Duration = time.Since(start)
	o.log.Info("year_done",
		slog.Int("ano", year),
		slog.Duration("duracao", res.Duration),
		slog.Int64("lidas", res.Read),
		slog.Int64("inseridas", res.Inserted),
		slog.Int64("ignoradas", res.Ignored),
	)
	return res, nil
}

func (o *Orchestrator) baseState(year int, prev *db.ImportState, zipSHA string) db.ImportState {
	now := time.Now().UTC()
	if prev == nil {
		return db.ImportState{
			Ano:           year,
			ZipSHA256:     zipSHA,
			Status:        "downloading",
			LastStep:      "download",
			RowsViagem:    0,
			RowsTrecho:    0,
			RowsPassagem:  0,
			RowsPagamento: 0,
			StartedAt:     &now,
			FinishedAt:    nil,
			ErrorMsg:      "",
		}
	}
	st := *prev
	st.Ano = year
	if st.StartedAt == nil {
		st.StartedAt = &now
	}
	if zipSHA != "" {
		st.ZipSHA256 = zipSHA
	}
	return st
}

func (o *Orchestrator) markYearError(ctx context.Context, st db.ImportState, step string, runErr error) {
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
		return "viagem"
	}
	switch prev.Status {
	case "importing":
		if next, ok := nextStep(prev.LastStep); ok {
			return next
		}
		if prev.LastStep == "pagamento" {
			return "pagamento"
		}
		return "viagem"
	case "error":
		if prev.LastStep == "" {
			return "viagem"
		}
		return prev.LastStep
	case "extracted":
		return "viagem"
	default:
		return "viagem"
	}
}

func nextStep(step string) (string, bool) {
	switch step {
	case "viagem":
		return "trecho", true
	case "trecho":
		return "passagem", true
	case "passagem":
		return "pagamento", true
	default:
		return "", false
	}
}

func stepIndex(step string) int {
	switch step {
	case "viagem":
		return 0
	case "trecho":
		return 1
	case "passagem":
		return 2
	case "pagamento":
		return 3
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

func (o *Orchestrator) zipPath(year int) string {
	return filepath.Join(o.zipsDir(), fmt.Sprintf("viagens_%d.zip", year))
}

func (o *Orchestrator) extractPath(year int) string {
	return filepath.Join(o.extractedDir(), fmt.Sprintf("%d", year))
}

func ensureDirs(dirs ...string) error {
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func (o *Orchestrator) ensureExtracted(year int, zipPath, extractPath string) (zipx.CSVFiles, error) {
	if o.cfg.OnlyImport {
		files, err := detectExtractedCSVs(extractPath)
		if err == nil {
			return files, nil
		}
	}
	if _, err := os.Stat(zipPath); err != nil {
		return zipx.CSVFiles{}, fmt.Errorf("ZIP nao encontrado para o ano %d em %s", year, zipPath)
	}
	files, err := extractCSVsFromZIP(zipPath, extractPath)
	if err != nil {
		return zipx.CSVFiles{}, err
	}
	o.log.Info("zip_extracted", slog.Int("ano", year), slog.String("dir", extractPath))
	return files, nil
}
