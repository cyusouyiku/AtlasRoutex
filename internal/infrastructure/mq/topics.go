package mq

import "atlas-routex/pkg/queue"

const (
TopicPlanRequested = "plan.requested"
TopicFeedbackSaved = "feedback.saved"
)

func DefaultTopics() []string { return []string{TopicPlanRequested, TopicFeedbackSaved} }

func NewInMemoryQueue() queue.Queue { return queue.NewInMemoryQueue() }
