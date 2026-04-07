package handlers

import "context"

type FeedbackHandler struct{}

func NewFeedbackHandler() *FeedbackHandler { return &FeedbackHandler{} }

func (h *FeedbackHandler) Handle(_ context.Context, _ []byte) error { return nil }
