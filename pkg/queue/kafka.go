package queue

import (
"context"
"sync"
)

type InMemoryQueue struct {
mu    sync.RWMutex
topic map[string]chan Message
}

func NewInMemoryQueue() *InMemoryQueue {
return &InMemoryQueue{topic: make(map[string]chan Message)}
}

func (q *InMemoryQueue) Publish(_ context.Context, msg Message) error {
q.mu.RLock()
ch, ok := q.topic[msg.Topic]
q.mu.RUnlock()
if !ok {
q.mu.Lock()
if q.topic[msg.Topic] == nil {
q.topic[msg.Topic] = make(chan Message, 128)
}
ch = q.topic[msg.Topic]
q.mu.Unlock()
}
ch <- msg
return nil
}

func (q *InMemoryQueue) Consume(_ context.Context, topic string) (<-chan Message, error) {
q.mu.Lock()
defer q.mu.Unlock()
if q.topic[topic] == nil {
q.topic[topic] = make(chan Message, 128)
}
return q.topic[topic], nil
}

func (q *InMemoryQueue) Close() error {
q.mu.Lock()
defer q.mu.Unlock()
for k, ch := range q.topic {
close(ch)
delete(q.topic, k)
}
return nil
}
