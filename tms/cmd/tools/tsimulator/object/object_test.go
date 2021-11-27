package object

import (
	"container/ring"
	"fmt"
	"math"
	"prisma/tms/nmea"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const deltaDistance = 0.1

func getVessel() Object {
	return Object{
		Mmsi:        235009802,
		Device:      "ais",
		Destination: "testDestination",
		Name:        "test",
		ETA:         "01020304",
		Type:        30,
		Pos: []PositionArrivalTime{
			{
				PositionSpeed: PositionSpeed{
					Latitude:  0,
					Longitude: 0,
					Speed:     10,
				},
			},
			{
				PositionSpeed: PositionSpeed{
					Latitude:  20.1,
					Longitude: 20.1,
					Speed:     10,
				},
			},
			{
				PositionSpeed: PositionSpeed{
					Latitude:  30.1,
					Longitude: 20.1,
					Speed:     30,
				},
			},
		},
		ReportPeriod:    5,
		historyPosition: ring.New(historyTrackSize),
	}
}

func getFewPosVessel() Object {
	return Object{
		Mmsi:        235009802,
		Device:      "ais",
		Destination: "testDestination",
		Name:        "test",
		ETA:         "01020304",
		Type:        30,
		Pos: []PositionArrivalTime{
			{
				PositionSpeed: PositionSpeed{
					Latitude:  0,
					Longitude: 0,
					Speed:     10,
				},
			},
		},
		ReportPeriod: 3,
	}
}

func TestSeaObject_Init(t *testing.T) {
	vessel := getVessel()
	vessel.InitMoving(vessel.ReportPeriod)
	location, _ := time.LoadLocation("")
	timeEtaExp := time.Date(0, 01, 02, 03, 04, 0, 0, location)
	assert.True(t, timeEtaExp.Sub(vessel.deviceObj.(*AIS).etaTime) == 0,
		"got: %s; expected: %s", vessel.deviceObj.(*AIS).etaTime, timeEtaExp)
	assert.Equal(t, 10.0, vessel.curPos.Speed)
	assert.NotZero(t, vessel.ReportPeriod)
	vessel.activePos = 0
	vessel.Pos[0].Speed = 124
	vessel.InitMoving(vessel.ReportPeriod)
	assert.NotEqual(t, float64(124), vessel.curPos.Speed)
	vessel = getFewPosVessel()
	vessel.InitMoving(vessel.ReportPeriod)
	vessel = getVessel()
	vessel.Pos[1].Speed = 0
	vessel.Pos[1].ArrivalTimeSeconds = 10
	vessel.Init()
	assert.Zero(t, vessel.steps)
	vessel.InitMoving(vessel.ReportPeriod)
	assert.NotZero(t, vessel.steps)
}

func TestSeaObject_GetMessagePosition(t *testing.T) {
	vessel := getVessel()
	vessel.InitMoving(vessel.ReportPeriod)
	data, err := vessel.GetMessagePosition()
	assert.NoError(t, err)
	sent, err := nmea.Parse(string(data))
	assert.NoError(t, err)
	nm, err := nmea.PopulateProtobuf(sent)
	assert.Nil(t, err)
	m := nm.GetVdm().M1371.GetPos()
	fmt.Println(string(data))
	fmt.Println(uint32(MessageTypePosition))
	fmt.Println(nm.GetVdm().M1371.MessageId)
	assert.Equal(t, uint32(MessageTypePosition), nm.GetVdm().M1371.MessageId)
	assert.Equal(t, uint32(235009802), nm.GetVdm().M1371.Mmsi)
	assert.Equal(t, encodeDegree(vessel.curPos.Longitude), m.Longitude)
	assert.Equal(t, encodeDegree(vessel.curPos.Latitude), m.Latitude)
}

func TestSeaObject_Move(t *testing.T) {
	vessel := getVessel()
	vessel.InitMoving(vessel.ReportPeriod)
	// Why decrement j on 1? Because a moving has to change a destination on the last step and do the one
	for i, j := 0, int(vessel.steps); i < j-1; i++ {
		vessel.Move(vessel.ReportPeriod)
	}
	assert.InDelta(t, vessel.Pos[1].Latitude, vessel.curPos.Latitude, deltaDistance)
	assert.InDelta(t, vessel.Pos[1].Longitude, vessel.curPos.Longitude, deltaDistance)
	vessel.InitMoving(vessel.ReportPeriod)
	for i, j := 0, int(vessel.steps); i < j-1; i++ {
		vessel.Move(vessel.ReportPeriod)
	}
	assert.InDelta(t, vessel.Pos[2].Latitude, vessel.curPos.Latitude, deltaDistance)
	assert.InDelta(t, vessel.Pos[2].Longitude, vessel.curPos.Longitude, deltaDistance)

	vessel.curPos = PositionArrivalTime{}
	vessel.Pos = make([]PositionArrivalTime, 0)
	vessel.Move(vessel.ReportPeriod) // index out of range ? :)

	// test 1 position
	// it should not move
	vessel = getVessel()
	vessel.Pos = vessel.Pos[:1]
	copiedVessel := vessel
	vessel.Move(vessel.ReportPeriod)
	assert.Equal(t, copiedVessel.curPos.Longitude, vessel.curPos.Longitude)
	assert.Equal(t, copiedVessel.curPos.Latitude, vessel.curPos.Latitude)

	// Positions have arrival time only. The end position should be reachable
	vessel = getVessel()
	vessel.Pos[1].Speed = 0
	vessel.Pos[1].ArrivalTimeSeconds = 3
	vessel.Move(vessel.ReportPeriod)
	assert.False(t, math.IsInf(vessel.steps, 0))

	// test a strange omnicom
	// it should move immediately
	vessel = getVessel()
	vessel.Pos = []PositionArrivalTime{
		{
			PositionSpeed: PositionSpeed{
				Latitude:  0.891826,
				Longitude: 104.0442829,
				Speed:     5,
			},
		}, {
			PositionSpeed: PositionSpeed{
				Latitude:  0.390855,
				Longitude: 104.557563,
			},
			ArrivalTimeSeconds: 100,
		}, {
			PositionSpeed: PositionSpeed{
				Latitude:  0.644656,
				Longitude: 104.084028,
				Speed:     5,
			},
			Damage: Medium,
			Collided: true,
		},
	}
	vessel.Move(vessel.ReportPeriod)
	assert.Equal(t, vessel.Pos[0].Longitude, vessel.curPos.Longitude)
	assert.Equal(t, vessel.Pos[0].Latitude, vessel.curPos.Latitude)
	// after it should move little bit
	vessel.Move(vessel.ReportPeriod)
	assert.InDelta(t, vessel.Pos[1].Longitude, vessel.curPos.Longitude, 1)
	assert.InDelta(t, vessel.Pos[1].Latitude, vessel.curPos.Latitude, 1)
	// damage
	cpPos := vessel.Pos
	vessel.InitMoving(1)
	vessel.Move(1)
	// here vessel should not move anymore
	assert.Equal(t, cpPos, vessel.Pos)
}

func TestSeaObject_GetHistoryFromTo(t *testing.T) {
	vessel := getVessel()
	vessel.InitMoving(vessel.ReportPeriod)
	lastPartCount := 0
	steps := int(vessel.steps)
	timeSleep := 0
	const sleepDuration = 20 * time.Millisecond
	var tFirst time.Time
	// Why decrement j on 1? Because a moving has to change a destination on the last step and do the one
	for i := 0; i < steps-1; i, lastPartCount = i+1, lastPartCount+1 {
		vessel.Move(vessel.ReportPeriod)
		if i%1000 == 1 {
			tFirst = vessel.historyPosition.Value.(*PositionSpeedTime).t
		}
		if i%1000 == 0 {
			timeSleep++
			lastPartCount = 0
			time.Sleep(sleepDuration)
		}
	}
	// we have whole history of the track since start the test
	r := vessel.GetHistoryFromTo(time.Now().Add(-1*time.Second), time.Now().Add(1*time.Second))
	assert.Len(t, r, historyTrackSize)

	// we don't have history for future
	r = vessel.GetHistoryFromTo(time.Now().Add(20*time.Second), time.Now())
	assert.Empty(t, r)

	// Get first part
	r = vessel.GetHistoryFromTo(tFirst.Add(-1*time.Microsecond), tFirst.Add(1*time.Microsecond))
	assert.Len(t, r, 1)

	// we are able to get a part of the history
	r = vessel.GetHistoryFromTo(time.Now().Add(-sleepDuration), time.Now())
	assert.Len(t, r, lastPartCount-1)
}


func TestSeaObject_GetMessageStaticInformation(t *testing.T) {
	vessel := getVessel()
	vessel.InitMoving(vessel.ReportPeriod)
	data, err := vessel.GetMessageStaticInformation()
	assert.NoError(t, err)
	sent, err := nmea.Parse(string(data))
	assert.NoError(t, err)
	nm, err := nmea.PopulateProtobuf(sent)
	assert.NoError(t, err)
	m := nm.GetVdm().M1371.GetStaticVoyage()
	assert.Equal(t, uint32(MessageTypeSRB), nm.GetVdm().M1371.MessageId)
	assert.Equal(t, uint32(235009802), nm.GetVdm().M1371.Mmsi)
	//assert.Equal(t, "test", m.Name)
	//assert.Equal(t, "test", m.CallSign)
	//assert.Equal(t, "testDestination", m.Destination)
	assert.Equal(t, uint32(4), m.EtaMinute)
	assert.Equal(t, uint32(3), m.EtaHour)
	assert.Equal(t, uint32(2), m.EtaDay)
	assert.Equal(t, uint32(1), m.EtaMonth)
}
