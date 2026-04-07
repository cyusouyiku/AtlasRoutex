package genetic

func MutateSwap(c Chromosome, i, j int) Chromosome {
out := Chromosome{Genes: append([]string(nil), c.Genes...)}
if i >= 0 && j >= 0 && i < len(out.Genes) && j < len(out.Genes) {
out.Genes[i], out.Genes[j] = out.Genes[j], out.Genes[i]
}
return out
}
