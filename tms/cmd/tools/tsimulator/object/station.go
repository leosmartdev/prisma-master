package object

import (
	"container/list"
	"math"
	"math/rand"
	"prisma/tms/log"
	"prisma/tms/nmea"
	"sync"
	"time"
)

// Structures for releasing a behavior of statuses of sea objects like radars.
// It has next restrictions:
// - only 0...99 numbers for objects
// - the status A can be only 10 seconds
// - the status L is the last status
// - After L the number should be free
type infoRadar struct {
	mu     sync.RWMutex
	status string
	number uint32
}

// Station is a structure contains info about AIS target
type Station struct {
	Device    string
	Latitude  float64
	Longitude float64
	Radius    float64
	Addr      string
	Period    int

	vdm        nmea.VDMO
	mmsi       uint32
	mux        sync.Mutex
	seaObjects map[string]Object // What seaObjects can be showed

	infoRadars   map[string]*infoRadar
	numberRadars list.List
}

// Return a new infoRadar, also it's controlling the first status
func newInfoRadar(number uint32) *infoRadar {
	nw := &infoRadar{
		status: "A",
		number: number,
	}

	go func() {
		timer := time.NewTimer(10 * time.Second)
		defer timer.Stop()
		<-timer.C
		if nw.GetStatus() != "L" {
			nw.SetStatus("T")
		}
	}()
	return nw
}

func (nw *infoRadar) SetStatus(status string) {
	nw.mu.Lock()
	defer nw.mu.Unlock()
	nw.status = status
}

func (nw *infoRadar) GetStatus() string {
	nw.mu.Lock()
	defer nw.mu.Unlock()
	return nw.status
}

func (nw *infoRadar) Leave() {
	nw.SetStatus("L")
}

// Init setups parameters for identification
func (st *Station) Init() {
	st.vdm = nmea.VDMO{
		BaseSentence: nmea.BaseSentence{
			SOS:    "!",
			Talker: "AI",
			Format: "VDM",
		},
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
	}
	st.mmsi = 970000000 + uint32(rand.Int31n(9999))
	st.infoRadars = make(map[string]*infoRadar)
	st.seaObjects = make(map[string]Object)
	for i := uint32(1); i < 100; i++ {
		st.numberRadars.PushBack(i)
	}
}

// GetMessageStaticInformation returns a static information about a station
func (st *Station) GetMessageStaticInformation() ([]byte, error) {
	statStruct := nmea.M13715{
		VDMO: st.vdm,
		CoreM13715: nmea.CoreM13715{
			MessageID:       MessageTypeSRB,
			RepeatIndicator: 0,
			Mmsi:            st.mmsi,
		},
	}
	encoded, _ := statStruct.Encode()
	return []byte(encoded), nil
}

func encodeNmToDegree(m float64) float64 {
	return m * math.Cos(0.1666)
}

func (st *Station) GetMessagePosition() ([]byte, error) {
	posStruct := nmea.M13711{
		VDMO: st.vdm,
		CoreM13711: nmea.CoreM13711{
			MessageID:          MessageTypePosition,
			RepeatIndicator:    0,
			Mmsi:               st.mmsi,
			NavigationalStatus: NavigationStatusAISSart,
			Latitude:           encodeDegree(st.Latitude),
			Longitude:          encodeDegree(st.Longitude),
		}}
	encoded, _ := posStruct.Encode()
	return []byte(encoded), nil
}

func (st *Station) IterateVisibleSeaObjects() []Object {
	st.mux.Lock()
	defer st.mux.Unlock()
	ret := make([]Object, 0, len(st.seaObjects))
	for _, obj := range st.seaObjects {
		ret = append(ret, obj)
	}
	return ret
}

// See is used for what seaObjects a station can see. The station will contain visible seaObjects into a slice
func (st *Station) See(seaObjectChannel <-chan Object) {
	for seaObject := range seaObjectChannel {
		// Are we inside this station?
		if st.Device == seaObject.Device &&
			math.Abs(seaObject.GetCurPos().Longitude-st.Longitude) < encodeNmToDegree(st.Radius) &&
			math.Abs(seaObject.GetCurPos().Latitude-st.Latitude) < encodeNmToDegree(st.Radius) {
			// If it's a radar
			if info, ok := st.infoRadars[seaObject.GetUniqName()]; st.Device == "radar" && (!ok || info.GetStatus() == "L") {
				if st.numberRadars.Len() == 0 {
					log.Warn("A lot of radars...!!! We can't see more")
					continue
				}

				numberRadar := st.numberRadars.Front()
				st.infoRadars[seaObject.GetUniqName()] = newInfoRadar(numberRadar.Value.(uint32))
				st.numberRadars.Remove(numberRadar)
				seaObject.Status = st.infoRadars[seaObject.GetUniqName()].GetStatus()
				seaObject.Number = st.infoRadars[seaObject.GetUniqName()].number
				log.Debug("obj: %v has number: %d", seaObject.Name, numberRadar.Value)
			} else if st.Device == "radar" && ok {
				seaObject.Status = info.GetStatus()
				seaObject.Number = st.infoRadars[seaObject.GetUniqName()].number
			}
			st.mux.Lock()
			st.seaObjects[seaObject.GetUniqName()] = seaObject
			st.mux.Unlock()
		} else if st.Device == seaObject.Device && st.Device == "radar" { // it means this radar has left
			if info, ok := st.infoRadars[seaObject.GetUniqName()]; ok && info.GetStatus() != "L" {
				info.Leave()
				seaObject.Status = info.GetStatus()
				st.numberRadars.PushBack(info.number)

				// delete information about this object
				delete(st.infoRadars, seaObject.GetUniqName())
				st.mux.Lock()
				delete(st.seaObjects, seaObject.GetUniqName())
				st.mux.Unlock()
			}
		} else {
			st.mux.Lock()
			delete(st.seaObjects, seaObject.GetUniqName())
			st.mux.Unlock()
		}
	}
}
