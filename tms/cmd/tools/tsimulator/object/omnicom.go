package object

import (
	"container/list"
	"errors"
	"sync"
	"time"

	"prisma/tms/cmd/tools/tsimulator/task"
	"prisma/tms/iridium"
	"prisma/tms/log"
	"prisma/tms/omnicom"
)

// These constants are used to determine the type for alerts
const (
	PU = iota
	PD
	BA
	IA
	NPF
	JBDA
	LMC
	DA
	AA
	TM
	LastTypeAlerting
)

const taskQueueSize = 50
const sleepForHandlingTask = 1 * time.Second
const unitForReportTime = time.Minute
const unitForGeoFenceReportTime = time.Minute
const sizePositionPerSocket = 10 // how mush positions will be sent on one socket

// ErrTooEarly is used to point out that a message cannot be sent
// because the timer for new messages was not expired
var ErrTooEarly = errors.New("too early")

type timerPositionGeofence struct {
	t *time.Ticker
	p []Position
}

// Omnicom implements features for omnicom's devices
type Omnicom struct {
	NullDevice
	mu                sync.Mutex
	object            Object
	hraw              []byte
	startTypeAlerting uint
	stopTypeAlerting  uint
	startAlerting     bool
	stopAlerting      bool
	queueTask         list.List
	handledCh         chan task.Result

	sentTestMode bool

	lastDatePosition omnicom.Date_Position
	mgfence          sync.Mutex
	geoFences        map[uint32]timerPositionGeofence
	defaultReport    *time.Ticker
	defaultTimeTimer time.Duration
	ar               *omnicom.AR
}

//NewOmnicom Return a new omnicom's device. Obj pointer is used for implementing common information
func NewOmnicom(obj Object, reportTime time.Duration) *Omnicom {
	MOH := iridium.MOHeader{
		MO_IEI:        0x01,
		MOHL:          28,
		CDR:           2578512475,
		IMEI:          obj.Imei,
		SessStatus:    "0",
		MOMSN:         15661,
		MTMSN:         375,
		TimeOfSession: 1475582020,
	}
	Hraw, err := MOH.EncodeMO()
	if err != nil {
		log.Error(err.Error())
	}

	if err != nil {
		log.Error(" %+v\n", err)
	}
	om := &Omnicom{
		object:           obj,
		hraw:             Hraw,
		handledCh:        make(chan task.Result, 128),
		defaultReport:    time.NewTicker(reportTime * unitForReportTime),
		defaultTimeTimer: reportTime * unitForReportTime,
		geoFences:        make(map[uint32]timerPositionGeofence),
	}
	go om.handlingTasks()
	return om
}

func (o *Omnicom) GetDataForIridiumNetwork() ([]byte, error) {
	o.mgfence.Lock()
	defer o.mgfence.Unlock()
	if len(o.geoFences) > 0 {
		for _, geofence := range o.geoFences {
			if o.object.curPos.Intersection(geofence.p) {
				select {
				case <-geofence.t.C:
				default:
					return nil, ErrTooEarly
				}
			}
		}
	} else {
		select {
		case <-o.defaultReport.C:
		default:
			return nil, ErrTooEarly
		}
	}
	return o.makeBuffSPR()
}

func (o *Omnicom) GetStopAlertingMessage() ([]byte, error) {
	if !o.stopAlerting {
		return nil, nil
	}
	o.stopAlerting = false
	return o.makeBuffStopAlerting()
}

func (o *Omnicom) GetStartAlertingMessage() ([]byte, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if !o.sentTestMode && !o.startAlerting {
		o.startAlerting = true
		o.startTypeAlerting = TM
		o.sentTestMode = true
	}
	if !o.startAlerting {
		return nil, nil
	}
	o.startAlerting = false
	return o.makeBuffStartAlerting(0)
}

func (o *Omnicom) StartAlerting(typeAlerting uint) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if typeAlerting >= LastTypeAlerting {
		return errors.New("undefined type of the alerting")
	}
	o.startTypeAlerting = typeAlerting
	o.startAlerting = true
	return nil
}

func (o *Omnicom) StopAlerting(typeAlerting uint) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if typeAlerting >= LastTypeAlerting {
		return errors.New("undefined type of the alerting")
	}
	o.stopTypeAlerting = typeAlerting
	o.stopAlerting = true
	return nil
}

func (o *Omnicom) UpdateInformation(object Object) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.object = object
}

func (o *Omnicom) AddTask(t interface{}) (len int, err error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.queueTask.Len() == taskQueueSize {
		return 0, task.ErrTaskQueueFull
	}
	o.queueTask.PushBack(t)
	return o.queueTask.Len(), nil
}

func (o *Omnicom) GetChannelHandledTask() <-chan task.Result {
	return o.handledCh
}

func (o *Omnicom) handlingTasks() {
	for {
		o.mu.Lock()
		for el := o.queueTask.Front(); el != nil; el = el.Next() {
			taskInfo := el.Value
			switch ti := taskInfo.(type) {
			case *task.MT:
				o.handleMT(ti)
			}
			o.queueTask.Remove(el)
		}
		o.mu.Unlock()
		time.Sleep(sleepForHandlingTask)
	}
}

func (o *Omnicom) handleMT(ti *task.MT) {
	switch t := ti.Payload.Omn.(type) {
	case *omnicom.RMH:
		o.rmhMessage(t)
	case *omnicom.UGP:
		o.globalParametersMessage(t)
	case *omnicom.RSM:
		log.Debug("received RSM %+v", t)
		switch t.Msg_to_Ask {
		case 0x00:
			o.requestAlertReport(t)
		case 0x01:
			o.sendLastPositionReport(t)
		case 0x02:
			o.object.RequestCurrentPos <- struct{}{}
			<-o.object.RequestCurrentPos
			o.sendCurrentPositionReport(t)
		case 0x03:
			o.requestGlobalParameters(t)
		default:
			log.Error("Request specific message with uknwn message to ask value %+v", t.Msg_to_Ask)
		}
	case *omnicom.UG_Polygon:
		o.geofenceMessage(t)
	case *omnicom.UIC:
		log.Debug("received a UIC %+v", t)
		o.unitIntervalChangeMessage(t)
	case *omnicom.TMA:
		log.Debug("received a TMA, not futher action is required")
	default:
		log.Error("Unknown task type: %T %v", t, ti)
	}
}
