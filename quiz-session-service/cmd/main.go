package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/richardktran/realtime-quiz/gen"
	"github.com/richardktran/realtime-quiz/pkg/cache/redis"
	"github.com/richardktran/realtime-quiz/pkg/discovery"
	"github.com/richardktran/realtime-quiz/pkg/discovery/consul"
	"github.com/richardktran/realtime-quiz/pkg/message-broker/kafka"
	idGenerationGateway "github.com/richardktran/realtime-quiz/quiz-session-service/internal/gateway/idgeneration"
	quizBankGateway "github.com/richardktran/realtime-quiz/quiz-session-service/internal/gateway/quizbank"
	grpcHandler "github.com/richardktran/realtime-quiz/quiz-session-service/internal/handler/grpc"
	"github.com/richardktran/realtime-quiz/quiz-session-service/internal/repository/postgres"
	"github.com/richardktran/realtime-quiz/quiz-session-service/internal/service/quizsession"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v3"
)

const serviceName = "quiz-session"

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

	registry, err := consul.NewRegistry("localhost:8500")
	if err != nil {
		panic(err)
	}

	instanceId := discovery.GenerateInstanceID(serviceName)
	ctx := context.Background()

	if err := registry.Register(ctx, instanceId, serviceName, fmt.Sprintf("localhost:%v", port)); err != nil {
		panic(err)
	}

	go func() {
		for {
			if err := registry.ReportHealthyState(instanceId, serviceName); err != nil {
				log.Printf("Failed to report healthy state: %v", err)
			}
			time.Sleep(5 * time.Second)
		}
	}()
	defer registry.Deregister(ctx, instanceId)

	producer, err := kafka.NewProducer()
	if err != nil {
		panic(err)
	}
	defer producer.Close()

	redisAddr := cfg.RedisConfig.Addr
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	cache, err := redis.New(redisAddr)
	if err != nil {
		panic(err)
	}
	defer cache.Close()

	idGenGW := idGenerationGateway.New(registry)
	quizBankGW := quizBankGateway.New(registry)

	repo, err := postgres.New(postgres.Config{
		Host:     cfg.DBConfig.Host,
		Port:     cfg.DBConfig.Port,
		User:     cfg.DBConfig.User,
		Password: cfg.DBConfig.Password,
		DBName:   cfg.DBConfig.DBName,
	})
	if err != nil {
		panic(err)
	}

	svc := quizsession.New(repo, producer, cache, quizBankGW)
	h := grpcHandler.New(svc, idGenGW)

	server := grpc.NewServer()
	reflection.Register(server)
	gen.RegisterQuizSessionServiceServer(server, h)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("%s gRPC listening on :%v", serviceName, port)
	if err := server.Serve(listener); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}
