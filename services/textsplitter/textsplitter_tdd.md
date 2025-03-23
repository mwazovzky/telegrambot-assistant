# TextSplitter Service Technical Design Document

## Overview

The TextSplitter service is responsible for splitting a multi-line message into smaller chunks, each adhering to a specified character limit. It ensures that the integrity of the input message is preserved while meeting the requirements defined in the Product Requirements Document (PRD).

---

## Key Features

1. **Validation**:

   - Ensures the input message meets the validation rules before processing.
   - Validates empty input, line length, and matching code block delimiters.

2. **Chunking**:

   - Splits the input message into chunks that do not exceed the specified character limit.
   - Preserves the structure of code blocks and ensures proper handling of delimiters.

3. **Edge Case Handling**:
   - Handles consecutive code block delimiters, malformed delimiters, and mixed content.
   - Ensures trailing newlines and whitespace-only lines are preserved.

---

## Algorithm

### 1. Input Validation (Integrated into `Split()`)

- **Empty Input**: If the input message is empty, return an error.
- **Line Length**: If any line exceeds the character limit, return an error.
- **Matching Code Block Delimiters**: Count the number of code block delimiters (` ``` `). If the count is odd, return an error.

### 2. Chunking Logic

- **Initialization**:

  - Trim trailing newlines from the input message.
  - Split the input message into lines.
  - Initialize an empty list of chunks (`chunks`).
  - Use a `currentChunk` string builder to accumulate lines for the current chunk.
  - Track whether the algorithm is inside a code block using a boolean flag (`inCodeBlock`).

- **Iterate Through Lines**:

  - For each line:
    - Calculate the size of the line (including the newline character).
    - Handle code block delimiters:
      - Toggle the `inCodeBlock` flag when encountering a delimiter.
      - Ensure the entire code block stays together in the same chunk whenever possible.
    - Handle lines inside and outside code blocks:
      - If adding the line exceeds the limit:
        - If inside a code block:
          - Add a closing delimiter (` ``` `) to the current chunk if it does not already end with one.
          - Flush the `currentChunk` to `chunks` and start a new chunk with an opening delimiter (` ``` `) if the next line is not a delimiter.
        - If outside a code block:
          - Flush the `currentChunk` to `chunks` and start a new chunk.
      - Otherwise, add the line to the `currentChunk`.

- **Finalization**:
  - Add the remaining content in `currentChunk` to `chunks`.
  - Ensure every chunk ends with a newline character.

---

## Implementation Details

### 1. Validation

The validation logic integrated into the `Split()` method performs the following checks:

- **Empty Input**: Returns an error if the input is empty.
- **Line Length**: Returns an error if any line exceeds the character limit.
- **Matching Code Block Delimiters**: Tracks the state of code blocks and ensures all delimiters are properly matched.

### 2. Split Method

The `Split` method implements the chunking logic:

- **Code Block Handling**:
  - Ensures that code blocks (including their delimiters and contents) are preserved in the same chunk whenever possible.
  - Handles consecutive code block delimiters as separate empty blocks.
  - Adds or removes delimiters as needed to maintain code block integrity when splitting inside a code block.
- **Chunk Size**:
  - Ensures that no chunk exceeds the specified character limit.
  - Moves lines or delimiters to the next chunk if adding them would exceed the limit.
- **Trailing Newline**:
  - Ensures every chunk ends with a newline character.

---

## Example

### Input

````
text\n```\ncode\nmore\n```\nmore text
````

### Output (Limit: 10)

````
[
  "text\n",
  "```\ncode\n```\n",
  "```\nmore\n```\n",
  "more text\n"
]
````

---

## Edge Cases

1. **Empty Input**:

   - Input: `""`
   - Output: Error (`validation error: empty text`).

2. **Single Line Exceeding Limit**:

   - Input: `"This line is too long"`
   - Output: Error (`validation error: line exceeds limit`).

3. **Unmatched Code Block Delimiters**:

   - Input: ` "```\ncode" `
   - Output: Error (`validation error: unmatched code block delimiters`).

4. **Consecutive Code Block Delimiters**:

   - Input: ` "```\n```\n```\n```" `
   - Output: ` ["```\n```\n```\n```\n"] `.

5. **Mixed Content**:

   - Input: ` "text\n```\ncode\n```\nmore text" `
   - Output: ` ["text\n```\ncode\n```\n", "more text\n"] `.

6. **Code Block at Chunk Boundary**:
   - Input: ` "text\n```\ncode\n```" `
   - Output: ` ["text\n", "```\ncode\n```\n"] `.

---

## Logging and Debugging

The `info` function is used for logging intermediate states during execution. It logs:

- The current line being processed.
- The state of the `currentChunk`.
- Whether the algorithm is inside a code block.

---

## Limitations

1. **Performance**:
   - The algorithm processes the input line by line, which may be inefficient for very large inputs.
2. **Strict Validation**:
   - The service strictly enforces validation rules, which may reject inputs with minor formatting issues.

---

## Future Improvements

1. **Performance Optimization**:
   - Use more efficient data structures for handling large inputs.
2. **Customizable Validation**:
   - Allow users to configure validation rules (e.g., disable strict code block matching).

---

This document provides a comprehensive overview of the `TextSplitter` service, including its design, implementation, and edge case handling. Let me know if further details are needed!
