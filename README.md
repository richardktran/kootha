# Setup Protobuf and gRPC
## Install protobuf compiler
For MacOS:
```bash
brew install protobuf
```
## Install protoc-gen-go
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## Update your PATH so that the protoc compiler can find the plugins:
```bash
export PATH="$PATH:$(go env GOPATH)/bin" >> ~/.zshrc
```

## Run Hashicorp Consul in Docker
```bash
docker run -d -p 8500:8500 -p 8600:8600/udp --name=dev-consul hashicorp/consul:latest agent -server -ui -node=server-1 -bootstrap-expect=1 -client=0.0.0.0
```

## Run Kafka in Docker
Build the image and run the container
```bash
cd docker/kafka
docker compose up -d
```
Access Kafka
```bash
docker exec -it kafka bash
cd /opt/kafka/bin/
```

Create Topic
```bash
./kafka-topics.sh --bootstrap-server localhost:9092 --create --topic test 
```

Send a message to the topic (Open a new terminal)
```bash
./kafka-console-producer.sh --bootstrap-server localhost:9092 --topic test
<Enter your message>
```
Consume messages from the topic (Open a new terminal)
```bash
./kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic test --from-beginning

```