package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

// SlogAdapter adapts log/slog to the bot.Logger and openai.Logger interfaces.
type SlogAdapter struct {
	log *slog.Logger
}

// New creates a SlogAdapter writing JSON-structured logs to stdout.
// The log level is controlled by the LOG_LEVEL env var (DEBUG, INFO, WARN, ERROR).
// Defaults to INFO.
func New() *SlogAdapter {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLevel(os.Getenv("LOG_LEVEL")),
	})
	return &SlogAdapter{log: slog.New(handler)}
}

func parseLevel(s string) slog.Level {
	switch strings.ToUpper(s) {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func (a *SlogAdapter) Info(ctx context.Context, message string, keyValues ...interface{}) error {
	a.log.InfoContext(ctx, message, keyValues...)
	return nil
}

func (a *SlogAdapter) Error(ctx context.Context, message string, keyValues ...interface{}) error {
	a.log.ErrorContext(ctx, message, keyValues...)
	return nil
}

func (a *SlogAdapter) Debug(ctx context.Context, message string, keyValues ...interface{}) error {
	a.log.DebugContext(ctx, message, keyValues...)
	return nil
}
