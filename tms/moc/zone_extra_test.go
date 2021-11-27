package moc

import (
	"prisma/tms"
	"prisma/tms/geojson"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZone_PolygonFromCircle(t *testing.T) {
	z := new(Zone)
	z.Area = &Area{
		Center: &tms.Point{
			Latitude:  1,
			Longitude: 2,
			Altitude:  3,
		},
		Radius: 5,
	}
	assert.Equal(t, &tms.Polygon{
		Lines: []*tms.LineString{
			{
				Points: []*tms.Point{
					{
						Latitude:  1,
						Longitude: 2,
						Altitude:  3,
					}, {
						Longitude: 5,
					},
				},
			},
		},
	}, z.PolygonFromCircle())
}

func TestZone_IsExcludedTrack(t *testing.T) {
	z := new(Zone)
	z.ExcludedVessels = []*Vessel{
		{
			Devices: []*Device{
				{
					Networks: []*Device_Network{
						{
							RegistryId: "test_registry_id",
						},
					},
				},
			},
		},
	}
	track := new(tms.Track)
	track.RegistryId = "test_registry_id"
	assert.True(t, z.IsExcludedTrack(track))
	track.RegistryId = "bad_test_registry_id"
	assert.False(t, z.IsExcludedTrack(track))
}

func TestZone_GeoJsonPolygonFromCircle(t *testing.T) {
	z := new(Zone)
	z.Area = &Area{
		Center: &tms.Point{
			Latitude:  1,
			Longitude: 2,
			Altitude:  3,
		},
		Radius: 5,
	}
	assert.Equal(t, &geojson.Polygon{
		Coordinates: [][]geojson.Position{
			{
				{
					X: 2,
					Y: 1,
					Z: 3,
				},
				{
					X: 5,
				},
			},
		},
	}, z.GeoJsonPolygonFromCircle())
}

func TestZone_BelongAreaToTrack(t *testing.T) {
	z := new(Zone)
	z.Area = &Area{
		RegistryId: "test1",
		TrackId:    "trackIdTest1",
	}
	trackTest1 := &tms.Track{RegistryId: "test1"}
	trackTest2 := &tms.Track{RegistryId: "test2"}
	trackTest3 := &tms.Track{Id: "trackIdTest1"}
	trackTest4 := &tms.Track{Id: "trackIdTest2"}
	trackTest5 := &tms.Track{Id: "trackIdTest2", RegistryId: "test1"}
	trackTest6 := &tms.Track{Id: "trackIdTest1", RegistryId: "test2"}
	assert.True(t, z.BelongAreaToTrack(trackTest1))
	assert.False(t, z.BelongAreaToTrack(trackTest2))
	assert.False(t, z.BelongAreaToTrack(trackTest3))
	assert.False(t, z.BelongAreaToTrack(trackTest4))
	assert.True(t, z.BelongAreaToTrack(trackTest5))
	assert.False(t, z.BelongAreaToTrack(trackTest6))

	z = new(Zone)
	z.Area = &Area{
		TrackId: "trackIdTest1",
	}
	assert.True(t, z.BelongAreaToTrack(trackTest3))
	assert.False(t, z.BelongAreaToTrack(trackTest4))
}

func TestZone_IntersectFlat(t *testing.T) {
	z := new(Zone)
	z.Poly = &tms.Polygon{
		Lines: []*tms.LineString{
			{
				Points: []*tms.Point{
					{
						Longitude: 1,
						Latitude:  1,
					}, {
						Longitude: 2,
						Latitude:  1,
					}, {
						Longitude: 1,
						Latitude:  2,
						Altitude:  2,
					},
					{
						Longitude: 1,
						Latitude:  1,
					},
				},
			},
		},
	}
	assert.True(t, z.Intersect(&geojson.Point{
		Coordinates: geojson.Position{X: 1.1, Y: 1.1},
	}))
	assert.False(t, z.Intersect(&geojson.Point{
		Coordinates: geojson.Position{X: 0.9, Y: 1},
	}))
	z.Area = &Area{
		Center: &tms.Point{
			Longitude: 1,
			Latitude:  1,
		},
		Radius: 1.3,
	}
	assert.True(t, z.Intersect(&geojson.Point{
		Coordinates: geojson.Position{X: 1.9, Y: 1.9},
	}))
	assert.False(t, z.Intersect(&geojson.Point{
		Coordinates: geojson.Position{X: 2, Y: 2.00001},
	}))
}

func TestZone_IntersectElevation(t *testing.T) {
	z := new(Zone)
	z.Poly = &tms.Polygon{
		Lines: []*tms.LineString{
			{
				Points: []*tms.Point{
					{
						Longitude: 1,
						Latitude:  1,
					}, {
						Longitude: 2,
						Latitude:  1,
					}, {
						Longitude: 1,
						Latitude:  2,
						Altitude:  2,
					},
					{
						Longitude: 1,
						Latitude:  1,
					},
				},
			},
		},
	}
	z.Elevation = &Elevation{
		Base: 100,
		Top:  1000,
	}
	assert.True(t, z.Intersect(&geojson.Point{
		Coordinates: geojson.Position{X: 1.1, Y: 1.1, Z: 300},
	}))
	assert.False(t, z.Intersect(&geojson.Point{
		Coordinates: geojson.Position{X: 1.1, Y: 1.1, Z: 3000},
	}))
	z.Area = &Area{
		Center: &tms.Point{
			Longitude: 1,
			Latitude:  1,
		},
		Radius: 1.3,
	}
	assert.True(t, z.Intersect(&geojson.Point{
		Coordinates: geojson.Position{X: 1.9, Y: 1.9, Z: 300},
	}))
	assert.False(t, z.Intersect(&geojson.Point{
		Coordinates: geojson.Position{X: 1.9, Y: 1.9, Z: 10},
	}))

}
