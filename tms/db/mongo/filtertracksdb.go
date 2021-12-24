package mongo

import (
	"fmt"
	"prisma/gogroup"
	"prisma/tms/client_api"
	"prisma/tms/db"
	"prisma/tms/moc"
	"unsafe"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// FilterTracksObjectType ...
const FilterTracksObjectType = "prisma.tms.moc.FilterTracks"

// const FilterTracksObjectType = "prisma.tms.moc.IncidentLogEntry"

// CollectionFilterTracks mongo collection
const CollectionFilterTracks = "filtertracks"

type FilterTracksDb struct {
	group  gogroup.GoGroup
	miscDb db.MiscDB
}

func NewMongoFilterTracksDb(misc db.MiscDB) db.FilterTracksDB {
	client, ok := misc.(*MongoMiscClient)
	if !ok {
		return nil
	}
	return &MongoMiscClient{
		dbconn: client.dbconn,
		ctxt:   client.ctxt,
	}
}

func (filtertracksDb *MongoMiscClient) GetFilterTracks(userId string) ([]*moc.FilterTracks, error) {
	// get filter tracks by user id
	req := db.GoRequest{
		ObjectType: FilterTracksObjectType,
	}
	sd, ti, err := filtertracksDb.resolveTable(&req)
	if err != nil {
		return nil, err
	}
	query := bson.M{
		"me.user": userId,
	}
	raw := []bson.Raw{}
	c := Coder{TypeData: sd}

	filterTracks := make([]*moc.FilterTracks, 0)

	err = filtertracksDb.dbconn.DB().C(ti.Name).Find(query).All(&raw)
	if err != mgo.ErrNotFound {
		for _, data := range raw {
			var obj DBMiscObject
			c.DecodeTo(data, unsafe.Pointer(&obj))
			filterTrack, ok := obj.Obj.(*moc.FilterTracks)
			if !ok {
				return nil, fmt.Errorf("Could not recover filter tracks object")
			}
			filterTracks = append(filterTracks, filterTrack)
		}
	}
	return filterTracks, nil
}

func (filtertracksDb *MongoMiscClient) SaveFilterTrack(filterTracks *moc.FilterTracks) (*client_api.UpsertResponse, error) {
	req := db.GoRequest{
		ObjectType: FilterTracksObjectType,
	}
	sd, ti, err := filtertracksDb.resolveTable(&req)
	if err != nil {
		return nil, err
	}
	query := bson.M{
		"me.user": filterTracks.User,
		"me.type": filterTracks.Type,
	}
	raw := &bson.Raw{}
	c := Coder{TypeData: sd}
	var obj DBMiscObject
	er := filtertracksDb.dbconn.DB().C(ti.Name).Find(query).One(raw)
	if er != mgo.ErrNotFound {
		c.DecodeTo(*raw, unsafe.Pointer(&obj))
		filterTrackObj, ok := obj.Obj.(*moc.FilterTracks)
		if !ok {
			return nil, db.ErrorNotFound
		}
		filterTracks.Id = filterTrackObj.Id
	}
	res, err := filtertracksDb.Upsert(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: FilterTracksObjectType,
			Obj: &db.GoObject{
				ID:   filterTracks.Id,
				Data: filterTracks,
			},
		},
		Ctxt: filtertracksDb.ctxt,
	})
	return res, err
}
