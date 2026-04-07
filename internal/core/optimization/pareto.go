package optimization

type Point struct {
Values []float64
}

func IsDominated(a, b Point) bool {
if len(a.Values) != len(b.Values) || len(a.Values) == 0 {
return false
}
strictlyBetter := false
for i := range a.Values {
if b.Values[i] > a.Values[i] {
return false
}
if b.Values[i] < a.Values[i] {
strictlyBetter = true
}
}
return strictlyBetter
}
