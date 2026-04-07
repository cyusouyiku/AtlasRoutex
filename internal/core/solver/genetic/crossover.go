package genetic

func OnePointCrossover(a, b Chromosome, point int) (Chromosome, Chromosome) {
if point < 0 {
point = 0
}
if point > len(a.Genes) {
point = len(a.Genes)
}
c1 := append(append([]string{}, a.Genes[:point]...), b.Genes[point:]...)
c2 := append(append([]string{}, b.Genes[:point]...), a.Genes[point:]...)
return Chromosome{Genes: c1}, Chromosome{Genes: c2}
}
