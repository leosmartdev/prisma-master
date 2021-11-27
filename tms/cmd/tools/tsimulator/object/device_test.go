package object

import (
	geot "prisma/tms/geo"
	"prisma/tms/iridium"
	"prisma/tms/nmea"
	"prisma/tms/omnicom"
	"testing"

	"github.com/json-iterator/go/assert"
)

var RadarObject = &Object{
	Device: "radar",
	Name:   "MyRadarForTest",
	Mmsi:   1,
	curPos: PositionArrivalTime{
		PositionSpeed: PositionSpeed{
			Latitude:  1.173931,
			Longitude: 102.041706,
			Speed:     0.5,
		},
	},
	Pos: []PositionArrivalTime{
		{
			PositionSpeed: PositionSpeed{
				Latitude:  1.173931,
				Longitude: 102.041706,
				Speed:     0.5,
			},
		},
		{
			PositionSpeed: PositionSpeed{
				Latitude:  1.123931,
				Longitude: 102.041706,
				Speed:     0.5,
			},
		},
	},
	activePos: 1,
	Number:    1,
	Status:    "A",
}

var AISObject = &Object{
	Device: "ais",
	Name:   "MyRadarForTest",
	Mmsi:   1,
	curPos: PositionArrivalTime{
		PositionSpeed: PositionSpeed{
			Latitude:  1.173931,
			Longitude: 102.041706,
			Speed:     12.5,
		},
	},
	Number: 1,
	Status: "A",
}

var OmnicomObject = &Object{
	Device: "omnicom",
	Name:   "MyRadarForTest",
	Mmsi:   1,
	curPos: PositionArrivalTime{
		PositionSpeed: PositionSpeed{
			Latitude:  1.173931,
			Longitude: 102.041706,
			Speed:     12.5,
		},
	},
	Number: 1,
	Status: "A",
	Imei:   "123456789012345",
}

func TestRadar_GetMessagePosition(t *testing.T) {
	radar := NewRadar(*RadarObject)
	_, err := radar.GetMessagePosition()
	assert.Error(t, err)
}

func TestRadar_GetMessageStaticInformation(t *testing.T) {
	radar := NewRadar(*RadarObject)
	_, err := radar.GetMessageStaticInformation()
	assert.Error(t, err)
}

func testRadarGeneratePosition(t *testing.T, positionRadar, positionStation Position) {
	radar := NewRadar(*RadarObject)
	radar.object.curPos.Latitude = positionRadar.Latitude
	radar.object.curPos.Longitude = positionRadar.Longitude
	message, err := radar.GetTrackedTargetMessage(positionStation.Latitude, positionStation.Longitude)
	assert.NoError(t, err)
	sent, _ := nmea.Parse(string(message))
	nm, err := nmea.PopulateProtobuf(sent)
	assert.NoError(t, err)
	lat, lon := geo.At(positionStation.Latitude, positionStation.Longitude, nm.Ttm.Distance, nm.Ttm.Bearing)
	assert.InDelta(t, positionRadar.Latitude, lat, 0.0005)
	assert.InDelta(t, positionRadar.Longitude, lon, 0.0005)

}

func TestRadar_GetTrackedTargetMessage(t *testing.T) {
	radar := NewRadar(*RadarObject)
	message, err := radar.GetTrackedTargetMessage(1.151931, 102.003754)
	assert.NoError(t, err)
	sent, err := nmea.Parse(string(message))
	assert.NoError(t, err)
	nm, err := nmea.PopulateProtobuf(sent)
	assert.NoError(t, err)
	assert.Equal(t, "A", nm.Ttm.Status)
	assert.Equal(t, 0.5, nm.Ttm.Speed)

	// test positions
	lat, lon, err := geot.FindPositionUsingHaversineAlg(1.151931, 102.003754, nm.Ttm.Distance, nm.Ttm.Bearing)
	assert.NoError(t, err)
	assert.InDelta(t, 1.173931, lat, 0.1)
	assert.InDelta(t, 102.141706, lon, 0.2)
	assert.Equal(t, float64(0), nm.Ttm.Course)

	testRadarGeneratePosition(t, Position{1.1, 1.1}, Position{1.1, 1.1})
	testRadarGeneratePosition(t, Position{0.6, 0.7}, Position{0.1, 0.1})
	testRadarGeneratePosition(t, Position{1.4, 100.4}, Position{1.5, 102.4})
	testRadarGeneratePosition(t, Position{0.3, 0.3}, Position{0.6, 0.6})
	testRadarGeneratePosition(t, Position{0.1, 0.3}, Position{0.4, 0.4})
	testRadarGeneratePosition(t, Position{0.1, 0.3}, Position{0.6, 0.6})
}

func TestAIS_GetTrackedTargetMessage(t *testing.T) {
	ais := NewAIS(*AISObject)
	_, err := ais.GetTrackedTargetMessage(1.151931, 102.003754)
	assert.Error(t, err)
}

func TestOmnicom_StartAlerting(t *testing.T) {
	deviceObj := NewOmnicom(*OmnicomObject, 1)
	assert.NoError(t, deviceObj.StartAlerting(PU))
	assert.Equal(t, uint(PU), deviceObj.startTypeAlerting)
	assert.NoError(t, deviceObj.StartAlerting(PD))
	assert.Equal(t, uint(PD), deviceObj.startTypeAlerting)
	assert.NoError(t, deviceObj.StartAlerting(BA))
	assert.Equal(t, uint(BA), deviceObj.startTypeAlerting)
	assert.NoError(t, deviceObj.StartAlerting(IA))
	assert.Equal(t, uint(IA), deviceObj.startTypeAlerting)
	assert.NoError(t, deviceObj.StartAlerting(NPF))
	assert.Equal(t, uint(NPF), deviceObj.startTypeAlerting)
	assert.NoError(t, deviceObj.StartAlerting(JBDA))
	assert.Equal(t, uint(JBDA), deviceObj.startTypeAlerting)
	assert.NoError(t, deviceObj.StartAlerting(LMC))
	assert.Equal(t, uint(LMC), deviceObj.startTypeAlerting)
	assert.NoError(t, deviceObj.StartAlerting(DA))
	assert.Equal(t, uint(DA), deviceObj.startTypeAlerting)
	assert.NoError(t, deviceObj.StartAlerting(AA))
	assert.Equal(t, uint(AA), deviceObj.startTypeAlerting)
	assert.NoError(t, deviceObj.StartAlerting(TM))
	assert.Equal(t, uint(TM), deviceObj.startTypeAlerting)
	assert.Error(t, deviceObj.StartAlerting(LastTypeAlerting+1))
}

func TestOmnicom_StopAlerting(t *testing.T) {
	deviceObj := NewOmnicom(*OmnicomObject, 1)
	assert.NoError(t, deviceObj.StopAlerting(PU))
	assert.Equal(t, uint(PU), deviceObj.stopTypeAlerting)
	assert.NoError(t, deviceObj.StopAlerting(PD))
	assert.Equal(t, uint(PD), deviceObj.stopTypeAlerting)
	assert.NoError(t, deviceObj.StopAlerting(BA))
	assert.Equal(t, uint(BA), deviceObj.stopTypeAlerting)
	assert.NoError(t, deviceObj.StopAlerting(IA))
	assert.Equal(t, uint(IA), deviceObj.stopTypeAlerting)
	assert.NoError(t, deviceObj.StopAlerting(NPF))
	assert.Equal(t, uint(NPF), deviceObj.stopTypeAlerting)
	assert.NoError(t, deviceObj.StopAlerting(JBDA))
	assert.Equal(t, uint(JBDA), deviceObj.stopTypeAlerting)
	assert.NoError(t, deviceObj.StopAlerting(LMC))
	assert.Equal(t, uint(LMC), deviceObj.stopTypeAlerting)
	assert.NoError(t, deviceObj.StopAlerting(DA))
	assert.Equal(t, uint(DA), deviceObj.stopTypeAlerting)
	assert.NoError(t, deviceObj.StopAlerting(AA))
	assert.Equal(t, uint(AA), deviceObj.stopTypeAlerting)
	assert.NoError(t, deviceObj.StopAlerting(TM))
	assert.Equal(t, uint(TM), deviceObj.stopTypeAlerting)
	assert.Error(t, deviceObj.StopAlerting(LastTypeAlerting+1))
}

func TestOmnicom_GetStartAlertingMessage(t *testing.T) {

	device := NewOmnicom(*OmnicomObject, 1)
	device.sentTestMode = true

	b, err := device.GetStartAlertingMessage()
	assert.Nil(t, err)
	assert.Nil(t, b)
	device.startTypeAlerting = LastTypeAlerting + 1
	device.startAlerting = true
	b, err = device.GetStartAlertingMessage()
	assert.Error(t, err)
	assert.Nil(t, b)

	device.startTypeAlerting = PU
	device.startAlerting = true
	b, err = device.GetStartAlertingMessage()
	assert.NoError(t, err)
	assert.NotNil(t, b)
	h, err := iridium.ParseMOHeader(b[3:])
	assert.NoError(t, err)
	omni, err := iridium.ParseMOPayload(b[h.MOHL+6:])
	assert.NoError(t, err)
	ar := omni.Omn.(*omnicom.AR)
	assert.Equal(t, omnicom.Power_Up{1, 1}, ar.Power_Up)
}

func TestOmnicom_GetStopAlertingMessage(t *testing.T) {
	device := NewOmnicom(*OmnicomObject, 1)
	b, err := device.GetStopAlertingMessage()
	assert.Nil(t, err)
	assert.Nil(t, b)
	device.stopTypeAlerting = LastTypeAlerting + 1
	device.stopAlerting = true
	b, err = device.GetStopAlertingMessage()
	assert.Error(t, err)
	assert.Nil(t, b)
	device.stopTypeAlerting = PU
	device.stopAlerting = true
	b, err = device.GetStopAlertingMessage()
	assert.NoError(t, err)
	assert.NotNil(t, b)
}
