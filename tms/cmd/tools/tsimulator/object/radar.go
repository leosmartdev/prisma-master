package object

import (
	"prisma/tms/nmea"
	"sync"
)

// Radar implements features for radar's devices
type Radar struct {
	NullDevice
	mu     sync.Mutex
	object Object
}

// NewRadar returns a new radar's device. Obj pointer is used for implementing common information
func NewRadar(obj Object) *Radar {
	return &Radar{
		object: obj,
	}
}

func (r *Radar) UpdateInformation(object Object) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.object = object
}

func (r *Radar) GetTrackedTargetMessage(latitude, longitude float64) ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	distance, bearing := geo.To(latitude, longitude, r.object.GetCurPos().Latitude, r.object.GetCurPos().Longitude)
	statStruct := nmea.TTM{
		BaseSentence: nmea.BaseSentence{
			SOS:    "!",
			Talker: "RA",
			Format: "TTM",
		},
		CoreTTM: nmea.CoreTTM{
			Status:             r.object.Status,
			Bearing:            bearing,
			Name:               r.object.Name,
			Distance:           distance,
			Number:             r.object.Number,
			Speed:              r.object.GetCurPos().Speed,
			Course:             r.object.GetBearing(),
			CourseRelative:     "T",
			SpeedDistanceUnits: "N",

			StatusValidity:             true,
			BearingValidity:            true,
			DistanceValidity:           true,
			NumberValidity:             true,
			NameValidity:               true,
			SpeedValidity:              true,
			CourseValidity:             true,
			CourseRelativeValidity:     true,
			SpeedDistanceUnitsValidity: true,
		},
	}
	encode, _ := statStruct.Encode()
	return []byte(encode), nil
}
