package consumer

import (
	"context"
	"encoding/json"
	"log"

	redisclient "github.com/richardktran/realtime-quiz/pkg/cache/redis"
	"github.com/richardktran/realtime-quiz/pkg/events"
	"github.com/richardktran/realtime-quiz/pkg/topics"
	"github.com/richardktran/realtime-quiz/notification-service/internal/ws"
)

type RedisFanoutSubscriber struct {
	redis *redisclient.Client
	hub   *ws.Hub
}

func NewRedisFanoutSubscriber(redis *redisclient.Client, hub *ws.Hub) *RedisFanoutSubscriber {
	return &RedisFanoutSubscriber{redis: redis, hub: hub}
}

func (s *RedisFanoutSubscriber) Start(ctx context.Context) {
	pubsub := s.redis.Subscribe(ctx, topics.NotificationFanout)
	ch := pubsub.Channel()

	go func() {
		<-ctx.Done()
		_ = pubsub.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}

			var fanout events.FanoutMessage
			if err := json.Unmarshal([]byte(msg.Payload), &fanout); err != nil {
				log.Printf("failed to decode fanout message: %v", err)
				continue
			}

			s.hub.Broadcast(fanout.SessionID, fanout.Type, fanout.Payload)
		}
	}
}
