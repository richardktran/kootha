package kafka

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Consumer struct {
	consumer *kafka.Consumer
}

// NewConsumerGroup creates a new Kafka consumer group.
func NewConsumerGroup(groupId string) (*Consumer, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"group.id":          groupId,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, err
	}

	return &Consumer{
		consumer: c,
	}, nil
}

// Consume starts consuming messages from the specified topics.
func (c *Consumer) Consume(ctx context.Context, topics []string, handler func(message []byte, metadata map[string]interface{}) error) error {
	err := c.consumer.SubscribeTopics(topics, nil)
	if err != nil {
		return err
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	run := true
	for run {
		select {
		case <-ctx.Done():
			log.Println("Context canceled, stopping consumer...")
			return nil
		case sig := <-sigChan:
			switch sig {
			case os.Interrupt:
				run = false
			}
		default:
			ev := c.consumer.Poll(100)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				metadata := map[string]interface{}{
					"topic":     *e.TopicPartition.Topic,
					"partition": e.TopicPartition.Partition,
					"offset":    e.TopicPartition.Offset,
					"timestamp": e.Timestamp,
					"headers":   e.Headers,
				}
				if err := handler(e.Value, metadata); err != nil {
					log.Printf("Error processing message: %v\n", err)
				}
			case kafka.Error:
				return e
			}
		}
	}

	return nil
}

// Close shuts down the consumer gracefully.
func (c *Consumer) Close() {
	c.consumer.Close()
}
