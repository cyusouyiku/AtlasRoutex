package planner

func SplitByRegion(candidates []string, chunk int) [][]string {
if chunk <= 0 {
chunk = 10
}
var out [][]string
for i := 0; i < len(candidates); i += chunk {
end := i + chunk
if end > len(candidates) {
end = len(candidates)
}
out = append(out, append([]string(nil), candidates[i:end]...))
}
return out
}
