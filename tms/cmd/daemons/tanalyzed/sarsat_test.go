package main

import (
	"prisma/gogroup"
	"prisma/tms"
	api "prisma/tms/client_api"
	"prisma/tms/db/mockdb"
	"prisma/tms/devices"
	"prisma/tms/moc"
	"prisma/tms/sar"
	"prisma/tms/util/ident"
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock"
	"github.com/stretchr/testify/assert"
)

func testSarsat(id string, beaconID string) *tms.Track {
	return &tms.Track{
		Id: id,
		Targets: []*tms.Target{
			&tms.Target{
				Type: devices.DeviceType_SARSAT,
				Sarmsg: &sar.SarsatMessage{
					SarsatAlert: &sar.SarsatAlert{
						Beacon: &sar.Beacon{
							HexId: beaconID,
						},
					},
				},
			},
		},
	}
}

func TestDefaultSarsatMessageNotice(t *testing.T) {

	ctxt := gogroup.New(nil, "")
	redisMock := redigomock.NewConn()
	redisPool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redisMock, nil
		},
	}
	n := NewNotifier(ctxt, nil, redisPool)
	n.notifydb = mockdb.NewNotifyDbStub()
	a := newSarsatStage(n, ctxt, nil)
	id := &tms.TargetID{
		Producer: &tms.SensorID{
			Site: 10,
			Eid:  10,
		},
		SerialNumber: &tms.TargetID_TimeSerial{&tms.TimeSerialNumber{
			Seconds: 99,
			Counter: 1,
		}},
	}
	update := api.TrackUpdate{
		Status: api.Status_Current,
		Track: &tms.Track{
			Id: "UNKNOWN",
			Targets: []*tms.Target{
				&tms.Target{
					Id:   id,
					Type: devices.DeviceType_SARSAT,
					Sarmsg: &sar.SarsatMessage{
						MessageType: sar.SarsatMessage_UNKNOWN,
					},
				},
			},
		},
	}

	hmsetCmd := redisMock.Command("HMSET")
	hmsetCmd.Expect("OK")
	a.analyze(update)
	assert.Equal(t, 1, redisMock.Stats(hmsetCmd))
}

func TestSarsatNotice(t *testing.T) {
	redisMock := redigomock.NewConn()
	ctxt := gogroup.New(nil, "")
	redisPool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redisMock, nil
		},
	}
	n := NewNotifier(ctxt, nil, redisPool)
	n.notifydb = mockdb.NewNotifyDbStub()
	a := newSarsatStage(n, ctxt, nil )

	update := api.TrackUpdate{
		Status: api.Status_Current,
		Track:  testSarsat("sarsat1", "1234"),
	}

	hmsetCmd := redisMock.Command("HMSET")
	hmsetCmd.Expect("OK")
	a.analyze(update)
	assert.Equal(t, 1, redisMock.Stats(hmsetCmd))
}

func TestNotSarsatNotice(t *testing.T) {
	redisMock := redigomock.NewConn()
	ctxt := gogroup.New(nil, "")
	redisPool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redisMock, nil
		},
	}
	n := NewNotifier(ctxt, nil, redisPool)
	n.notifydb = mockdb.NewNotifyDbStub()
	a := newSartStage(n)

	update := api.TrackUpdate{
		Status: api.Status_Current,
		Track:  testSarsat("sarsat1", "1234"),
	}

	hmsetCmd := redisMock.Command("HMSET")
	hmsetCmd.Expect("OK")
	a.analyze(update)
	assert.Equal(t, 0, redisMock.Stats(hmsetCmd))
}

func TestNotSarsatNoticeTimeout(t *testing.T) {
	redisMock := redigomock.NewConn()
	ctxt := gogroup.New(nil, "")
	redisPool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redisMock, nil
		},
	}
	n := NewNotifier(ctxt, nil, redisPool)
	n.notifydb = mockdb.NewNotifyDbStub()
	a := newSarsatStage(n, ctxt, nil )

	update := api.TrackUpdate{
		Status: api.Status_Current,
		Track:  testSarsat("sarsat1", "1234"),
	}
	hmsetCmd := redisMock.Command("HMSET")
	hmsetCmd.Expect("OK")

	a.analyze(update)
	assert.Equal(t, 1, redisMock.Stats(hmsetCmd))

	update = api.TrackUpdate{
		Status: api.Status_Timeout,
		Track:  testSarsat("sarsat1", "1234"),
	}
	sarsat1 := ident.With("event", moc.Notice_Sarsat).With("beacon", "1234").Hash()
	note := &moc.Notice{
		NoticeId: sarsat1,
		Action:   moc.Notice_ACK,
	}
	redisMock.Command("HGET", getRedisKeyByNote(note), "data").Expect(note.String())
	hdelCmd := redisMock.Command("DEL")
	hdelCmd.Expect("OK")
	a.analyze(update)
	assert.Equal(t, 1, redisMock.Stats(hdelCmd))
}
