// Radar do Povo ETL - https://radardopovo.com
package logger

import (
	"log/slog"
	"os"
)

func New(verbose bool, json bool) *slog.Logger {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{Level: level}
	var handler slog.Handler
	if json {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler).With(
		"source", "radardopovo.com",
		"service", "cpgf-etl",
	)
}
