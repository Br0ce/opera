package monitor

import (
	"io"
	"log/slog"
	"os"
)

func NewLogger(debug bool) *slog.Logger {
	var level = new(slog.LevelVar)
	lg := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	lg = lg.With("service", slog.StringValue(serviceName))
	if debug {
		level.Set(slog.LevelDebug)
	}
	return lg
}

func NewTestLogger(debug bool) *slog.Logger {
	var level = new(slog.LevelVar)
	lg := slog.New(slog.NewTextHandler(
		getWriter(debug),
		&slog.HandlerOptions{Level: level, AddSource: true}))

	if debug {
		level.Set(slog.LevelDebug)
	}

	return lg
}

func getWriter(debug bool) io.Writer {
	if debug {
		return os.Stdout
	}
	return io.Discard
}
