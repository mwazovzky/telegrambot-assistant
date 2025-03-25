package textsplitter

import (
	"fmt"
	"strings"
)

const newLine = "\n"
const codeBlock = "```"

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
	inCodeBlock := false

	text = strings.TrimRight(text, newLine)
	lines := strings.Split(text, newLine)

	for _, line := range lines {
		if len(line) > s.limit {
			return nil, fmt.Errorf("validation error: line exceeds limit")
		}

		if currentChunk.Len()+len(line)+1 > s.limit {
			chunkEndsWithCodeBlock := strings.HasSuffix(currentChunk.String(), codeBlock+newLine)

			if inCodeBlock && !chunkEndsWithCodeBlock {
				currentChunk.WriteString(codeBlock + newLine)
			}

			if inCodeBlock && chunkEndsWithCodeBlock {
				str := currentChunk.String()
				str = strings.TrimSuffix(str, codeBlock+newLine)
				currentChunk.Reset()
				currentChunk.WriteString(str)
			}

			chunks = append(chunks, currentChunk.String())
			currentChunk.Reset()

			if inCodeBlock && line != codeBlock {
				currentChunk.WriteString(codeBlock + newLine)
			}

			if inCodeBlock && line == codeBlock {
				inCodeBlock = !inCodeBlock
				continue
			}
		}

		if line == codeBlock {
			inCodeBlock = !inCodeBlock
		}

		currentChunk.WriteString(line + newLine)
	}

	// append the last chunk
	if len(currentChunk.String()) != 0 {
		chunks = append(chunks, currentChunk.String())
	}

	// if inCodeBlock {
	// 	return nil, fmt.Errorf("validation error: unmatched code block delimiters")
	// }

	return chunks, nil
}
