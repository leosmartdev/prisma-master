package object

import (
	"prisma/tms/log"
	"prisma/tms/nmea"
	"sync"
	"time"
)

// AIS implements features for AIS devices
type AIS struct {
	NullDevice
	mu      sync.Mutex
	object  Object
	etaTime time.Time
}

// NewAIS returns a new AIS device. Obj pointer is used for implementing common information
func NewAIS(obj Object) *AIS {
	ais := new(AIS)
	var err error
	ais.etaTime, err = time.Parse("01021504", obj.ETA)
	if err != nil {
		log.Warn("Bad ETA time. Set now(). sea object: %s", obj.Name)
		ais.etaTime = time.Now()
	}
	ais.object = obj
	return ais
}

func (*AIS) GetSentence() nmea.BaseSentence {
	return nmea.BaseSentence{
		SOS:    "!",
		Talker: "AI",
		Format: "VDM",
	}
}

func (a *AIS) GetMessageStaticInformation() ([]byte, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	statStruct := nmea.M13715{
		VDMO: nmea.VDMO{
			BaseSentence: a.GetSentence(),
			CoreVDMO: nmea.CoreVDMO{
				SentenceCount:         1,
				SentenceCountValidity: true,
				SentenceIndex:         1,
				SentenceIndexValidity: true,
				SeqMsgID:              0,
				SeqMsgIDValidity:      true,
				Channel:               "A",
				ChannelValidity:       true,
			},
		},
		CoreM13715: nmea.CoreM13715{
			MessageID:        MessageTypeSRB,
			RepeatIndicator:  0,
			Mmsi:             a.object.Mmsi,
			Name:             a.object.Name,
			CallSign:         a.object.Name,
			Destination:      a.object.Destination,
			ShipAndCargoType: a.object.Type,
			EtaMonth:         uint32(a.etaTime.Month()),
			EtaDay:           uint32(a.etaTime.Day()),
			EtaHour:          uint32(a.etaTime.Hour()),
			EtaMinute:        uint32(a.etaTime.Minute()),
		},
	}
	encoded, _ := statStruct.Encode()
	return []byte(encoded), nil
}

func (a *AIS) UpdateInformation(object Object) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.object = object
}

func (a *AIS) GetMessagePosition() ([]byte, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	posStruct := nmea.M13711{
		VDMO: nmea.VDMO{
			BaseSentence: a.GetSentence(),
			CoreVDMO: nmea.CoreVDMO{
				SentenceCount:         1,
				SentenceCountValidity: true,
				SentenceIndex:         1,
				SentenceIndexValidity: true,
				SeqMsgID:              0,
				SeqMsgIDValidity:      true,
				Channel:               "A",
				ChannelValidity:       true,
			},
		},
		CoreM13711: nmea.CoreM13711{
			MessageID:          MessageTypePosition,
			RepeatIndicator:    0,
			Mmsi:               a.object.Mmsi,
			NavigationalStatus: a.object.NavigationStatus,
			Latitude:           encodeDegree(a.object.GetCurPos().Latitude),
			Longitude:          encodeDegree(a.object.GetCurPos().Longitude),
			SpeedOverGround:    uint32(a.object.GetCurPos().Speed * 10),
			CourseOverGround:   a.object.GetCourse(),
		}}
	encoded, _ := posStruct.Encode()
	return []byte(encoded), nil
}
