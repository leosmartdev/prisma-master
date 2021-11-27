package object

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"math"
)

func TestPositionArrivalTime_ComputeZeroSpeed(t *testing.T) {
	p := PositionArrivalTime{
		PositionSpeed: PositionSpeed{
			Longitude: 0,
			Latitude: 0,
			Speed: 0,
		},
	}
	nextp := PositionArrivalTime{
		PositionSpeed: PositionSpeed{
			Longitude: 1,
			Latitude: 1,
			Speed: 0,
		},
		ArrivalTimeSeconds: 10,
	}
	p.ComputeZeroSpeed(nextp)
	assert.NotZero(t, p.Speed)
	prevSpeed := p.Speed

	nextp.ArrivalTimeSeconds = 0
	p.ComputeZeroSpeed(nextp)
	assert.Equal(t,prevSpeed, p.Speed)

	p.Speed = 0
	nextp.ArrivalTimeSeconds = -1
	p.ComputeZeroSpeed(nextp)
	assert.True(t, math.IsInf(p.Speed, 1))
}

func TestPosition_Intersection(t *testing.T) {
	p := Position{
		Longitude: 1,
		Latitude: 1,
	}
	zone := []Position{
		{0,0},
		{0,2},
		{2,2},
		{2,0},
		{0,0},
	}
	assert.True(t, p.Intersection(zone))
	p.Latitude = 3
	assert.False(t, p.Intersection(zone))
}
