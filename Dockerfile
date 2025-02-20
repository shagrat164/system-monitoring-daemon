# Собираем в гошке
FROM golang:1.22 as build

ENV BIN_FILE /opt/monitoring-daemon/monitoring-daemon-app
ENV CODE_DIR /go/src/

WORKDIR ${CODE_DIR}

# Кэшируем слои с модулями
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . ${CODE_DIR}

# Собираем статический бинарник Go (без зависимостей на Си API),
# иначе он не будет работать в alpine образе.
ARG LDFLAGS
RUN CGO_ENABLED=0 go build \
        -ldflags "$LDFLAGS" \
        -o ${BIN_FILE} cmd/daemon/*

# На выходе тонкий образ
FROM alpine:3.9

LABEL SERVICE="monitoring-daemon"
LABEL MAINTAINERS="dprotsvetov@ya.ru"

ENV BIN_FILE "/opt/monitoring-daemon/monitoring-daemon-app"
COPY --from=build ${BIN_FILE} ${BIN_FILE}

ENV CONFIG_FILE /etc/monitoring-daemon/config.toml
COPY ./configs/config.toml ${CONFIG_FILE}

CMD ${BIN_FILE} -config ${CONFIG_FILE}