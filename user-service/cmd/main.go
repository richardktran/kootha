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
	idGenerationGateway "github.com/richardktran/realtime-quiz/user-service/internal/gateway/idgeneration/grpc"
	grpcHandler "github.com/richardktran/realtime-quiz/user-service/internal/handler/grpc"
	"github.com/richardktran/realtime-quiz/user-service/internal/repository/memory"
	"github.com/richardktran/realtime-quiz/user-service/internal/service/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v3"
)

const serviceName = "user-service"

type serviceConfig struct {
	APIConfig apiConfig `yaml:"api"`
}

type apiConfig struct {
	Port string `yaml:"port"`
}

func main() {
	f, err := os.Open("user-service/configs/base.yaml")
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

	registry, err := consul.NewRegistry("localhost:8500")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	instanceId := discovery.GenerateInstanceID(serviceName)

	if err := registry.Register(ctx, instanceId, serviceName, fmt.Sprintf("localhost:%v", port)); err != nil {
		panic(err)
	}

	go func() {
		for {
			if err := registry.ReportHealthyState(instanceId, serviceName); err != nil {
				log.Printf("Failed to report healthy state: %v", err)
				time.Sleep(5 * time.Second)
			}
		}
	}()
	defer registry.Deregister(ctx, instanceId)

	idGenerationGateway := idGenerationGateway.New(registry)

	repo := memory.New()
	svc := user.New(repo)
	h := grpcHandler.New(svc, idGenerationGateway)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	reflection.Register(server)
	gen.RegisterUserServiceServer(server, h)

	if err := server.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
