run-quiz-session:
	@echo "Running quiz session..."
	@go run quiz-session-service/cmd/*.go
run-id-generation:
	@echo "Running id generation service..."
	@go run id-generation-service/cmd/*.go

run-user-service:
	@echo "Running user service..."
	@go run user-service/cmd/*.go

run-all:
	@echo "Running all services..."
	@go run quiz-session-service/cmd/*.go &
	@go run id-generation-service/cmd/*.go &
	@go run user-service/cmd/*.go

protoc:
	@echo "Generating protobuf files..."
	@protoc -I=api --go_out=. --go-grpc_out=. api/*.proto

start-consul:
	@echo "Starting consul..."
	@docker run -d -p 8500:8500 -p 8600:8600/udp --name=dev-consul hashicorp/consul:latest agent -server -ui -node=server-1 -bootstrap-expect=1 -client=0.0.0.0

start-kafka:
	@echo "Starting kafka..."
	@docker compose -f docker/kafka/docker-compose.yml up -d