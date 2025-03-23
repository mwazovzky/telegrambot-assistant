package repository

import (
	"github.com/mwazovzky/assistant"
)

type ThreadRepository interface {
	ThreadExists(tid string) (bool, error)
	CreateThread(tid string) error
	AppendMessage(tid string, msg assistant.Message) error
	GetMessages(tid string) ([]assistant.Message, error)
}
