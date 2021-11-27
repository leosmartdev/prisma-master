package nmealib

import (
	"prisma/tms"
	"prisma/tms/nmea"
	"prisma/tms/util/ident"

	"testing"
)

var ni *NmeaIdentify

func setup() {
	ident.Clock = func() int64 { return 0 }
	ni = NewNmeaIdentify(&tms.SensorID{
		Site: 0,
		Eid:  0,
	})
}

func teardown() {
	ident.Clock = ident.Now
}

func newTrack(number int, status string) *tms.Track {
	target := &tms.Target{
		Nmea: &nmea.Nmea{
			Ttm: &nmea.Ttm{
				NumberValidity: true,
				Number:         uint32(number),
				StatusValidity: true,
				Status:         status,
			},
		},
	}
	track := &tms.Track{Targets: []*tms.Target{target}}
	return track
}

func genTrackID(seconds int) string {
	return ident.
		With("seconds", seconds).
		With("counter", 1).
		With("site", 0).
		With("eid", 0).
		Hash()
}

func TestNewRadarTarget(t *testing.T) {
	setup()
	defer func() { teardown() }()
	track := newTrack(20, "T")
	want := genTrackID(0)
	got, err := ni.getTtmID(track, track.Targets[0].Nmea.Ttm)
	if want != got {
		t.Fatalf("got %v ; want %v (err %v)", got, want, err)
	}
}

func TestUpdatedRadarTarget(t *testing.T) {
	setup()
	defer func() { teardown() }()
	newTrack(20, "T")
	ident.Clock = func() int64 { return 1 }
	track2 := newTrack(20, "T")
	want := genTrackID(1)
	got, err := ni.getTtmID(track2, track2.Targets[0].Nmea.Ttm)
	if want != got {
		t.Fatalf("got %v ; want %v (err %v)", got, want, err)
	}
}

func TestLostRadarTarget(t *testing.T) {
	setup()
	defer func() { teardown() }()
	newTrack(20, "T")
	ident.Clock = func() int64 { return 1 }
	newTrack(20, "L")
	want := 0
	got := len(ni.ttm_ids)
	if want != got {
		t.Fatalf("got %v targets ; wanted %v", want, got)
	}
}

func TestLostRadarTargetButNotTracking(t *testing.T) {
	setup()
	defer func() { teardown() }()
	track := newTrack(20, "L")
	want := ""
	got, err := ni.getTtmID(track, track.Targets[0].Nmea.Ttm)
	if want != got {
		t.Fatalf("got %v ; want %v (err %v)", got, want, err)
	}
}
