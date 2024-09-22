package repository

import (
	"fmt"

	openai "github.com/mwazovzky/assistant"
)

type ThreadRepository struct {
	data map[string][]openai.Message
}

func NewThreadRepository() *ThreadRepository {
	data := make(map[string][]openai.Message)
	return &ThreadRepository{data}
}

func (tr *ThreadRepository) CreateThread(tid string) error {
	tr.data[tid] = []openai.Message{}
	return nil
}

func (tr *ThreadRepository) AppendMessage(tid string, msg openai.Message) error {
	messages, ok := tr.data[tid]
	if !ok {
		return fmt.Errorf("thread [%s] does not exist", tid)
	}

	tr.data[tid] = append(messages, msg)
	return nil
}

func (tr *ThreadRepository) GetMessages(tid string) ([]openai.Message, error) {
	messages, ok := tr.data[tid]
	if !ok {
		return nil, fmt.Errorf("thread [%s] does not exist", tid)
	}

	return messages, nil
}
