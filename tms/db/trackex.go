package db

import (
	"prisma/tms"

	"github.com/globalsign/mgo/bson"
)

// TrackExDb interface implements track ext queries
type TrackExDb interface {
	Get() ([]tms.TrackExtension, error)
	Insert(tms.TrackExtension) error
	Remove(string) error
	Startup() (int, error)
	Update(tms.TrackExtension) error
	GetTrack(filter bson.M) (*tms.Track, error)
}
