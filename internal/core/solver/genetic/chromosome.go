package genetic

type Chromosome struct {
Genes []string
Score float64
}

func NewChromosome(genes []string) Chromosome {
return Chromosome{Genes: append([]string(nil), genes...)}
}
