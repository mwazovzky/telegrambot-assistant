package repository

import (
	"encoding/json"

	openai "github.com/mwazovzky/assistant"
)

type CacheClient interface {
	Get(key string) (string, error)
	Set(key string, value string) error
}

type CacheRepository struct {
	client CacheClient
}

func NewCachedRepository(client CacheClient) *CacheRepository {
	return &CacheRepository{client}
}

func (tr *CacheRepository) CreateThread(key string) error {
	messages := []openai.Message{}

	value, err := encode(messages)
	if err != nil {
		return err
	}

	return tr.client.Set(key, string(value))
}

func (tr *CacheRepository) AppendMessage(key string, msg openai.Message) error {
	messages, err := tr.GetMessages(key)
	if err != nil {
		return err
	}

	messages = append(messages, msg)
	value, err := json.Marshal(messages)
	if err != nil {
		return err
	}

	err = tr.client.Set(key, string(value))
	if err != nil {
		return err
	}

	return nil
}

func (tr *CacheRepository) GetMessages(key string) ([]openai.Message, error) {
	value, err := tr.client.Get(key)
	if err != nil {
		return nil, err
	}

	return decode(value)
}

func encode(messages []openai.Message) (string, error) {
	value, err := json.Marshal(messages)
	if err != nil {
		return "", err
	}

	return string(value), err
}

func decode(value string) ([]openai.Message, error) {
	var messages []openai.Message
	err := json.Unmarshal([]byte(value), &messages)
	if err != nil {
		return nil, err
	}
	return messages, err
}
