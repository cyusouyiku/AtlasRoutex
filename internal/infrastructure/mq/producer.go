package mq

import (
"context"

"atlas-routex/pkg/queue"
)

type Producer struct { Q queue.Queue }

func NewProducer(q queue.Queue) *Producer { return &Producer{Q: q} }

func (p *Producer) Publish(ctx context.Context, topic string, body []byte) error {
return p.Q.Publish(ctx, queue.Message{Topic: topic, Body: body})
}
