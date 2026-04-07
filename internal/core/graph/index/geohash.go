package index

import "fmt"

func EncodeGeohash(lat, lng float64, precision int) string {
return fmt.Sprintf("%.*f:%.*f", precision, lat, precision, lng)
}
