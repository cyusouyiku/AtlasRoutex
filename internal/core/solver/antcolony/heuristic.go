package antcolony

func Heuristic(distance float64) float64 {
if distance <= 0 {
return 0
}
return 1.0 / distance
}
