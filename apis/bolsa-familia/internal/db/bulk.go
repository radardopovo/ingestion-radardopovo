// Radar do Povo ETL - https://radardopovo.com
package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/lib/pq"
)

type BulkTableSpec struct {
	Table                   string
	Columns                 []string
	ConflictTarget          string
	UpsertColumns           []string
	CountConflictsAsIgnored bool
}

func BeginBulkTx(ctx context.Context, db *sql.DB) (*sql.Tx, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	if err := SetupBulkTransaction(ctx, tx); err != nil {
		tx.Rollback()
		return nil, err
	}
	return tx, nil
}

func InsertChunk(ctx context.Context, tx *sql.Tx, spec BulkTableSpec, rows [][]any, batchSize int, verbose bool, log *slog.Logger) (int64, error) {
	if len(rows) == 0 {
		return 0, nil
	}
	maxRowsPerStmt := safeInsertBatchSize(batchSize, len(spec.Columns))
	var total int64
	for start := 0; start < len(rows); start += maxRowsPerStmt {
		end := start + maxRowsPerStmt
		if end > len(rows) {
			end = len(rows)
		}
		query, args := buildInsertSQL(spec, rows[start:end])
		if verbose && log != nil {
			log.Debug("insert batch", slog.String("table", spec.Table), slog.Int("rows", end-start), slog.String("sql", query))
		}
		res, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return total, err
		}
		aff, err := res.RowsAffected()
		if err != nil {
			return total, err
		}
		total += aff
	}
	return total, nil
}

func safeInsertBatchSize(requested, cols int) int {
	if requested <= 0 {
		requested = 1
	}
	if cols <= 0 {
		return requested
	}
	// PostgreSQL limite de parametros por statement.
	const maxParams = 65535
	limit := maxParams / cols
	if limit < 1 {
		limit = 1
	}
	if requested > limit {
		return limit
	}
	return requested
}

func CopyChunk(ctx context.Context, tx *sql.Tx, spec BulkTableSpec, rows [][]any, verbose bool, log *slog.Logger) (int64, error) {
	if len(rows) == 0 {
		return 0, nil
	}
	tmp := fmt.Sprintf("tmp_%s_%d", spec.Table, time.Now().UnixNano())
	createTmp := fmt.Sprintf(
		`CREATE TEMP TABLE %s ON COMMIT DROP AS SELECT %s FROM %s WHERE 1=0`,
		quoteIdent(tmp),
		joinIdentifiers(spec.Columns),
		quoteIdent(spec.Table),
	)
	if _, err := tx.ExecContext(ctx, createTmp); err != nil {
		return 0, err
	}

	stmt, err := tx.PrepareContext(ctx, pq.CopyIn(tmp, spec.Columns...))
	if err != nil {
		return 0, err
	}
	for _, r := range rows {
		if _, err := stmt.ExecContext(ctx, r...); err != nil {
			stmt.Close()
			return 0, err
		}
	}
	if _, err := stmt.ExecContext(ctx); err != nil {
		stmt.Close()
		return 0, err
	}
	if err := stmt.Close(); err != nil {
		return 0, err
	}

	mergeSQL := buildMergeSQL(spec, tmp)
	if verbose && log != nil {
		log.Debug("copy merge", slog.String("table", spec.Table), slog.Int("rows", len(rows)), slog.String("sql", mergeSQL))
	}
	res, err := tx.ExecContext(ctx, mergeSQL)
	if err != nil {
		return 0, err
	}
	inserted, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return inserted, nil
}

func buildInsertSQL(spec BulkTableSpec, rows [][]any) (string, []any) {
	var sb strings.Builder
	args := make([]any, 0, len(rows)*len(spec.Columns))
	sb.WriteString("INSERT INTO ")
	sb.WriteString(quoteIdent(spec.Table))
	sb.WriteString(" (")
	sb.WriteString(joinIdentifiers(spec.Columns))
	sb.WriteString(") VALUES ")
	argPos := 1
	for i, row := range rows {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString("(")
		for j := range row {
			if j > 0 {
				sb.WriteString(",")
			}
			sb.WriteString(fmt.Sprintf("$%d", argPos))
			argPos++
		}
		sb.WriteString(")")
		args = append(args, row...)
	}
	sb.WriteString(" ")
	sb.WriteString(conflictClause(spec))
	return sb.String(), args
}

func buildMergeSQL(spec BulkTableSpec, tempTable string) string {
	return fmt.Sprintf(
		`INSERT INTO %s (%s)
		SELECT %s FROM %s
		%s`,
		quoteIdent(spec.Table),
		joinIdentifiers(spec.Columns),
		joinIdentifiers(spec.Columns),
		quoteIdent(tempTable),
		conflictClause(spec),
	)
}

func conflictClause(spec BulkTableSpec) string {
	if len(spec.UpsertColumns) == 0 {
		return fmt.Sprintf("ON CONFLICT (%s) DO NOTHING", quoteIdent(spec.ConflictTarget))
	}
	sets := make([]string, 0, len(spec.UpsertColumns))
	for _, c := range spec.UpsertColumns {
		sets = append(sets, fmt.Sprintf("%s = EXCLUDED.%s", quoteIdent(c), quoteIdent(c)))
	}
	return fmt.Sprintf("ON CONFLICT (%s) DO UPDATE SET %s", quoteIdent(spec.ConflictTarget), strings.Join(sets, ", "))
}

func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

func joinIdentifiers(cols []string) string {
	out := make([]string, len(cols))
	for i, c := range cols {
		out[i] = quoteIdent(c)
	}
	return strings.Join(out, ",")
}
