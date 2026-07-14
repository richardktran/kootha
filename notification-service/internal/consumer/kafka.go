package consumer

import (
	"context"
	"encoding/json"
	"log"

	"github.com/richardktran/realtime-quiz/notification-service/internal/fanout"
	"github.com/richardktran/realtime-quiz/pkg/events"
	"github.com/richardktran/realtime-quiz/pkg/message-broker/kafka"
	"github.com/richardktran/realtime-quiz/pkg/topics"
)

const groupID = "notification-service"

type KafkaConsumer struct {
	consumer *kafka.Consumer
	fanout   *fanout.Publisher
}

func NewKafkaConsumer(fanout *fanout.Publisher) (*KafkaConsumer, error) {
	consumer, err := kafka.NewConsumerGroup(groupID)
	if err != nil {
		return nil, err
	}

	return &KafkaConsumer{
		consumer: consumer,
		fanout:   fanout,
	}, nil
}

func (c *KafkaConsumer) Close() {
	c.consumer.Close()
}

func (c *KafkaConsumer) Start(ctx context.Context) error {
	kafkaTopics := []string{
		topics.UserJoinedQuiz,
		topics.SessionStart,
		topics.ChangeQuestion,
		topics.RankingUpdated,
		topics.QuestionResult,
		topics.SessionEnd,
	}

	return c.consumer.Consume(ctx, kafkaTopics, func(message []byte, metadata map[string]interface{}) error {
		topic, _ := metadata["topic"].(string)
		return c.handleMessage(ctx, topic, message)
	})
}

func (c *KafkaConsumer) handleMessage(ctx context.Context, topic string, message []byte) error {
	switch topic {
	case topics.UserJoinedQuiz:
		var evt events.UserJoined
		if err := json.Unmarshal(message, &evt); err != nil {
			log.Printf("failed to unmarshal user-joined-quiz: %v", err)
			return nil
		}
		return c.fanout.Publish(ctx, evt.SessionID, "participant_joined", map[string]interface{}{
			"participant": map[string]interface{}{
				"id":    evt.UserID,
				"name":  evt.Name,
				"score": 0,
			},
		})

	case topics.SessionStart:
		var evt events.SessionStart
		if err := json.Unmarshal(message, &evt); err != nil {
			log.Printf("failed to unmarshal session-start: %v", err)
			return nil
		}
		return c.fanout.Publish(ctx, evt.SessionID, "question_started", map[string]interface{}{
			"questionIndex": evt.QuestionIndex,
			"question":      evt.Question,
			"status":        "in-progress",
		})

	case topics.ChangeQuestion:
		var evt events.ChangeQuestion
		if err := json.Unmarshal(message, &evt); err != nil {
			log.Printf("failed to unmarshal change-question: %v", err)
			return nil
		}
		return c.fanout.Publish(ctx, evt.SessionID, "question_started", map[string]interface{}{
			"questionIndex": evt.QuestionIndex,
			"question":      evt.Question,
		})

	case topics.RankingUpdated:
		var evt events.RankingUpdated
		if err := json.Unmarshal(message, &evt); err != nil {
			log.Printf("failed to unmarshal ranking-updated: %v", err)
			return nil
		}
		return c.fanout.Publish(ctx, evt.SessionID, "leaderboard_update", map[string]interface{}{
			"participants": evt.Participants,
		})

	case topics.QuestionResult:
		var evt events.QuestionResult
		if err := json.Unmarshal(message, &evt); err != nil {
			log.Printf("failed to unmarshal question-result: %v", err)
			return nil
		}
		return c.fanout.Publish(ctx, evt.SessionID, "question_result", map[string]interface{}{
			"questionId":    evt.QuestionID,
			"questionIndex": evt.QuestionIndex,
			"correctAnswer": evt.CorrectAnswer,
			"reason":        evt.Reason,
		})

	case topics.SessionEnd:
		var evt events.SessionEnd
		if err := json.Unmarshal(message, &evt); err != nil {
			log.Printf("failed to unmarshal session-end: %v", err)
			return nil
		}
		return c.fanout.Publish(ctx, evt.SessionID, "quiz_ended", map[string]interface{}{
			"participants": evt.Participants,
		})
	}

	return nil
}
