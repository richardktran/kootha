# Realtime Quiz (Kootha)

## Architecture

```
Web ──REST──► API Gateway (:8080) ──gRPC──► User / QuizSession / (via Consul)
Web ──WS────► Notification (:8086) ──gRPC──► QuizSession
QuizSession ──► QuizBank, Redis, Kafka
Leaderboard ──► Kafka answer-submitted → Redis leaderboard → ranking-updated
Notification ──► Kafka events → Redis Pub/Sub → WS broadcast
```

## Prerequisites

- Go 1.22+, Docker, [migrate](https://github.com/golang-migrate/migrate), Consul image
- Frontend: Bun (or npm)
- `protoc` + `protoc-gen-go` / `protoc-gen-go-grpc` for regenerating APIs

## One-time setup

```bash
# 1. Infrastructure
make start-consul
make start-infra

# 2. Create quiz_session DB if needed (first time only)
docker exec -it quiz-session-db psql -U richardktran -c "CREATE DATABASE quiz_session;"

# 3. Migrations
make migrate-all

# 4. (Optional) regenerate protos after editing api/*.proto
make protoc
```

## Run the stack

```bash
# Terminal A — all Go services + consumers
make run-all

# Terminal B — Next.js
cd web && bun install && bun run dev
```

Or start services individually: `make run-api-gateway`, `make run-notification`, etc.

## Ports

| Service | Port |
|---------|------|
| API Gateway (REST) | 8080 |
| Notification (WS) | 8086 |
| User gRPC | 8082 |
| ID Generation gRPC | 8083 |
| QuizSession gRPC | 8084 |
| QuizBank gRPC | 8085 |
| Redis | 6379 |
| Quiz Session Postgres | 5433 |
| Quiz Bank Postgres | 5434 |
| Kafka | 9092 |
| Consul | 8500 |
| Web | 3000 |

## Play smoke test

1. Open two browsers at http://localhost:3000
2. Enter names → Create room in one → Join with room ID in the other
3. Host clicks **Start Quiz** → answer → host **Next Question** until finished
