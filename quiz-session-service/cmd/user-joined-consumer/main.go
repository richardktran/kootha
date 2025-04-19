package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/richardktran/realtime-quiz/pkg/message-broker/kafka"
	"github.com/richardktran/realtime-quiz/pkg/topics"
	"github.com/richardktran/realtime-quiz/quiz-session-service/internal/repository/postgres"
	"github.com/richardktran/realtime-quiz/quiz-session-service/internal/workers"
	"gopkg.in/yaml.v3"
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

type userJoinedData struct {
	SessionId string `json:"sessionId"`
	UserId    string `json:"userId"`
}

func main() {
	fmt.Println("Start user joined consumer...")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	go func() {
		<-sigChan
		log.Println("Received interrupt signal, shutting down...")
		cancel()
	}()

	f, err := os.Open("quiz-session-service/configs/base.yaml")

	if err != nil {
		panic(err)
	}
	defer f.Close()

	var cfg serviceConfig

	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		panic(err)
	}

	consumer, err := kafka.NewConsumerGroup(groupId)
	topics := []string{
		topics.UserJoinedQuiz,
	}

	if err != nil {
		panic(err)
	}

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

	worker := workers.NewUserJoinedWorker(repo)

	handler := func(message []byte, metadata map[string]interface{}) error {
		select {
		case <-ctx.Done():
			log.Println("Handler context canceled")
			return ctx.Err()
		default:
		}

		log.Printf("Consume the data: %s", string(message))
		log.Printf("Metadata: topic=%v, partition=%v, offset=%v\n", metadata["topic"], metadata["partition"], metadata["offset"])

		var joinedData userJoinedData
		if err := json.Unmarshal(message, &joinedData); err != nil {
			log.Println("Unmarshal error: ", err.Error())
		}

		err := worker.JoinQuiz(ctx, joinedData.SessionId, joinedData.UserId)
		if err != nil {
			log.Println("Error while joining the quiz: ", err.Error())
		}

		return nil
	}

	fmt.Println("Start consume...")

	if err := consumer.Consume(ctx, topics, handler); err != nil {
		log.Fatalf("Error while consuming messages: %v", err)
	}

	log.Println("Consumer shut down gracefully.")
}
