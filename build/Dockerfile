# Собираем в гошке
FROM golang:1.24 as build

ENV BIN_FILE /opt/sysmondaemon/sysmondaemon-app
ENV CODE_DIR /go/src/

WORKDIR ${CODE_DIR}

# Кэшируем слои с модулями
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . ${CODE_DIR}

# Собираем статический бинарник Go (без зависимостей на Си API)
ARG LDFLAGS
RUN CGO_ENABLED=0 go build \
    -ldflags "$LDFLAGS" \
    -o ${BIN_FILE} cmd/daemon/*

FROM ubuntu:18.04

LABEL MAINTAINER="Prostvetov Danila <dprotsvetov@ya.ru>"

ENV BIN_FILE "/opt/sysmondaemon/sysmondaemon-app"
ENV CONFIG_FILE /etc/sysmondaemon/config.toml

COPY --from=build ${BIN_FILE} ${BIN_FILE}

COPY ./configs/config.toml ${CONFIG_FILE}

# Порт
EXPOSE 50051

# Запуск
CMD ${BIN_FILE} -config ${CONFIG_FILE} -port 50051