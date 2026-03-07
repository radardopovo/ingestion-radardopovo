// Radar do Povo ETL - https://radardopovo.com
package etl

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/radardopovo/cpgf-etl/internal/csvx"
	"github.com/radardopovo/cpgf-etl/internal/db"
)

type rowParser func(get fieldGetter) ([]any, bool)
type fieldGetter func(names ...string) string

func (o *Orchestrator) runCSVImport(
	ctx context.Context,
	period int,
	label string,
	csvPath string,
	spec db.BulkTableSpec,
	requiredHeaders []string,
	parser rowParser,
) (TableStats, error) {
	start := time.Now()
	stream, err := csvx.Open(csvPath)
	if err != nil {
		return TableStats{}, err
	}
	defer stream.Close()

	if err := stream.Require(requiredHeaders...); err != nil {
		return TableStats{}, fmt.Errorf("[%d-%s] %w", period, label, err)
	}

	writerWorkers := o.cfg.WriterWorkers
	if writerWorkers < 1 {
		writerWorkers = 1
	}

	ctxRun, cancel := context.WithCancel(ctx)
	defer cancel()

	rows := make(chan []any, o.cfg.BatchSize*4*writerWorkers)
	errCh := make(chan error, writerWorkers+1)
	done := make(chan struct{})

	var readCount atomic.Int64
	var insertedCount atomic.Int64
	var parseIgnoredCount atomic.Int64
	var conflictIgnoredCount atomic.Int64

	sendErr := func(e error) {
		if e == nil {
			return
		}
		select {
		case errCh <- e:
		default:
		}
		cancel()
	}

	var wg sync.WaitGroup
	wg.Add(1 + writerWorkers)

	go func() {
		defer wg.Done()
		defer close(rows)
		for {
			select {
			case <-ctxRun.Done():
				return
			default:
			}

			record, e := stream.Reader.Read()
			if e == io.EOF {
				return
			}
			if e != nil {
				sendErr(e)
				return
			}
			readCount.Add(1)
			get := makeGetter(stream, record)
			vals, ok := parser(get)
			if !ok {
				parseIgnoredCount.Add(1)
				continue
			}
			select {
			case rows <- vals:
			case <-ctxRun.Done():
				return
			}
		}
	}()

	useInsertMode := o.cfg.InsertMode
	for i := 0; i < writerWorkers; i++ {
		go func() {
			defer wg.Done()
			chunk := make([][]any, 0, o.cfg.ChunkSize)

			flush := func() error {
				if len(chunk) == 0 {
					return nil
				}
				var inserted int64
				if o.cfg.DryRun {
					inserted = int64(len(chunk))
				} else {
					tx, e := db.BeginBulkTx(ctxRun, o.db)
					if e != nil {
						return e
					}
					committed := false
					defer func() {
						if !committed {
							_ = tx.Rollback()
						}
					}()
					if useInsertMode {
						inserted, e = db.InsertChunk(ctxRun, tx, spec, chunk, o.cfg.BatchSize, o.cfg.Verbose, o.log)
					} else {
						inserted, e = db.CopyChunk(ctxRun, tx, spec, chunk, o.cfg.Verbose, o.log)
					}
					if e != nil {
						return e
					}
					if e := tx.Commit(); e != nil {
						return e
					}
					committed = true
				}
				insertedCount.Add(inserted)
				if spec.CountConflictsAsIgnored && inserted < int64(len(chunk)) {
					conflictIgnoredCount.Add(int64(len(chunk)) - inserted)
				}
				chunk = chunk[:0]
				return nil
			}

			for {
				select {
				case <-ctxRun.Done():
					return
				case r, ok := <-rows:
					if !ok {
						if e := flush(); e != nil {
							sendErr(e)
						}
						return
					}
					chunk = append(chunk, r)
					if len(chunk) >= o.cfg.ChunkSize {
						if e := flush(); e != nil {
							sendErr(e)
							return
						}
					}
				}
			}
		}()
	}

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		lastAt := start
		var lastInserted int64

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				read := readCount.Load()
				inserted := insertedCount.Load()
				ignored := parseIgnoredCount.Load() + conflictIgnoredCount.Load()

				now := time.Now()
				totalElapsed := now.Sub(start).Seconds()
				totalSpeed := 0.0
				if totalElapsed > 0 {
					totalSpeed = float64(inserted) / totalElapsed
				}

				deltaElapsed := now.Sub(lastAt).Seconds()
				instantSpeed := 0.0
				if deltaElapsed > 0 {
					instantSpeed = float64(inserted-lastInserted) / deltaElapsed
				}
				lastAt = now
				lastInserted = inserted

				o.log.Info("csv_progress",
					slog.String("arquivo", fmt.Sprintf("%d-%s", period, label)),
					slog.Int64("lidas", read),
					slog.Int64("inseridas", inserted),
					slog.Int64("ignoradas", ignored),
					slog.Float64("rows_s", totalSpeed),
					slog.Float64("rows_s_inst", instantSpeed),
					slog.Int("fila", len(rows)),
					slog.Bool("insert_mode", useInsertMode),
					slog.Int("writers", writerWorkers),
				)
			}
		}
	}()

	wg.Wait()
	close(done)
	close(errCh)

	var firstErr error
	for e := range errCh {
		if e != nil && firstErr == nil {
			firstErr = e
		}
	}
	if firstErr != nil {
		return TableStats{}, firstErr
	}

	stats := TableStats{
		Read:     readCount.Load(),
		Inserted: insertedCount.Load(),
		Ignored:  parseIgnoredCount.Load() + conflictIgnoredCount.Load(),
		Duration: time.Since(start),
	}
	speed := 0.0
	if stats.Duration > 0 {
		speed = float64(stats.Inserted) / stats.Duration.Seconds()
	}
	o.log.Info("csv_done",
		slog.String("arquivo", fmt.Sprintf("%d-%s", period, label)),
		slog.Int64("lidas", stats.Read),
		slog.Int64("inseridas", stats.Inserted),
		slog.Int64("ignoradas", stats.Ignored),
		slog.Float64("rows_s", speed),
		slog.Float64("rows_min", speed*60),
		slog.Bool("insert_mode", useInsertMode),
		slog.Int("writers", writerWorkers),
	)
	return stats, nil
}

func makeGetter(stream *csvx.Stream, row []string) fieldGetter {
	return func(names ...string) string {
		for _, n := range names {
			idx, ok := stream.HeaderMap[csvx.NormalizeHeader(n)]
			if !ok || idx >= len(row) {
				continue
			}
			return cleanString(row[idx])
		}
		return ""
	}
}

func cleanString(v string) string {
	v = strings.TrimPrefix(v, "\uFEFF")
	v = strings.TrimSpace(v)
	v = strings.ReplaceAll(v, "\u00a0", " ")
	return strings.TrimSpace(v)
}

func nullableString(v string) any {
	v = cleanString(v)
	if v == "" {
		return nil
	}
	return v
}

func intOrNil(v string) any {
	v = cleanString(v)
	if v == "" {
		return nil
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return nil
	}
	return i
}

func decimalOrNil(v string) any {
	v = cleanString(v)
	if v == "" || strings.EqualFold(v, "Sem informacao") || strings.EqualFold(v, "S/I") {
		return nil
	}
	if strings.Contains(v, ",") {
		v = strings.ReplaceAll(v, ".", "")
	}
	v = strings.ReplaceAll(v, ",", ".")
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return nil
	}
	return fmt.Sprintf("%.2f", f)
}
