package main

import (
	"testing"
	"prisma/gogroup"
	"prisma/tms"
	api "prisma/tms/client_api"
	"prisma/tms/db/mockdb"
	"prisma/tms/nmea"
	"prisma/tms/moc"
	"prisma/tms/util/ident"
	"github.com/rafaeljusto/redigomock"
	"github.com/stretchr/testify/assert"
	"github.com/gomodule/redigo/redis"
)

func testAis(id string, mmsi uint32) *tms.Track {
	return &tms.Track{
		Id: id,
		Targets: []*tms.Target{
			&tms.Target{
				Nmea: &nmea.Nmea{
					Vdm: &nmea.Vdm{
						M1371: &nmea.M1371{
							Mmsi: mmsi,
						},
					},
				},
			},
		},
	}
}
func TestSartNotice(t *testing.T) {
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
		Track:  testAis("sart1", 970000000),
	}

	hmsetCmd := redisMock.Command("HMSET")
	hmsetCmd.Expect(true)
	a.analyze(update)
	assert.Equal(t, 1, redisMock.Stats(hmsetCmd))
}

func TestNotSartNotice(t *testing.T) {
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
		Track:  testAis("sart1", 120000000),
	}
	hmsetCmd := redisMock.Command("HMSET")
	hmsetCmd.Expect(true)
	a.analyze(update)
	assert.Equal(t, 0, redisMock.Stats(hmsetCmd))
}

func TestNotSartNoticeTimeout(t *testing.T) {
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
		Track:  testAis("sart1", 970000000),
	}

	hmsetCmd := redisMock.Command("HMSET")
	hmsetCmd.Expect(true)
	a.analyze(update)
	assert.Equal(t, 1, redisMock.Stats(hmsetCmd))

	update = api.TrackUpdate{
		Status: api.Status_Timeout,
		Track:  testAis("sart1", 970000000),
	}
	sart1 := ident.With("event", moc.Notice_Sart).With("mmsi", "970000000").Hash()
	note := &moc.Notice{
		NoticeId: sart1,
		Action:   moc.Notice_ACK,
	}
	redisMock.Command("HGET", getRedisKeyByNote(note), "data").Expect(note.String())
	hdelCmd := redisMock.Command("DEL")
	hdelCmd.Expect("OK")
	a.analyze(update)
	assert.Equal(t, 1, redisMock.Stats(hdelCmd))
}
