package web

import (
	"testing"
	"prisma/tms/cmd/tools/tsimulator/object"
	"github.com/stretchr/testify/assert"
	"prisma/tms/iridium"
	"prisma/tms/omnicom"
	"prisma/gogroup"
	"prisma/tms/test/context"
	"net"
	"time"
)

func TestDirectIPMT_Listen(t *testing.T) {
	const imei = "123456789012345"
	const addr = ":65532"
	const UCMID = "1234"

	obj := object.NewObject()
	obj.Device = "omnicom"
	obj.Imei = imei
	obj.Pos = []object.PositionArrivalTime{
		{
			PositionSpeed: object.PositionSpeed{
				Longitude: 0,
				Latitude:  0,
			},
		},
		{
			PositionSpeed: object.PositionSpeed{
				Longitude: 1,
				Latitude:  1,
			},
		},
	}
	control := object.NewObjectControl([]object.Object{*obj}, nil)
	dipmt := NewDirectIPMT(control)
	group := gogroup.New(context.Background(), "test_dip")

	// setup the listener
	go dipmt.Listen(group, addr)
	defer group.Cancel(nil)
	time.Sleep(500 * time.Millisecond)
	con, err := net.Dial("tcp", addr)
	assert.NoError(t, err)

	// prepare body
	hmt := iridium.MTHeader{
		IMEI:                  imei,
		UniqueClientMessageID: UCMID,
		IEI:                   0x01,
		MTflag:                0x02,
		MTHL:                  0x01,
	}
	hraw, err := hmt.Encode()
	assert.NoError(t, err)
	rmh := omnicom.RMH{
		Date_Interval: omnicom.Date_Interval{
			Start: omnicom.Date{},
			Stop: omnicom.Date{
				Year: 99,
			},
		},
		Header: 0x31,
	}
	payload := iridium.MPayload{IEI: 0x02, Omn: &rmh}
	praw, err := payload.Encode()
	assert.NoError(t, err)
	msgLength := byte(uint16(len(praw) + len(hraw)))
	var raw []byte
	raw = append(raw, 0x01, msgLength>>8, msgLength)
	raw = append(raw, hraw...)
	raw = append(raw, praw...)
	// send an mt message to get hpr
	_, err = con.Write(raw)
	assert.NoError(t, err)
	// read data with ok status
	rbuff := make([]byte, 128)
	con.Read(rbuff)

	// here we have a response from the "server". Parse and be sure we got a success response
	mtR := iridium.MTResponse{}
	mtR.Parse(rbuff)
	assert.Equal(t, imei, mtR.IMEI)
	assert.Equal(t, UCMID, mtR.UniqueClientMessageID)
	assert.Equal(t, iridium.SessionStatus(1), mtR.MTMessageStatus)
}

func TestDirectIPMT_handleMT(t *testing.T) {
	const imei = "123456789012345"
	ar := &omnicom.AR{
		Header: 0x02,
	}
	om := object.NewObject()
	om.Device = "omnicom"
	om.Imei = imei
	dmt := NewDirectIPMT(object.NewObjectControl([]object.Object{*om}, nil))

	// Should not be crashed, should return an error
	_, err := dmt.handleMT(nil)
	assert.Error(t, err)

	mth := iridium.MTHeader{
		IMEI:                  imei,
		UniqueClientMessageID: "asdf",
		MTflag:                0x02,
		MTHL:                  0x01,
		IEI:                   0x01,
	}
	// not enough bytes
	hraw, err := mth.Encode()
	assert.NoError(t, err)
	_, err = dmt.handleMT(hraw)
	assert.Error(t, err)

	payload := iridium.MPayload{IEI: 0x02, Omn: ar}
	praw, err := payload.Encode()
	assert.NoError(t, err)
	// enough bytes header + payload
	var raw []byte
	msgLength := byte(uint16(len(praw) + len(hraw)))
	raw = append(raw, 0x01, msgLength>>8, msgLength)
	raw = append(raw, hraw...)
	raw = append(raw, praw...)
	_, err = dmt.handleMT(raw)
	assert.NoError(t, err)

	// we have to get an error for wrong imei
	mth.IMEI = "2" + imei[1:]
	hraw, err = mth.Encode()
	assert.NoError(t, err)
	_, err = dmt.handleMT(hraw)
	assert.Error(t, err)
	msgLength = byte(uint16(len(praw) + len(hraw)))
	raw = raw[:0]
	raw = append(raw, 0x01, msgLength>>8, msgLength)
	raw = append(raw, hraw...)
	raw = append(raw, praw...)
	_, err = dmt.handleMT(raw)
	assert.Error(t, err)
}
