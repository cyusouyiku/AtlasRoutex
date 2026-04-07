package planner

func ReplanIncremental(oldDays [][]string, changedIDs []string) [][]string {
out := make([][]string, len(oldDays))
for i := range oldDays {
out[i] = append([]string(nil), oldDays[i]...)
}
if len(out) == 0 {
out = [][]string{{}}
}
out[0] = append(out[0], changedIDs...)
return out
}
