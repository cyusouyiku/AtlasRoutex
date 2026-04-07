package hybrid

func EnsembleBest(candidates [][]string) []string {
if len(candidates) == 0 {
return nil
}
best := candidates[0]
for _, c := range candidates {
if len(c) > len(best) {
best = c
}
}
return append([]string(nil), best...)
}
