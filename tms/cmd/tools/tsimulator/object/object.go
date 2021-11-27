// Package object is fabric to create different seaobjects that will be controlled by tsimulator.
package object

import (
	"container/ring"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"prisma/tms/log"
	"strconv"
	"strings"
	"time"

	"prisma/tms/cmd/tools/tsimulator/task"

	"github.com/StefanSchroeder/Golang-Ellipsoid/ellipsoid"
)

var (
	geo        = ellipsoid.Init("WGS84", ellipsoid.Degrees, ellipsoid.Nm, ellipsoid.LongitudeIsSymmetric, ellipsoid.BearingNotSymmetric)
	randomizer = rand.New(rand.NewSource(time.Now().UnixNano()))
	// ErrHolden is for determine an object is holden and unreachable yet
	ErrHolden = errors.New("this object is holden")
)

const (
	// A Class A AIS unit broadcasts the following information every 2 to 10 seconds while underway...
	sleepGetInfoPosition = 20
	// the Class A AIS unit broadcasts the following information every 6 minutes.
	sleepGetInfoStaticInformation = 6
	// Target messaging how often to send
	sleepGetTrackedTargetMessage = 10
)

// TimeConfig will contain information about reporting periods for different messages
type TimeConfig struct {
	SleepGetInfoPosition          uint32 `json:"sleep_get_info_position"`
	SleepMoveVessels              uint32 `json:"sleep_move_vessels"`
	SleepGetInfoStaticInformation uint32 `json:"sleep_get_info_static_information"`
	SleepGetTrackedTargetMessage  uint32 `json:"sleep_get_tracked_target_message"`
}

const (
	// a default direct between a valid pointer and an error pointer for sarsat beacons. It uses NM
	defaultDirectErrorPointer = 10
	// history size for whole a track
	historyTrackSize = 150
)

// ObjectCommunicator should be used for any sea objects
// that should communicate with something(e.g. iridium, stations)
type ObjectCommunicator interface {
	Informer
	// Return a tracked target message which provided by radar's devices
	// latitude and longitude are positions of a station in rad
	GetTrackedTargetMessage(latitude, longitude float64) ([]byte, error)
	// Return a message for starting an alerting or error
	GetStartAlertingMessage() ([]byte, error)
	// Return a message for stopping an alerting or error
	GetStopAlertingMessage() ([]byte, error)
	// Stating an alert
	StartAlerting(uint) error
	// Stopping an alert
	StopAlerting(uint) error
	// Return a chan of tasks - bytes for sending to somewhere
	GetChannelHandledTask() <-chan task.Result
	// Get information about an object for sending via IridiumNetwork
	GetDataForIridiumNetwork() ([]byte, error)
	// Generate a message for current position of an alert
	GetPositionAlertingMessage() ([]byte, error)
	// Add a task to an object
	AddTask(t interface{}) (len int, err error)
}

// Object is an object which has several "devices" - features
type Object struct {
	// This field is for web
	Id int

	// Common information for all devices
	Device           string
	Mmsi             uint32
	Pos              []PositionArrivalTime
	ReportPeriod     uint32
	Name             string
	NavigationStatus uint32      `json:"navigation_status"`
	TimeConfig       *TimeConfig `json:"time_config"`
	historyPosition  *ring.Ring
	damaged          string // affects on behavior on an object after middle will be some fines on this object
	startLive        time.Time

	// Tickers for issuing messages
	TickerTTM               *time.Ticker `json:"-"`
	TickerPosition          *time.Ticker `json:"-"`
	TickerStaticInformation *time.Ticker `json:"-"`
	TickerMoving            *time.Ticker `json:"-"`

	// Time of previous ticker
	TimeOfLastMove time.Time

	// Ais information
	Destination string
	ETA         string
	Type        uint32

	// Radar information
	Number uint32 `json:"-"`
	Status string `json:"-"`

	// Omnicom information
	Imei              string
	ImeiG3            string           `json:"imei_g3"`
	OmnicomId         uint32           `json:"omnicom_id"`
	ReportTimer       <-chan time.Time `json:"-"` // it leaks
	RequestCurrentPos chan struct{}    `json:"-"`

	// Sarsat information
	Unlocated    bool    `json:"unlocated,omitempty"`
	Located      bool    `json:"located,omitempty"`
	Elemental    []Elemental
	DOA          Position
	BeaconId     string  `json:"beacon_id"`
	ErrorPointNM float64 `json:"error_point_nm"`

	// information for inside working
	deviceObj DeviceCommunicator
	curPos    PositionArrivalTime
	activePos int
	deltaD    float64
	steps     float64
	distance  float64
	bearing   float64
	course    uint32
}

// NewObject return new seaobject
func NewObject() *Object {
	return &Object{}
}

func (s *Object) Init() {
	if s.Name == "" {
		s.Name = strconv.Itoa(time.Now().Nanosecond())
	}
	if s.BeaconId == "" {
		s.BeaconId = "B029C2900D97591"
	}
	if s.Imei == "" {
		s.Imei = (s.Name + "P00234010030450")[0:15]
	}
	if s.Mmsi == 0 {
		s.Mmsi = 235009800 + uint32(randomizer.Int31n(math.MaxInt32))
	}
	if s.ReportPeriod == 0 {
		log.Info("ReportPeriod is 0, set 2")
		s.ReportPeriod = 2
	}
	if s.NavigationStatus != NavigationStatusAISSart && s.NavigationStatus != NavigationStatusAISSartTesting {
		s.NavigationStatus = 0
	}
	s.setupTickers()
	if s.ErrorPointNM == 0 {
		s.ErrorPointNM = defaultDirectErrorPointer
	}
	s.SetupDevice()
	if s.historyPosition == nil {
		s.historyPosition = ring.New(historyTrackSize)
	}
}

func (s *Object) setupTickers() {
	// don't resetup
	if s.TickerMoving != nil {
		return
	}
	if s.TimeConfig == nil {
		s.TimeConfig = new(TimeConfig)
	}
	if s.TimeConfig.SleepGetInfoPosition == 0 {
		s.TimeConfig.SleepGetInfoPosition = sleepGetInfoPosition
	}
	if s.TimeConfig.SleepGetInfoStaticInformation == 0 {
		s.TimeConfig.SleepGetInfoStaticInformation = sleepGetInfoStaticInformation
	}
	if s.TimeConfig.SleepGetTrackedTargetMessage == 0 {
		s.TimeConfig.SleepGetTrackedTargetMessage = sleepGetTrackedTargetMessage
	}
	if s.TimeConfig.SleepMoveVessels == 0 {
		s.TimeConfig.SleepMoveVessels = s.TimeConfig.SleepGetInfoPosition
	}
	s.TickerPosition = time.NewTicker(time.Duration(s.TimeConfig.SleepGetInfoPosition) * time.Second)
	s.TickerStaticInformation = time.NewTicker(time.Duration(s.TimeConfig.SleepGetInfoStaticInformation) * time.Second)
	s.TickerTTM = time.NewTicker(time.Duration(s.TimeConfig.SleepGetTrackedTargetMessage) * time.Second)
	s.TickerMoving = time.NewTicker(time.Duration(s.TimeConfig.SleepMoveVessels) * time.Second)
}

// InitMoving initialises vars like speed, steps and etc.
// It needs for reinit also
func (s *Object) InitMoving(period uint32) {
	s.Init()
	s.initCurPos()
	s.deltaD = s.curPos.Speed * (float64(period) / 60)
	if len(s.Pos) == 0 {
		return
	}
	s.distance, s.bearing =
		geo.To(s.curPos.Latitude, s.curPos.Longitude,
			s.Pos[s.activePos].Latitude, s.Pos[s.activePos].Longitude)
	s.steps = s.distance / s.deltaD
	s.course = uint32(s.bearing)
}

func (s *Object) SetupDevice() {
	if s.deviceObj != nil {
		return
	}
	switch strings.ToLower(s.Device) {
	default:
		log.Warn("Undefined a device: %s. Setup a null device", s.Device)
		s.deviceObj = NewNullDevice()
	case "radar":
		s.deviceObj = NewRadar(*s)
	case "ais", "sart":
		s.deviceObj = NewAIS(*s)
	case "omnicom":
		s.RequestCurrentPos = make(chan struct{})
		s.deviceObj = NewOmnicom(*s, time.Duration(s.ReportPeriod))
	case "sarsat":
		s.deviceObj = NewSarsat(*s)
	case "spidertracks":
		s.deviceObj = NewSpiderTrack(*s)
	}
}

// Move is for moving a sea object to destination position, like a life cycle
// If we arrived then repeat a way
func (s *Object) Move(period uint32) {
	s.TimeOfLastMove = time.Now()
	if s.isHolden() {
		log.Info("objects is holden still it will be alive %v", s.startLive.Sub(time.Now()))
		return
	}
	if s.damaged != Low && s.damaged != "" {
		log.Info("object is damaged with %s level", s.damaged)
		return
	}
	// Need init?
	if s.steps < 1 {
		s.InitMoving(period)
	}
	if len(s.Pos) <= 1 {
		return
	}
	// it means we need to move immediately
	var p PositionArrivalTime
	if s.steps >= 1 {
		p.Latitude, p.Longitude = geo.At(s.curPos.Latitude, s.curPos.Longitude, s.deltaD, s.bearing)
		s.curPos.Latitude, s.curPos.Longitude = p.Latitude, p.Longitude
		_, s.bearing = geo.To(p.Latitude, p.Longitude, s.Pos[s.activePos].Latitude, s.Pos[s.activePos].Longitude)
		s.steps--
	} else {
		p = s.Pos[s.activePos]
	}
	s.historyPosition = s.historyPosition.Next()
	s.historyPosition.Value = NewPositionSpeedTime(p.PositionSpeed)

}

func (s *Object) GetHistoryFromTo(from, to time.Time) (res []*PositionSpeedTime) {
	if s.historyPosition == nil {
		return nil
	}
	for pos, start := s.historyPosition, s.historyPosition; pos != nil && pos.Value != nil; {

		if pst, ok := pos.Value.(*PositionSpeedTime); !ok {
			log.Warn("Bad type for an interface")
		} else if pst.t.After(from) && pst.t.Before(to) {
			res = append(res, pst)
		}
		pos = pos.Prev()
		if pos == start {
			break
		}
	}
	return
}

func (s Object) GetDevice() DeviceCommunicator {
	s.deviceObj.UpdateInformation(s)
	return s.deviceObj
}

func (s Object) GetCurPos() PositionArrivalTime {
	return s.curPos
}

func (s Object) GetBearing() float64 {
	return s.bearing
}

func (s Object) GetCourse() uint32 {
	return s.course
}

func (s *Object) SetCurPos(pos PositionArrivalTime) {
	s.curPos = pos
}

// GetUniqName returns an unique name of this object
func (s Object) GetUniqName() string {
	return fmt.Sprintf("%s%d%s", s.Imei, s.Mmsi, s.Name)
}

func (s Object) GetMessagePosition() ([]byte, error) {
	if s.isHolden() {
		return nil, ErrHolden
	}
	s.deviceObj.UpdateInformation(s)
	return s.deviceObj.GetMessagePosition()
}

func (s Object) GetMessageStaticInformation() ([]byte, error) {
	s.deviceObj.UpdateInformation(s)
	return s.deviceObj.GetMessageStaticInformation()
}

func (s Object) GetTrackedTargetMessage(latitude, longitude float64) ([]byte, error) {
	s.deviceObj.UpdateInformation(s)
	return s.deviceObj.GetTrackedTargetMessage(latitude, longitude)
}

func (s Object) GetDataForIridiumNetwork() ([]byte, error) {
	if s.isHolden() {
		return nil, ErrHolden
	}
	s.deviceObj.UpdateInformation(s)
	return s.deviceObj.GetDataForIridiumNetwork()
}

func (s Object) GetStartAlertingMessage() ([]byte, error) {
	return s.deviceObj.GetStartAlertingMessage()
}

func (s Object) GetStopAlertingMessage() ([]byte, error) {
	return s.deviceObj.GetStopAlertingMessage()
}

func (s Object) StartAlerting(typeAlerting uint) error {
	return s.deviceObj.StartAlerting(typeAlerting)
}

func (s Object) StopAlerting(typeAlerting uint) error {
	return s.deviceObj.StopAlerting(typeAlerting)
}

func (s Object) GetPositionAlertingMessage() ([]byte, error) {
	if s.isHolden() {
		return nil, ErrHolden
	}
	return s.deviceObj.GetPositionAlertingMessage()
}

func (s Object) AddTask(t interface{}) (len int, err error) {
	return s.deviceObj.AddTask(t)
}

func (s Object) GetChannelHandledTask() <-chan task.Result {
	return s.deviceObj.GetChannelHandledTask()
}

// isHolden return the state of current object. The object can be holden
// by setting time for first position
func (s *Object) isHolden() bool {
	return s.startLive.After(time.Now())
}

func (s *Object) initCurPos() {
	if len(s.Pos) < 1 {
		return
	}
	// choose a next destination
	s.curPos = s.Pos[s.activePos]
	// if current position is collided and affects on this object then stop computing new position
	if s.curPos.Damage != "" && s.curPos.Collided {
		s.damaged = s.curPos.Damage
		return
	}
	s.activePos++
	// we have finished?
	if s.activePos == len(s.Pos) {
		s.startLive = time.Time{}
		s.curPos = s.Pos[0]
		if len(s.Pos) != 1 {
			s.activePos = 1
		} else {
			s.activePos = 0
		}
	}
	if s.startLive.IsZero() {
		s.startLive = time.Now().Add(time.Duration(s.curPos.ArrivalTimeSeconds) * time.Second)
	}
	// We need to compute zero speed using sleep get info position
	// In 1 minute it will be called n(1 minute - 60 / sleepGetInfoPosition) times
	// We need to compute steps in totally to get expected result
	nextPos := s.Pos[s.activePos]
	if nextPos.ArrivalTimeSeconds != 0 {
		nextPos.ArrivalTimeSeconds *= 60 / int(s.TimeConfig.SleepMoveVessels)
	}
	if s.curPos.ArrivalTimeSeconds != 0 {
		s.curPos.ArrivalTimeSeconds *= 60 / int(s.TimeConfig.SleepMoveVessels)
	}
	s.curPos.ComputeZeroSpeed(nextPos)
	if s.Pos[s.activePos].ArrivalTimeSeconds == 0 &&
		(s.curPos.Speed == 0 || (s.curPos.Speed > 123 && s.Device != "spidertracks")) {
		log.Info("Speed is %f, set randomly", s.curPos.Speed)
		s.curPos.Speed = float64(randomizer.Int31n(150)) / 10
	}
}
