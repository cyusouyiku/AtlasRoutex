package optimization

type ObjectiveBreakdown struct {
DistanceScore float64
BudgetScore   float64
PreferenceScore float64
}

func TotalScore(b ObjectiveBreakdown, w Weights) float64 {
return b.DistanceScore*w.Distance + b.BudgetScore*w.Budget + b.PreferenceScore*w.Preference
}
