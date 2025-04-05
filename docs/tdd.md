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
    Info(message string, keyValues ...interface{}) error
    Error(message string, keyValues ...interface{}) error
    Debug(message string, keyValues ...interface{}) error
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

#### Key Data Structures

```go
type Bot struct {
    botApi        BotAPI
    name          string
    userChats     []string
    groupChats    []int64
    splitter      Splitter
    logger        Logger
    useShowMore   bool
    pendingChunks map[string]*ChunkQueue
    chunksMutex   sync.RWMutex
}

type ChunkQueue struct {
    Chunks     []string
    Position   int
    OriginalID int
}
```

### 2. AI Assistant Service

The Assistant Service interfaces with OpenAI's API to generate responses:

- Manages conversation context
- Processes user queries
- Provides natural language responses

### 3. Text Splitter Service

The Text Splitter Service handles breaking long responses into Telegram-compatible chunks:

- Respects Telegram's message size limits
- Maintains semantic coherence when possible
- Avoids breaking messages at awkward points

### 4. Storage Service

The Redis-backed Storage Service provides:

- Persistence for conversation history
- Caching of responses to improve performance
- Session management for users

### 5. Configuration Service

The Configuration Service manages all app settings:

- Environment-based configuration
- Secure management of API tokens and keys
- Feature flags like "Show More" functionality

### 6. Chunk Storage Service

The Chunk Storage Service handles storing and retrieving message chunks for the "Show More" functionality:

- Provides an abstract interface for storage implementations
- Manages message chunks across multiple conversations
- Supports retrieval of chunks for incremental display

#### Key Interfaces

```go
type ChunkStorage interface {
    StoreChunks(chatID int64, username string, messageID int, chunks []string)
    GetNextChunk(chatID int64, username string) (chunk string, originalID int, hasMore bool, exists bool)
    HasChunks(chatID int64, username string) bool
}
```

#### Implementations

1. **InMemoryChunkStorage**: Default implementation that stores chunks in memory
   - Fast access but doesn't persist across restarts
   - Suitable for development or low-traffic bots
2. **RedisChunkStorage**: (Future implementation)

   - Distributed storage supporting multiple bot instances
   - Built-in TTL support for automatic cleanup
   - Persistence across service restarts

3. **DatabaseChunkStorage**: (Future implementation)
   - Long-term retention and analytics
   - Support for complex queries and historical data

## Key Workflows

### 1. Message Processing Flow

1. Telegram update received via webhook or long polling
2. Bot controller checks permission (authorized user/group)
3. If authorized, message is parsed and processed
4. Query is sent to AI Assistant service
5. Response is received and passed to Text Splitter
6. Response chunks are either:
   - All sent at once (if UseShowMore = false)
   - First chunk sent with "Show More" button (if UseShowMore = true)
7. If using "Show More", remaining chunks stored in queue

### 2. "Show More" Button Flow

1. User clicks "Show More" button
2. Callback query received by Bot service
3. Bot requests next chunk from ChunkStorage service
4. If chunk exists:
   - Next chunk sent to user
   - If more chunks exist, new "Show More" button attached
5. If no chunk exists or error occurs:
   - Error message sent to user

## Configuration Parameters

| Parameter          | Environment Variable   | Default | Description                               |
| ------------------ | ---------------------- | ------- | ----------------------------------------- |
| Bot Name           | TELEGRAM_BOT_NAME      | -       | Bot's display name in Telegram            |
| API Token          | TELEGRAM_API_TOKEN     | -       | Telegram Bot API token                    |
| Authorized Users   | TELEGRAM_USER_CHATS    | -       | Comma-separated list of allowed users     |
| Authorized Groups  | TELEGRAM_GROUP_CHATS   | -       | Comma-separated list of allowed group IDs |
| Message Size Limit | TELEGRAM_MESSAGE_LIMIT | -       | Maximum message size in characters        |
| Show More Feature  | TELEGRAM_SHOW_MORE     | true    | Enable/disable Show More button feature   |

## Error Handling

The bot implements comprehensive error handling with specific strategies for:

1. **Permission Errors** - Unauthorized access attempts are logged but not responded to
2. **AI Service Errors** - Temporary failures trigger retry logic
3. **Parsing Errors** - Malformed input is logged with relevant context
4. **API Errors** - Network or service issues are handled gracefully

## Testing Strategy

1. **Unit Tests** - For core logic in each component
2. **Integration Tests** - For interactions between components
3. **Mock Services** - For isolating components during testing
4. **End-to-End Tests** - With a complete test environment

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

1. **Structured Logging** - Using CloudLog/Loki for centralized log management
2. **Performance Metrics** - Response times, queue depths, API calls
3. **Error Tracking** - Automated alerting for critical failures
4. **Usage Analytics** - User engagement and feature utilization

## Security Considerations

1. **API Token Management** - Secure storage in environment variables
2. **User Authentication** - Strict validation of authorized users
3. **Rate Limiting** - Protection against abuse or DoS
4. **Data Privacy** - No unnecessary storage of user data

## Future Technical Considerations

1. **Scaling Strategy** - Horizontal scaling for high-load scenarios
2. **Database Migration** - Potential move to more robust persistence
3. **Plugin System** - Architecture to support third-party extensions
4. **API Gateway** - For advanced request routing and management

## Migration Paths

### Chunk Storage Migration

As the bot's usage grows, the following storage migration path is recommended:

1. **Phase 1** (Development/Low Traffic): InMemoryChunkStorage

   - Simple implementation, minimal dependencies
   - No persistence across restarts
   - Limited by memory constraints

2. **Phase 2** (Production/Medium Traffic): RedisChunkStorage

   - TTL-based expiration for automatic resource management
   - Supports horizontal scaling across multiple bot instances
   - Persistence across service restarts

3. **Phase 3** (High Traffic/Analytics): DatabaseChunkStorage
   - Long-term data retention
   - Support for usage analytics and reporting
   - More robust transaction management

## Scaling Considerations

### Chunk Storage Scaling

1. **Memory Management**: Move from in-memory to Redis-based storage
2. **TTL Policies**: Implement automatic expiration for inactive conversations
3. **Horizontal Scaling**: Use distributed storage to support multiple bot instances
4. **Failure Recovery**: Ensure chunks can be recovered or gracefully expire after crashes
