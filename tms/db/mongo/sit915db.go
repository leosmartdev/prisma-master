package mongo

import (
	"fmt"
	"prisma/tms/db"
	"prisma/tms/moc"
	"unsafe"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// ObjectTypes
const (
	Sit915ObjectType = "prisma.tms.moc.Sit915"
)

// MongoDB collections
const CollectionSit915 = "sit915"

func NewSit915Db(misc db.MiscDB) db.Sit915DB {
	client, ok := misc.(*MongoMiscClient)
	if !ok {
		return nil
	}
	return &MongoMiscClient{
		dbconn: client.dbconn,
		ctxt:   client.ctxt,
	}
}

func (sit915Db *MongoMiscClient) UpsertSit915(sit915 *moc.Sit915) error {
	_, err := sit915Db.Upsert(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: Sit915ObjectType,
			Obj: &db.GoObject{
				ID:   sit915.Id,
				Data: sit915,
			},
		},
		Ctxt: sit915Db.ctxt,
		Time: &db.TimeKeeper{},
	})

	return err
}

func (sit915Db *MongoMiscClient) FindAllSit915s(params ...int) ([]*moc.Sit915, error) {
	req := db.GoRequest{
		ObjectType: Sit915ObjectType,
	}
	structData, tableInfo, err := sit915Db.resolveTable(&req)
	if err != nil {
		return nil, err
	}

	var query bson.M
	if len(params) == 0 {
		query = nil
	} else if len(params) > 1 {
		startDateTime := params[0]
		endDateTime := params[1]
		direction := params[2]

		query = bson.M{
			"me.timestamp.seconds": bson.M{
				"$gte": startDateTime,
				"$lt":  endDateTime,
			},
		}

		if direction == 1 {
			query["me.status"] = moc.Sit915_SENT.String()
		} else if direction == 3 {
			query["me.status"] = moc.Sit915_PENDING.String()
		} else if direction == 4 {
			query["me.status"] = moc.Sit915_FAILED.String()
		}
	}

	raw := []bson.Raw{}
	corder := Coder{TypeData: structData}

	sit915s := make([]*moc.Sit915, 0)

	err = sit915Db.dbconn.DB().C(tableInfo.Name).Find(query).Sort("-me.timestamp.seconds").All(&raw)
	if err != mgo.ErrNotFound {
		for _, data := range raw {
			var obj DBMiscObject
			corder.DecodeTo(data, unsafe.Pointer(&obj))
			sit915, ok := obj.Obj.(*moc.Sit915)
			if !ok {
				return nil, fmt.Errorf("Could not recover SIT 915 object")
			}
			sit915.Id = obj.Id.Hex()
			sit915s = append(sit915s, sit915)
		}
	}

	return sit915s, err
}

func (sit915Db *MongoMiscClient) FindOneSit915(id string) (*moc.Sit915, error) {
	var sit915 *moc.Sit915

	res, err := sit915Db.Get(db.GoMiscRequest{
		Req: &db.GoRequest{
			Obj: &db.GoObject{
				ID: id,
			},
			ObjectType: Sit915ObjectType,
		},
		Ctxt: sit915Db.ctxt,
		Time: &db.TimeKeeper{},
	})

	if err == nil {
		var sit915s = make([]*moc.Sit915, 0)

		for _, sit915Datum := range res {
			if sit915, ok := sit915Datum.Contents.Data.(*moc.Sit915); ok {
				sit915s = append(sit915s, sit915)
			}
		}

		if len(sit915s) > 0 {
			sit915 = sit915s[0]
		} else {
			err = db.ErrorNotFound
		}
	}

	return sit915, err
}
