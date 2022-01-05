FROM golang:1.17-buster as builder

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . ./

RUN chmod +x ./tests.sh
RUN ./tests.sh

RUN go build -v -o botapp cmd/bot/main.go

FROM debian:buster-slim

ENV TZ="Europe/Moscow"
ARG cal_token
ARG cal_url
ARG bot_token
ARG bot_admin_group_id
ENV CAL_TOKEN=$cal_token
ENV CAL_URL=$cal_url
ENV BOT_TOKEN=$bot_token
ENV BOT_ADMIN_GROUP_ID=$bot_admin_group_id

RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/botapp /app/botapp

CMD ["/app/botapp"]
