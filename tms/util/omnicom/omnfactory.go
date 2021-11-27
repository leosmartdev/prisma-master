// Package omnicom provides extra functions for omnicom devices.
package omnicom

import (
	"math/rand"
	"prisma/tms/moc"
	. "prisma/tms/omnicom"
	"time"
)

// NewPbUic is a factory function that returns *omnicom.Omni containing an omnicom.Uic structure
func NewPbUic(rpi uint32) *Omni {
	now := time.Now()
	rand.Seed(now.Unix())
	id := uint32(rand.Int63n(4095))
	date := CreateOmnicomDate(now.UTC())
	return &Omni{
		Omnicom: &Omni_Uic{Uic: &Uic{
			Header:        []byte{0x32},
			Date:          &Dt{Year: date.Year, Month: date.Month, Day: date.Day, Minute: date.Minute},
			New_Reporting: rpi,
			ID_Msg:        id,
		},
		},
	}
}

// NewPbRsm is a factory function that returns *omnicom.Omni containing an omnicom.Rsm structure
func NewPbRsm(ask uint32) *Omni {
	now := time.Now()
	rand.Seed(now.Unix())
	id := uint32(rand.Int63n(4095))
	date := CreateOmnicomDate(now.UTC())
	return &Omni{
		Omnicom: &Omni_Rsm{
			Rsm: &Rsm{
				Header:    []byte{0x33},
				ID_Msg:    id,
				Date:      &Dt{Year: date.Year, Month: date.Month, Day: date.Day, Minute: date.Minute},
				MsgTo_Ask: ask,
			},
		},
	}
}

// NewPbAa is a factory function ...
func NewPbAa() *Omni {
	date := CreateOmnicomDate(time.Now().UTC())
	return &Omni{
		Omnicom: &Omni_Aa{
			Aa: &Aa{
				Header: []byte{0x45},
				Date:   &Dt{Year: date.Year, Month: date.Month, Day: date.Day, Minute: date.Minute},
			},
		},
	}
}

// NewPbDgf is a factory function ...
func NewPbDgf(zoneID uint32) *Omni {
	// Upload delete-geofence get geofence ack message back from the beacon with error_type to evaluate the status of the upload
	//Error type :
	//0 : ok for :
	//  « Upload Geofence(0x035) : new record
	//  « Delete GEO fence(0x037) : delete succes
	//1 : buffer geofence full for :
	//  « Upload Geofence(0x035)
	//2 : no find GEO ID:
	//	« Delete GEO fence(0x37)
	//3 : ok for :
	//  « Upload Geofence(0x035) : update record
	// Error type with value 0 can be associated to two different events, and in order to avoid a useless database read
	// to figure which event it is Msg_ID for geo-fence uploads will be even random number,
	// and Msg_ID for geo-fence deletes will be an odd random number
	now := time.Now()
	rand.Seed(now.Unix())
	id := 2 + (2*uint32(rand.Int63())+1)%4094
	date := CreateOmnicomDate(now.UTC())
	return &Omni{
		Omnicom: &Omni_Dg{
			Dg: &Dg{
				Header: []byte{0x37},
				Date:   &Dt{Year: date.Year, Month: date.Month, Day: date.Day, Minute: date.Minute},
				Msg_ID: id,
				GEO_ID: zoneID,
			},
		},
	}
}

// NewPbUgf is a factory fucntion ...
func NewPbUgf(z *moc.Zone, pty uint32, act bool, stg Stg) *Omni {
	if z == nil {
		return nil
	}
	// Upload geo-fence get geofence ack message back from the beacon with error_type to evaluate the status of the upload
	//Error type :
	//0 : ok for :
	//  « Upload Geofence(0x035) : new record
	//  « Delete GEO fence(0x037) : delete succes
	//1 : buffer geofence full for :
	//  « Upload Geofence(0x035)
	//2 : no find GEO ID:
	//	« Delete GEO fence(0x037)
	//3 : ok for :
	//  « Upload Geofence(0x035) : update record
	// Error type with value 0 can be associated to two different events, and in order to avoid a useless database read
	// to figure which event it is Msg_ID for geo-fence uploads will be even random number,
	// and Msg_ID for geo-fence deletes will be an odd random number
	now := time.Now()
	rand.Seed(now.Unix())
	id := 2 + 2*uint32(rand.Int63())%4094
	date := CreateOmnicomDate(now.UTC())
	var activated uint32
	if act {
		activated = 1
	}
	if z.Poly != nil && z.Poly.Lines != nil {
		var npt uint32
		var pos []*Pos
		for _, line := range z.Poly.Lines {
			for _, p := range line.Points {
				pos = append(pos, &Pos{
					Latitude:  p.Latitude,
					Longitude: p.Longitude,
				})
				npt++
			}
		}
		return &Omni{Omnicom: &Omni_Ugpolygon{
			Ugpolygon: &UGPolygon{
				Header:       []byte{0x35},
				Msg_ID:       id,
				Date:         &Dt{Year: date.Year, Month: date.Month, Day: date.Day, Minute: date.Minute},
				GEO_ID:       z.ZoneId,
				TYPE:         0,
				Priority:     pty,
				Activated:    activated,
				Setting:      &stg,
				Number_Point: npt,
				Shape:        0,
				Position:     pos,
				NAME:         []byte(z.Name),
			},
		},
		}
	} else if z.Area != nil {
		return &Omni{Omnicom: &Omni_Ugcircle{
			Ugcircle: &UGCircle{
				Header:       []byte{0x35},
				Msg_ID:       id,
				Date:         &Dt{Year: date.Year, Month: date.Month, Day: date.Day, Minute: date.Minute},
				GEO_ID:       z.ZoneId,
				TYPE:         0,
				Priority:     pty,
				Activated:    activated,
				Setting:      &stg,
				Number_Point: 0,
				Shape:        1,
				Position: &PositionRadius{
					Longitude: float32(z.Area.Center.Longitude),
					Latitude:  float32(z.Area.Center.Latitude),
					Radius:    float32(z.Area.Radius),
				},
				NAME: []byte(z.Name),
			},
		},
		}
	}
	return nil
}
