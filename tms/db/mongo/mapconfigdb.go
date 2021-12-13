package mongo

import (
	"prisma/gogroup"
	"prisma/tms/client_api"
	"prisma/tms/db"
	"prisma/tms/moc"
	"unsafe"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// MapConfigObjectType ...
const MapConfigObjectType = "prisma.tms.moc.MapConfig"

// const MapConfigObjectType = "prisma.tms.moc.IncidentLogEntry"

// CollectionMapConfig mongo collection
const CollectionMapConfig = "mapconfig"

type MapConfigDb struct {
	group  gogroup.GoGroup
	miscDb db.MiscDB
}

func NewMongoMapConfigDb(misc db.MiscDB) db.MapConfigDB {
	client, ok := misc.(*MongoMiscClient)
	if !ok {
		return nil
	}
	return &MongoMiscClient{
		dbconn: client.dbconn,
		ctxt:   client.ctxt,
	}
}

func (mapconfigDb *MongoMiscClient) FindAllMapConfig() ([]db.GoGetResponse, error) {
	res, err := mapconfigDb.Get(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: MapConfigObjectType,
		},
		Ctxt: mapconfigDb.ctxt,
		Time: &db.TimeKeeper{},
	})

	return res, err
}

func (mapconfigDb *MongoMiscClient) SaveMapConfig(mapconfig *moc.MapConfig) (*client_api.UpsertResponse, error) {
	req := db.GoRequest{
		ObjectType: MapConfigObjectType,
	}
	sd, ti, err := mapconfigDb.resolveTable(&req)
	if err != nil {
		return nil, err
	}
	query := bson.M{
		"me.key": mapconfig.Key,
	}
	raw := &bson.Raw{}
	c := Coder{TypeData: sd}
	var obj DBMiscObject
	er := mapconfigDb.dbconn.DB().C(ti.Name).Find(query).One(raw)
	if er != mgo.ErrNotFound {
		c.DecodeTo(*raw, unsafe.Pointer(&obj))
		trackTimeoutConfig, ok := obj.Obj.(*moc.MapConfig)
		if !ok {
			return nil, db.ErrorNotFound
		}
		mapconfig.Id = trackTimeoutConfig.Id
	}
	res, err := mapconfigDb.Upsert(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: MapConfigObjectType,
			Obj: &db.GoObject{
				ID:   mapconfig.Id,
				Data: mapconfig,
			},
		},
		Ctxt: mapconfigDb.ctxt,
	})
	return res, err
	// else {
	// 	query := bson.M{
	// 		"me.key": mapconfig.Key,
	// 	}

	// 	update := bson.M{
	// 		"$set": bson.M{
	// 			"me.value": mapconfig.Value,
	// 		},
	// 	}

	// 	err := mapconfigDb.dbconn.DB().C(CollectionMapConfig).Update(query, update)
	// 	return err;
	// }
}
