![test](https://github.com/mwazovzky/telegrambot-assistant/actions/workflows/test.yml/badge.svg)

# Telegram Bot Assistant

## Dev

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

## Deploy

Create image

```
docker build --platform=linux/amd64 -t mwazovzky/telegrambot-assistant .
```

Store image to docker hub

```
docker push mwazovzky/telegrambot-assistant
```

Start app

```
docker compose -f docker-compose.prod.yml up
```
