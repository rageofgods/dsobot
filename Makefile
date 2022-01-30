.PHONY:
.SILENT:

#VERSION=`git describe --tags`
VERSION=`git rev-parse --short HEAD`
BUILD=`date +%FT%T%z`

build:
	go build -ldflags "-X main.Version=${VERSION} -X main.Build=${BUILD}" -o ./.bin/bot cmd/bot/main.go
docker_build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags "-X main.Version=${VERSION} -X main.Build=${BUILD} -w -s -extldflags '-static'" -a \
		-o botapp cmd/bot/main.go
run: build
	./.bin/bot
test:
	go test ./... -v
lint:
	golangci-lint run