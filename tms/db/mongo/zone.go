package mongo

import (
	"fmt"
	"prisma/tms/db"
	"prisma/tms/moc"
	"unsafe"

	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// ZoneObjectType ...
const ZoneObjectType = "prisma.tms.moc.Zone"

// NewMongoZoneMiscData ...
func NewMongoZoneMiscData(misc db.MiscDB) db.ZoneDB {
	client, ok := misc.(*MongoMiscClient)
	if !ok {
		return nil
	}
	c := &MongoMiscClient{
		dbconn: client.dbconn,
		ctxt:   client.ctxt,
	}
	return c
}

// GetOne returns a zone that corresponds to the omnicom geo_ID
func (z *MongoMiscClient) GetOne(omnID uint32) (*moc.Zone, error) {
	req := db.GoRequest{
		ObjectType: ZoneObjectType,
	}
	sd, ti, err := z.resolveTable(&req)
	if err != nil {
		return nil, err
	}
	query := bson.M{"me.zone_id": omnID, "etime": bson.M{"$gte": time.Now()}}
	raw := &bson.Raw{}
	c := Coder{TypeData: sd}
	var obj DBMiscObject
	err = z.dbconn.DB().C(ti.Name).Find(query).One(raw)
	if err != mgo.ErrNotFound {
		c.DecodeTo(*raw, unsafe.Pointer(&obj))
	}
	zone, ok := obj.Obj.(*moc.Zone)
	if !ok {
		return nil, fmt.Errorf("Could not recover Zone object")
	}

	return zone, nil
}
