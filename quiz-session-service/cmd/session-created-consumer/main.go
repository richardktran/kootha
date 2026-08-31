package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/richardktran/realtime-quiz/pkg/config"
	"github.com/richardktran/realtime-quiz/pkg/env"
	"github.com/richardktran/realtime-quiz/pkg/message-broker/kafka"
	"github.com/richardktran/realtime-quiz/pkg/topics"
	"github.com/richardktran/realtime-quiz/quiz-session-service/internal/repository/postgres"
	"github.com/richardktran/realtime-quiz/quiz-session-service/internal/workers"
	"github.com/richardktran/realtime-quiz/quiz-session-service/pkg/model"
)

var groupId = "session-created"

type serviceConfig struct {
	DBConfig dbConfig `yaml:"db"`
}

type dbConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

func main() {
	fmt.Println("Start session created consumer...")
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
	if err := config.LoadYAML("quiz-session-service/configs/base.yaml", &cfg); err != nil {
		panic(err)
	}

	consumer, err := kafka.NewConsumerGroup(groupId)
	if err != nil {
		panic(err)
	}

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

	worker := workers.NewQuizCreatedWorker(repo)

	handler := func(message []byte, metadata map[string]interface{}) error {
		select {
		case <-ctx.Done():
			log.Println("Handler context canceled")
			return ctx.Err()
		default:
		}

		var session model.QuizSession
		if err := json.Unmarshal(message, &session); err != nil {
			log.Println("Unmarshal error: ", err.Error())
		}

		_, err := worker.StoreQuizSession(ctx, &session)

		if err != nil {
			log.Println("Error while storing the session: ", err.Error())
		}

		log.Printf("Consume the data: %s", string(message))
		log.Printf("Metadata: topic=%v, partition=%v, offset=%v\n", metadata["topic"], metadata["partition"], metadata["offset"])

		return nil
	}

	fmt.Println("Start consume...")

	if err := consumer.Consume(ctx, []string{topics.QuizSessionCreated}, handler); err != nil {
		log.Fatalf("Error while consuming messages: %v", err)
	}

	log.Println("Consumer shut down gracefully.")
}
