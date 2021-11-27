package moc

import (
	"errors"
	"math"
	"prisma/tms"
	"prisma/tms/geojson"
)

// GeoJsonBBoxFromCircle is for getting bbox from the circle of a zone
// https://social.msdn.microsoft.com/Forums/sqlserver/en-US/46ec0b60-ec57-46bf-9873-9a16620b3f63/convert-circle-to-polygon?forum=sqlspatial
func (z *Zone) GeoJsonBBoxFromCircle() (*geojson.BBox, error) {
	if z.Area == nil || z.Area.Center == nil {
		return nil, errors.New("Need to provide area")
	}
	var bearing float64
	d := degreeToNM(z.Area.Radius) / 6371.0
	lat, lon, alt := z.Area.Center.Latitude*math.Pi/180, z.Area.Center.Longitude*math.Pi/180, z.Area.Center.Altitude
	var points []geojson.Position
	var point geojson.Position
	for i := 0; i <= 360; i++ {
		bearing = 2 * float64(i) * math.Pi / 360 //rad
		point.Y = math.Asin(math.Sin(lat)*math.Cos(d) + math.Cos(lat)*math.Sin(d)*math.Cos(bearing))
		point.X = ((lon + math.Atan2(math.Sin(bearing)*math.Sin(d)*
			math.Cos(lat), math.Cos(d)-math.Sin(lat)*math.Sin(point.Y))) * 180) / math.Pi
		point.Y = (point.Y * 180) / math.Pi
		point.Z = alt
		points = append(points, point)
	}
	polygon := &geojson.Polygon{
		Coordinates: [][]geojson.Position{points},
	}
	b := polygon.CalculatedBBox()
	return &b, nil
}

func (z *Zone) GetBBox() geojson.BBox {
	if z.Area != nil && z.Area.Center != nil {
		bbox, _ := z.GeoJsonBBoxFromCircle()
		return *bbox
	}
	return z.Polygon().CalculatedBBox()
}

// GeoJsonPolygonFromCircle computes polygon from the circle of a zone(center and radius)
func (z *Zone) GeoJsonPolygonFromCircle() *geojson.Polygon {
	if z.Area == nil || z.Area.Center == nil {
		return nil
	}
	return &geojson.Polygon{
		Coordinates: [][]geojson.Position{
			{
				{
					X: z.Area.Center.Longitude,
					Y: z.Area.Center.Latitude,
					Z: z.Area.Center.Altitude,
				},
				{
					X: z.Area.Radius,
				},
			},
		},
	}
}

func degreeToNM(degree float64) float64 {
	return degree * 111.325
}

// IsExcludedTrack is used to determine this track should be excluded for this zone or not
func (z *Zone) IsExcludedTrack(track *tms.Track) bool {
	for _, vessel := range z.ExcludedVessels {
		for _, device := range vessel.Devices {
			for _, network := range device.Networks {
				if network.RegistryId == track.RegistryId {
					return true
				}
			}
		}
	}
	return false
}

func (z *Zone) PolygonFromCircle() *tms.Polygon {
	if z.Area == nil || z.Area.Center == nil {
		return nil
	}
	return &tms.Polygon{
		Lines: []*tms.LineString{
			{
				Points: []*tms.Point{
					{
						Latitude:  z.Area.Center.Latitude,
						Longitude: z.Area.Center.Longitude,
						Altitude:  z.Area.Center.Altitude,
					},
					{
						Longitude: z.Area.Radius,
					},
				},
			},
		},
	}
}

// SetPolygonSelf is used to avoid assigning to the poly directly from client sides
// You should use it. It computes polygon and assign to the one.
func (z *Zone) SetPolygonSelf() {
	if z.Area == nil || z.Area.Center == nil {
		return
	}
	z.Poly = z.PolygonFromCircle()
}

func (z *Zone) Polygon() *geojson.Polygon {
	var coordinates [][]geojson.Position
	poly := z.Poly
	for _, lineString := range poly.Lines {
		var points []geojson.Position
		for _, point := range lineString.Points {
			position := geojson.Position{
				X: point.Longitude,
				Y: point.Latitude,
				Z: point.Altitude,
			}

			points = append(points, position)
		}
		coordinates = append(coordinates, points)
	}
	polygon := &geojson.Polygon{
		Coordinates: coordinates,
	}
	b := polygon.CalculatedBBox()
	polygon.BBox = &b
	return polygon
}

func (z *Zone) BelongAreaToTrack(track *tms.Track) bool {
	id := track.Id
	if track.RegistryId != "" {
		id = track.RegistryId
	}
	return z.Area != nil && z.GetAreaID() == id
}

func (z *Zone) GetAreaID() (id string) {
	switch {
	case z.Area == nil:
	case z.Area.RegistryId != "":
		id = z.Area.RegistryId
	case z.Area.TrackId != "":
		id = z.Area.TrackId
	}
	return
}

// Intersect is a function that check is a geojson point is neaby an area, this function
// also evaluates if the area has an elevation, and verifies if the point is inside a given box
func (z *Zone) Intersect(p *geojson.Point) bool {
	var isElevation bool
	if z.Elevation != nil {
		isElevation = (z.Elevation.Base <= p.Coordinates.Z && p.Coordinates.Z <= z.Elevation.Top)
	} else {
		isElevation = true
	}
	if z.Area != nil && z.Area.Center != nil {
		return (p.Nearby(z.Area.Center.ToGeo(), degreeToNM(z.Area.Radius)*1000) && isElevation)
	}
	return (z.Polygon().Nearby(p.Coordinates, 0) && isElevation)
}
