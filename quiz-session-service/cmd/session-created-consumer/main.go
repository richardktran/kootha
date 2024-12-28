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
	"github.com/richardktran/realtime-quiz/quiz-session-service/internal/repository/memory"
	"github.com/richardktran/realtime-quiz/quiz-session-service/internal/workers"
	"github.com/richardktran/realtime-quiz/quiz-session-service/pkg/model"
)

var groupId = "session-created"

func main() {
	fmt.Println("Start session created consumer...")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	go func() {
		<-sigChan
		log.Println("Received interrupt signal, shutting down...")
		cancel()
	}()

	consumer, err := kafka.NewConsumerGroup(groupId)
	topics := []string{
		topics.QuizSessionCreated,
	}

	if err != nil {
		panic(err)
	}

	repo := memory.New()
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

	if err := consumer.Consume(ctx, topics, handler); err != nil {
		log.Fatalf("Error while consuming messages: %v", err)
	}

	log.Println("Consumer shut down gracefully.")
}
