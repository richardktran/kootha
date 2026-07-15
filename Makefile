.PHONY: start-infra start-consul start-kafka start-quiz-session-db migrate-all protoc run-all stop-all start-consumer-all create-kafka-topics

PID_DIR := .run
PID_FILE := $(PID_DIR)/services.pids

# Ports used by local Go services (api-gateway, microservices, notification WS)
SERVICE_PORTS := 8080 8082 8083 8084 8085 8086

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
	@$(MAKE) run-all

# Start all Go services in the foreground process group.
# Ctrl+C (or `make stop-all` from another terminal) stops everything.
# PIDs are tracked in $(PID_FILE) so stop-all can clean up orphans.
run-all:
	@mkdir -p $(PID_DIR)
	@$(MAKE) stop-all >/dev/null 2>&1 || true
	@echo "Starting all services (Ctrl+C or \`make stop-all\` to stop)..."
	@set -e; \
	pids=""; \
	cleanup() { \
		echo ""; \
		echo "Stopping all services..."; \
		for pid in $$pids; do \
			kill $$pid 2>/dev/null || true; \
			pkill -P $$pid 2>/dev/null || true; \
		done; \
		rm -f $(PID_FILE); \
		$(MAKE) stop-all >/dev/null 2>&1 || true; \
		echo "All services stopped."; \
	}; \
	trap cleanup INT TERM EXIT; \
	go run id-generation-service/cmd/*.go & pids="$$pids $$!"; \
	go run user-service/cmd/*.go & pids="$$pids $$!"; \
	go run quiz-bank-service/cmd/main.go & pids="$$pids $$!"; \
	go run quiz-session-service/cmd/main.go & pids="$$pids $$!"; \
	go run api-gateway/cmd/main.go & pids="$$pids $$!"; \
	go run notification-service/cmd/main.go & pids="$$pids $$!"; \
	go run leaderboard-service/cmd/main.go & pids="$$pids $$!"; \
	go run quiz-session-service/cmd/session-created-consumer/*.go & pids="$$pids $$!"; \
	go run quiz-session-service/cmd/user-joined-consumer/*.go & pids="$$pids $$!"; \
	echo $$pids > $(PID_FILE); \
	echo "PIDs: $$pids"; \
	echo "Frontend: cd web && bun run dev"; \
	wait

# Kill every local Go service started by run-all (and any orphans).
stop-all:
	@echo "Stopping all Kootha Go services..."
	@if [ -f $(PID_FILE) ]; then \
		for pid in $$(cat $(PID_FILE)); do \
			kill $$pid 2>/dev/null || true; \
			pkill -P $$pid 2>/dev/null || true; \
		done; \
		rm -f $(PID_FILE); \
	fi
	@pkill -f 'id-generation-service/cmd' 2>/dev/null || true
	@pkill -f 'user-service/cmd' 2>/dev/null || true
	@pkill -f 'quiz-bank-service/cmd' 2>/dev/null || true
	@pkill -f 'quiz-session-service/cmd' 2>/dev/null || true
	@pkill -f 'api-gateway/cmd' 2>/dev/null || true
	@pkill -f 'notification-service/cmd' 2>/dev/null || true
	@pkill -f 'leaderboard-service/cmd' 2>/dev/null || true
	@for port in $(SERVICE_PORTS); do \
		pids=$$(lsof -tiTCP:$$port -sTCP:LISTEN 2>/dev/null || true); \
		if [ -n "$$pids" ]; then \
			echo "  freeing port $$port ($$pids)"; \
			kill $$pids 2>/dev/null || true; \
			sleep 0.2; \
			kill -9 $$pids 2>/dev/null || true; \
		fi; \
	done
	@rm -f $(PID_FILE)
	@echo "All services stopped."
