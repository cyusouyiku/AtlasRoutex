package grpc

import "context"

type PlanService struct{}

func NewPlanService() *PlanService { return &PlanService{} }

func (s *PlanService) Ping(_ context.Context) string { return "pong" }
