BINARY_NAME := "./bin/monitoring-daemon"
DOCKER_IMG="monitoring-daemon:latest"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

# Сборка бинарного файла
build:
	go build -v -o $(BINARY_NAME) -ldflags "$(LDFLAGS)" ./cmd/daemon

run: build
	$(BINARY_NAME) -config ./configs/config.toml

version: build
	$(BINARY_NAME) version

# Запуск тестов
test:
	go test -race -count 100 ./...

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.64.5

# Линтинг
lint: install-lint-deps
	golangci-lint run ./...

# Очистка
clean:
	rm -rf ./bin
	go clean

docker-build:
	docker build -t $(DOCKER_IMG) .

docker-run:
	docker run --rm $(DOCKER_IMG)

.PHONY: build run version test lint clean
