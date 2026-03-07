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

	"github.com/radardopovo/bolsa-familia-etl/internal/config"
	"github.com/radardopovo/bolsa-familia-etl/internal/db"
	"github.com/radardopovo/bolsa-familia-etl/internal/httpx"
	"github.com/radardopovo/bolsa-familia-etl/internal/zipx"
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

type PeriodStats struct {
	Period   int
	Read     int64
	Inserted int64
	Ignored  int64
	Duration time.Duration
	Skipped  bool
}

func New(cfg config.Config, log *slog.Logger, database *sql.DB, client *httpx.Client) *Orchestrator {
	return &Orchestrator{cfg: cfg, log: log, db: database, client: client}
}

func (o *Orchestrator) Run(ctx context.Context) error {
	periods := o.cfg.Periods(time.Now())
	if len(periods) == 0 {
		return fmt.Errorf("nenhum periodo para processar")
	}

	o.log.Info("import_started",
		slog.Int("from_period", periods[0]),
		slog.Int("to_period", periods[len(periods)-1]),
		slog.String("mode", o.cfg.ModeLabel()),
		slog.Int("batch", o.cfg.BatchSize),
		slog.Int("chunk", o.cfg.ChunkSize),
		slog.Int("writers", o.cfg.WriterWorkers),
		slog.Bool("dry_run", o.cfg.DryRun),
	)

	startAll := time.Now()
	var totalRead, totalInserted, totalIgnored int64
	var processed int

	for _, period := range periods {
		ps, err := o.processPeriod(ctx, period)
		if err != nil {
			return err
		}
		if ps.Skipped {
			continue
		}
		processed++
		totalRead += ps.Read
		totalInserted += ps.Inserted
		totalIgnored += ps.Ignored
	}

	totalDuration := time.Since(startAll)
	rowsPerSec := 0.0
	rowsPerMin := 0.0
	if totalDuration > 0 {
		rowsPerSec = float64(totalInserted) / totalDuration.Seconds()
		rowsPerMin = rowsPerSec * 60
	}

	o.log.Info("import_finished",
		slog.Int("periodos_processados", processed),
		slog.Int64("total_linhas", totalRead),
		slog.Int64("total_inseridas", totalInserted),
		slog.Int64("total_ignoradas", totalIgnored),
		slog.Duration("tempo_total", totalDuration),
		slog.Float64("velocidade_rows_s", rowsPerSec),
		slog.Float64("velocidade_rows_min", rowsPerMin),
	)
	return nil
}

func (o *Orchestrator) processPeriod(ctx context.Context, period int) (res PeriodStats, err error) {
	start := time.Now()
	res.Period = period

	if err := ensureDirs(o.dataDir(), o.zipsDir(), o.extractedDir()); err != nil {
		return res, err
	}

	var st *db.ImportState
	if o.db != nil && !o.cfg.DryRun {
		st, err = db.GetImportState(ctx, o.db, period)
		if err != nil {
			return res, err
		}
	}

	zipPath := o.zipPath(period)
	extractPath := o.extractPath(period)
	if st != nil && st.Status == "done" && !o.cfg.Force && !o.cfg.OnlyDownload {
		if localSHA, ok := fileSHA256IfExists(zipPath); ok && localSHA == st.ZipSHA256 {
			o.log.Info("period_skipped_cache", slog.Int("periodo", period), slog.String("zip_sha256", localSHA))
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
			o.log.Info("period_use_extracted_cache", slog.Int("periodo", period), slog.String("dir", extractPath))
		}
	}

	if o.cfg.OnlyImport {
		if localSHA, ok := fileSHA256IfExists(zipPath); ok {
			zipSHA = localSHA
		}
	} else if !haveExtracted {
		state := o.baseState(period, st, zipSHA)
		state.Status = "downloading"
		state.LastStep = "download"
		state.ErrorMsg = ""
		state.FinishedAt = nil
		if err := o.persistImportState(ctx, state); err != nil {
			return res, err
		}

		zipSHA, err = o.ensurePeriodZIP(ctx, period, zipPath, st)
		if err != nil {
			o.markPeriodError(ctx, state, "download", err)
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
		o.log.Info("period_reimport", slog.Int("periodo", period), slog.Bool("force", o.cfg.Force))
		if err := db.PurgePeriod(ctx, o.db, period); err != nil {
			return res, err
		}
	}
	if reimport {
		st = nil
	}

	if !haveExtracted {
		files, err = o.ensureExtracted(period, zipPath, extractPath)
		if err != nil {
			state := o.baseState(period, st, zipSHA)
			o.markPeriodError(ctx, state, "extract", err)
			return res, err
		}
	}

	state := o.baseState(period, st, zipSHA)
	state.ZipSHA256 = zipSHA
	state.Status = "extracted"
	state.LastStep = "extract"
	state.ErrorMsg = ""
	state.FinishedAt = nil
	if err := o.persistImportState(ctx, state); err != nil {
		return res, err
	}

	if o.cfg.OnlyDownload {
		o.log.Info("period_download_only_done", slog.Int("periodo", period), slog.String("zip_sha256", zipSHA))
		return res, nil
	}

	state.Status = "importing"
	state.LastStep = "bolsa_familia"
	state.ErrorMsg = ""
	if err := o.persistImportState(ctx, state); err != nil {
		return res, err
	}

	stats, runErr := o.importBolsaFamilia(ctx, period, files.BolsaFamilia)
	if runErr != nil {
		o.markPeriodError(ctx, state, "bolsa_familia", runErr)
		return res, runErr
	}

	state.RowsBolsaFamilia = stats.Inserted
	finishedAt := time.Now().UTC()
	state.Status = "done"
	state.LastStep = "bolsa_familia"
	state.ErrorMsg = ""
	state.FinishedAt = &finishedAt
	if err := o.persistImportState(ctx, state); err != nil {
		return res, err
	}

	res.Read = stats.Read
	res.Inserted = stats.Inserted
	res.Ignored = stats.Ignored
	res.Duration = time.Since(start)
	o.log.Info("period_done",
		slog.Int("periodo", period),
		slog.Duration("duracao", res.Duration),
		slog.Int64("lidas", res.Read),
		slog.Int64("inseridas", res.Inserted),
		slog.Int64("ignoradas", res.Ignored),
	)
	return res, nil
}

func (o *Orchestrator) baseState(period int, prev *db.ImportState, zipSHA string) db.ImportState {
	now := time.Now().UTC()
	if prev == nil {
		return db.ImportState{
			Period:           period,
			ZipSHA256:        zipSHA,
			Status:           "downloading",
			LastStep:         "download",
			RowsBolsaFamilia: 0,
			StartedAt:        &now,
			FinishedAt:       nil,
			ErrorMsg:         "",
		}
	}
	st := *prev
	st.Period = period
	if st.StartedAt == nil {
		st.StartedAt = &now
	}
	if zipSHA != "" {
		st.ZipSHA256 = zipSHA
	}
	return st
}

func (o *Orchestrator) markPeriodError(ctx context.Context, st db.ImportState, step string, runErr error) {
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

func (o *Orchestrator) dataDir() string { return o.cfg.DataDir }
func (o *Orchestrator) zipsDir() string { return filepath.Join(o.dataDir(), "zips") }
func (o *Orchestrator) extractedDir() string { return filepath.Join(o.dataDir(), "extracted") }

func (o *Orchestrator) zipPath(period int) string {
	return filepath.Join(o.zipsDir(), fmt.Sprintf("novo_bolsa_familia_%d.zip", period))
}

func (o *Orchestrator) extractPath(period int) string {
	return filepath.Join(o.extractedDir(), fmt.Sprintf("%d", period))
}

func ensureDirs(dirs ...string) error {
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func (o *Orchestrator) ensureExtracted(period int, zipPath, extractPath string) (zipx.CSVFiles, error) {
	if o.cfg.OnlyImport {
		files, err := detectExtractedCSVs(extractPath)
		if err == nil {
			return files, nil
		}
	}
	if _, err := os.Stat(zipPath); err != nil {
		return zipx.CSVFiles{}, fmt.Errorf("ZIP nao encontrado para o periodo %d em %s", period, zipPath)
	}
	files, err := extractCSVsFromZIP(zipPath, extractPath)
	if err != nil {
		return zipx.CSVFiles{}, err
	}
	o.log.Info("zip_extracted", slog.Int("periodo", period), slog.String("dir", extractPath))
	return files, nil
}
