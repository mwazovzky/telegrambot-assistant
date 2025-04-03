# TextSplitter Service Requirements

The TextSplitter service splits a multi-line message into smaller chunks, each adhering to a specified character limit. Below are the detailed requirements:

## Input

- A message consisting of multiple lines. Each line ends with a newline character (`\n`).
- A character limit (integer) defines the maximum size of each chunk.

## Output

- An array of strings (`[]string`), where each string is a chunk.

## Validation Rules

1. **Empty Input**:
   If the input message is empty, return an error.

2. **Single Line Exceeding Limit**:
   If a single line exceeds the character limit, return an error.

3. **Validation of Character Limit**:
   The service does not need to validate the provided character limit.

## Processing Rules

1. **Chunk Size**:
   Each chunk must not exceed the specified character limit, including newline characters.

2. **Exact Limit Handling**:
   If adding a newline character to the last line of a chunk exceeds the limit, move the line to the next chunk.

3. **Trailing Newline**:
   Ensure every chunk ends with a newline character, even if a line must be moved to the next chunk to meet this requirement.

4. **Whitespace-Only Lines**:
   Preserve lines containing only whitespace as they are.
