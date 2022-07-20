package mathx

import (
	"math"
)

const earthRadius = 6371

// GetHaversineDistance finds the distance (in km) between two latitude/longitude
// pairs using the Haversine formula.
func GetHaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	dlat := degreesToRadian(lat2 - lat1)
	dlon := degreesToRadian(lon2 - lon1)

	a := math.Sin(dlat/2)*math.Sin(dlat/2) + math.Cos(degreesToRadian(lat1))*
		math.Cos(degreesToRadian(lat2))*math.Sin(dlon/2)*math.Sin(dlon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	d := earthRadius * c

	return d
}

func degreesToRadian(d float64) float64 {
	return d * math.Pi / 180
}
