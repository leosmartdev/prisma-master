package main

import (
	"prisma/tms"
	api "prisma/tms/client_api"
	"prisma/tms/db/mockdb"
	"prisma/tms/devices"
	"prisma/tms/tmsg"
	"prisma/tms/util/clock"
	"reflect"
	"testing"
	"time"
)

type extenderFixtures struct {
	s         *trackExtenderStage
	clk       *clock.Mock
	db        *mockdb.TrackExDb
	tsiClient *tmsg.MockTsiClient
}

var extenderTests = []struct {
	name string
	fn   func(*testing.T, extenderFixtures)
}{
	{name: "testStartWatching", fn: testStartWatching},
	{name: "testUpdate", fn: testUpdate},
	{name: "testTimeout", fn: testTimeout},
	{name: "testExtend", fn: testExtend},
	{name: "testExpire", fn: testExpire},
	{name: "testManualNotExpire", fn: testManualNotExpire},
}

func TestExtender(t *testing.T) {
	for _, test := range extenderTests {
		s := newTrackExtenderStage()
		f := extenderFixtures{
			s:         s,
			clk:       &clock.Mock{},
			db:        mockdb.NewTrackExDbStub(),
			tsiClient: tmsg.NewTsiClientStub(),
		}
		s.clk = f.clk
		s.db = f.db
		s.tsiClient = f.tsiClient
		s.initialized = true

		t.Run(test.name, func(t *testing.T) {
			test.fn(t, f)
		})
	}
}

func testStartWatching(t *testing.T, f extenderFixtures) {
	var time0 time.Time
	track := testSarsat("sarsat1", "beacon1")
	f.s.analyze(api.TrackUpdate{
		Status: api.Status_Current,
		Track:  track,
	})
	if len(f.s.watching) != 1 {
		t.Fatal("expecting a watch")
	}

	want := tms.TrackExtension{
		Track:   track,
		Updated: time0,
		Next:    time0.Add(10 * time.Minute),
		Expires: time0.Add(12 * time.Hour),
	}
	have := f.s.watching["sarsat1"]
	if !reflect.DeepEqual(want, have) {
		t.Errorf("\n want %+v \n have %+v", want, have)
	}
}

func testUpdate(t *testing.T, f extenderFixtures) {
	var time0 time.Time
	time1 := time0.Add(5 * time.Minute)

	track := testSarsat("sarsat1", "beacon1")
	f.s.analyze(api.TrackUpdate{
		Status: api.Status_Current,
		Track:  track,
	})
	f.clk.MockNow = time1
	f.s.analyze(api.TrackUpdate{
		Status: api.Status_Current,
		Track:  track,
	})

	want := tms.TrackExtension{
		Track:   track,
		Updated: time1,
		Next:    time1.Add(10 * time.Minute),
		Expires: time1.Add(12 * time.Hour),
	}
	have := f.s.watching["sarsat1"]
	if !reflect.DeepEqual(want, have) {
		t.Errorf("\n want %+v \n have %+v", want, have)
	}

}

func testTimeout(t *testing.T, f extenderFixtures) {
	track := testSarsat("sarsat1", "beacon1")
	f.s.analyze(api.TrackUpdate{
		Status: api.Status_Current,
		Track:  track,
	})
	f.s.analyze(api.TrackUpdate{
		Status: api.Status_Timeout,
		Track:  track,
	})

	if len(f.s.watching) != 0 {
		t.Fatalf("not expecting a watch")
	}
}

func testExtend(t *testing.T, f extenderFixtures) {
	var time0 time.Time
	time1 := time0.Add(12 * time.Minute)
	tmsg.GClient = tmsg.NewTsiClientStub()
	track := testSarsat("sarsat1", "beacon1")
	f.s.analyze(api.TrackUpdate{
		Status: api.Status_Current,
		Track:  track,
	})
	f.clk.MockNow = time1
	f.s.check()

	if len(f.s.watching) != 1 {
		t.Fatal("expecting a watch")
	}

	want := tms.TrackExtension{
		Track:   track,
		Updated: time0,
		Next:    time1.Add(10 * time.Minute),
		Expires: time0.Add(12 * time.Hour),
	}
	have := f.s.watching["sarsat1"]
	if !reflect.DeepEqual(want, have) {
		t.Errorf("\n want %+v \n have %+v", want, have)
	}
}

func testExpire(t *testing.T, f extenderFixtures) {
	var time0 time.Time
	time1 := time0.Add(13 * time.Hour)

	track := testSarsat("sarsat1", "beacon1")
	f.s.analyze(api.TrackUpdate{
		Status: api.Status_Current,
		Track:  track,
	})
	f.clk.MockNow = time1
	f.s.check()

	if len(f.s.watching) != 0 {
		t.Fatal("expecting no watches")
	}
}

func testManualNotExpire(t *testing.T, f extenderFixtures) {
	var time0 time.Time
	time1 := time0.Add(1300 * time.Hour)

	track := testManual("manual")
	f.s.analyze(api.TrackUpdate{
		Status: api.Status_Current,
		Track:  track,
	})
	f.clk.MockNow = time1
	f.s.check()

	if len(f.s.watching) != 1 {
		t.Fatal("expecting a watch")
	}
}

func testManual(id string) *tms.Track {
	return &tms.Track{
		Id: id,
		Targets: []*tms.Target{
			&tms.Target{
				Type: devices.DeviceType_Manual,
			},
		},
	}
}
