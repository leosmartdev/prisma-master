// Package coordsys provides extra functions for projections.
package coordsys

import (
	"math"

	"prisma/tms/geojson"
)

const mercatorPole = 20037508.34

type C struct {
	Forward     Converter
	Inverse     Converter
	Bounds      geojson.BBox
	WGS84Bounds geojson.BBox
	EPSG        string
}

func (c C) String() string {
	return c.EPSG
}

// Web Mercator
var EPSG3857 = C{
	EPSG: "EPSG:3857",
	Forward: func(p geojson.Position) geojson.Position {
		x := mercatorPole / 180.0 * p.X
		y := math.Log(math.Tan((90.0+p.Y)*math.Pi/360.0)) / math.Pi * mercatorPole
		y = math.Max(-mercatorPole, math.Min(y, mercatorPole))

		return geojson.Position{
			X: x,
			Y: y,
		}
	},
	Inverse: func(p geojson.Position) geojson.Position {
		x := p.X * 180.0 / mercatorPole
		y := 180.0 / math.Pi * (2*math.Atan(math.Exp((p.Y/mercatorPole)*math.Pi)) - math.Pi/2.0)

		return geojson.Position{
			X: x,
			Y: y,
		}
	},
	Bounds:      geojson.New2DBBox(-20026376.39, -20048966.10, 20026376.39, 20048966.10),
	WGS84Bounds: geojson.New2DBBox(-180.0, -85.06, 180.0, 85.06),
}

var WebMercator = EPSG3857

var EPSG4326 = C{
	EPSG:        "EPSG:4326",
	Forward:     func(p geojson.Position) geojson.Position { return p },
	Inverse:     func(p geojson.Position) geojson.Position { return p },
	Bounds:      geojson.New2DBBox(-180, -90, 180, 90),
	WGS84Bounds: geojson.New2DBBox(-180, -90, 180, 90),
}

var WGS84 = EPSG4326
