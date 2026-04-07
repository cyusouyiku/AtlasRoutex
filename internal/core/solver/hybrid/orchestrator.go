package hybrid

func Orchestrate(paths ...[]string) []string {
if len(paths) == 0 {
return nil
}
out := make([]string, 0)
for _, p := range paths {
out = append(out, p...)
}
return out
}
