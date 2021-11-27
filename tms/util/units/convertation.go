// Package units contains functions that convert between different units.
package units

import "math"

// FromMetersSecondToKnots converts meters/second to knots
func FromMetersSecondToKnots(ms float64) float64 {
	return ms * 1.94384
}

// FromKnotsToMetersSecond converts knots to meters/second
func FromKnotsToMetersSecond(ms float64) float64 {
	return ms * 0.514444
}

// DistanceGeoID computes meters between to GEO points
func DistanceGeoID(lat1, lon1, lat2, lon2 float64) float64 {
	// Convert degrees to radians
	rlat1 := lat1 * math.Pi / 180.0
	rlon1 := lon1 * math.Pi / 180.0

	rlat2 := lat2 * math.Pi / 180.0
	rlon2 := lon2 * math.Pi / 180.0

	dlat := rlat2 - rlat1
	dlon := rlon2 - rlon1

	R := float64(6371)

	a :=
		math.Sin(dlat/2) * math.Sin(dlat/2) +
			math.Cos(rlat1) * math.Cos(rlat2) *
				math.Sin(dlon/2) * math.Sin(dlon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	d := R * c
	return d * 1000
}
