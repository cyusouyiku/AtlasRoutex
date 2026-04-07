package dto

import (
"atlas-routex/internal/application/planner"
)

func ToPlanInput(req PlanRequest) *planner.PlanInput {
return &planner.PlanInput{
UserID:       req.UserID,
ItineraryName: req.Name,
City:         req.City,
Tags:         req.Tags,
TotalBudget:  req.TotalBudget,
}
}
