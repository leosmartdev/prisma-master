package object

import (
	"sync"
	"time"
	"fmt"
	"encoding/xml"

	"prisma/tms/cmd/daemons/tmccd/lib"
)

// max range between points
var maxRangeLocatedPosition = encodeDegree(20)

const (
	// This is a time mask which will received by tmccd
	layoutTimeSarsatMessages = "2006-01-02T15:04:05.000Z"
	// this header should be used instead of from library, cause tmccd waits this header only
	xmlHeader = `<?xml version="1.0" ?>`
	// some symbols before an xml message. It required by a protocol
	messageHeader = "/25501 00000/5030/17 299 1532\n/122/503A/012/01\n"
	// some symbols before an xml message. It required by a protocol
	messageFooter = "\n/LASSIT\n/ENDMSG"
	// a broken message format will be sent to the server to be sure the application can handle broken messages
	brokenMessageTpl = "I’m in distress, please help. my position is maybe lat: %f, lon: %f"
)

// Doppler is a position for dopplers
type Doppler Position

// Elemental is an object that should be found
type Elemental struct {
	DopplerA Doppler
	DopplerB Doppler
}

// Sarsat is a device which can generate different messages for defining own positions
// It uses as a sarsat beacon and a value of the one is issuing a rescuing alert
type Sarsat struct {
	NullDevice
	mu     sync.Mutex
	object Object

	realPos    Position
	locatedPos []Position

	muticker            sync.Mutex
	isBrokenMessageSent bool // is used to send a message with a broken format
}

// NewSarsat returns a new sarsat's device. Obj pointer is used for implementing common information
func NewSarsat(obj Object) *Sarsat {
	sarsat := &Sarsat{
		object: obj,
	}
	return sarsat
}

func (s *Sarsat) UpdateInformation(object Object) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.object = object
}

// This return a message. The message describe an alert without our positions
func (s *Sarsat) unLocatedMessage() ([]byte, error) {
	message := lib.TopMessage{
		EnvelopeHeader: lib.EnvelopeHeader{
			Dest:   "2770",
			Number: int32(time.Now().Nanosecond()),
			Date:   time.Now().Format(layoutTimeSarsatMessages),
			Orig:   "2570",
		},
		Message: lib.Message{
			UnlocatedAlertMessage: lib.UnlocatedAlertMessage{
				Header: lib.Header{
					SiteId: 27258,
					Beacon: s.object.BeaconId,
				},
				Tca:         time.Now().Format(layoutTimeSarsatMessages),
				Satellite:   "7",
				OrbitNumber: "20203",
			},
		},
	}
	return xml.MarshalIndent(message, "\n", " ")
}

func (s *Sarsat) generateErrorPoint() (latitude, longitude float64) {
	latitude = s.object.GetCurPos().Latitude + encodeNmToDegree(randomizer.Float64()*s.object.ErrorPointNM)
	longitude = s.object.GetCurPos().Longitude + encodeNmToDegree(randomizer.Float64()*s.object.ErrorPointNM)
	return
}

// This return a message. The message describe an alert with approximate positions
func (s *Sarsat) locatedMessage() ([]byte, error) {
	message := lib.TopMessage{
		EnvelopeHeader: lib.EnvelopeHeader{
			Dest:   "2770",
			Number: int32(time.Now().Nanosecond()),
			Date:   time.Now().Format(layoutTimeSarsatMessages),
		},
		Message: lib.Message{
			IncidentAlertMessage: lib.IncidentAlertMessage{
				Header: lib.Header{
					SiteId: 27258,
					Beacon: s.object.BeaconId,
				},
			},
		},
	}
	latitude, longitude := s.generateErrorPoint()
	message.Message.IncidentAlertMessage.Elemental = append(message.Message.IncidentAlertMessage.Elemental,
		lib.Elemental{
			Satellite:   "10",
			OrbitNumber: "48374",
			Tca:         time.Now().Format(layoutTimeSarsatMessages),
			DopplerA: lib.Doppler{
				Location: lib.Locate{
					Latitude:  s.object.GetCurPos().Latitude,
					Longitude: s.object.GetCurPos().Longitude,
				},
			},
			DopplerB: lib.Doppler{
				Location: lib.Locate{
					Latitude:  latitude,
					Longitude: longitude,
				},
			},
		},
	)
	return xml.MarshalIndent(message, "\n", " ")
}

// This return a message. The message describes an alert with a particular position
func (s *Sarsat) confirmedMessage() ([]byte, error) {
	message := lib.TopMessage{
		EnvelopeHeader: lib.EnvelopeHeader{
			Dest:   "2770",
			Number: int32(time.Now().Nanosecond()),
			Date:   time.Now().Format(layoutTimeSarsatMessages),
		},
		Message: lib.Message{
			ResolvedAlertMessage: lib.ResolvedAlertMessage{
				Header: lib.Header{
					SiteId: 27258,
					Beacon: s.object.BeaconId,
				},
				Composite: lib.Composite{
					Location: lib.Locate{
						Latitude:  s.object.GetCurPos().Latitude,
						Longitude: s.object.GetCurPos().Longitude,
					},
					Duration: "PT1155M",
				},
				Elemental: []lib.Elemental{
					lib.Elemental{
						Satellite:   "10",
						OrbitNumber: "48374",
						Tca:         time.Now().Format(layoutTimeSarsatMessages),
					},
				},
			},
		},
	}
	return xml.MarshalIndent(message, "\n", " ")
}

func (s *Sarsat) GetPositionAlertingMessage() ([]byte, error) {
	s.muticker.Lock()
	defer s.muticker.Unlock()
	if !s.isBrokenMessageSent {
		s.isBrokenMessageSent = true
		return []byte(fmt.Sprintf(brokenMessageTpl, s.object.curPos.Latitude, s.object.curPos.Longitude)), nil
	}
	var (
		xmlBody []byte
		err     error
	)
	switch {
	case s.object.Unlocated:
		xmlBody, err = s.unLocatedMessage()
	case s.object.Located:
		xmlBody, err = s.locatedMessage()
	default:
		xmlBody, err = s.confirmedMessage()
	}
	xmlBody = []byte(messageHeader + xmlHeader + string(xmlBody) + messageFooter)
	return xmlBody, err
}
