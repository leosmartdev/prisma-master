// Package feature provides feature structure.
package feature

import (
	"prisma/tms/util/coordsys"
	"prisma/tms/geojson"
)

type F struct {
	geojson.Feature
}

func New(id interface{}, geom geojson.Object, props map[string]interface{}) *F {
	return &F{
		Feature: geojson.Feature{
			ID:         id,
			Geometry:   geom,
			Properties: props,
		},
	}
}

func NewRef(id interface{}) *F {
	return &F{
		Feature: geojson.Feature{
			ID: id,
		},
	}
}

func (f *F) FromWGS84(crs coordsys.C) *F {
	newGeom := coordsys.FromWGS84(crs, f.Geometry)
	newProps := make(map[string]interface{})
	for k, v := range f.Properties {
		newProps[k] = v
	}
	newProps["crs"] = crs.EPSG
	return New(f.ID, newGeom, f.Properties)
}
