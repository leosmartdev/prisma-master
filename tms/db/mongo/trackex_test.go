package mongo

import (
	"context"
	"prisma/gogroup"
	"reflect"
	"testing"
	"time"

	"prisma/tms"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

func TestGetTrack(t *testing.T) {
	ctx := gogroup.New(context.Background(), "trackex_test")
	data, err := mgo.ParseURL("localhost:27017")
	session, err := mgo.DialWithTimeout("localhost:27017", 2*time.Second)
	if err != nil && err.Error() == "no reachable servers" {
		t.Skip("run this test when mongod is up")
		return
	}
	session.Close()
	data.Timeout = 1 * time.Second
	mClient, err := NewMongoClient(ctx, data, nil)
	if err != nil {
		t.Error("can not ininitalize new mongo client")
	}
	tex := NewTrackExDb(ctx, mClient)
	regID := bson.NewObjectId().Hex()
	trackID := bson.NewObjectId().Hex()
	TrackEx := tms.TrackExtension{
		Track: &tms.Track{
			RegistryId: regID,
			Id:         trackID,
		},
		Updated: time.Now(),
		Expires: time.Now().Add(100 * time.Minute),
		Next:    time.Now().Add(10 * time.Minute),
		Count:   2,
	}
	err = tex.Insert(TrackEx)
	if err != nil {
		t.Errorf("can not insert TrackEx: %+v", err)
	}
	track, err := tex.GetTrack(bson.M{"Track.registry_id": regID})
	if err != nil {
		t.Errorf("Can not recover track: %+v", err)
	}
	ok := reflect.DeepEqual(TrackEx.Track, track)
	if !ok {
		t.Errorf("Inserted and recovered track are not identical, %+v != %+v", TrackEx.Track, track)
	}
	err = tex.Remove(TrackEx.Track.Id)
	if err != nil {
		t.Errorf("Could not remove trackex document with track id %+v: %+v", TrackEx.Track.Id, err)
	}
}
