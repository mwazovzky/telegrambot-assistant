package logger_test

import (
	"context"
	"testing"

	"telegrambot-assistant/services/logger"
)

func TestSlogAdapterMethodsReturnNil(t *testing.T) {
	adapter := logger.New()
	ctx := context.Background()

	if err := adapter.Info(ctx, "test info", "key", "value"); err != nil {
		t.Errorf("Info returned non-nil error: %v", err)
	}
	if err := adapter.Error(ctx, "test error", "key", "value"); err != nil {
		t.Errorf("Error returned non-nil error: %v", err)
	}
	if err := adapter.Debug(ctx, "test debug", "key", "value"); err != nil {
		t.Errorf("Debug returned non-nil error: %v", err)
	}
}
