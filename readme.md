![CI](https://github.com/mwazovzky/telegrambot-assistant/actions/workflows/test.yml/badge.svg)

# Telegram Bot Assistant

A Telegram bot that provides AI-powered assistance in private and group chats, leveraging OpenAI's models to respond to user queries.

## Features

- **Private Chat Support**: Direct interactions with authorized users
- **Group Chat Support**: Responds when mentioned by name in authorized groups
- **Smart Response Handling**: Two modes for handling long responses:
  - **Show More Button**: Interactive "Show More" button for incremental reading
  - **Full Response**: Send all chunks at once for immediate access
- **Configurable**: Easily adjustable through environment variables
- **Persistent Context**: Conversation history stored in Redis
- **Secure**: Access control for both private chats and groups

## Internal Design

The bot uses several components:

- **Bot**: Main class that handles Telegram interactions
- **Splitter**: Interface for text splitting with implementations:
  - **BasicSplitter**: Line-based splitter for breaking long messages into chunks
- **ChunkStorage**: Interface for storing and retrieving message chunks:
  - **InMemoryChunkStorage**: Default implementation that stores chunks in memory
- **Assistant**: Interfaces with the OpenAI API to generate responses


## Usage

### Private Chat

Send a message directly to the bot in a private chat. The bot will respond if you're in the authorized users list.

### Group Chat

In group chats, invite bot to the chat, mention the bot by name at the beginning of your message.

### Error Handling

The bot provides helpful error messages in case of issues, specifically when the AI service is unavailable, you'll receive a notification.

These error message helps user understand that something is wrong instead of leaving user wondering why the bot isn't responding.

## Setup

### Prerequisites

- Go 1.23 or higher
- Docker & Docker Compose (for containerized deployment)
- OpenAI API key
- Telegram Bot Token (from @BotFather)
- Redis (for conversation persistence)

### Environment Variables

Create a `.env` file in the project root with the following configuration:

```env
# Telegram Configuration
TELEGRAM_BOT_NAME=your_bot_name
TELEGRAM_API_TOKEN=your_telegram_token
TELEGRAM_USER_CHATS=user1,user2
TELEGRAM_GROUP_CHATS=-100123456789,-100987654321
TELEGRAM_MESSAGE_LIMIT=4096
TELEGRAM_SHOW_MORE=true

# OpenAI Configuration
OPENAI_API_URL=https://api.openai.com/v1
OPENAI_API_KEY=your_openai_api_key
OPENAI_MODEL=gpt-4o-mini
OPENAI_ASSISTANT_NAME=Assistant
OPENAI_ASSISTANT_ROLE="You are a helpful assistant."

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your_redis_password
REDIS_EXPIRATION_TIME=24h

# Loki Logging
LOKI_URL=http://localhost:3100
LOKI_USERNAME=loki
LOKI_AUTH_TOKEN=your_loki_token
```

Please refer to `.env.example` for additional configuration parameters.

## Development

### Local Development

Run the application locally using Docker Compose:

```bash
docker compose -f docker-compose.dev.yml build
docker compose -f docker-compose.dev.yml up -d
```

### Testing

Run the test suite:

```bash
# Run test
go test ./... -v
# Generate test coverage
./coverage.sh
# Review test coverage
open coverage.html
```

## Deployment

After a PR is merged into `main`:

1. GitHub Actions automatically runs tests and builds a new Docker image (`mwazovzky/telegrambot-assistant:latest`) pushed to Docker Hub.
2. Wait for the [CI](https://github.com/mwazovzky/telegrambot-assistant/actions/workflows/test.yml) workflow to complete.
3. SSH into the production server and run:

```bash
cd /path/to/telegrambot-assistant
docker compose pull
docker compose up -d --force-recreate
docker image prune -f
```

4. Verify the bot is running:

```bash
docker compose logs -f app
```
