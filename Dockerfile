FROM golang:1.22 AS builder
WORKDIR /app
COPY . .
RUN go mod tidy
RUN make build

FROM ubuntu:18.04
WORKDIR /app
COPY --from=builder /app/bin/daemon .
CMD ["./daemon", "-port=50051"]