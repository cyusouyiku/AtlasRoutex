package queue

import "context"

type Message struct {
Topic string
Key   string
Body  []byte
}

type Queue interface {
Publish(ctx context.Context, msg Message) error
Consume(ctx context.Context, topic string) (<-chan Message, error)
Close() error
}
