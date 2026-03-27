package bot

import (
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

func TestBasicSplitter_Split_BoundaryLengthLine(t *testing.T) {
	// Line "abcde" (5 chars) + newline = 6, which equals the limit.
	// Should produce a single chunk without overflow or empty chunks.
	splitter := NewBasicSplitter(6)
	input := "abcde"
	expected := []string{
		"abcde\n",
	}
	result, err := splitter.Split(input)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)

	// Two boundary-length lines should produce two separate chunks, no empty chunks.
	input = "abcde\n12345"
	expected = []string{
		"abcde\n",
		"12345\n",
	}
	result, err = splitter.Split(input)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestBasicSplitter_LineExceedsLimitWithNewline(t *testing.T) {
	// Line "abcdef" (6 chars) + newline = 7, which exceeds limit of 6.
	splitter := NewBasicSplitter(6)
	input := "abcdef"
	result, err := splitter.Split(input)
	assert.Nil(t, result)
	assert.EqualError(t, err, "validation error: line exceeds limit")
}

