################
# Build GO app #
################
FROM golang:1.17-buster as builder

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . ./

RUN chmod +x ./tests.sh
RUN ./tests.sh

RUN make docker_build

#########################
# Get certs and TZ data #
#########################
FROM alpine:3.15.0 as certer
RUN apk update && apk add ca-certificates tzdata

################################################
# Use scratch image to reduce final image size #
################################################
FROM scratch

ENV TZ="Europe/Moscow"
ARG cal_token
ARG cal_url
ARG bot_token
ARG bot_admin_group_id
ENV CAL_TOKEN=$cal_token
ENV CAL_URL=$cal_url
ENV BOT_TOKEN=$bot_token
ENV BOT_ADMIN_GROUP_ID=$bot_admin_group_id

COPY --from=certer /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=certer /usr/share/zoneinfo /usr/share/zoneinfo/

COPY --from=builder /app/botapp /app/botapp

CMD ["/app/botapp"]
