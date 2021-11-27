package db

import (
	"prisma/tms/feature"
	"prisma/tms/moc"
	"reflect"
)

func init() {
	FeatureConverters[reflect.TypeOf(&moc.Zone{})] = ConvertZoneToFeature
}

func ConvertZoneToFeature(obj *GoObject) (*feature.F, error) {
	zone := obj.Data.(*moc.Zone)
	poly := zone.Poly.ToGeo()
	shape := "polygon"
	if zone.Area != nil && zone.Area.Center != nil {
		shape = "circle"
		poly = *zone.GeoJsonPolygonFromCircle()
	}
	feat := feature.New(
		obj.ID,
		poly,
		map[string]interface{}{
			"databaseId":         obj.ID,
			"type":               "zone",
			"name":               zone.Name,
			"shape":              shape,
			"area":               zone.Area,
			"description":        zone.Description,
			"fillColor":          zone.FillColor,
			"fillPattern":        zone.FillPattern,
			"strokeColor":        zone.StrokeColor,
			"createAlertOnEnter": zone.CreateAlertOnEnter,
			"createAlertOnExit":  zone.CreateAlertOnExit,
		},
	)
	return feat, nil
}
