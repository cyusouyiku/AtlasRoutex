package utils

func Clamp(v, min, max float64) float64 {
if v < min {
return min
}
if v > max {
return max
}
return v
}
