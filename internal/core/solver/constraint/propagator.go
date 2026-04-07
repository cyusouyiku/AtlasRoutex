package constraint

func Propagate(domains map[string][]string, fixed map[string]string) map[string][]string {
out := make(map[string][]string, len(domains))
for k, v := range domains {
out[k] = append([]string(nil), v...)
}
for k, chosen := range fixed {
out[k] = []string{chosen}
}
return out
}
