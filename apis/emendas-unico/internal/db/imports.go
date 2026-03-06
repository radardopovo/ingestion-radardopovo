// Radar do Povo ETL - https://radardopovo.com
package db

import (
	"context"
	"database/sql"
	"time"
)

type ImportState struct {
	DatasetKey    string
	ZipSHA256     string
	Status        string
	LastStep      string
	RowsEmendas   int64
	RowsFavorecido int64
	RowsConvenio  int64
	StartedAt     *time.Time
	FinishedAt    *time.Time
	ErrorMsg      string
}

func GetImportState(ctx context.Context, db *sql.DB, datasetKey string) (*ImportState, error) {
	const q = `SELECT dataset_key, zip_sha256, status, last_step, rows_emendas, rows_favorecido, rows_convenio, started_at, finished_at, error_msg
	FROM imports WHERE dataset_key = $1`
	row := db.QueryRowContext(ctx, q, datasetKey)

	var st ImportState
	var lastStep sql.NullString
	var startedAt sql.NullTime
	var finishedAt sql.NullTime
	var errMsg sql.NullString

	err := row.Scan(
		&st.DatasetKey,
		&st.ZipSHA256,
		&st.Status,
		&lastStep,
		&st.RowsEmendas,
		&st.RowsFavorecido,
		&st.RowsConvenio,
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
		dataset_key, zip_sha256, status, last_step,
		rows_emendas, rows_favorecido, rows_convenio,
		started_at, finished_at, error_msg
	) VALUES (
		$1, $2, $3, $4,
		$5, $6, $7,
		$8, $9, $10
	)
	ON CONFLICT (dataset_key) DO UPDATE SET
		zip_sha256 = EXCLUDED.zip_sha256,
		status = EXCLUDED.status,
		last_step = EXCLUDED.last_step,
		rows_emendas = EXCLUDED.rows_emendas,
		rows_favorecido = EXCLUDED.rows_favorecido,
		rows_convenio = EXCLUDED.rows_convenio,
		started_at = EXCLUDED.started_at,
		finished_at = EXCLUDED.finished_at,
		error_msg = EXCLUDED.error_msg`

	_, err := db.ExecContext(ctx, q,
		st.DatasetKey,
		st.ZipSHA256,
		st.Status,
		nullIfEmpty(st.LastStep),
		st.RowsEmendas,
		st.RowsFavorecido,
		st.RowsConvenio,
		st.StartedAt,
		st.FinishedAt,
		nullIfEmpty(st.ErrorMsg),
	)
	return err
}

func PurgeDataset(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := SetupBulkTransaction(ctx, tx); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM emendas_convenios`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM emendas_por_favorecido`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM emendas`); err != nil {
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
