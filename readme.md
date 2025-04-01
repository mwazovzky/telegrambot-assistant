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

## Deploy

```bash
docker compose -f docker-compose.prod.yml down
docker image ls
docker image rm {hash}
docker compose -f docker-compose.prod.yml up -d
```
