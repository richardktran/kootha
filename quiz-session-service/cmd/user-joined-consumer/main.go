package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/richardktran/realtime-quiz/pkg/message-broker/kafka"
	"github.com/richardktran/realtime-quiz/pkg/topics"
)

var groupId = "user-joined"

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

	consumer, err := kafka.NewConsumerGroup(groupId)
	topics := []string{
		topics.UserJoinedQuiz,
	}

	if err != nil {
		panic(err)
	}

	handler := func(message []byte, metadata map[string]interface{}) error {
		select {
		case <-ctx.Done():
			log.Println("Handler context canceled")
			return ctx.Err()
		default:
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
