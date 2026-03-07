// Radar do Povo ETL - https://radardopovo.com
package db

import (
	"context"
	"database/sql"
	"time"
)

type ImportState struct {
	Period    int
	ZipSHA256 string
	Status    string
	LastStep  string
	RowsCPGF  int64
	StartedAt *time.Time
	FinishedAt *time.Time
	ErrorMsg  string
}

func GetImportState(ctx context.Context, db *sql.DB, period int) (*ImportState, error) {
	const q = `SELECT periodo_yyyymm, zip_sha256, status, last_step, rows_cpgf, started_at, finished_at, error_msg
	FROM imports_cpgf WHERE periodo_yyyymm = $1`
	row := db.QueryRowContext(ctx, q, period)

	var st ImportState
	var lastStep sql.NullString
	var startedAt sql.NullTime
	var finishedAt sql.NullTime
	var errMsg sql.NullString

	err := row.Scan(
		&st.Period,
		&st.ZipSHA256,
		&st.Status,
		&lastStep,
		&st.RowsCPGF,
		&startedAt,
		&finishedAt,
		&errMsg,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if lastStep.Valid {
		st.LastStep = lastStep.String
	}
	if startedAt.Valid {
		t := startedAt.Time
		st.StartedAt = &t
	}
	if finishedAt.Valid {
		t := finishedAt.Time
		st.FinishedAt = &t
	}
	if errMsg.Valid {
		st.ErrorMsg = errMsg.String
	}
	return &st, nil
}

func UpsertImportState(ctx context.Context, db *sql.DB, st ImportState) error {
	const q = `
	INSERT INTO imports_cpgf (
		periodo_yyyymm, zip_sha256, status, last_step,
		rows_cpgf, started_at, finished_at, error_msg
	) VALUES (
		$1, $2, $3, $4,
		$5, $6, $7, $8
	)
	ON CONFLICT (periodo_yyyymm) DO UPDATE SET
		zip_sha256 = EXCLUDED.zip_sha256,
		status = EXCLUDED.status,
		last_step = EXCLUDED.last_step,
		rows_cpgf = EXCLUDED.rows_cpgf,
		started_at = EXCLUDED.started_at,
		finished_at = EXCLUDED.finished_at,
		error_msg = EXCLUDED.error_msg`

	_, err := db.ExecContext(ctx, q,
		st.Period,
		st.ZipSHA256,
		st.Status,
		nullIfEmpty(st.LastStep),
		st.RowsCPGF,
		st.StartedAt,
		st.FinishedAt,
		nullIfEmpty(st.ErrorMsg),
	)
	return err
}

func PurgePeriod(ctx context.Context, db *sql.DB, period int) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := SetupBulkTransaction(ctx, tx); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM cpgf WHERE periodo_yyyymm = $1`, period); err != nil {
		return err
	}
	return tx.Commit()
}

func nullIfEmpty(v string) any {
	if v == "" {
		return nil
	}
	return v
}
