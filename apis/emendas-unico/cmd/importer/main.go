// Radar do Povo ETL - https://radardopovo.com
package main

import (
	"context"
	"database/sql"
	stdlog "log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/radardopovo/emendas-etl/internal/config"
	"github.com/radardopovo/emendas-etl/internal/db"
	"github.com/radardopovo/emendas-etl/internal/etl"
	"github.com/radardopovo/emendas-etl/internal/httpx"
	"github.com/radardopovo/emendas-etl/internal/logger"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load(os.Args[1:], time.Now())
	if err != nil {
		stdlog.Fatal(err)
	}

	runID := time.Now().UTC().Format("20060102T150405.000000000")
	log := logger.New(cfg.Verbose, cfg.LogJSON).With("run_id", runID)

	client := httpx.NewClient(time.Duration(cfg.HTTPTimeoutSec)*time.Second, cfg.HTTPMaxRetries, log)

	var sqlDB *sql.DB
	if !cfg.DryRun {
		sqlDB, err = db.Connect(ctx, cfg)
		if err != nil {
			log.Error("db_connect_failed", "error", err)
			os.Exit(1)
		}
		defer sqlDB.Close()

		if err := db.Migrate(ctx, sqlDB); err != nil {
			log.Error("db_migrate_failed", "error", err)
			os.Exit(1)
		}
	} else {
		log.Info("dry_run_enabled")
	}

	orch := etl.New(cfg, log, sqlDB, client)
	if err := orch.Run(ctx); err != nil {
		log.Error("import_failed", "error", err)
		os.Exit(1)
	}
}
