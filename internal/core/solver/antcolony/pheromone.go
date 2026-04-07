package antcolony

type PheromoneMap map[string]float64

func Evaporate(m PheromoneMap, rate float64) {
for k, v := range m {
m[k] = v * (1 - rate)
}
}
