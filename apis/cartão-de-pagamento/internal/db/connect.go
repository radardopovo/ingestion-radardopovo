// Radar do Povo ETL - https://radardopovo.com
package db

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/radardopovo/cpgf-etl/internal/config"
	_ "github.com/lib/pq"
)

func Connect(ctx context.Context, cfg config.Config) (*sql.DB, error) {
	caCert, err := os.ReadFile(cfg.DBSSLCA)
	if err != nil {
		return nil, fmt.Errorf("falha lendo DB_SSL_CA: %w", err)
	}
	pool := x509.NewCertPool()
	if ok := pool.AppendCertsFromPEM(caCert); !ok {
		return nil, fmt.Errorf("CA invalido em DB_SSL_CA: %s", cfg.DBSSLCA)
	}
	tlsCfg := &tls.Config{
		RootCAs:    pool,
		MinVersion: tls.VersionTLS12,
		ServerName: cfg.DBHost,
	}
	_ = tlsCfg

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=verify-full sslrootcert=%s TimeZone=UTC connect_timeout=10 application_name=%s",
		pqQuote(cfg.DBHost),
		cfg.DBPort,
		pqQuote(cfg.DBUser),
		pqQuote(cfg.DBPass),
		pqQuote(cfg.DBName),
		pqQuote(cfg.DBSSLCA),
		pqQuote("radardopovo-cpgf-etl"),
	)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxLifetime(cfg.DBConnMaxLifetime())
	db.SetConnMaxIdleTime(1 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		db.Close()
		return nil, fmt.Errorf("falha no ping do banco: %w", err)
	}

	if err := SetupSession(pingCtx, db); err != nil {
		db.Close()
		return nil, fmt.Errorf("falha no setup da sessao: %w", err)
	}
	return db, nil
}

type execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func SetupSession(ctx context.Context, ex execer) error {
	if _, err := ex.ExecContext(ctx, "SET TIME ZONE 'UTC'"); err != nil {
		return err
	}
	if _, err := ex.ExecContext(ctx, "SET client_encoding = 'UTF8'"); err != nil {
		return err
	}
	if _, err := ex.ExecContext(ctx, "SET idle_in_transaction_session_timeout = '60s'"); err != nil {
		return err
	}
	return nil
}

func SetupBulkTransaction(ctx context.Context, tx *sql.Tx) error {
	if err := SetupSession(ctx, tx); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "SET synchronous_commit = off"); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "SET work_mem = '64MB'"); err != nil {
		return err
	}
	return nil
}

func pqQuote(v string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `'`, `\'`)
	return "'" + replacer.Replace(v) + "'"
}
