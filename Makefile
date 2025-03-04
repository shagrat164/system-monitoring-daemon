BINARY_NAME := "./bin/monitoring-daemon"
BINARY_CLIENT_NAME := "./bin/monitoring-client"
DOCKER_IMG="monitoring-daemon:latest"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

all: lint test build build-client

build:
	go build -v -o $(BINARY_NAME) ./cmd/daemon

build-client:
	go build -v -o $(BINARY_CLIENT_NAME) ./cmd/client

run: build
	$(BINARY_NAME) -config ./configs/config.toml -port 50051

run-client: build-client
	$(BINARY_CLIENT_NAME) -i 10 -d 20

version: build
	$(BINARY_NAME) version

test:
	go test -race -count 100 -timeout 35m ./...

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.64.5

lint: install-lint-deps
	golangci-lint run ./...

proto:
	protoc --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative proto/monitoring.proto

clean:
	rm -rf ./bin

# Сборка образа Docker
build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

# Запуск образа Docker
run-img: build-img
	docker run -p 50051:50051 $(DOCKER_IMG)

.PHONY: all build run test lint proto build-client run-client