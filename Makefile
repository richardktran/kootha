# =============================================================================
# Kootha — Realtime Quiz Platform
# =============================================================================
# Usage:  make help
# Quick:  make start   → infra + migrate + all services
#         make stop    → stop Go services
#         make down    → stop services + infrastructure
# =============================================================================

.DEFAULT_GOAL := help
SHELL := /bin/bash

# --- Directories ------------------------------------------------------------
ROOT     := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
BIN_DIR  := $(ROOT)/bin
PID_DIR  := $(ROOT)/.run
PID_FILE := $(PID_DIR)/services.pids

# --- Docker Compose ---------------------------------------------------------
COMPOSE_INFRA := docker/infra/docker-compose.yml
COMPOSE_KAFKA := docker/kafka/docker-compose.yml
COMPOSE_QSESS := quiz-session-service/docker/database.yml

# --- Database ---------------------------------------------------------------
DB_USER := richardktran
DB_PASS := password

DSN_QUIZ_SESSION := postgres://$(DB_USER):$(DB_PASS)@localhost:5433/quiz_session?sslmode=disable
DSN_QUIZ_BANK    := postgres://$(DB_USER):$(DB_PASS)@localhost:5434/quiz_bank?sslmode=disable

MIGRATE_QUIZ_SESSION := quiz-session-service/internal/database/migrations
MIGRATE_QUIZ_BANK    := quiz-bank-service/internal/database/migrations

# --- Services ---------------------------------------------------------------
# name|go-package (relative to module root)
SERVICES := \
	api-gateway|./api-gateway/cmd \
	id-generation-service|./id-generation-service/cmd \
	user-service|./user-service/cmd \
	quiz-bank-service|./quiz-bank-service/cmd \
	quiz-session-service|./quiz-session-service/cmd \
	notification-service|./notification-service/cmd \
	leaderboard-service|./leaderboard-service/cmd

CONSUMERS := \
	session-created-consumer|./quiz-session-service/cmd/session-created-consumer \
	user-joined-consumer|./quiz-session-service/cmd/user-joined-consumer

SERVICE_PORTS := 8080 8082 8083 8084 8085 8086

# Colors (optional; ignored if terminal has no color)
CYAN  := \033[36m
GREEN := \033[32m
BOLD  := \033[1m
RESET := \033[0m

.PHONY: help start stop down restart \
	infra-up infra-down infra-status \
	consul-up compose-up redis-up db-up kafka-up kafka-topics \
	migrate migrate-quiz-session migrate-quiz-bank migrate-create \
	protoc deps \
	build build-clean \
	run run-api-gateway run-id-generation run-user-service \
	run-quiz-bank run-quiz-session run-notification run-leaderboard \
	run-session-created-consumer run-user-joined-consumer \
	web web-install \
	status clean

# =============================================================================
# Help
# =============================================================================

help: ## Show this help
	@printf "$(BOLD)Kootha$(RESET) — realtime quiz platform\n\n"
	@printf "$(BOLD)Usage:$(RESET)  make $(CYAN)<target>$(RESET)\n\n"
	@awk 'BEGIN {FS = ":.*##"; printf "$(BOLD)Targets:$(RESET)\n"} \
		/^##@/ {printf "\n  $(BOLD)%s$(RESET)\n", substr($$0, 5)} \
		/^[a-zA-Z0-9_-]+:.*?##/ {printf "  $(CYAN)%-28s$(RESET) %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@printf "\n"

##@ Quick start

start: infra-up migrate run ## Start infrastructure, migrate DBs, and run all services

stop: ## Stop all local Go services (keeps Docker infra running)
	@echo "Stopping Go services..."
	@if [ -f "$(PID_FILE)" ]; then \
		for pid in $$(cat "$(PID_FILE)"); do \
			kill $$pid 2>/dev/null || true; \
			pkill -P $$pid 2>/dev/null || true; \
		done; \
		rm -f "$(PID_FILE)"; \
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
	@rm -f "$(PID_FILE)"
	@echo "$(GREEN)All Go services stopped.$(RESET)"

down: stop infra-down ## Stop Go services and tear down infrastructure

restart: stop start ## Restart the full stack

##@ Infrastructure

infra-up: consul-up compose-up kafka-up ## Start Consul, Redis, Postgres, Kafka + topics
	@echo "Waiting for dependencies..."
	@sleep 3
	@$(MAKE) --no-print-directory kafka-topics
	@echo "$(GREEN)Infrastructure is ready.$(RESET)"

infra-down: ## Stop all Docker infrastructure
	@echo "Stopping infrastructure..."
	@docker compose -f $(COMPOSE_INFRA) down 2>/dev/null || true
	@docker compose -f $(COMPOSE_KAFKA) down 2>/dev/null || true
	@docker compose -f $(COMPOSE_QSESS) down 2>/dev/null || true
	@docker rm -f dev-consul 2>/dev/null || true
	@echo "$(GREEN)Infrastructure stopped.$(RESET)"

infra-status: ## Show status of infra containers
	@echo "$(BOLD)Docker containers:$(RESET)"
	@docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" \
		--filter name=kootha-redis \
		--filter name=quiz-bank-db \
		--filter name=user-db \
		--filter name=quiz-session-db \
		--filter name=kafka \
		--filter name=dev-consul 2>/dev/null || true

consul-up: ## Start Consul (service discovery)
	@echo "Starting Consul..."
	@docker inspect dev-consul >/dev/null 2>&1 \
		&& docker start dev-consul >/dev/null \
		|| docker run -d \
			-p 8500:8500 -p 8600:8600/udp \
			--name=dev-consul \
			hashicorp/consul:latest \
			agent -server -ui -node=server-1 -bootstrap-expect=1 -client=0.0.0.0 >/dev/null
	@echo "  Consul UI → http://localhost:8500"

compose-up: ## Start Redis + all Postgres instances
	@echo "Starting Redis and Postgres..."
	@docker compose -f $(COMPOSE_INFRA) up -d
	@docker compose -f $(COMPOSE_QSESS) up -d

kafka-up: ## Start Kafka
	@echo "Starting Kafka..."
	@docker compose -f $(COMPOSE_KAFKA) up -d

kafka-topics: ## Ensure Kafka topics exist
	@echo "Ensuring Kafka topics..."
	@go run ./cmd/create-kafka-topics

##@ Database

migrate: migrate-quiz-session migrate-quiz-bank ## Run all database migrations

migrate-quiz-session: ## Migrate quiz-session database
	@echo "Migrating quiz_session..."
	@migrate -path=$(MIGRATE_QUIZ_SESSION) -database "$(DSN_QUIZ_SESSION)" up

migrate-quiz-bank: ## Migrate quiz-bank database
	@echo "Migrating quiz_bank..."
	@migrate -path=$(MIGRATE_QUIZ_BANK) -database "$(DSN_QUIZ_BANK)" up

migrate-create: ## Create a new migration: make migrate-create name=add_foo service=quiz-session
	@test -n "$(name)" || (echo "Usage: make migrate-create name=<name> service=quiz-session|quiz-bank"; exit 1)
	@case "$(service)" in \
		quiz-session) dir="$(MIGRATE_QUIZ_SESSION)" ;; \
		quiz-bank)    dir="$(MIGRATE_QUIZ_BANK)" ;; \
		*) echo "service must be quiz-session or quiz-bank"; exit 1 ;; \
	esac; \
	migrate create -ext=sql -dir=$$dir -seq $(name)

##@ Codegen & dependencies

protoc: ## Regenerate gRPC / protobuf code from api/*.proto
	@echo "Generating protobuf stubs..."
	@protoc -I=api --go_out=. --go-grpc_out=. api/*.proto
	@echo "$(GREEN)Protobuf generation complete.$(RESET)"

deps: ## Download Go module dependencies
	@go mod download
	@go mod tidy

##@ Build

build: ## Build all service binaries into ./bin
	@mkdir -p "$(BIN_DIR)"
	@echo "Building services → $(BIN_DIR)"
	@for entry in $(SERVICES); do \
		name=$${entry%%|*}; \
		pkg=$${entry##*|}; \
		echo "  $$name"; \
		go build -o "$(BIN_DIR)/$$name" $$pkg; \
	done
	@for entry in $(CONSUMERS); do \
		name=$${entry%%|*}; \
		pkg=$${entry##*|}; \
		echo "  $$name"; \
		go build -o "$(BIN_DIR)/$$name" $$pkg; \
	done
	@echo "$(GREEN)Build complete.$(RESET)"

build-clean: ## Remove built binaries
	@rm -rf "$(BIN_DIR)"
	@echo "Removed $(BIN_DIR)"

##@ Run (development)

run: ## Run all Go services + consumers (Ctrl+C to stop)
	@mkdir -p "$(PID_DIR)"
	@$(MAKE) --no-print-directory stop >/dev/null 2>&1 || true
	@echo "$(BOLD)Starting all services$(RESET) (Ctrl+C or \`make stop\` to stop)..."
	@set -e; \
	pids=""; \
	cleanup() { \
		echo ""; \
		echo "Stopping all services..."; \
		for pid in $$pids; do \
			kill $$pid 2>/dev/null || true; \
			pkill -P $$pid 2>/dev/null || true; \
		done; \
		rm -f "$(PID_FILE)"; \
		$(MAKE) --no-print-directory stop >/dev/null 2>&1 || true; \
		echo "$(GREEN)All services stopped.$(RESET)"; \
	}; \
	trap cleanup INT TERM EXIT; \
	go run ./id-generation-service/cmd & pids="$$pids $$!"; \
	go run ./user-service/cmd & pids="$$pids $$!"; \
	go run ./quiz-bank-service/cmd & pids="$$pids $$!"; \
	go run ./quiz-session-service/cmd & pids="$$pids $$!"; \
	go run ./api-gateway/cmd & pids="$$pids $$!"; \
	go run ./notification-service/cmd & pids="$$pids $$!"; \
	go run ./leaderboard-service/cmd & pids="$$pids $$!"; \
	go run ./quiz-session-service/cmd/session-created-consumer & pids="$$pids $$!"; \
	go run ./quiz-session-service/cmd/user-joined-consumer & pids="$$pids $$!"; \
	echo $$pids > "$(PID_FILE)"; \
	echo "PIDs: $$pids"; \
	echo "Frontend: make web   (or: cd web && bun run dev)"; \
	wait

run-api-gateway: ## Run API gateway only
	@go run ./api-gateway/cmd

run-id-generation: ## Run ID generation service only
	@go run ./id-generation-service/cmd

run-user-service: ## Run user service only
	@go run ./user-service/cmd

run-quiz-bank: ## Run quiz bank service only
	@go run ./quiz-bank-service/cmd

run-quiz-session: ## Run quiz session service only
	@go run ./quiz-session-service/cmd

run-notification: ## Run notification service only
	@go run ./notification-service/cmd

run-leaderboard: ## Run leaderboard service only
	@go run ./leaderboard-service/cmd

run-session-created-consumer: ## Run session-created Kafka consumer
	@go run ./quiz-session-service/cmd/session-created-consumer

run-user-joined-consumer: ## Run user-joined Kafka consumer
	@go run ./quiz-session-service/cmd/user-joined-consumer

##@ Frontend

web-install: ## Install frontend dependencies (Bun)
	@cd web && bun install

web: ## Start Next.js dev server
	@cd web && bun run dev

##@ Utilities

status: infra-status ## Show infra status and listening service ports
	@echo ""
	@echo "$(BOLD)Service ports:$(RESET)"
	@for port in $(SERVICE_PORTS); do \
		pids=$$(lsof -tiTCP:$$port -sTCP:LISTEN 2>/dev/null || true); \
		if [ -n "$$pids" ]; then \
			printf "  :%-5s $(GREEN)LISTEN$(RESET)  pid %s\n" "$$port" "$$pids"; \
		else \
			printf "  :%-5s —\n" "$$port"; \
		fi; \
	done

clean: stop build-clean ## Stop services and remove binaries / PID files
	@rm -rf "$(PID_DIR)"
	@echo "$(GREEN)Clean complete.$(RESET)"

# -----------------------------------------------------------------------------
# Backward-compatible aliases (older docs / muscle memory)
# -----------------------------------------------------------------------------
.PHONY: start-infra start-consul start-kafka start-quiz-session-db \
	migrate-all create-kafka-topics run-all stop-all start-consumer-all \
	create-migrate-quiz-session-db

start-infra: infra-up
start-consul: consul-up
start-kafka: kafka-up
start-quiz-session-db:
	@docker compose -f $(COMPOSE_QSESS) up -d
migrate-all: migrate
create-kafka-topics: kafka-topics
run-all: run
stop-all: stop
start-consumer-all: run
create-migrate-quiz-session-db:
	@$(MAKE) migrate-create name=$(name) service=quiz-session

# Quiet aliases (still invokable; omitted from help)
redis-up: compose-up
db-up: compose-up
