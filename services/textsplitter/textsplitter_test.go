package textsplitter

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTextSplitter_Split_EmptyInput(t *testing.T) {
	splitter := NewTextSplitter(10)
	result, err := splitter.Split("")
	assert.Nil(t, result)
	assert.EqualError(t, err, "validation error: empty text")
}

func TestTextSplitter_LineExceedsLimit(t *testing.T) {
	splitter := NewTextSplitter(10)
	input := "line is too long"
	result, err := splitter.Split(input)
	assert.Nil(t, result)
	assert.EqualError(t, err, "validation error: line exceeds limit")
}

// func TestTextSplitter_Split_UnmatchedCodeBlockDelimiters(t *testing.T) {
// 	splitter := NewTextSplitter(20)
// 	input := "```\ncode"
// 	result, err := splitter.Split(input)
// 	assert.Nil(t, result)
// 	assert.EqualError(t, err, "validation error: unmatched code block delimiters")
// }

func TestTextSplitter_Split_WithoutCodeBlock(t *testing.T) {
	splitter := NewTextSplitter(15)
	input := "line1\nline2\nline3"
	expected := []string{
		"line1\nline2\n",
		"line3\n",
	}
	result, err := splitter.Split(input)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestTextSplitter_Split_SimpleCodeBlock(t *testing.T) {
	splitter := NewTextSplitter(15)
	input := "```\ncode\n```"
	expected := []string{
		"```\ncode\n```\n",
	}
	result, err := splitter.Split(input)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestTextSplitter_Split_InsideCodeBlock(t *testing.T) {
	splitter := NewTextSplitter(10)
	input := "```\ncode\nmore\n```\n"
	expected := []string{
		"```\ncode\n```\n",
		"```\nmore\n```\n",
	}
	result, err := splitter.Split(input)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestTextSplitter_ConsecutiveCodeBlocks(t *testing.T) {
	splitter := NewTextSplitter(20)
	input := "```\n```\n```\n```"
	expected := []string{
		"```\n```\n```\n```\n",
	}
	result, err := splitter.Split(input)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestTextSplitter_MixedContent(t *testing.T) {
	splitter := NewTextSplitter(20)
	input := "text\n```\ncode\n```\nmore text"
	expected := []string{
		"text\n```\ncode\n```\n",
		"more text\n",
	}
	result, err := splitter.Split(input)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestTextSplitter_CodeBlockAtChunkBoundary(t *testing.T) {
	splitter := NewTextSplitter(10)
	input := "text\n```\ncode\n```"
	expected := []string{
		"text\n",
		"```\ncode\n```\n",
	}
	result, err := splitter.Split(input)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestTextSplitter_Split_WhitespaceOnlyLines(t *testing.T) {
	splitter := NewTextSplitter(15)
	input := "line1\n   \nline2"
	expected := []string{
		"line1\n   \n",
		"line2\n",
	}
	result, err := splitter.Split(input)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestTextSplitter_Split_TrailingNewline(t *testing.T) {
	splitter := NewTextSplitter(10)
	input := "line1\nline2"
	expected := []string{
		"line1\n",
		"line2\n",
	}
	result, err := splitter.Split(input)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestTextSplitter_Split_MalformedCodeBlockDelimiters(t *testing.T) {
	splitter := NewTextSplitter(20)
	input := "``\n"
	expected := []string{
		"``\n",
	}
	result, err := splitter.Split(input)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestTextSplitter_Split_Example(t *testing.T) {
	data, err := os.ReadFile("example.txt")
	assert.NoError(t, err)

	splitter := NewTextSplitter(300)
	result, err := splitter.Split(string(data))
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 4)
}
