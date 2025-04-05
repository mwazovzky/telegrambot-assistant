// Renamed from: /Users/alex/code/telegrambot-assistant/services/bot/textsplitter.go
package bot

import (
	"fmt"
	"strings"
)

const newLine = "\n"

// BasicSplitter implements a simple line-based text splitter
type BasicSplitter struct {
	limit int
}

// NewBasicSplitter creates a new BasicSplitter with the specified size limit
func NewBasicSplitter(limit int) *BasicSplitter {
	return &BasicSplitter{limit: limit}
}

func (s *BasicSplitter) Split(text string) ([]string, error) {
	if len(text) == 0 {
		return nil, fmt.Errorf("validation error: empty text")
	}

	var chunks []string
	var currentChunk strings.Builder

	text = strings.TrimRight(text, newLine)
	lines := strings.Split(text, newLine)

	for _, line := range lines {
		if len(line) > s.limit {
			return nil, fmt.Errorf("validation error: line exceeds limit")
		}

		if currentChunk.Len()+len(line)+1 > s.limit {
			chunks = append(chunks, currentChunk.String())
			currentChunk.Reset()
		}

		currentChunk.WriteString(line + newLine)
	}

	// append the last chunk
	if len(currentChunk.String()) != 0 {
		chunks = append(chunks, currentChunk.String())
	}

	return chunks, nil
}
