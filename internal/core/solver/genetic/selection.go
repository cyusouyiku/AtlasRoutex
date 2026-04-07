package genetic

func SelectBest(pop []Chromosome, k int) []Chromosome {
if k > len(pop) {
k = len(pop)
}
out := append([]Chromosome(nil), pop...)
for i := 0; i < len(out); i++ {
for j := i + 1; j < len(out); j++ {
if out[j].Score > out[i].Score {
out[i], out[j] = out[j], out[i]
}
}
}
return out[:k]
}
