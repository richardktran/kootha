start-quiz-session-db:
	@echo "Starting postgres..."
	@docker compose -f quiz-session-service/docker/database.yml up -d

run-quiz-session:
	@echo "Running quiz session..."
	@make start-quiz-session-db
	@go run quiz-session-service/cmd/*.go
	
run-id-generation:
	@echo "Running id generation service..."
	@go run id-generation-service/cmd/*.go

run-user-service:
	@echo "Running user service..."
	@go run user-service/cmd/*.go

run-session-created-consumer:
	@echo "Running session created consumer..."
	@go run quiz-session-service/cmd/session-created-consumer/*.go
run-user-joined-consumer:
	@echo "Running user joined consumer..."
	@go run quiz-session-service/cmd/user-joined-consumer/*.go

run-all:
	@echo "Running all services..."
	@go run quiz-session-service/cmd/*.go &
	@go run id-generation-service/cmd/*.go &
	@go run user-service/cmd/*.go

start-consumer-all:
	@echo "Running all consumers..."
	@go run quiz-session-service/cmd/session-created-consumer/*.go &
	@go run quiz-session-service/cmd/user-joined-consumer/*.go

protoc:
	@echo "Generating protobuf files..."
	@protoc -I=api --go_out=. --go-grpc_out=. api/*.proto

start-consul:
	@echo "Starting consul..."
	@docker run -d -p 8500:8500 -p 8600:8600/udp --name=dev-consul hashicorp/consul:latest agent -server -ui -node=server-1 -bootstrap-expect=1 -client=0.0.0.0

start-kafka:
	@echo "Starting kafka..."
	@docker compose -f docker/kafka/docker-compose.yml up -d


create-migrate-quiz-session-db:
	@echo "Creating migrate quiz session db..."
	@migrate create -ext=sql -dir=quiz-session-service/internal/database/migrations -seq $1

migrate-quiz-session-db:
	@echo "Migrating quiz session db..."
	@migrate -path=quiz-session-service/internal/database/migrations -database "postgres://richardktran:password@localhost:5433/quiz_session?sslmode=disable" up
