package main

import (
	"fmt"
	"io"
	"testing"
	"time"

	"prisma/gogroup"
	"prisma/tms"
	api "prisma/tms/client_api"
	"prisma/tms/db"
	"prisma/tms/db/mockdb"
	"prisma/tms/geojson/rtree"
	"prisma/tms/moc"
	"prisma/tms/util/clock"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/gomodule/redigo/redis"
	"github.com/json-iterator/go/assert"
	"github.com/rafaeljusto/redigomock"
)

var _ = fmt.Println

type notifierFixtures struct {
	ctxt    gogroup.GoGroup
	n       *Notices
	mdb     *mockdb.NotifyDb
	redisDB *redigomock.Conn
	clock   *clock.Mock
	time1   time.Time
	time2   time.Time

	ptime1 *timestamp.Timestamp
	ptime2 *timestamp.Timestamp
}

var tests = []struct {
	name string
	fn   func(*testing.T, *notifierFixtures)
}{
	{name: "NoticeClearToNew", fn: noticeClearToNew},
	{name: "NoticeNewHit", fn: noticeNewHit},
	{name: "NoticeNewMiss", fn: noticeNewMiss},
	{name: "NoticeNewToAck", fn: noticeNewToAck},
	{name: "NoticeAckHit", fn: noticeAckHit},
	{name: "NoticeAckMiss", fn: noticeAckMiss},
	{name: "NoticeReap", fn: noticeReap},
}

func TestNotifier(t *testing.T) {
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			redisMock := redigomock.NewConn()
			ctxt := gogroup.New(nil, "")
			redisPool := &redis.Pool{
				Dial: func() (redis.Conn, error) {
					return redisMock, nil
				},
			}
			f := &notifierFixtures{
				ctxt:    ctxt,
				n:       NewNotifier(ctxt, nil, redisPool),
				mdb:     mockdb.NewNotifyDbStub(),
				clock:   &clock.Mock{},
				time1:   time.Unix(1, 0),
				time2:   time.Unix(2, 0),
				redisDB: redisMock,
			}
			f.n.notifydb = f.mdb
			f.n.clk = f.clock
			f.clock.MockNow = f.time1

			ptime1, _ := ptypes.TimestampProto(f.time1)
			ptime2, _ := ptypes.TimestampProto(f.time2)
			f.ptime1 = ptime1
			f.ptime2 = ptime2

			test.fn(t, f)
			ctxt.Cancel(io.EOF)
		})
	}
}

func noticeClearToNew(t *testing.T, f *notifierFixtures) {
	note := &moc.Notice{DatabaseId: "testId1", NoticeId: "test1"}
	f.mdb.On("Create", note).Return(nil)
	noteCopy := *note
	noteCopy.CreatedTime = &timestamp.Timestamp{Seconds: 1}
	noteCopy.UpdatedTime = &timestamp.Timestamp{Seconds: 1}
	noteCopy.Action = moc.Notice_NEW
	hmsetCmd := f.redisDB.Command("HMSET", getRedisKeyByNote(note), "data", noteCopy.String(), "count_updates", 0)
	hmsetCmd.Expect("OK")

	assert.NoError(t, f.n.Notify(note, true))
	assert.True(t, hmsetCmd.Called)
	want := moc.Notice_NEW
	have := note.Action
	if want != have {
		t.Errorf("\n want %v \n have %v", want, have)
	}

	want2, _ := ptypes.TimestampProto(time.Unix(1, 0))
	have2 := note.CreatedTime
	if want2.Seconds != have2.Seconds {
		t.Errorf("\n want %v \n have %v", want2, have2)
	}

	want3, _ := ptypes.TimestampProto(time.Unix(1, 0))
	have3 := note.UpdatedTime
	if want3.Seconds != have3.Seconds {
		t.Errorf("\n want %v \n have %v", want3, have3)
	}

	f.mdb.AssertCalled(t, "Create", note)
}

func noticeNewHit(t *testing.T, f *notifierFixtures) {
	note := &moc.Notice{DatabaseId: "db1", NoticeId: "test1"}
	f.mdb.On("Create", note).Return(nil)
	f.mdb.On("UpdateTime", "db1", f.ptime2).Return(nil)

	noteCopy := *note
	noteCopy.CreatedTime = &timestamp.Timestamp{Seconds: 1}
	noteCopy.UpdatedTime = &timestamp.Timestamp{Seconds: 1}
	noteCopy.Action = moc.Notice_NEW
	f.redisDB.Command("HMSET", getRedisKeyByNote(note), "data", noteCopy.String(), "count_updates", 0).Expect("OK")
	assert.NoError(t, f.n.Notify(note, true))

	f.redisDB.Command("HGET", getRedisKeyByNote(note), "data").Expect(noteCopy.String()).Expect(noteCopy.String())
	f.clock.MockNow = f.time2
	noteCopy.UpdatedTime.Seconds = 2
	updateTimeCommand := f.redisDB.Command("HSET", getRedisKeyByNote(note), "data", noteCopy.String())
	updateTimeCommand.Expect("OK")
	f.redisDB.Command("HINCRBY").Expect(int64(1))

	assert.NoError(t, f.n.Notify(note, true))
	{
		want := moc.Notice_NEW
		have := note.Action
		if want != have {
			t.Errorf("\n want %v \n have %v", want, have)
		}
	}
	{
		want := f.ptime1
		have := note.CreatedTime
		if want.Seconds != have.Seconds {
			t.Errorf("\n want %v \n have %v", want, have)
		}
	}
	assert.Equal(t, 2, f.redisDB.Stats(updateTimeCommand))
}

func noticeNewMiss(t *testing.T, f *notifierFixtures) {
	note := &moc.Notice{DatabaseId: "db1", NoticeId: "test1"}
	f.mdb.On("Create", note).Return(nil)
	f.mdb.On("AckWait", "db1", f.ptime2).Return(nil)
	f.mdb.On("UpdateTime", "db1", f.ptime2).Return(nil)

	noteCopy := *note
	noteCopy.CreatedTime = &timestamp.Timestamp{Seconds: 1}
	noteCopy.UpdatedTime = &timestamp.Timestamp{Seconds: 1}
	noteCopy.Action = moc.Notice_NEW
	hmsetCmd := f.redisDB.Command("HMSET", getRedisKeyByNote(note), "data", noteCopy.String(), "count_updates", 0)
	hmsetCmd.Expect("OK")
	assert.NoError(t, f.n.Notify(note, true))

	f.redisDB.Command("HGET", getRedisKeyByNote(note), "data").Expect(noteCopy.String())

	f.clock.MockNow = f.time2
	noteCopy.UpdatedTime.Seconds = int64(f.time2.Second())
	noteCopy.Action = moc.Notice_ACK_WAIT
	updateAck := f.redisDB.Command("HSET", getRedisKeyByNote(note), "data", noteCopy.String())
	updateAck.Expect("OK")

	assert.NoError(t, f.n.Notify(note, false))

	assert.Equal(t, 2, f.redisDB.Stats(updateAck))
	{
		want := f.ptime1
		have := note.CreatedTime
		if want.Seconds != have.Seconds {
			t.Errorf("\n want %v \n have %v", want, have)
		}
	}
	assert.True(t, hmsetCmd.Called)

}

func noticeNewToAck(t *testing.T, f *notifierFixtures) {
	stream := db.NewNoticeStream()
	f.mdb.On("GetPersistentStream").Return(stream, nil)

	note := &moc.Notice{DatabaseId: "db1", NoticeId: "test1"}
	f.mdb.On("Create", note).Return(nil)
	f.mdb.On("UpdateTime", "db1", f.ptime2).Return(nil)

	stream.InitialLoadDone <- true
	f.n.Init()
	f.n.Start()
	defer f.ctxt.Cancel(nil)

	noteCopy := *note
	noteCopy.CreatedTime = &timestamp.Timestamp{Seconds: 1}
	noteCopy.UpdatedTime = &timestamp.Timestamp{Seconds: 1}
	noteCopy.Action = moc.Notice_NEW
	redisData := map[string]interface{}{
		"data":          noteCopy.String(),
		"count_updates": 0,
	}
	f.redisDB.Command("HMSET", getRedisKeyByNote(note), redisData).Expect("OK")

	f.n.Notify(note, true)
	ack := &moc.Notice{
		DatabaseId: "db1",
		NoticeId:   "test1",
		Action:     moc.Notice_ACK,
		AckTime:    f.ptime2,
	}
	f.redisDB.Command("HGET", getRedisKeyByNote(note), "data").Expect(note.String())
	commandUpdateNotice := f.redisDB.Command("HSET", getRedisKeyByNote(note), "data", ack.String())
	commandUpdateNotice.Expect("OK")
	stream.Updates <- ack
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 1, f.redisDB.Stats(commandUpdateNotice))
}

func noticeAckHit(t *testing.T, f *notifierFixtures) {
	note := &moc.Notice{DatabaseId: "db1", NoticeId: "test1"}
	prev := &moc.Notice{DatabaseId: "db1", NoticeId: "test1", Action: moc.Notice_ACK}

	f.mdb.On("UpdateTime", "db1", f.ptime2).Return(nil)
	f.redisDB.Command("HGET", getRedisKeyByNote(prev), "data").Expect(prev.String())
	f.redisDB.Command("EXISTS", getRedisKeyByNote(prev)).Expect(true)
	f.redisDB.Command("HINCRBY", getRedisKeyByNote(prev), "count_updates", 1).Expect(int64(1))
	copyNote := *prev
	copyNote.UpdatedTime = f.ptime2
	updateAckNote := f.redisDB.Command("HSET", getRedisKeyByNote(&copyNote), "data", copyNote.String())
	updateAckNote.Expect("OK")

	f.clock.MockNow = f.time2
	f.n.Notify(note, true)

	assert.Equal(t, 2, f.redisDB.Stats(updateAckNote))
}

func noticeAckMiss(t *testing.T, f *notifierFixtures) {
	note := &moc.Notice{DatabaseId: "db1", NoticeId: "test1"}
	prev := &moc.Notice{DatabaseId: "db1", NoticeId: "test1", Action: moc.Notice_ACK}

	f.redisDB.Command("HGET", getRedisKeyByNote(prev), "data").Expect(prev.String())
	f.mdb.On("Clear", "db1", f.ptime2).Return(nil)
	delCommand := f.redisDB.Command("DEL", getRedisKeyByNote(prev))
	delCommand.Expect("OK")

	f.clock.MockNow = f.time2
	f.n.Notify(note, false)

	assert.Equal(t, 1, f.redisDB.Stats(delCommand))

}

func noticeReap(t *testing.T, f *notifierFixtures) {
	stream := db.NewNoticeStream()
	f.mdb.On("GetPersistentStream").Return(stream, nil)

	now := time.Now()
	f.n.clk = &clock.Real{}
	pnow, _ := ptypes.TimestampProto(now)
	item := &moc.Notice{
		DatabaseId:  "db1",
		NoticeId:    "test1",
		Action:      moc.Notice_ACK,
		UpdatedTime: pnow,
	}

	f.mdb.On("TimeoutBySliceId", []string{item.DatabaseId}).Return(1, nil)
	f.n.reapTicker = time.NewTicker(100 * time.Millisecond)
	f.n.ackTimeout = -50 * time.Millisecond
	f.redisDB.Command("HGET", getRedisKeyByNote(item), "data").Expect(item.String())
	scanCmd := f.redisDB.Command("SCAN", 0, "MATCH", redisNotifierName+":*")
	scanCmd.
		Expect([]interface{}{int64(0), []interface{}{getRedisKeyByNote(item)}}).
		Expect([]interface{}{int64(0), []interface{}{}})
	delCmd := f.redisDB.Command("DEL", getRedisKeyByNote(item))
	delCmd.Expect("OK")

	stream.InitialLoadDone <- true
	f.n.Init()
	f.n.Start()
	defer f.ctxt.Cancel(io.EOF)
	time.Sleep(1 * time.Second)
	assert.True(t, scanCmd.Called)
	time.Sleep(150 * time.Millisecond)
	assert.Equal(t, 1, f.redisDB.Stats(delCmd))
}

func testMakeTrack(id string, x float64, y float64) *tms.Track {
	return &tms.Track{
		Id: id,
		Targets: []*tms.Target{
			&tms.Target{
				Position: &tms.Point{
					Latitude:  y,
					Longitude: x,
				},
			},
		},
	}
}

func TestRTree(t *testing.T) {
	n := NewNotifier(nil, nil, nil)

	n.updateTrack(api.TrackUpdate{
		Status: api.Status_Current,
		Track:  testMakeTrack("t1", 10, 10),
	})
	n.updateTrack(api.TrackUpdate{
		Status: api.Status_Current,
		Track:  testMakeTrack("t2", -10, -10),
	})
	n.updateTrack(api.TrackUpdate{
		Status: api.Status_Current,
		Track:  testMakeTrack("t1", 15, 15),
	})

	{
		want := 2
		have := n.TrackTree().Count()
		if want != have {
			t.Errorf("\n want %v \n have %v", want, have)
		}
	}

	items := make([]rtree.Item, 0)
	n.TrackTree().Search(5, 5, 0, 20, 20, 0, func(item rtree.Item) bool {
		items = append(items, item)
		return true
	})
	{
		want := 1
		have := len(items)
		if want != have {
			t.Errorf("\n want %v \n have %v", want, have)
		}
	}

	{
		want := "t1"
		have := items[0].(*tms.Track).Id
		if want != have {
			t.Errorf("\n want %v \n have %v", want, have)
		}
	}

}
