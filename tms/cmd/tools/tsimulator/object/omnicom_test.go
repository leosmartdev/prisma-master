package object

import (
	"context"
	"prisma/tms/cmd/tools/tsimulator/task"
	"prisma/tms/iridium"
	"prisma/tms/omnicom"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOmnicom_AddTask(t *testing.T) {
	o := NewOmnicom(*NewObject(), 1)
	assert.Equal(t, 0, o.queueTask.Len())
	o.AddTask(&task.MT{
		Header: iridium.MTHeader{},
		Payload: iridium.MPayload{
			Omn: &omnicom.RMH{},
		},
	})
	assert.Equal(t, 1, o.queueTask.Len())
}

func TestOmnicom_handleMT(t *testing.T) {
	s := NewObject()
	s.Device = "omnicom-vms"
	s.Pos = []PositionArrivalTime{
		{
			PositionSpeed: PositionSpeed{
				Latitude:  0,
				Longitude: 0,
			},
		},
		{
			PositionSpeed: PositionSpeed{
				Latitude:  1,
				Longitude: 1,
			},
		},
	}
	o := NewOmnicom(*s, 1)
	s.Move(s.ReportPeriod)
	o.UpdateInformation(*s)
	o.handleMT(&task.MT{
		Header: iridium.MTHeader{
			IMEI:                  "123456789012345",
			IEI:                   0x01,
			MTflag:                0x01,
			UniqueClientMessageID: "1234",
			//MTHL:                  0x01,
		},
		Payload: iridium.MPayload{
			IEI: 0x01,
			Omn: &omnicom.RMH{
				Date:   omnicom.Date{},
				ID_Msg: 0x02,
				Date_Interval: omnicom.Date_Interval{
					Start: omnicom.Date{
						Year:  0,
						Month: uint32(time.January),
						Day:   1,
					},
					Stop: omnicom.Date{
						Year:  uint32(time.Now().Year()+2) - 2000,
						Month: uint32(time.January),
						Day:   1,
					},
				},
				Header:  2,
				CRC:     2,
				Padding: 2,
			},
		},
	})

	tp := o.object.historyPosition.Value.(*PositionSpeedTime)
	ch := o.GetChannelHandledTask()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	select {
	case <-ctx.Done():
		assert.Fail(t, "did not receive results from a channel")
	case data := <-ch:
		histPos, err := omnicom.Parse(data.Data[6:]) // header is nil
		assert.NoError(t, err)
		histPosHPR := histPos.(*omnicom.HPR)
		assert.Equal(t, uint32(tp.t.Year()-2000), histPosHPR.Data_Report[0].Date_Position.Year)
		assert.Equal(t, uint32(tp.t.Month()), histPosHPR.Data_Report[0].Date_Position.Month)
		assert.Equal(t, uint32(tp.t.Day()), histPosHPR.Data_Report[0].Date_Position.Day)
		assert.InDelta(t, tp.p.Latitude, histPosHPR.Data_Report[0].Date_Position.Latitude, 0.0001)
		assert.InDelta(t, tp.p.Longitude, histPosHPR.Data_Report[0].Date_Position.Longitude, 0.0001)
	}

	o.handleMT(&task.MT{
		Header: iridium.MTHeader{
			IMEI:                  "123456789012345",
			IEI:                   0x01,
			MTflag:                0x01,
			UniqueClientMessageID: "1234",
		},
		Payload: iridium.MPayload{
			IEI: 0x01,
			Omn: &omnicom.UGP{
				ID_Msg: 0x02,
			},
		},
	})
	ctx, cancelq := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelq()
	select {
	case <-ctx.Done():
		assert.Fail(t, "did not receive results from a channel")
	case data := <-ch:
		gp, err := omnicom.Parse(data.Data[6:])
		assert.NoError(t, err)
		stgp := gp.(*omnicom.GP)
		assert.Equal(t, uint32(2), stgp.ID_Msg)
	}
}

func TestOmnicom_SplittingMessages(t *testing.T) {
	s := NewObject()
	s.Device = "omnicom"
	s.Pos = []PositionArrivalTime{
		{},
		{
			 PositionSpeed: PositionSpeed{
				 Longitude: 1,
				 Latitude:  104,
			 },
		},
	}
	for i := 0; i < sizePositionPerSocket+10; i++ {
		s.Move(s.ReportPeriod)
	}
	s.deviceObj.UpdateInformation(*s)
	s.AddTask(&task.MT{
		Payload: iridium.MPayload{
			Omn: &omnicom.RMH{
				Header: 2,
				Date_Interval: omnicom.Date_Interval{
					Start: omnicom.Date{},
					Stop: omnicom.Date{
						Year: 99,
					},
				},
			},
		},
	})
	ch := s.GetChannelHandledTask()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	select {
	case <-ctx.Done():
		assert.Fail(t, "did not receive results from a channel")
	case <-ch:
	}
	select {
	case <-ctx.Done():
		assert.Fail(t, "did not receive results from a channel")
	case res := <-ch:
		// we pass bytes to get payload cause we care about payload only
		hpr, err := omnicom.Parse(res.Data[37:])
		assert.NoError(t, err)
		return
		assert.InDelta(t, s.curPos.Longitude, hpr.(*omnicom.HPR).Data_Report[0].Date_Position.Longitude, 0.001)
		assert.InDelta(t, s.curPos.Latitude, hpr.(*omnicom.HPR).Data_Report[0].Date_Position.Latitude, 0.001)
	}
}

func TestOmnicom_handlingTasks(t *testing.T) {
	s := NewObject()
	s.Device = "omnicom"
	s.Pos = []PositionArrivalTime{
		{},
		{
			PositionSpeed: PositionSpeed{
				Longitude: 1,
				Latitude:  1,
			},
		},
	}
	s.Move(s.ReportPeriod)
	s.deviceObj.UpdateInformation(*s)
	s.AddTask(&task.MT{
		Payload: iridium.MPayload{
			Omn: &omnicom.RMH{
				Header: 2,
				Date_Interval: omnicom.Date_Interval{
					Start: omnicom.Date{},
					Stop: omnicom.Date{
						Year: 99,
					},
				},
			},
		},
	})
	ch := s.GetChannelHandledTask()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	select {
	case <-ctx.Done():
		assert.Fail(t, "did not receive results from a channel")
	case res := <-ch:
		// we pass bytes to get payload cause we care about payload only
		hpr, err := omnicom.Parse(res.Data[37:])
		assert.NoError(t, err)
		assert.InDelta(t, s.curPos.Longitude, hpr.(*omnicom.HPR).Data_Report[0].Date_Position.Longitude, 0.001)
		assert.InDelta(t, s.curPos.Latitude, hpr.(*omnicom.HPR).Data_Report[0].Date_Position.Latitude, 0.001)
	}
}
