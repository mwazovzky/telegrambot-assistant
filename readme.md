# Telegram Bot Assistant

## Start app

```
export TELEGRAM_HTTP_API_TOKEN={bot-token}
export BOT_NAME={name}
export ALLOWED_CHATS={list-chat-ids}
export OPENAI_API_KEY={open-api-key}

docker-compose build
docker-compose up -d
```

## Testing

```
go test ./... -v
go test ./.../parser -v
```
