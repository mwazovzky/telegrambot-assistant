![test](https://github.com/mwazovzky/telegrambot-assistant/actions/workflows/test.yml/badge.svg)
![build-docker-image](https://github.com/mwazovzky/telegrambot-assistant/actions/workflows/build-docker-image.yml/badge.svg)

# Telegram Bot Assistant

## Dev

```bash
# set env variables
export VAR_NAME=VAR_VALUE
docker compose build
docker compose up -d
```

## Testing

```bash
go test ./... -v
go test ./.../parser -v
# generate coverage report
./testing/coverage.sh
```

## Deploy

Create image

```bash
docker build --platform=linux/amd64 -t mwazovzky/telegrambot-assistant .
```

Store image to docker hub

```bash
docker push mwazovzky/telegrambot-assistant
```

Start app

```bash
# set env variables
export VAR_NAME=VAR_VALUE
docker image rm {hash}
docker compose -f docker-compose.prod.yml up -d
```
