package openai

import (
	"context"
	"fmt"
	"time"

	"telegrambot-assistant/services/repository"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/responses"
	"github.com/openai/openai-go/shared"
)

// ResponseClient defines the interface for creating OpenAI responses.
type ResponseClient interface {
	New(ctx context.Context, params responses.ResponseNewParams, opts ...option.RequestOption) (*responses.Response, error)
}

// Logger defines a minimal logging interface.
type Logger interface {
	Error(ctx context.Context, message string, keyValues ...interface{}) error
}

// Assistant implements bot.Assistant using the OpenAI Responses API.
type Assistant struct {
	client         ResponseClient
	model          string
	instructions   string
	store          repository.ResponseStore
	logger         Logger
	requestTimeout time.Duration
}

// NewAssistant creates a new Assistant with the given dependencies.
func NewAssistant(client ResponseClient, model string, instructions string, store repository.ResponseStore, logger Logger, requestTimeout time.Duration) *Assistant {
	return &Assistant{
		client:         client,
		model:          model,
		instructions:   instructions,
		store:          store,
		logger:         logger,
		requestTimeout: requestTimeout,
	}
}

func (a *Assistant) Ask(username string, request string) (string, error) {
	params := responses.ResponseNewParams{
		Model:        shared.ResponsesModel(a.model),
		Instructions: openai.String(a.instructions),
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(request),
		},
	}

	// Chain to previous response for conversation continuity
	prevID, err := a.store.GetResponseID(username)
	if err == nil && prevID != "" {
		params.PreviousResponseID = openai.String(prevID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), a.requestTimeout)
	defer cancel()

	resp, err := a.client.New(ctx, params)
	if err != nil {
		return "", fmt.Errorf("openai response error: %w", err)
	}

	// Store response ID for next turn; failure is non-fatal — the user
	// still gets the response, but the next turn won't chain.
	if err := a.store.SetResponseID(username, resp.ID); err != nil {
		a.logger.Error(context.Background(), "Failed to store response ID", "user", username, "error", err)
	}

	return resp.OutputText(), nil
}
