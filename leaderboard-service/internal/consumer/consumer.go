package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/richardktran/realtime-quiz/leaderboard-service/internal/service/leaderboard"
	"github.com/richardktran/realtime-quiz/pkg/events"
	"github.com/richardktran/realtime-quiz/pkg/message-broker/kafka"
	"github.com/richardktran/realtime-quiz/pkg/topics"
)

const groupID = "leaderboard"

type Consumer struct {
	consumer *kafka.Consumer
	producer *kafka.Producer
	service  *leaderboard.Service
}

func New(producer *kafka.Producer, service *leaderboard.Service) (*Consumer, error) {
	c, err := kafka.NewConsumerGroup(groupID)
	if err != nil {
		return nil, fmt.Errorf("create kafka consumer: %w", err)
	}

	return &Consumer{
		consumer: c,
		producer: producer,
		service:  service,
	}, nil
}

func (c *Consumer) Run(ctx context.Context) error {
	kafkaTopics := []string{
		topics.AnswerSubmitted,
		topics.SessionEnd,
	}

	return c.consumer.Consume(ctx, kafkaTopics, c.handleMessage)
}

func (c *Consumer) Close() {
	c.consumer.Close()
}

func (c *Consumer) handleMessage(message []byte, metadata map[string]interface{}) error {
	topic, _ := metadata["topic"].(string)

	switch topic {
	case topics.AnswerSubmitted:
		return c.handleAnswerSubmitted(context.Background(), message)
	case topics.SessionEnd:
		return c.handleSessionEnd(context.Background(), message)
	default:
		log.Printf("ignoring unknown topic: %s", topic)
		return nil
	}
}

func (c *Consumer) handleAnswerSubmitted(ctx context.Context, message []byte) error {
	var evt events.AnswerSubmitted
	if err := json.Unmarshal(message, &evt); err != nil {
		return fmt.Errorf("unmarshal answer-submitted: %w", err)
	}

	if err := c.service.HandleAnswer(ctx, evt); err != nil {
		return fmt.Errorf("handle answer: %w", err)
	}

	return c.publishRanking(ctx, evt.SessionID, false)
}

func (c *Consumer) handleSessionEnd(ctx context.Context, message []byte) error {
	var evt events.SessionEnd
	if err := json.Unmarshal(message, &evt); err != nil {
		return fmt.Errorf("unmarshal session-end: %w", err)
	}

	if err := c.publishRanking(ctx, evt.SessionID, true); err != nil {
		return err
	}

	return nil
}

func (c *Consumer) publishRanking(ctx context.Context, sessionID string, final bool) error {
	participants, err := c.service.GetRanking(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("get ranking: %w", err)
	}

	evt := events.RankingUpdated{
		SessionID:    sessionID,
		Participants: participants,
	}

	payload, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("marshal ranking-updated: %w", err)
	}

	if err := c.producer.Produce(ctx, topics.RankingUpdated, payload); err != nil {
		return fmt.Errorf("publish ranking-updated: %w", err)
	}

	if final {
		if err := c.producer.Produce(ctx, topics.QuizCompleted, payload); err != nil {
			return fmt.Errorf("publish quiz-completed: %w", err)
		}
	}

	return nil
}
