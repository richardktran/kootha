package messagebroker

import "context"

type Producer interface {
	Produce(ctx context.Context, topic string, message []byte) error
	Close()
}

type Consumer interface {
	Consume(ctx context.Context, topics []string, handler func(message []byte, metadata map[string]interface{}) error) error
	Close()
}
