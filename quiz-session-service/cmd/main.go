package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/richardktran/realtime-quiz/gen"
	"github.com/richardktran/realtime-quiz/pkg/discovery"
	"github.com/richardktran/realtime-quiz/pkg/discovery/consul"
	"github.com/richardktran/realtime-quiz/pkg/message-broker/kafka"
	idGenerationGateway "github.com/richardktran/realtime-quiz/quiz-session-service/internal/gateway/idgeneration"
	grpcHandler "github.com/richardktran/realtime-quiz/quiz-session-service/internal/handler/grpc"
	"github.com/richardktran/realtime-quiz/quiz-session-service/internal/repository/memory"
	"github.com/richardktran/realtime-quiz/quiz-session-service/internal/service/quizsession"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v3"
)

const serviceName = "quiz-session"

type serviceConfig struct {
	APIConfig apiConfig `yaml:"api"`
}

type apiConfig struct {
	Port string `yaml:"port"`
}

func main() {
	f, err := os.Open("quiz-session-service/configs/base.yaml")

	if err != nil {
		panic(err)
	}
	defer f.Close()

	var cfg serviceConfig

	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		panic(err)
	}

	port := cfg.APIConfig.Port
	log.Printf("Starting the %s service with port %v...", serviceName, port)

	// Service registry
	registry, err := consul.NewRegistry("localhost:8500")
	if err != nil {
		panic(err)
	}

	// Register the service
	instanceId := discovery.GenerateInstanceID(serviceName)
	ctx := context.Background()

	if err := registry.Register(ctx, instanceId, serviceName, fmt.Sprintf("localhost:%v", port)); err != nil {
		panic(err)
	}

	// Health check
	go func() {
		for {
			if err := registry.ReportHealthyState(instanceId, serviceName); err != nil {
				log.Printf("Failed to report healthy state: %v", err)
			}
			time.Sleep(5 * time.Second)
		}
	}()
	defer registry.Deregister(ctx, instanceId)

	// Kafka producer
	producer, err := kafka.NewProducer()
	if err != nil {
		panic(err)
	}
	defer producer.Close()

	idGenerationGateway := idGenerationGateway.New(registry)
	repo := memory.New()
	svc := quizsession.New(repo, producer)
	h := grpcHandler.New(svc, idGenerationGateway)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	reflection.Register(server)
	gen.RegisterQuizSessionServiceServer(server, h)

	if err := server.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
