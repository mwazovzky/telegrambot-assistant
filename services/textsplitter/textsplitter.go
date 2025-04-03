package textsplitter

import (
	"fmt"
	"strings"
)

const newLine = "\n"

type TextSplitter struct {
	limit int
}

func NewTextSplitter(limit int) *TextSplitter {
	return &TextSplitter{limit: limit}
}

func (s *TextSplitter) Split(text string) ([]string, error) {
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
