.PHONY: start-infra start-consul start-kafka start-quiz-session-db migrate-all protoc run-all start-consumer-all create-kafka-topics

start-infra:
	@echo "Starting Redis, QuizBank DB, User DB..."
	@docker compose -f docker/infra/docker-compose.yml up -d
	@echo "Starting Kafka..."
	@docker compose -f docker/kafka/docker-compose.yml up -d
	@echo "Starting quiz-session DB..."
	@docker compose -f quiz-session-service/docker/database.yml up -d
	@echo "Ensuring Kafka topics..."
	@sleep 2
	@$(MAKE) create-kafka-topics

create-kafka-topics:
	@echo "Creating Kafka topics..."
	@go run cmd/create-kafka-topics/main.go

start-consul:
	@echo "Starting consul..."
	@docker run -d -p 8500:8500 -p 8600:8600/udp --name=dev-consul hashicorp/consul:latest agent -server -ui -node=server-1 -bootstrap-expect=1 -client=0.0.0.0 || true

start-kafka:
	@echo "Starting kafka..."
	@docker compose -f docker/kafka/docker-compose.yml up -d

start-quiz-session-db:
	@echo "Starting postgres..."
	@docker compose -f quiz-session-service/docker/database.yml up -d

migrate-quiz-session-db:
	@echo "Migrating quiz session db..."
	@migrate -path=quiz-session-service/internal/database/migrations -database "postgres://richardktran:password@localhost:5433/quiz_session?sslmode=disable" up

migrate-quiz-bank-db:
	@echo "Migrating quiz bank db..."
	@migrate -path=quiz-bank-service/internal/database/migrations -database "postgres://richardktran:password@localhost:5434/quiz_bank?sslmode=disable" up

migrate-all: migrate-quiz-session-db migrate-quiz-bank-db

create-migrate-quiz-session-db:
	@echo "Creating migrate quiz session db..."
	@migrate create -ext=sql -dir=quiz-session-service/internal/database/migrations -seq $(name)

protoc:
	@echo "Generating protobuf files..."
	@protoc -I=api --go_out=. --go-grpc_out=. api/*.proto

run-id-generation:
	@go run id-generation-service/cmd/*.go

run-user-service:
	@go run user-service/cmd/*.go

run-quiz-bank:
	@go run quiz-bank-service/cmd/main.go

run-quiz-session:
	@go run quiz-session-service/cmd/main.go

run-api-gateway:
	@go run api-gateway/cmd/main.go

run-notification:
	@go run notification-service/cmd/main.go

run-leaderboard:
	@go run leaderboard-service/cmd/main.go

run-session-created-consumer:
	@go run quiz-session-service/cmd/session-created-consumer/*.go

run-user-joined-consumer:
	@go run quiz-session-service/cmd/user-joined-consumer/*.go

start-consumer-all:
	@echo "Running session consumers..."
	@go run quiz-session-service/cmd/session-created-consumer/*.go &
	@go run quiz-session-service/cmd/user-joined-consumer/*.go &
	@go run leaderboard-service/cmd/main.go &

# Starts all Go services (assumes infra + consul + migrations already done).
run-all:
	@echo "Starting all services..."
	@go run id-generation-service/cmd/*.go &
	@go run user-service/cmd/*.go &
	@go run quiz-bank-service/cmd/main.go &
	@go run quiz-session-service/cmd/main.go &
	@go run api-gateway/cmd/main.go &
	@go run notification-service/cmd/main.go &
	@go run leaderboard-service/cmd/main.go &
	@go run quiz-session-service/cmd/session-created-consumer/*.go &
	@go run quiz-session-service/cmd/user-joined-consumer/*.go &
	@echo "Services launching in background. Frontend: cd web && bun run dev"
	@wait
