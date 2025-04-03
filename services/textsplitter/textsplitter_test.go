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

func TestTextSplitter_Split_BasicChunking(t *testing.T) {
	splitter := NewTextSplitter(10)
	input := "line1\nline2\nline3"
	expected := []string{
		"line1\n",
		"line2\n",
		"line3\n",
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

func TestTextSplitter_Split_Example(t *testing.T) {
	data, err := os.ReadFile("example.txt")
	assert.NoError(t, err)

	splitter := NewTextSplitter(300)
	result, err := splitter.Split(string(data))
	assert.NoError(t, err)
	assert.NotNil(t, result)
}
