package object

import (
	"prisma/tms/nmea"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func getStation() Station {
	return Station{
		Device:    "radar",
		Longitude: 1.0,
		Latitude:  1.0,
		mmsi:      970000000,
		Radius:    0.5,
		Addr:      ":9000",
	}
}

func TestNewInfoRadar(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	nm := newInfoRadar(1)
	assert.Equal(t, "A", nm.status)
	time.Sleep(2 * time.Second)
	assert.Equal(t, "A", nm.status)
	time.Sleep(9 * time.Second)
	assert.Equal(t, "T", nm.status)
}

func TestInfoRadar_Leave(t *testing.T) {
	nm := newInfoRadar(1)
	nm.Leave()
	assert.Equal(t, "L", nm.status)
	if testing.Short() {
		t.Skip()
	}
	time.Sleep(10 * time.Second)
	assert.Equal(t, "L", nm.status)
}

func TestStation_Init(t *testing.T) {
	station := getStation()
	station.Init()
	assert.InDelta(t, 970000000, station.mmsi, 999999)
	assert.Equal(t, 99, station.numberRadars.Len())
}

func TestStation_See(t *testing.T) {
	station := getStation()
	station.Init()
	vessels := []Object{
		{
			Device: "radar",
			Name:   "good",
			Mmsi:   1,
			curPos: PositionArrivalTime{
				PositionSpeed: PositionSpeed{
					Longitude: 1.0,
					Latitude:  1.0,
				},
			},
		},
		{
			Device: "radar",
			Name:   "bad",
			Mmsi:   2,
			curPos: PositionArrivalTime{
				PositionSpeed: PositionSpeed{
					Longitude: 1.5,
					Latitude:  1.5,
				},
			},
		}, {
			Device: "radar",
			Name:   "bad1",
			Mmsi:   3,
			curPos: PositionArrivalTime{
				PositionSpeed: PositionSpeed{
					Longitude: 1.0,
					Latitude:  1.5,
				},
			},
		}, {
			Device: "radar",
			Name:   "bad2",
			Mmsi:   4,
			curPos: PositionArrivalTime{
				PositionSpeed: PositionSpeed{
					Longitude: 1.5,
					Latitude:  1.0,
				},
			},
		}, {
			Device: "ais",
			Name:   "bad3",
			Mmsi:   5,
			curPos: PositionArrivalTime{
				PositionSpeed: PositionSpeed{
					Longitude: 1.0,
					Latitude:  1.0,
				},
			},
		},
	}
	sendChan := make(chan Object)
	go station.See(sendChan)
	for i := range vessels {
		vessels[i].Init()
		sendChan <- vessels[i]
	}

	assert.Len(t, station.seaObjects, 1)
	assert.Equal(t, "good", station.seaObjects[vessels[0].GetUniqName()].Name)
	assert.Equal(t, uint32(2), station.numberRadars.Front().Value.(uint32))
	assert.Equal(t, 98, station.numberRadars.Len())
	vessels[0].curPos.Longitude = 99
	sendChan <- vessels[0]
	time.Sleep(1 * time.Second)
	assert.Len(t, station.seaObjects, 0)
	assert.Equal(t, uint32(1), station.numberRadars.Back().Value.(uint32))
	assert.Equal(t, 99, station.numberRadars.Len())
}


func TestStation_GetMessageAIPosition(t *testing.T) {
	station := getStation()
	station.Init()
	data, err := station.GetMessagePosition()
	assert.NoError(t, err)
	sent, err := nmea.Parse(string(data))
	assert.NoError(t, err)
	nm, err := nmea.PopulateProtobuf(sent)
	assert.Nil(t, err)
	m := nm.GetVdm().M1371.GetPos()
	assert.Equal(t, uint32(MessageTypePosition), nm.GetVdm().M1371.MessageId)
	assert.Equal(t, uint32(station.mmsi), nm.GetVdm().M1371.Mmsi)
	assert.Equal(t, encodeDegree(station.Longitude), m.Longitude)
	assert.Equal(t, encodeDegree(station.Latitude), m.Latitude)
}

func TestStation_GetMessageAIStaticInformation(t *testing.T) {
	station := getStation()
	station.Init()
	data, err := station.GetMessageStaticInformation()
	assert.NoError(t, err)
	sent, err := nmea.Parse(string(data))
	assert.NoError(t, err)
	nm, err := nmea.PopulateProtobuf(sent)
	assert.NoError(t, err)
	assert.Equal(t, uint32(MessageTypeSRB), nm.GetVdm().M1371.MessageId)
	assert.Equal(t, uint32(station.mmsi), nm.GetVdm().M1371.Mmsi)
}
