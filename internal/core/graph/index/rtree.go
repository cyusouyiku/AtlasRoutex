package index

type Rect struct {
MinX float64
MinY float64
MaxX float64
MaxY float64
}

func (r Rect) Contains(x, y float64) bool {
return x >= r.MinX && x <= r.MaxX && y >= r.MinY && y <= r.MaxY
}
