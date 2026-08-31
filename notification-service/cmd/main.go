package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/richardktran/realtime-quiz/notification-service/internal/consumer"
	"github.com/richardktran/realtime-quiz/notification-service/internal/fanout"
	quizsessionGW "github.com/richardktran/realtime-quiz/notification-service/internal/gateway/quizsession"
	"github.com/richardktran/realtime-quiz/notification-service/internal/ws"
	redisclient "github.com/richardktran/realtime-quiz/pkg/cache/redis"
	"github.com/richardktran/realtime-quiz/pkg/config"
	"github.com/richardktran/realtime-quiz/pkg/cors"
	"github.com/richardktran/realtime-quiz/pkg/discovery"
	"github.com/richardktran/realtime-quiz/pkg/env"
	"github.com/richardktran/realtime-quiz/pkg/health"
	"github.com/richardktran/realtime-quiz/pkg/message-broker/kafka"
	"github.com/richardktran/realtime-quiz/pkg/registry"
	"github.com/richardktran/realtime-quiz/pkg/topics"
)

const serviceName = "notification"

type serviceConfig struct {
	APIConfig   apiConfig   `yaml:"api"`
	RedisConfig redisConfig `yaml:"redis"`
}

type apiConfig struct {
	Port string `yaml:"port"`
}

type redisConfig struct {
	Addr string `yaml:"addr"`
}

func main() {
	var cfg serviceConfig
	if err := config.LoadYAML("notification-service/configs/base.yaml", &cfg); err != nil {
		panic(err)
	}

	port := env.OverlayString(fmt.Sprint(cfg.APIConfig.Port), "API_PORT")
	log.Printf("starting %s service on port %s...", serviceName, port)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("received interrupt signal, shutting down...")
		cancel()
	}()

	reg, err := registry.New()
	if err != nil {
		panic(err)
	}

	stop, err := discovery.StartHeartbeat(ctx, reg, serviceName, discovery.AdvertiseAddr(port))
	if err != nil {
		panic(err)
	}
	defer stop()

	redisAddr := env.OverlayString(cfg.RedisConfig.Addr, "REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redis, err := redisclient.New(redisAddr)
	if err != nil {
		panic(err)
	}
	defer redis.Close()

	hub := ws.NewHub(redis)
	fanoutPublisher := fanout.NewPublisher(redis)
	quizGW := quizsessionGW.New(reg)
	wsServer := ws.NewServer(hub, quizGW, fanoutPublisher)

	kafkaTopics := []string{
		topics.UserJoinedQuiz,
		topics.SessionStart,
		topics.ChangeQuestion,
		topics.RankingUpdated,
		topics.QuestionResult,
		topics.SessionEnd,
	}
	if err := kafka.EnsureTopics(kafkaTopics, 1, 1); err != nil {
		log.Printf("ensure kafka topics: %v (continuing)", err)
	}

	kafkaConsumer, err := consumer.NewKafkaConsumer(fanoutPublisher)
	if err != nil {
		panic(err)
	}
	defer kafkaConsumer.Close()

	go func() {
		if err := kafkaConsumer.Start(ctx); err != nil {
			log.Printf("kafka consumer error: %v", err)
		}
	}()

	redisSubscriber := consumer.NewRedisFanoutSubscriber(redis, hub)
	go redisSubscriber.Start(ctx)

	mux := http.NewServeMux()
	health.MountHTTP(mux)
	mux.HandleFunc("/ws", wsServer.HandleWebSocket)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("WebSocket server listening on %s/ws", addr)
	if err := http.ListenAndServe(addr, cors.Middleware(mux)); err != nil {
		log.Fatalf("failed to start HTTP server: %v", err)
	}
}
