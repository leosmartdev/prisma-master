package main

import (
	"prisma/tms"
	api "prisma/tms/client_api"
	"prisma/tms/db"
	"prisma/tms/geojson/rtree"
	"prisma/tms/log"
	"prisma/tms/moc"
	"testing"

	"github.com/json-iterator/go/assert"
	"github.com/stretchr/testify/mock"
)

type mockNotifier struct {
	mock.Mock
}

func (m *mockNotifier) Notify(note *moc.Notice, ruleMatch bool) error {
	args := m.Called(note, ruleMatch)
	return args.Error(0)
}

func (m *mockNotifier) TrackMap() map[string]*tms.Track {
	args := m.Called()
	return args.Get(0).(map[string]*tms.Track)
}

func (m *mockNotifier) TrackTree() *rtree.RTree {
	args := m.Called()
	return args.Get(0).(*rtree.RTree)
}

type zoneFixtures struct {
	n         *mockNotifier
	s         *zoneStage
	trackTree *rtree.RTree
}

func testTrack(id string, x float64, y float64) (*tms.Track, api.TrackUpdate) {
	track := &tms.Track{
		Id: id,
		Targets: []*tms.Target{
			{
				Position: &tms.Point{
					Latitude:  y,
					Longitude: x,
				},
			},
		},
	}
	update := api.TrackUpdate{
		Status: api.Status_Current,
		Track:  track,
	}
	return track, update
}

func testZone(id string, x1, y1, x2, y2 float64) (*moc.Zone, db.GoGetResponse) {
	zone := &moc.Zone{
		Poly: &tms.Polygon{
			Lines: []*tms.LineString{
				&tms.LineString{
					Points: []*tms.Point{
						&tms.Point{
							Latitude:  y1,
							Longitude: x1,
						},
						&tms.Point{
							Latitude:  y2,
							Longitude: x1,
						},
						&tms.Point{
							Latitude:  y2,
							Longitude: x2,
						},
						&tms.Point{
							Latitude:  y1,
							Longitude: x2,
						},
						&tms.Point{
							Latitude:  y1,
							Longitude: x1,
						},
					},
				},
			},
		},
	}

	update := db.GoGetResponse{
		Status: api.Status_Current,
		Contents: &db.GoObject{
			ID:   id,
			Data: zone,
		},
	}
	return zone, update
}

var zoneTests = []struct {
	name string
	fn   func(*testing.T, *zoneFixtures)
}{
	{name: "EnterZone", fn: testEnterZone},
	{name: "EnterZoneNoNotice", fn: testEnterZoneNoNotice},
	{name: "ExitZone", fn: testExitZone},
	{name: "ExitZoneNoNotice", fn: testExitZoneNoNotice},
	{name: "NewZone", fn: testNewZone},
	{name: "NewEmptyZone", fn: testNewZoneEmpty},
	{name: "ZoneUpdate", fn: testZoneUpdate},
	{name: "ZoneArea", fn: testUpdatePoliesByTrack},
}

func TestZoneStage(t *testing.T) {
	for _, test := range zoneTests {
		t.Run(test.name, func(t *testing.T) {
			f := &zoneFixtures{}
			f.trackTree = rtree.New()
			f.n = &mockNotifier{}
			f.n.On("TrackTree").Return(f.trackTree)
			f.n.On("Notify", mock.AnythingOfType("*moc.Notice"), mock.AnythingOfType("bool")).Return(nil)
			f.s = newZoneStage(f.n)
			f.s.zones = make(map[string]*moc.Zone)
			f.s.members = make(map[string]map[string]*tms.Track)
			f.s.tracer = log.GetTracer("zones")
			f.s.initialized = true
			test.fn(t, f)
		})
	}
}

func testEnterZone(t *testing.T, f *zoneFixtures) {
	zone1, uzone1 := testZone("z1", 10, 10, 20, 20)
	_, utrack1 := testTrack("t1", 15, 15)
	zone1.CreateAlertOnEnter = true
	f.s.updateZone(uzone1)
	f.s.analyze(utrack1)
	f.n.AssertCalled(t, "Notify", mock.AnythingOfType("*moc.Notice"), true)
}

func testEnterZoneNoNotice(t *testing.T, f *zoneFixtures) {
	zone1, uzone1 := testZone("z1", 10, 10, 20, 20)
	_, utrack1 := testTrack("t1", 15, 15)
	zone1.CreateAlertOnEnter = false
	f.s.updateZone(uzone1)
	f.s.analyze(utrack1)
	f.n.AssertNotCalled(t, "Notify", mock.AnythingOfType("*moc.Notice"), true)
}

func testExitZone(t *testing.T, f *zoneFixtures) {
	zone1, uzone1 := testZone("z1", 10, 10, 20, 20)
	_, utrack1 := testTrack("t1", 15, 15)
	zone1.CreateAlertOnExit = true
	f.s.updateZone(uzone1)
	f.s.analyze(utrack1)
	f.n.AssertNotCalled(t, "Notify", mock.AnythingOfType("*moc.Notice"), true)

	_, utrack1 = testTrack("t1", 25, 25)
	f.s.analyze(utrack1)
	f.n.AssertCalled(t, "Notify", mock.AnythingOfType("*moc.Notice"), true)
}

func testExitZoneNoNotice(t *testing.T, f *zoneFixtures) {
	zone1, uzone1 := testZone("z1", 10, 10, 20, 20)
	_, utrack1 := testTrack("t1", 15, 15)
	zone1.CreateAlertOnExit = false
	f.s.updateZone(uzone1)
	f.s.analyze(utrack1)
	f.n.AssertNotCalled(t, "Notify", mock.AnythingOfType("*moc.Notice"), true)

	_, utrack1 = testTrack("t1", 25, 25)
	f.s.analyze(utrack1)
	f.n.AssertNotCalled(t, "Notify", mock.AnythingOfType("*moc.Notice"), true)
}

func testNewZone(t *testing.T, f *zoneFixtures) {
	zone1, uzone1 := testZone("z1", 10, 10, 20, 20)
	track1, _ := testTrack("t1", 15, 15)
	zone1.CreateAlertOnEnter = true
	f.trackTree.Insert(track1)
	f.s.updateZone(uzone1)
	f.n.AssertCalled(t, "Notify", mock.AnythingOfType("*moc.Notice"), true)
}

func testUpdatePoliesByTrack(t *testing.T, f *zoneFixtures) {
	zone1, uzone1 := testZone("z1", 10, 10, 20, 20)
	track1, utrack1 := testTrack("test1", 15, 15)
	zone1.CreateAlertOnEnter = true
	f.trackTree.Insert(track1)
	uzone1.Contents.Data.(*moc.Zone).Area = &moc.Area{
		TrackId: "test1",
		Radius:  0.5,
	}
	f.s.updateZone(uzone1)
	assert.Contains(t, f.s.relAreaZone[track1.Id], uzone1.Contents.ID)
	f.s.analyze(utrack1)
}

func testNewZoneEmpty(t *testing.T, f *zoneFixtures) {
	zone1, uzone1 := testZone("z1", 10, 10, 20, 20)
	track1, _ := testTrack("t1", 25, 25)
	zone1.CreateAlertOnEnter = true
	f.trackTree.Insert(track1)
	f.s.updateZone(uzone1)
	f.n.AssertNotCalled(t, "Notify", mock.AnythingOfType("*moc.Notice"), true)
}

func testZoneUpdate(t *testing.T, f *zoneFixtures) {
	zone1, uzone1 := testZone("z1", 10, 10, 20, 20)
	track1, _ := testTrack("t1", 25, 25)
	zone1.CreateAlertOnEnter = true
	f.trackTree.Insert(track1)
	f.s.updateZone(uzone1)
	f.n.AssertNotCalled(t, "Notify", mock.AnythingOfType("*moc.Notice"), true)

	zone1, uzone1 = testZone("z1", 20, 20, 30, 30)
	zone1.CreateAlertOnEnter = true
	f.s.updateZone(uzone1)
	f.n.AssertCalled(t, "Notify", mock.AnythingOfType("*moc.Notice"), true)
}
