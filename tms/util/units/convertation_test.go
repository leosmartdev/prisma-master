package units

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestFromKnotsToMetersSecond(t *testing.T) {
	assert.InDelta(t, 51.4444, FromKnotsToMetersSecond(100), 0.001)
	assert.Zero(t, FromKnotsToMetersSecond(0))
	assert.InDelta(t, 0.514444, FromKnotsToMetersSecond(1), 0.001)
}

func TestFromMetersSecondToKnots(t *testing.T) {
	assert.InDelta(t, 194.384, FromMetersSecondToKnots(100), 0.001)
	assert.Zero(t, FromMetersSecondToKnots(0))
	assert.InDelta(t, 1.94384, FromMetersSecondToKnots(1), 0.001)
}

// http://www.onlineconversion.com/map_greatcircle_distance.htm
func TestDistanceGeoID(t *testing.T) {
	assert.InDelta(t, 157.24938127194397*1000, DistanceGeoID(0, 0, 1, 1), 0.1)
}
