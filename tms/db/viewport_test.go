package db

import (
	"math"
	"prisma/tms/feature"
	"testing"

	"prisma/tms/geojson"
)

const (
	Epsilon = 0.00001
)

func TestInHeatmapAIS(t *testing.T) {
	f := feature.New(nil, nil, map[string]interface{}{
		"mmsi": "123456789",
	})
	if !inHeatmap(f) {
		t.Errorf("Expected AIS to be in heatmap")
	}
}

func TestNoInHeatmapSART(t *testing.T) {
	f := feature.New(nil, nil, map[string]interface{}{
		"mmsi": "973456789",
	})
	if inHeatmap(f) {
		t.Errorf("Expected SART not to be in heatmap")
	}
}

func TestInHeatmapRadar(t *testing.T) {
	f := feature.New(nil, nil, map[string]interface{}{
		"type": "track:Radar",
	})
	if !inHeatmap(f) {
		t.Errorf("Expected radar to be in heatmap")
	}
}

func TestOthersNotInHeatmap(t *testing.T) {
	f := feature.New(nil, nil, nil)
	if inHeatmap(f) {
		t.Errorf("Expected others not to be in heatmap")
	}
}

func TestZoneNotInHeatmap(t *testing.T) {
	f := feature.New("zone:58ab2c6077f94b22f5bb3669", nil, map[string]interface{}{
		"type": "zone",
	})
	if inHeatmap(f) {
		t.Errorf("Expected zone not to be in heatmap")
	}
}

func TestPointToCell(t *testing.T) {
	point := geojson.Point{
		Coordinates: geojson.Position{X: 14.23, Y: 16.75},
	}
	bbox := geojson.New2DBBox(10, 10, 20, 20)
	bounds := NewBounds(bbox, 10)
	cell := PointToCell(point, bounds)
	if cell.Col != 4 || cell.Row != 6 {
		t.Errorf("Expecting (4 6) but got (%v %v)", cell.Col, cell.Row)
	}
}

func TestCellToPoint(t *testing.T) {
	cell := Cell{Col: 4, Row: 6}
	bbox := geojson.New2DBBox(10, 10, 20, 20)
	bounds := NewBounds(bbox, 10)
	x, y := CellToPoint(cell, bounds)
	if math.Abs(x-14.5) > Epsilon || math.Abs(y-16.5) > Epsilon {
		t.Errorf("Expecting (14.5 16.5) but got (%v %v)", x, y)
	}
}
