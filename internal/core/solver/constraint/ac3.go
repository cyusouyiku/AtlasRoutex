package constraint

func ArcConsistent(domains map[string][]string, arcs [][2]string) bool {
for _, a := range arcs {
if len(domains[a[0]]) == 0 || len(domains[a[1]]) == 0 {
return false
}
}
return true
}
