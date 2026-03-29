# TelegramBot Assistant - Technical Design Document

## Architecture Overview

The TelegramBot Assistant follows a modular architecture with clear separation of concerns between components:

```ascii
┌─────────────────┐      ┌───────────────┐      ┌───────────────────┐
│  Telegram Bot   │◄────►│ Core Bot      │◄────►│ AI Assistant      │
│  API            │      │ Controller    │      │ Service           │
└─────────────────┘      └───────────────┘      └───────────────────┘
                                │                         │
                                ▼                         │
                         ┌───────────────┐                │
                         │ Text Splitter │                │
                         └───────────────┘                │
                                │                         │
                                ▼                         │
                         ┌───────────────┐                │
                         │ Chunk Storage │                │
                         └───────────────┘                │
                                                          │
┌─────────────────┐      ┌───────────────┐                │
│  Config Service │◄────►│ Redis Storage │◄───────────────┘
└─────────────────┘      └───────────────┘
```

## Core Components

### 1. Bot Service (`services/bot`)

The Bot Service handles direct interactions with Telegram's API and is responsible for:

- Processing incoming messages from users
- Formatting and sending responses back to users
- Managing the "Show More" functionality
- Handling user authorization and chat permissions

#### Key Interfaces

```go
type Assistant interface {
    Ask(username string, request string) (response string, err error)
}

type Splitter interface {
    Split(text string) ([]string, error)
}

type BotAPI interface {
    Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
    GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel
}

type Logger interface {
    Info(ctx context.Context, message string, keyValues ...interface{}) error
    Error(ctx context.Context, message string, keyValues ...interface{}) error
    Debug(ctx context.Context, message string, keyValues ...interface{}) error
}
```

#### Bot Configuration

```go
type BotConfig struct {
    Name        string
    UserChats   []string
    GroupChats  []int64
    UseShowMore bool
}
```

### 2. AI Assistant Service (`services/openai`)

The Assistant Service interfaces with OpenAI's Responses API:

- Uses `previous_response_id` for server-side conversation state management
- Stores only the latest response ID per user in Redis (via `ResponseStore`)
- Re-sends system instructions each turn (instructions don't persist across chained responses)

```go
type ResponseStore interface {
    GetResponseID(key string) (string, error)
    SetResponseID(key string, responseID string) error
}
```

### 3. Text Splitter Service

The Text Splitter Service handles breaking long responses into Telegram-compatible chunks:

- Respects Telegram's message size limits
- Line-based splitting to avoid breaking messages mid-line

### 4. Storage Service (`services/storage`)

The Redis-backed Storage Service provides:

- Persistence for OpenAI response IDs (conversation continuity)
- TTL-based expiration for automatic cleanup

### 5. Configuration Service (`services/config`)

The Configuration Service manages all app settings:

- Environment-based configuration
- Secure management of API tokens and keys
- Feature flags like "Show More" functionality

### 6. Chunk Storage Service

The Chunk Storage Service handles storing and retrieving message chunks for the "Show More" functionality:

- Provides an abstract interface for storage implementations
- Manages message chunks across multiple conversations
- Cleans up after all chunks are delivered

#### Key Interfaces

```go
type ChunkStorage interface {
    StoreChunks(chatID int64, username string, messageID int, chunks []string)
    GetNextChunk(chatID int64, username string) (chunk string, originalID int, hasMore bool, exists bool)
    HasChunks(chatID int64, username string) bool
    Clear(chatID int64, username string)
}
```

#### Implementation

**InMemoryChunkStorage**: Stores chunks in memory with mutex-based thread safety. Chunks are cleaned up after the last chunk is delivered to Telegram.

## Key Workflows

### 1. Message Processing Flow

1. Telegram update received via long polling
2. Bot controller checks permission (authorized user/group)
3. If authorized, message is parsed and processed
4. Query is sent to AI Assistant service
5. Assistant loads previous response ID from Redis, calls OpenAI Responses API with `previous_response_id`
6. Response text is passed to Text Splitter
7. Response chunks are either:
   - All sent at once (if UseShowMore = false)
   - First chunk sent with "Show More" button (if UseShowMore = true)
8. If using "Show More", remaining chunks stored in queue

### 2. "Show More" Button Flow

1. User clicks "Show More" button
2. Callback query received by Bot service
3. Bot requests next chunk from ChunkStorage service
4. If chunk exists:
   - Next chunk sent to user
   - If more chunks exist, new "Show More" button attached
   - If no more chunks, storage is cleaned up
5. If no chunk exists or error occurs:
   - Error message sent to user

## Configuration Parameters

| Parameter              | Environment Variable     | Default | Description                                               |
| ---------------------- | ------------------------ | ------- | --------------------------------------------------------- |
| Bot Name               | TELEGRAM_BOT_NAME        | -       | Bot's display name in Telegram                            |
| API Token              | TELEGRAM_API_TOKEN       | -       | Telegram Bot API token                                    |
| Authorized Users       | TELEGRAM_USER_CHATS      | -       | Comma-separated list of allowed users                     |
| Authorized Groups      | TELEGRAM_GROUP_CHATS     | -       | Comma-separated list of allowed group IDs                 |
| Message Size Limit     | TELEGRAM_MESSAGE_LIMIT   | -       | Maximum message size in characters                        |
| Show More Feature      | TELEGRAM_SHOW_MORE       | true    | Enable/disable Show More button feature                   |
| OpenAI API Key         | OPENAI_API_KEY           | -       | OpenAI API authentication key                             |
| OpenAI Model           | OPENAI_MODEL             | -       | Model to use (e.g. gpt-4o-mini)                           |
| Assistant Name         | OPENAI_ASSISTANT_NAME    | -       | Assistant's display name                                  |
| Assistant Role         | OPENAI_ASSISTANT_ROLE    | -       | System instructions for the assistant                     |
| OpenAI Request Timeout | OPENAI_REQUEST_TIMEOUT   | 30s     | Timeout for OpenAI requests (duration, e.g. 30s, 1m, 2m) |
| Redis Host             | REDIS_HOST               | -       | Redis server hostname                                     |
| Redis Port             | REDIS_PORT               | -       | Redis server port                                         |
| Redis Password         | REDIS_PASSWORD           | -       | Redis authentication password                             |
| Redis TTL              | REDIS_EXPIRATION_TIME    | -       | TTL for stored response IDs                               |
| Loki URL               | LOKI_URL                 | -       | Loki logging endpoint                                     |
| Loki Username          | LOKI_USERNAME            | -       | Loki authentication username                              |
| Loki Token             | LOKI_AUTH_TOKEN          | -       | Loki authentication token                                 |

## Error Handling

The bot implements error handling with specific strategies for:

1. **Permission Errors** - Unauthorized access attempts are logged but not responded to
2. **AI Service Errors** - Failures are reported to the user with a friendly message
3. **Parsing Errors** - Malformed input is logged with relevant context
4. **API Errors** - Network or service issues are handled gracefully

## Testing Strategy

1. **Unit Tests** - For core logic in each component
2. **Integration Tests** - For interactions between components
3. **Mock Services** - For isolating components during testing

## Deployment Architecture

```ascii
┌─────────────────────────────────────────┐
│              Docker Container           │
│                                         │
│  ┌─────────────┐       ┌─────────────┐  │
│  │ Telegram    │       │ Redis       │  │
│  │ Bot Service │◄─────►│ Container   │  │
│  └─────────────┘       └─────────────┘  │
│         │                               │
└─────────┼───────────────────────────────┘
          │
          ▼
┌─────────────────┐       ┌─────────────┐
│ OpenAI API      │       │ Loki        │
│                 │       │ Logging     │
└─────────────────┘       └─────────────┘
```

## Monitoring and Logging

1. **Structured Logging** - Using CloudLog/Loki for centralized log management (async, non-blocking)
2. **Error Tracking** - Errors logged with chat ID and context for debugging

## Security Considerations

1. **API Token Management** - Secure storage in environment variables
2. **User Authentication** - Strict validation of authorized users
3. **Data Privacy** - User messages and AI responses are not logged
