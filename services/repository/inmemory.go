package repository

import (
	"fmt"

	openai "github.com/mwazovzky/assistant"
)

type InmemoryRepository struct {
	data map[string][]openai.Message
}

func NewInmemoryRepository() *InmemoryRepository {
	data := make(map[string][]openai.Message)
	return &InmemoryRepository{data}
}

func (tr *InmemoryRepository) ThreadExists(tid string) (bool, error) {
	_, ok := tr.data[tid]
	if !ok {
		return false, nil
	}

	return true, nil
}

func (tr *InmemoryRepository) CreateThread(tid string) error {
	tr.data[tid] = []openai.Message{}
	return nil
}

func (tr *InmemoryRepository) AppendMessage(tid string, msg openai.Message) error {
	messages, ok := tr.data[tid]
	if !ok {
		return fmt.Errorf("thread [%s] does not exist", tid)
	}

	tr.data[tid] = append(messages, msg)
	return nil
}

func (tr *InmemoryRepository) GetMessages(tid string) ([]openai.Message, error) {
	messages, ok := tr.data[tid]
	if !ok {
		return nil, fmt.Errorf("thread [%s] does not exist", tid)
	}

	return messages, nil
}
