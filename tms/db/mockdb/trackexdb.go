package mockdb

import (
	"prisma/tms"

	"github.com/globalsign/mgo/bson"
	"github.com/stretchr/testify/mock"
)

type TrackExDb struct {
	mock.Mock
}

func (d *TrackExDb) Get() ([]tms.TrackExtension, error) {
	args := d.Called()
	return args.Get(0).([]tms.TrackExtension), args.Error(1)
}

func (d *TrackExDb) Insert(t tms.TrackExtension) error {
	args := d.Called(t)
	return args.Error(0)
}

func (d *TrackExDb) Startup() (int, error) {
	args := d.Called()
	return 0, args.Error(0)
}

func (d *TrackExDb) Update(t tms.TrackExtension) error {
	args := d.Called(t)
	return args.Error(0)
}

func (d *TrackExDb) Remove(trackID string) error {
	args := d.Called(trackID)
	return args.Error(0)
}
func (d *TrackExDb) GetTrack(filter bson.M) (*tms.Track, error) {
	args := d.Called(filter)
	return nil, args.Error(0)
}

func NewTrackExDbStub() *TrackExDb {
	db := &TrackExDb{}
	db.On("Insert", mock.Anything).Return(nil)
	db.On("Remove", mock.Anything).Return(nil)
	db.On("Update", mock.Anything).Return(nil)
	db.On("GetTrack", mock.Anything).Return(nil)
	return db
}
