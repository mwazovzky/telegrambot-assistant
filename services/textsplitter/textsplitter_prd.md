# TextSplitter Service Requirements

The TextSplitter service splits a multi-line message into smaller chunks, each adhering to a specified character limit. Below are the detailed requirements:

## Input

- A message consisting of multiple lines. Each line ends with a newline character (`\n`).
- Lines may contain text or code block delimiters (` ``` `).
- A character limit (integer) defines the maximum size of each chunk.

## Output

- An array of strings (`[]string`), where each string is a chunk.

## Validation Rules

1. **Empty Input**:
   If the input message is empty, return an error.

2. **Single Line Exceeding Limit**:
   If a single line exceeds the character limit, return an error.

3. **Matching Code Block Delimiters**:
   Ensure that every code block start delimiter (` ``` `) has a corresponding end delimiter. If not, return an error.

4. **Validation of Character Limit**:
   The service does not need to validate the provided character limit.

## Processing Rules

1. **Chunk Size**:
   Each chunk must not exceed the specified character limit, including newline characters.

2. **Code Block Delimiters**:
   2.1. Code block delimiters (` ``` `) must be preserved.  
   2.2. If a chunk contains a code block start delimiter but not the corresponding end delimiter, the end delimiter must be added to the end of the chunk.  
   2.3. The next chunk must start with the code block start delimiter.  
   2.4. Avoid duplicating delimiters:  
    2.4.1. Do not add a code block end delimiter if the last line of the chunk already ends with it.  
    2.4.2. Do not add a code block start delimiter if the first line of the chunk already starts with it.

3. **Consecutive Code Block Delimiters**:
   Treat consecutive code block delimiters (e.g., ` ```\n```\n`) as separate empty code blocks.

4. **Exact Limit Handling**:
   4.1. If adding a newline character to the last line of a chunk exceeds the limit, move the line to the next chunk.  
   4.2. If adding a code block delimiter as specified in items 2.2 and 2.3 exceeds the limit, move the line to the next chunk.

5. **Malformed Code Block Delimiters**:
   5.1. Treat malformed delimiters (e.g., ` `` ` or ` ```` `) as regular text. Only ` ``` ` is considered a valid code block delimiter.

6. **Mixed Content**:
   6.1. A chunk may contain both text and code blocks. Preserve the order of content as it appears in the input.

7. **Trailing Newline**:
   7.1. Ensure every chunk ends with a newline character, even if a line must be moved to the next chunk to meet this requirement.

8. **Whitespace-Only Lines**:
   8.1. Preserve lines containing only whitespace as they are.
