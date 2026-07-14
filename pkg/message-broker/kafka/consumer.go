package kafka

import (
	"context"
	"log"
	"strings"
	"time"

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
// Retries subscription when topics are not yet created.
func (c *Consumer) Consume(ctx context.Context, topics []string, handler func(message []byte, metadata map[string]interface{}) error) error {
	for {
		err := c.consumer.SubscribeTopics(topics, nil)
		if err == nil {
			break
		}
		log.Printf("Kafka subscribe failed (%v); retrying in 2s...", err)
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(2 * time.Second):
		}
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Context canceled, stopping consumer...")
			return nil
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
				// Don't fatal on missing topics / transient broker errors.
				if e.Code() == kafka.ErrUnknownTopic || e.Code() == kafka.ErrUnknownTopicOrPart ||
					strings.Contains(e.String(), "Unknown topic") {
					log.Printf("Kafka topic not ready yet: %v (will keep polling)", e)
					continue
				}
				if e.IsFatal() {
					return e
				}
				log.Printf("Kafka non-fatal error: %v", e)
			}
		}
	}
}

// Close shuts down the consumer gracefully.
func (c *Consumer) Close() {
	c.consumer.Close()
}

// EnsureTopics creates topics if they do not already exist.
func EnsureTopics(topics []string, numPartitions int, replicationFactor int) error {
	admin, err := kafka.NewAdminClient(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
	})
	if err != nil {
		return err
	}
	defer admin.Close()

	if numPartitions <= 0 {
		numPartitions = 1
	}
	if replicationFactor <= 0 {
		replicationFactor = 1
	}

	specs := make([]kafka.TopicSpecification, 0, len(topics))
	for _, t := range topics {
		specs = append(specs, kafka.TopicSpecification{
			Topic:             t,
			NumPartitions:     numPartitions,
			ReplicationFactor: replicationFactor,
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	results, err := admin.CreateTopics(ctx, specs)
	if err != nil {
		return err
	}
	for _, r := range results {
		if r.Error.Code() != kafka.ErrNoError && r.Error.Code() != kafka.ErrTopicAlreadyExists {
			log.Printf("create topic %s: %v", r.Topic, r.Error)
		} else {
			log.Printf("topic ready: %s", r.Topic)
		}
	}
	return nil
}
