package logger

import (
	"context"
	"log/slog"
	"os"
)

// SlogAdapter adapts log/slog to the bot.Logger and openai.Logger interfaces.
type SlogAdapter struct {
	log *slog.Logger
}

// New creates a SlogAdapter writing JSON-structured logs to stdout.
func New() *SlogAdapter {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	return &SlogAdapter{log: slog.New(handler)}
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
