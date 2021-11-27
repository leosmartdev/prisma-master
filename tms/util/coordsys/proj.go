package coordsys

import (
	"fmt"

	"prisma/tms/geojson"
)

type Converter func(p geojson.Position) geojson.Position

func Proj(source C, target C, geom geojson.Object) geojson.Object {
	wgs84 := convert(source.Inverse, geom)
	return convert(target.Forward, wgs84)
}

func ToWGS84(crs C, geom geojson.Object) geojson.Object {
	return Proj(crs, WGS84, geom)
}

func FromWGS84(crs C, geom geojson.Object) geojson.Object {
	return Proj(WGS84, crs, geom)
}

func convert(fn Converter, geom geojson.Object) geojson.Object {
	if geom != nil {
		// Project
		switch g := geom.(type) {
		case *geojson.Point:
			return geojson.Point{
				Coordinates: convertP(fn, g.Coordinates),
			}
		case geojson.Point:
			return geojson.Point{
				Coordinates: convertP(fn, g.Coordinates),
			}
		case geojson.MultiPoint:
			return geojson.MultiPoint{
				Coordinates: convertA1(fn, g.Coordinates),
			}
		case *geojson.MultiPoint:
			return geojson.MultiPoint{
				Coordinates: convertA1(fn, g.Coordinates),
			}
		case geojson.Polygon:
			return geojson.Polygon{
				Coordinates: convertA2(fn, g.Coordinates),
			}
		case *geojson.Polygon:
			return geojson.Polygon{
				Coordinates: convertA2(fn, g.Coordinates),
			}
		case *geojson.MultiPolygon:
			return geojson.MultiPolygon{
				Coordinates: convertA3(fn, g.Coordinates),
			}
		case geojson.MultiPolygon:
			return geojson.MultiPolygon{
				Coordinates: convertA3(fn, g.Coordinates),
			}
		default:
			panic(fmt.Sprintf("unknown type: %+v", geom))
		}
	}
	return nil
}

func convertP(fn Converter, p geojson.Position) geojson.Position {
	return fn(p)
}

func convertA1(fn Converter, coords []geojson.Position) []geojson.Position {
	ret := make([]geojson.Position, len(coords))
	for i := range coords {
		ret[i] = convertP(fn, coords[i])
	}
	return ret
}

func convertA2(fn Converter, coords [][]geojson.Position) [][]geojson.Position {
	ret := make([][]geojson.Position, len(coords))
	for i := range coords {
		ret[i] = convertA1(fn, coords[i])
	}
	return ret
}

func convertA3(fn Converter, coords [][][]geojson.Position) [][][]geojson.Position {
	ret := make([][][]geojson.Position, len(coords))
	for i := range coords {
		ret[i] = convertA2(fn, coords[i])
	}
	return ret
}
