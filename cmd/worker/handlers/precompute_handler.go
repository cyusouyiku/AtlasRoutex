package handlers

import "context"

type PrecomputeHandler struct{}

func NewPrecomputeHandler() *PrecomputeHandler { return &PrecomputeHandler{} }

func (h *PrecomputeHandler) Handle(_ context.Context, _ []byte) error { return nil }
