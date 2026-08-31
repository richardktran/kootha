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
	"github.com/richardktran/realtime-quiz/pkg/events"
	"github.com/richardktran/realtime-quiz/pkg/message-broker/kafka"
	"github.com/richardktran/realtime-quiz/pkg/topics"
	"github.com/richardktran/realtime-quiz/quiz-session-service/internal/repository/postgres"
	"github.com/richardktran/realtime-quiz/quiz-session-service/internal/workers"
)

var groupId = "user-joined"

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
	fmt.Println("Start user joined consumer...")
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

	worker := workers.NewUserJoinedWorker(repo)

	handler := func(message []byte, metadata map[string]interface{}) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var joinedData events.UserJoined
		if err := json.Unmarshal(message, &joinedData); err != nil {
			log.Println("Unmarshal error: ", err.Error())
			return nil
		}

		if err := worker.JoinQuiz(ctx, joinedData.SessionID, joinedData.UserID, joinedData.Name); err != nil {
			log.Println("Error while joining the quiz: ", err.Error())
		}

		return nil
	}

	if err := consumer.Consume(ctx, []string{topics.UserJoinedQuiz}, handler); err != nil {
		log.Fatalf("Error while consuming messages: %v", err)
	}
}
