package mapapi

import "context"

func DistanceMatrix(ctx context.Context, c *Client, coords [][2]float64) ([][]float64, error) {
m := make([][]float64, len(coords))
for i := range coords {
m[i] = make([]float64, len(coords))
for j := range coords {
if i == j { continue }
d, err := c.Distance(ctx, coords[i][0], coords[i][1], coords[j][0], coords[j][1])
if err != nil { return nil, err }
m[i][j] = d
}
}
return m, nil
}
