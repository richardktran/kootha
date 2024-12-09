run-quiz-session:
	@echo "Running quiz session..."
	@go run quiz-session-service/cmd/*.go
run-id-generation:
	@echo "Running id generation service..."
	@go run id-generation-service/cmd/*.go

run-user-service:
	@echo "Running user service..."
	@go run user-service/cmd/*.go

protoc:
	@echo "Generating protobuf files..."
	@protoc -I=api --go_out=. --go-grpc_out=. api/*.proto