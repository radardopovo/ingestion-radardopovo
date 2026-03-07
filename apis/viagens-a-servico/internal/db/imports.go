// Radar do Povo ETL - https://radardopovo.com
package db

import (
	"context"
	"database/sql"
	"time"
)

type ImportState struct {
	Ano           int
	ZipSHA256     string
	Status        string
	LastStep      string
	RowsViagem    int64
	RowsTrecho    int64
	RowsPassagem  int64
	RowsPagamento int64
	StartedAt     *time.Time
	FinishedAt    *time.Time
	ErrorMsg      string
}

func GetImportState(ctx context.Context, db *sql.DB, year int) (*ImportState, error) {
	const q = `SELECT ano, zip_sha256, status, last_step, rows_viagem, rows_trecho, rows_passagem, rows_pagamento, started_at, finished_at, error_msg
	FROM imports WHERE ano = $1`
	row := db.QueryRowContext(ctx, q, year)

	var st ImportState
	var lastStep sql.NullString
	var startedAt sql.NullTime
	var finishedAt sql.NullTime
	var errMsg sql.NullString

	err := row.Scan(
		&st.Ano,
		&st.ZipSHA256,
		&st.Status,
		&lastStep,
		&st.RowsViagem,
		&st.RowsTrecho,
		&st.RowsPassagem,
		&st.RowsPagamento,
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
	INSERT INTO imports (
		ano, zip_sha256, status, last_step,
		rows_viagem, rows_trecho, rows_passagem, rows_pagamento,
		started_at, finished_at, error_msg
	) VALUES (
		$1, $2, $3, $4,
		$5, $6, $7, $8,
		$9, $10, $11
	)
	ON CONFLICT (ano) DO UPDATE SET
		zip_sha256 = EXCLUDED.zip_sha256,
		status = EXCLUDED.status,
		last_step = EXCLUDED.last_step,
		rows_viagem = EXCLUDED.rows_viagem,
		rows_trecho = EXCLUDED.rows_trecho,
		rows_passagem = EXCLUDED.rows_passagem,
		rows_pagamento = EXCLUDED.rows_pagamento,
		started_at = EXCLUDED.started_at,
		finished_at = EXCLUDED.finished_at,
		error_msg = EXCLUDED.error_msg`

	_, err := db.ExecContext(ctx, q,
		st.Ano,
		st.ZipSHA256,
		st.Status,
		nullIfEmpty(st.LastStep),
		st.RowsViagem,
		st.RowsTrecho,
		st.RowsPassagem,
		st.RowsPagamento,
		st.StartedAt,
		st.FinishedAt,
		nullIfEmpty(st.ErrorMsg),
	)
	return err
}

func PurgeYear(ctx context.Context, db *sql.DB, year int) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := SetupBulkTransaction(ctx, tx); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM pagamentos WHERE ano = $1`, year); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM passagens WHERE ano = $1`, year); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM trechos WHERE ano = $1`, year); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM viagens WHERE ano = $1`, year); err != nil {
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
