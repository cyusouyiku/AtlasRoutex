package dto

type PlanRequest struct {
UserID      string   `json:"user_id"`
Name        string   `json:"name"`
City        string   `json:"city"`
StartDate   string   `json:"start_date"`
EndDate     string   `json:"end_date"`
Tags        []string `json:"tags"`
TotalBudget float64  `json:"total_budget"`
}
