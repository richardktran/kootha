package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/richardktran/realtime-quiz/leaderboard-service/internal/consumer"
	"github.com/richardktran/realtime-quiz/leaderboard-service/internal/service/leaderboard"
	"github.com/richardktran/realtime-quiz/pkg/cache/redis"
	"github.com/richardktran/realtime-quiz/pkg/config"
	"github.com/richardktran/realtime-quiz/pkg/env"
	"github.com/richardktran/realtime-quiz/pkg/message-broker/kafka"
)

type serviceConfig struct {
	Redis redisConfig `yaml:"redis"`
}

type redisConfig struct {
	Addr string `yaml:"addr"`
}

func main() {
	fmt.Println("Starting leaderboard service...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received interrupt signal, shutting down...")
		cancel()
	}()

	var cfg serviceConfig
	if err := config.LoadYAML("leaderboard-service/configs/base.yaml", &cfg); err != nil {
		panic(err)
	}

	redisAddr := env.OverlayString(cfg.Redis.Addr, "REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisClient, err := redis.New(redisAddr)
	if err != nil {
		panic(err)
	}
	defer redisClient.Close()

	producer, err := kafka.NewProducer()
	if err != nil {
		panic(err)
	}
	defer producer.Close()

	svc := leaderboard.New(redisClient)
	c, err := consumer.New(producer, svc)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	log.Println("Leaderboard consumer running...")
	if err := c.Run(ctx); err != nil {
		log.Fatalf("consumer error: %v", err)
	}

	log.Println("Leaderboard service shut down gracefully.")
}
