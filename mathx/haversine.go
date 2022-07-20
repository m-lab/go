package mathx

import (
	"math"
)

const earthRadius = 6371

// GetHaversineDistance finds the distance (in km) between two latitude/longitude
// pairs using the Haversine formula.
// For more details, see http://en.wikipedia.org/wiki/Haversine_formula.
func GetHaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	dlat1 := degreesToRadian(lat1)
	dlon1 := degreesToRadian(lon1)
	dlat2 := degreesToRadian(lat2)
	dlon2 := degreesToRadian(lon2)

	diffLat := dlat2 - dlat1
	diffLon := dlon2 - dlon1
	sinDiffLat := math.Sin(diffLat / 2)
	sinDiffLon := math.Sin(diffLon / 2)

	a := sinDiffLat*sinDiffLat + math.Cos(dlat1)*
		math.Cos(dlat2)*sinDiffLon*sinDiffLon
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	d := earthRadius * c

	return d
}

func degreesToRadian(d float64) float64 {
	return d * math.Pi / 180
}
