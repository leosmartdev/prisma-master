package tms

import (
	"strings"

	geo "prisma/tms/geojson"
)

func (p *Point) ToGeo() geo.Position {
	return geo.Position{
		X: p.Longitude,
		Y: p.Latitude,
		Z: p.Altitude,
	}
}

func (p *LineString) ToGeo() geo.LineString {
	var coords []geo.Position
	for _, point := range p.Points {
		coords = append(coords, point.ToGeo())
	}
	return geo.LineString{
		Coordinates: coords,
	}
}

func (p *Polygon) ToGeo() geo.Polygon {
	var coords [][]geo.Position
	for _, line := range p.Lines {
		coords = append(coords, line.ToGeo().Coordinates)
	}
	return geo.Polygon{
		Coordinates: coords,
	}
}

func (msg *TsiMessage) Type() string {
	if msg.Body == nil {
		return ""
	}
	return strings.TrimPrefix(msg.Body.TypeUrl, "type.googleapis.com/")
}
