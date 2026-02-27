package logging

import (
	"log/slog"
	"os"
	"strings"
)

func New(level string) *slog.Logger {
	l := new(slog.LevelVar)
	switch strings.ToLower(level) {
	case "debug":
		l.Set(slog.LevelDebug)
	case "warn":
		l.Set(slog.LevelWarn)
	case "error":
		l.Set(slog.LevelError)
	default:
		l.Set(slog.LevelInfo)
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: l}))
}
