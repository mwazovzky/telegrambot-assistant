package assistant

import (
	"fmt"

	openaiassistant "github.com/mwazovzky/assistant"
)

type OpenAiClient interface {
	GetThread(username string) ([]openaiassistant.Message, error)
	CreateThread(username string) error
	Post(username, req string) (string, error)
}

type Assistant struct {
	client OpenAiClient
}

func NewAssistant(client OpenAiClient) *Assistant {
	return &Assistant{
		client: client,
	}
}

func (a *Assistant) Ask(req string, username string) (string, error) {
	_, err := a.client.GetThread(username)
	if err != nil {
		err = a.client.CreateThread(username)
		if err != nil {
			return "", fmt.Errorf("cannot get or create thread: %w", err)
		}
	}

	res, err := a.client.Post(username, req)
	if err != nil {
		return "", fmt.Errorf("cannot post a question: %w", err)
	}

	return res, nil
}
