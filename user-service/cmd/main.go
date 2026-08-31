package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/richardktran/realtime-quiz/gen"
	"github.com/richardktran/realtime-quiz/pkg/config"
	"github.com/richardktran/realtime-quiz/pkg/discovery"
	"github.com/richardktran/realtime-quiz/pkg/env"
	"github.com/richardktran/realtime-quiz/pkg/health"
	"github.com/richardktran/realtime-quiz/pkg/registry"
	idGenerationGateway "github.com/richardktran/realtime-quiz/user-service/internal/gateway/idgeneration/grpc"
	grpcHandler "github.com/richardktran/realtime-quiz/user-service/internal/handler/grpc"
	"github.com/richardktran/realtime-quiz/user-service/internal/repository/postgres"
	"github.com/richardktran/realtime-quiz/user-service/internal/service/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const serviceName = "user-service"

type serviceConfig struct {
	APIConfig apiConfig `yaml:"api"`
	DBConfig  dbConfig  `yaml:"db"`
}

type apiConfig struct {
	Port string `yaml:"port"`
}

type dbConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

func main() {
	var cfg serviceConfig
	if err := config.LoadYAML("user-service/configs/base.yaml", &cfg); err != nil {
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

	idGenGW := idGenerationGateway.New(reg)

	repo, err := postgres.New(postgres.Config{
		Host:     env.OverlayString(cfg.DBConfig.Host, "DB_HOST"),
		Port:     env.OverlayString(fmt.Sprint(cfg.DBConfig.Port), "DB_PORT"),
		User:     env.OverlayString(cfg.DBConfig.User, "DB_USER"),
		Password: env.OverlayString(cfg.DBConfig.Password, "DB_PASSWORD"),
		DBName:   env.OverlayString(cfg.DBConfig.DBName, "DB_NAME"),
	})
	if err != nil {
		panic(err)
	}

	svc := user.New(repo)
	h := grpcHandler.New(svc, idGenGW)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	health.RegisterGRPC(server)
	reflection.Register(server)
	gen.RegisterUserServiceServer(server, h)

	if err := server.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
