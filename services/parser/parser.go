package parser

import (
	"fmt"
	"slices"
	"strings"
)

type Parser struct {
	botName       string
	botChat       int64
	assignedChats []int64
}

func NewParser(botName string, botChat int64, assignedChats []int64) *Parser {
	return &Parser{botName, botChat, assignedChats}
}

func (p *Parser) Parse(chatID int64, txt string) (string, error) {
	if chatID == p.botChat {
		return txt, nil
	}

	if !slices.Contains(p.assignedChats, chatID) {
		return "", fmt.Errorf("can not process chat")
	}

	// messages in assigned chats must address bot by name
	if !strings.HasPrefix(txt, p.botName) {
		return "", fmt.Errorf("can not process request")
	}

	trimmedSymbols := "!, "

	return strings.TrimLeft(strings.TrimPrefix(txt, p.botName), trimmedSymbols), nil
}
