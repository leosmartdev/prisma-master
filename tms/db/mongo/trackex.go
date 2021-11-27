package mongo

import (
	"prisma/gogroup"
	"prisma/tms"
	"prisma/tms/db"
	"time"

	"github.com/golang/protobuf/jsonpb"

	"reflect"
	"unsafe"

	"github.com/globalsign/mgo/bson"
)

type MongoTrackExDb struct {
	dbconn *MongoClient
	ctxt   gogroup.GoGroup
	m      *jsonpb.Marshaler
}

func NewTrackExDb(ctxt gogroup.GoGroup, client *MongoClient) db.TrackExDb {
	d := &MongoTrackExDb{
		dbconn: client,
		ctxt:   ctxt,
		m:      &jsonpb.Marshaler{},
	}
	return d
}

func (d *MongoTrackExDb) Insert(ex tms.TrackExtension) error {
	db := d.dbconn.DB()
	defer d.dbconn.Release(db)
	dbex, err := ex.Db()
	if err != nil {
		return err
	}
	c := Coder{TypeData: NewStructData(
		reflect.TypeOf(tms.TrackExtension{}),
		NoMap)}
	raw := c.Encode(unsafe.Pointer(dbex))
	return db.C("trackex").Insert(raw)
}

func (d *MongoTrackExDb) Update(ex tms.TrackExtension) error {
	db := d.dbconn.DB()
	defer d.dbconn.Release(db)
	dbex, err := ex.Db()
	if err != nil {
		return err
	}
	c := Coder{TypeData: NewStructData(
		reflect.TypeOf(tms.TrackExtension{}),
		NoMap)}
	raw := c.Encode(unsafe.Pointer(dbex))
	_, err = db.C("trackex").Upsert(bson.M{"Track.id": ex.Track.Id}, raw)
	return err
}

func (d *MongoTrackExDb) Remove(trackID string) error {
	db := d.dbconn.DB()
	defer d.dbconn.Release(db)
	_, err := db.C("trackex").RemoveAll(bson.M{"Track.id": trackID})
	return err
}

// Startup run when we lunch the extender in order to clean up old expired tracks
// the function returns the number of removed tracks or an error
func (d *MongoTrackExDb) Startup() (int, error) {
	db := d.dbconn.DB()
	defer d.dbconn.Release(db)
	// delete all the tracks that expired at start up
	// expired track have expiration time in the past
	// and count is less than 1.
	changes, err := db.C("trackex").RemoveAll(
		bson.M{
			"Expires": bson.M{
				"$lt": time.Now(),
			},
			"Count": bson.M{
				"$lt": 1,
			},
		})

	return changes.Removed, err
}

// Get will load all trackex documents
func (d *MongoTrackExDb) Get() ([]tms.TrackExtension, error) {
	db := d.dbconn.DB()
	defer d.dbconn.Release(db)
	results := make([]tms.TrackExtension, 0)
	iter := db.C("trackex").Find(nil).Iter()
	var raw bson.Raw
	for iter.Next(&raw) {
		c := Coder{TypeData: NewStructData(
			reflect.TypeOf(tms.TrackExtension{}),
			NoMap)}
		resp := new(tms.TrackExtension)
		c.DecodeTo(raw, unsafe.Pointer(resp))
		results = append(results, *resp)
	}
	return results, iter.Close()

}

// GetTrack takes a bson filter and return a track after running the filter against trackex collection
func (d *MongoTrackExDb) GetTrack(filter bson.M) (*tms.Track, error) {
	db := d.dbconn.DB()
	defer d.dbconn.Release(db)
	raw := new(bson.Raw)
	c := Coder{TypeData: NewStructData(reflect.TypeOf(tms.TrackExtension{}), NoMap)}
	err := db.C("trackex").Find(filter).One(&raw)
	if err != nil {
		return nil, err
	}
	resp := new(tms.TrackExtension)
	c.DecodeTo(*raw, unsafe.Pointer(resp))
	return resp.Track, err
}
