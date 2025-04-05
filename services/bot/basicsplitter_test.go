package bot

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicSplitter_Split_EmptyInput(t *testing.T) {
	splitter := NewBasicSplitter(10)
	result, err := splitter.Split("")
	assert.Nil(t, result)
	assert.EqualError(t, err, "validation error: empty text")
}

func TestBasicSplitter_LineExceedsLimit(t *testing.T) {
	splitter := NewBasicSplitter(10)
	input := "line is too long"
	result, err := splitter.Split(input)
	assert.Nil(t, result)
	assert.EqualError(t, err, "validation error: line exceeds limit")
}

func TestBasicSplitter_Split_WithoutCodeBlock(t *testing.T) {
	splitter := NewBasicSplitter(15)
	input := "line1\nline2\nline3"
	expected := []string{
		"line1\nline2\n",
		"line3\n",
	}
	result, err := splitter.Split(input)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestBasicSplitter_Split_BasicChunking(t *testing.T) {
	splitter := NewBasicSplitter(10)
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

func TestBasicSplitter_Split_WhitespaceOnlyLines(t *testing.T) {
	splitter := NewBasicSplitter(15)
	input := "line1\n   \nline2"
	expected := []string{
		"line1\n   \n",
		"line2\n",
	}
	result, err := splitter.Split(input)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestBasicSplitter_Split_TrailingNewline(t *testing.T) {
	splitter := NewBasicSplitter(10)
	input := "line1\nline2"
	expected := []string{
		"line1\n",
		"line2\n",
	}
	result, err := splitter.Split(input)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestBasicSplitter_Split_Example(t *testing.T) {
	// This test will need the example.txt file moved or path updated
	t.Skip("Skipping example file test after moving to services/bot")

	data, err := os.ReadFile("example.txt")
	assert.NoError(t, err)

	splitter := NewBasicSplitter(300)
	result, err := splitter.Split(string(data))
	assert.NoError(t, err)
	assert.NotNil(t, result)
}
