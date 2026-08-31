package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/richardktran/realtime-quiz/gen"
	grpcHandler "github.com/richardktran/realtime-quiz/id-generation-service/internal/handler/grpc"
	"github.com/richardktran/realtime-quiz/id-generation-service/internal/service/idgeneration"
	"github.com/richardktran/realtime-quiz/pkg/config"
	"github.com/richardktran/realtime-quiz/pkg/discovery"
	"github.com/richardktran/realtime-quiz/pkg/env"
	"github.com/richardktran/realtime-quiz/pkg/health"
	"github.com/richardktran/realtime-quiz/pkg/registry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const serviceName = "id-generation-service"

type serviceConfig struct {
	APIConfig apiConfig `yaml:"api"`
}

type apiConfig struct {
	Port string `yaml:"port"`
}

func main() {
	var cfg serviceConfig
	if err := config.LoadYAML("id-generation-service/configs/base.yaml", &cfg); err != nil {
		panic(err)
	}

	port := env.OverlayString(fmt.Sprint(cfg.APIConfig.Port), "API_PORT")
	log.Printf("Starting the %s service with port %v...", serviceName, port)

	reg, err := registry.New()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	stop, err := discovery.StartHeartbeat(ctx, reg, serviceName, discovery.AdvertiseAddr(port))
	if err != nil {
		panic(err)
	}
	defer stop()

	svc := idgeneration.New()
	h := grpcHandler.New(svc)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	health.RegisterGRPC(grpcServer)
	reflection.Register(grpcServer)
	gen.RegisterIdGenerationServiceServer(grpcServer, h)

	if err := grpcServer.Serve(listener); err != nil {
		panic(err)
	}
}
