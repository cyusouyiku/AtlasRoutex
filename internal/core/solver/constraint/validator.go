package constraint

func ValidateDayAssignments(days [][]string) bool {
for _, d := range days {
seen := map[string]struct{}{}
for _, id := range d {
if _, ok := seen[id]; ok {
return false
}
seen[id] = struct{}{}
}
}
return true
}
