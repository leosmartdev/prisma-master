package object

import (
	"time"

	"prisma/tms/geojson"
	"prisma/tms/util/units"
	"prisma/tms/log"
	"math"
)

// timeStart is used to determine arrival position
var timeStart = time.Now()

const (
	Low    = "low"
	Medium = "medium"
	High   = "high"
)

// PositionArrivalTime determines a direction and speed for the direction
type PositionSpeed struct {
	Latitude  float64
	Longitude float64
	Altitude  float64
	Speed     float64
}

// PositionSpeedTime determines when a seaobject had a direction and speed
type PositionSpeedTime struct {
	p PositionSpeed
	t time.Time
}

// PositionArrivalTime provides computing speed using the arrival time of a next point
type PositionArrivalTime struct {
	PositionSpeed
	ArrivalTimeSeconds int `json:"arrival_time_seconds"`
	Damage             string
	Collided           bool
}

// ComputeZeroSpeed is used to determine speed using ArrivalTime.
// Also if speed is not 0 then this speed is going to be used
// If ArrivalTimeSeconds is 0 then not changes will be with speed and no issued messages to log
// http://www.ridgesolutions.ie/index.php/2013/11/14/algorithm-to-calculate-speed-from-two-gps-latitude-and-longitude-points-and-time-difference/
func (p *PositionArrivalTime) ComputeZeroSpeed(nextPoints PositionArrivalTime) {
	if nextPoints.Speed != 0 {
		return
	}
	if nextPoints.ArrivalTimeSeconds == 0 {
		return
	}
	dist := units.DistanceGeoID(p.Latitude, p.Longitude, nextPoints.Latitude, nextPoints.Longitude)
	times := timeStart.Add(time.Duration(nextPoints.ArrivalTimeSeconds) * time.Second)
	if times.Before(timeStart) {
		log.Warn("time of arriving is less then 0")
		p.Speed = math.Inf(1)
		return
	}
	p.Speed = units.FromMetersSecondToKnots(dist / times.Sub(timeStart).Seconds())
}

// NewPositionSpeedTime concat position and speed with current time
func NewPositionSpeedTime(p PositionSpeed) *PositionSpeedTime {
	return &PositionSpeedTime{
		p: p,
		t: time.Now(),
	}
}

// Position is used to determine a position for an object
type Position struct {
	Latitude  float64
	Longitude float64
}

// Intersection determines an intersection between two points
func (p *PositionSpeed) Intersection(points []Position) bool {
	return (&Position{
		Latitude:  p.Latitude,
		Longitude: p.Longitude,
	}).Intersection(points)
}

// Intersection determines an intersection between two points
func (point *Position) Intersection(points []Position) bool {
	var gpoints []geojson.Position
	for i := range points {
		gpoints = append(gpoints, geojson.Position{
			X: points[i].Longitude,
			Y: points[i].Latitude,
		})
	}
	polygon := &geojson.Polygon{
		Coordinates: [][]geojson.Position{gpoints},
	}
	return polygon.Nearby(geojson.Position{
		X: point.Longitude,
		Y: point.Latitude,
	}, 0)
}
