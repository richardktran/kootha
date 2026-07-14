package fanout

import (
	"context"
	"encoding/json"

	redisclient "github.com/richardktran/realtime-quiz/pkg/cache/redis"
	"github.com/richardktran/realtime-quiz/pkg/events"
	"github.com/richardktran/realtime-quiz/pkg/topics"
)

type Publisher struct {
	redis *redisclient.Client
}

func NewPublisher(redis *redisclient.Client) *Publisher {
	return &Publisher{redis: redis}
}

func (p *Publisher) Publish(ctx context.Context, sessionID, eventType string, payload interface{}) error {
	msg := events.FanoutMessage{
		SessionID: sessionID,
		Type:      eventType,
		Payload:   payload,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return p.redis.Publish(ctx, topics.NotificationFanout, string(data))
}
