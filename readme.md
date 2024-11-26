![test](https://github.com/mwazovzky/telegrambot-assistant/actions/workflows/test.yml/badge.svg)
![build-docker-image](https://github.com/mwazovzky/telegrambot-assistant/actions/workflows/build-docker-image.yml/badge.svg)

# Telegram Bot Assistant

## Dev

```
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
docker image rm {hash}
docker compose -f docker-compose.prod.yml up -d
```
