![test](https://github.com/mwazovzky/telegrambot-assistant/actions/workflows/test.yml/badge.svg)
![build-docker-image](https://github.com/mwazovzky/telegrambot-assistant/actions/workflows/build-docker-image.yml/badge.svg)

# Telegram Bot Assistant

## Dev

```bash
docker compose build
docker compose up -d
```

## Testing

```bash
go test ./... -v
# review test coverage
./coverage.sh
open coverage.html
```

## Config

Please refer to .env.example for config parameters.

## Build

Create image

```bash
docker build --platform=linux/amd64 -t mwazovzky/telegrambot-assistant .
docker push mwazovzky/telegrambot-assistant
```

## Run locally

```bash
docker compose -f docker-compose.dev.yml build
docker compose -f docker-compose.dev.yml up -d
```

## Deploy

```bash
docker compose -f docker-compose.prod.yml pull
docker compose -f docker-compose.prod.yml up -d --force-recreate
docker image prune
```
