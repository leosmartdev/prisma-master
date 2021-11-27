package coordsys

import (
	"math"
	"testing"

	"prisma/tms/geojson"
)

const epsilon = 0.0000001

const WGS84X = 10.0
const WGS84Y = 20.0
const WebMercX = 1.1131949077777779e+06
const WebMercY = 2.2730309266712805e+06

func TestProjForward(t *testing.T) {
	from := geojson.Point{
		Coordinates: geojson.Position{
			X: WGS84X,
			Y: WGS84Y,
		},
	}
	to := Proj(WGS84, WebMercator, from).(geojson.Point)
	wantX := WebMercX
	wantY := WebMercY

	if math.Abs(to.Coordinates.X-wantX) > epsilon {
		t.Fatalf("want %v ; got %v", wantX, to.Coordinates.X)
	}
	if math.Abs(to.Coordinates.Y-wantY) > epsilon {
		t.Fatalf("want %v ; got %v", wantY, to.Coordinates.Y)
	}
}

func TestProjInverse(t *testing.T) {
	from := geojson.Point{
		Coordinates: geojson.Position{
			X: WebMercX,
			Y: WebMercY,
		},
	}
	to := Proj(WebMercator, WGS84, from).(geojson.Point)
	wantX := WGS84X
	wantY := WGS84Y

	if math.Abs(to.Coordinates.X-wantX) > epsilon {
		t.Fatalf("want %v ; got %v", wantX, to.Coordinates.X)
	}
	if math.Abs(to.Coordinates.Y-wantY) > epsilon {
		t.Fatalf("want %v ; got %v", wantY, to.Coordinates.Y)
	}
}
