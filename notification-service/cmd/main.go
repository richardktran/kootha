package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/richardktran/realtime-quiz/notification-service/internal/consumer"
	"github.com/richardktran/realtime-quiz/notification-service/internal/fanout"
	quizsessionGW "github.com/richardktran/realtime-quiz/notification-service/internal/gateway/quizsession"
	"github.com/richardktran/realtime-quiz/notification-service/internal/ws"
	redisclient "github.com/richardktran/realtime-quiz/pkg/cache/redis"
	"github.com/richardktran/realtime-quiz/pkg/discovery"
	"github.com/richardktran/realtime-quiz/pkg/discovery/consul"
	"github.com/richardktran/realtime-quiz/pkg/message-broker/kafka"
	"github.com/richardktran/realtime-quiz/pkg/topics"
	"gopkg.in/yaml.v3"
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

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Upgrade, Connection")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Header.Get("Upgrade") == "websocket" {
			w.Header().Set("Upgrade", "websocket")
			w.Header().Set("Connection", "Upgrade")
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	f, err := os.Open("notification-service/configs/base.yaml")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var cfg serviceConfig
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		panic(err)
	}

	port := cfg.APIConfig.Port
	log.Printf("starting %s service on port %s...", serviceName, port)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		log.Println("received interrupt signal, shutting down...")
		cancel()
	}()

	registry, err := consul.NewRegistry("localhost:8500")
	if err != nil {
		panic(err)
	}

	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("localhost:%s", port)); err != nil {
		panic(err)
	}
	defer registry.Deregister(ctx, instanceID)

	go func() {
		for {
			if err := registry.ReportHealthyState(instanceID, serviceName); err != nil {
				log.Printf("failed to report healthy state: %v", err)
			}
			time.Sleep(5 * time.Second)
		}
	}()

	redis, err := redisclient.New(cfg.RedisConfig.Addr)
	if err != nil {
		panic(err)
	}
	defer redis.Close()

	hub := ws.NewHub(redis)
	fanoutPublisher := fanout.NewPublisher(redis)
	quizGW := quizsessionGW.New(registry)
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
	mux.HandleFunc("/ws", corsMiddleware(http.HandlerFunc(wsServer.HandleWebSocket)).ServeHTTP)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("WebSocket server listening on %s/ws", addr)
	if err := http.ListenAndServe(addr, corsMiddleware(mux)); err != nil {
		log.Fatalf("failed to start HTTP server: %v", err)
	}
}
