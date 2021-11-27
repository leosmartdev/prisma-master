package object

import (
	"time"

	"prisma/tms/spidertracks"
	"prisma/tms/util/units"
	"sync"
)

const (
	SysId   = "spidertracks"
	Version = "2.32"
	DataCtr = "simulator"
	Fix     = "3D"
	Source  = "GPS"
)

type Spidertrack struct {
	NullDevice

	mu     sync.Mutex
	object Object
}

func NewSpiderTrack(object Object) *Spidertrack {
	return &Spidertrack{
		object: object,
	}
}

func (sp *Spidertrack) GetACPos() spidertracks.AcPos {
	return spidertracks.AcPos{
		DataCtr:         DataCtr,
		Speed:           int(units.FromKnotsToMetersSecond(sp.object.GetCurPos().Speed)),
		Heading:         int(sp.object.GetCourse()),
		DateTime:        time.Now().Format(spidertracks.TimeLayout),
		Fix:             Fix,
		Altitude:        int(sp.object.GetCurPos().Altitude),
		Lat:             sp.object.GetCurPos().Latitude,
		Long:            sp.object.GetCurPos().Longitude,
		DataCtrDateTime: time.Now().Format(spidertracks.TimeLayout),
		Source:          Source,
		UnitID:          sp.object.Imei,
		Esn:             sp.object.Imei,
	}
}

func (sp *Spidertrack) UpdateInformation(object Object) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.object = object
}
