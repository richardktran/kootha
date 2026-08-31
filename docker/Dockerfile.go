# Multi-stage Go service image. Build from the repository root:
#
#   docker build -f docker/Dockerfile.go \
#     --build-arg SERVICE_PKG=./api-gateway/cmd \
#     --build-arg CGO=0 \
#     -t kootha/api-gateway:latest .
#
# Kafka-backed services need CGO=1 and -tags dynamic so they link against
# Debian's librdkafka instead of confluent-kafka-go's vendored amd64 .a
# (that vendor lib fails to link on Apple Silicon / linux/arm64).

ARG GO_VERSION=1.22.4
FROM golang:${GO_VERSION}-bookworm AS builder

ARG CGO=0
ARG SERVICE_PKG

RUN if [ "$CGO" = "1" ]; then \
      echo 'Acquire::Retries "5";' > /etc/apt/apt.conf.d/80-retries && \
      for i in 1 2 3 4 5; do apt-get update && break || sleep 5; done && \
      apt-get install -y --no-install-recommends \
        pkg-config build-essential librdkafka-dev && \
      rm -rf /var/lib/apt/lists/*; \
    fi

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# -tags dynamic: use pkg-config/system librdkafka (multi-arch). The bundled
# static library in confluent-kafka-go v1.9.2 is linux/amd64 only (ELF EM: 62).
RUN if [ "$CGO" = "1" ]; then \
      CGO_ENABLED=1 GOOS=linux go build -tags dynamic -o /out/service ${SERVICE_PKG}; \
    else \
      CGO_ENABLED=0 GOOS=linux go build -o /out/service ${SERVICE_PKG}; \
    fi

FROM debian:bookworm-slim

ARG CGO=0

RUN echo 'Acquire::Retries "5";' > /etc/apt/apt.conf.d/80-retries && \
    for i in 1 2 3 4 5; do apt-get update && break || sleep 5; done && \
    apt-get install -y --no-install-recommends ca-certificates && \
    if [ "$CGO" = "1" ]; then apt-get install -y --no-install-recommends librdkafka1; fi && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /out/service /app/service
COPY --from=builder /src/api-gateway/configs ./api-gateway/configs
COPY --from=builder /src/user-service/configs ./user-service/configs
COPY --from=builder /src/id-generation-service/configs ./id-generation-service/configs
COPY --from=builder /src/quiz-bank-service/configs ./quiz-bank-service/configs
COPY --from=builder /src/quiz-session-service/configs ./quiz-session-service/configs
COPY --from=builder /src/notification-service/configs ./notification-service/configs
COPY --from=builder /src/leaderboard-service/configs ./leaderboard-service/configs

ENTRYPOINT ["/app/service"]
