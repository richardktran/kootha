package kafka

import "github.com/richardktran/realtime-quiz/pkg/env"

func brokers() string {
	return env.GetString("KAFKA_BROKERS", "localhost:9092")
}
