package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/richardktran/realtime-quiz/gen"
	"github.com/richardktran/realtime-quiz/pkg/cache/redis"
	"github.com/richardktran/realtime-quiz/pkg/config"
	"github.com/richardktran/realtime-quiz/pkg/discovery"
	"github.com/richardktran/realtime-quiz/pkg/env"
	"github.com/richardktran/realtime-quiz/pkg/health"
	"github.com/richardktran/realtime-quiz/pkg/registry"
	grpcHandler "github.com/richardktran/realtime-quiz/quiz-bank-service/internal/handler/grpc"
	"github.com/richardktran/realtime-quiz/quiz-bank-service/internal/repository/postgres"
	"github.com/richardktran/realtime-quiz/quiz-bank-service/internal/service/quizbank"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const serviceName = "quiz-bank"

type serviceConfig struct {
	APIConfig   apiConfig   `yaml:"api"`
	DBConfig    dbConfig    `yaml:"db"`
	RedisConfig redisConfig `yaml:"redis"`
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

type redisConfig struct {
	Addr string `yaml:"addr"`
}

func main() {
	var cfg serviceConfig
	if err := config.LoadYAML("quiz-bank-service/configs/base.yaml", &cfg); err != nil {
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

	redisAddr := env.OverlayString(cfg.RedisConfig.Addr, "REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	cache, err := redis.New(redisAddr)
	if err != nil {
		panic(err)
	}
	defer cache.Close()

	svc := quizbank.New(repo, cache)
	h := grpcHandler.New(svc)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	health.RegisterGRPC(server)
	reflection.Register(server)
	gen.RegisterQuizBankServiceServer(server, h)

	if err := server.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
