// Package angles provides functions and constants to work with angles.
package geo

import (
	"errors"
	"fmt"
	"math"
)

const (
	DegreeToRadian       = math.Pi / 180
	RadianToDegree       = 1.0 / DegreeToRadian
	Mean_Earth_Radius_NM = 3440.0
	// AIS Longitude-Invalid value
	Longitude_Na  = 181
	Longitude_Min = -180
	Longitude_Max = 180
	// AIS Latitude-Invalid value
	Latitude_Na  = 91
	Latitude_Min = -90
	Latitude_Max = 90
	// AIS Course-Invalid value
	Course_Na  = 360.0
	Course_Min = 0.0
	Course_Max = 359.9
	// AIS Speed-Invalid value
	Speed_Na  = 102.3
	Speed_Min = 0.0
	Speed_Max = 102.2
	// AIS Rateofturn-Invalid value
	Rot_Na  = -128
	Rot_Min = -127
	Rot_Max = 127
	// AIS Heading-Invalid value
	Heading_Na  = 511
	Heading_Min = 0
	Heading_Max = 359
)

type Vector struct {
	X, Y float64
}

// Converts course (in degrees) and speed (in anything) to a vector
func FromCourseSpeed(course float64, speed float64) Vector {
	crad := math.Mod(course, 360) * DegreeToRadian
	return Vector{
		X: math.Sin(crad) * speed,
		Y: math.Cos(crad) * speed,
	}
}

// Returns the course (in degrees)
func (v Vector) ToCourseSpeed() (course float64, speed float64) {
	course = math.Mod(RadianToDegree*math.Atan2(v.X, v.Y), 360)
	speed = math.Sqrt(v.Y*v.Y + v.X*v.X)
	return
}

func (v Vector) Add(o Vector) Vector {
	return Vector{
		X: v.X + o.X,
		Y: v.Y + o.Y,
	}
}

func (v Vector) Mul(s float64) Vector {
	return Vector{
		X: s * v.X,
		Y: s * v.Y,
	}
}

func (v Vector) String() string {
	return fmt.Sprintf("%e, %e", v.X, v.Y)
}

func AvgVec(vecs ...Vector) Vector {
	tot := Vector{}
	for _, v := range vecs {
		tot = tot.Add(v)
	}

	return Vector{
		X: tot.X / float64(len(vecs)),
		Y: tot.Y / float64(len(vecs)),
	}
}

func FindPositionUsingHaversineAlg(src_lat_deg float64, src_lon_deg float64, distance_nm float64, true_bearing_deg float64) (float64, float64, error) {
	var err error

	if (src_lat_deg < -90) || (src_lat_deg > 90) || (src_lon_deg < -180) || (src_lon_deg > 180) || (true_bearing_deg < -180) || (true_bearing_deg > 360) || (distance_nm < 0) {
		err = errors.New("Input value out of bounds")
		return 0.0, 0.0, err
	}

	distance_rad := distance_nm / Mean_Earth_Radius_NM

	for true_bearing_deg < 0 {
		true_bearing_deg += 360.0
	}

	bearing_rad := true_bearing_deg * math.Pi / 180.0
	src_lat_rad := src_lat_deg * math.Pi / 180.0
	src_lon_rad := src_lon_deg * math.Pi / 180.0
	lat_rad := math.Asin(math.Sin(src_lat_rad)*math.Cos(distance_rad) + math.Cos(src_lat_rad)*math.Sin(distance_rad)*math.Cos(bearing_rad))
	lon_rad := src_lon_rad + math.Atan2(math.Sin(bearing_rad)*math.Sin(distance_rad)*math.Cos(src_lat_rad), math.Cos(distance_rad)-math.Sin(src_lat_rad)*math.Sin(lat_rad))
	calculated_lat := lat_rad * 180.0 / math.Pi
	calculated_lon := lon_rad * 180.0 / math.Pi

	if math.Abs(calculated_lat) < math.SmallestNonzeroFloat64 {
		calculated_lat = 0.0
	}

	if math.Abs(calculated_lon) < math.SmallestNonzeroFloat64 {
		calculated_lon = 0.0
	}

	if (calculated_lat < -90.0) || (calculated_lat > 90.0) || (calculated_lon < -180.0) || (calculated_lon > 180.0) {
		err = errors.New("Output value out of bounds")
		return 0.0, 0.0, err
	}

	return calculated_lat, calculated_lon, nil
}

func Arcmin2Decimal(value float64, direction string) (float64, error) {
	var err error
	val_deg := int(value / 100.0)
	val_min := float64(value - (float64(val_deg) * 100.0))
	new_val := float64(val_deg) + (val_min / 60.0)
	if (direction[0:1] == "S") || (direction[0:1] == "W") || (direction[0:1] == "s") || (direction[0:1] == "w") {
		new_val = -new_val
	} else if (direction[0:1] == "N") || (direction[0:1] == "E") || (direction[0:1] == "n") || (direction[0:1] == "e") {
	} else {
		err = errors.New("Unknown direction")
	}

	if (new_val > 180.0) || (new_val < -180.0) {
		err = errors.New("Lat/lon value of range")
	}

	return new_val, err
}
