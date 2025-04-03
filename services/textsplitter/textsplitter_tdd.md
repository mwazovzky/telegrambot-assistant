# TextSplitter Service Technical Design Document

## Overview

The TextSplitter service is responsible for splitting a multi-line message into smaller chunks, each adhering to a specified character limit. This document describes the current implementation which handles basic text chunking functionality.

---

## Key Features

1. **Validation**:

   - Ensures the input message is not empty.
   - Validates that no line exceeds the character limit.

2. **Chunking**:

   - Splits the input message into chunks that do not exceed the specified character limit.
   - Preserves the structure of individual lines.

3. **Edge Case Handling**:
   - Ensures trailing newlines and whitespace-only lines are preserved.

---

## Algorithm

### 1. Input Validation

- **Empty Input**: If the input message is empty, return an error.
- **Line Length**: If any line exceeds the character limit, return an error.

### 2. Chunking Logic

- **Initialization**:

  - Trim trailing newlines from the input message.
  - Split the input message into lines.
  - Initialize an empty list of chunks (`chunks`).
  - Use a `currentChunk` string builder to accumulate lines for the current chunk.

- **Iterate Through Lines**:

  - For each line:
    - Calculate the size of the line (including the newline character).
    - If adding the line would exceed the limit, start a new chunk.
    - Otherwise, add the line to the current chunk.

- **Finalization**:
  - Add the remaining content in `currentChunk` to `chunks`.
  - Each chunk naturally ends with a newline character due to how lines are added.

---

## Implementation Details

### 1. Validation

The validation logic integrated into the `Split()` method performs the following checks:

- **Empty Input**: Returns an error if the input is empty.
- **Line Length**: Returns an error if any line exceeds the character limit.

### 2. Split Method

The `Split` method implements the chunking logic:

- **Line Handling**:
  - Processes the input message line by line.
  - Ensures that each chunk doesn't exceed the character limit.
- **Chunk Size**:
  - Calculates the size of each line including the newline character.
  - Starts a new chunk when adding a line would exceed the limit.
- **Trailing Newline**:
  - Each line added to a chunk ends with a newline character.

---

## Example

### Input

```
line1\nline2\nline3
```

### Output (Limit: 10)

```
[
  "line1\n",
  "line2\n",
  "line3\n"
]
```

---

## Edge Cases

1. **Empty Input**:

   - Input: `""`
   - Output: Error (`validation error: empty text`).

2. **Single Line Exceeding Limit**:

   - Input: `"This line is too long"`
   - Output: Error (`validation error: line exceeds limit`).

3. **Whitespace-Only Lines**:
   - Input: `"line1\n   \nline2"`
   - Output: `["line1\n   \n", "line2\n"]` (assuming this chunking fits the limit).

---

## Limitations

1. **Performance**:
   - The algorithm processes the input line by line, which may be inefficient for very large inputs.

---

## Future Improvements

1. **Performance Optimization**:
   - Use more efficient data structures for handling large inputs.
2. **Additional Configuration Options**:
   - Allow configuration of how line breaks should be handled.
   - Add options for different splitting strategies.

---

This document provides an overview of the current `TextSplitter` service implementation, including its design, algorithms, and edge case handling.
