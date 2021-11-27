package tms

import (
	"testing"

	"github.com/golang/protobuf/ptypes/wrappers"
)

func TestLookupDevID(t *testing.T) {
	tt := []struct {
		name  string
		tgt   *Target
		devID string
	}{
		{"Imei dev ID", &Target{Imei: &wrappers.StringValue{Value: "123456789012345"}}, "123456789012345"},
		{"Mmsi dev ID", &Target{Mmsi: "970000001"}, "970000001"},
		{"NodeId dev ID", &Target{Nodeid: &wrappers.StringValue{Value: "hex12345t09"}}, "hex12345t09"},
		{"N/A dev ID", &Target{}, ""},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.tgt.LookupDevID() != tc.devID {
				t.Errorf("%s look up failed because %+v does have device id with value: %s", tc.name, tc.tgt, tc.devID)
			}
		})
	}
}

func TestPoint(t *testing.T) {
	track := &Track{
		Targets: []*Target{
			&Target{
				Position: &Point{
					Latitude:  10,
					Longitude: 20,
					Altitude:  100,
				},
			},
		},
	}
	pt := track.Point()
	if pt.Coordinates.X != 20 || pt.Coordinates.Y != 10 || pt.Coordinates.Z != 100 {
		t.Errorf("incorrect geojson point population from track: %+v != %+v", track.Targets[0].Position, pt)
	}
}

func TestRectPoint(t *testing.T) {
	track := &Track{
		Targets: []*Target{
			&Target{
				Position: &Point{
					Latitude:  10,
					Longitude: 20,
				},
			},
		},
	}
	x1, y1, z1, x2, y2, z2 := track.Rect()
	if x1 != 20 || y1 != 10 || z1 != 0 || x2 != 20 || y2 != 10 || z2 != 0 {
		t.Errorf("incorrect rect: %v %v %v %v %v %v", x1, y1, z1, x2, y2, z2)
	}
}

func TestRectMultiPoint(t *testing.T) {
	track := &Track{
		Targets: []*Target{
			&Target{
				Positions: []*Point{
					&Point{
						Latitude:  11,
						Longitude: 21,
					},
					&Point{
						Latitude:  19,
						Longitude: 29,
					},
				},
			},
		},
	}
	x1, y1, z1, x2, y2, z2 := track.Rect()
	if x1 != 21 || y1 != 11 || z1 != 0 || x2 != 29 || y2 != 19 || z2 != 0 {
		t.Errorf("incorrect rect: %v %v %v %v %v %v", x1, y1, z1, x2, y2, z2)
	}
}
