package optimization

type Weights struct {
Distance   float64
Budget     float64
Preference float64
}

func DefaultWeights() Weights {
return Weights{Distance: 0.4, Budget: 0.3, Preference: 0.3}
}
