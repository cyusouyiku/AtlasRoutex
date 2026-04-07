package mq

import (
"context"

"atlas-routex/pkg/queue"
)

type Consumer struct { Q queue.Queue }

func NewConsumer(q queue.Queue) *Consumer { return &Consumer{Q: q} }

func (c *Consumer) Consume(ctx context.Context, topic string) (<-chan queue.Message, error) {
return c.Q.Consume(ctx, topic)
}
