package mongo

import (
	"prisma/gogroup"
	"prisma/tms"
	tmsdb "prisma/tms/db"
	"reflect"
	"unsafe"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const CollectionActivity = "activity"

type MongoActivityClient struct {
	dbconn *MongoClient     // Connection to DB
	reg    tmsdb.RegistryDB // Connection to registry
	ctxt   gogroup.GoGroup  // Execution context
}

func NewMongoActivities(ctxt gogroup.GoGroup, dbconn *MongoClient) tmsdb.ActivityDB {
	client := &MongoActivityClient{
		dbconn: dbconn,
		reg:    NewMongoRegistry(ctxt, dbconn),
		ctxt:   ctxt,
	}
	return client
}

func (activityDb *MongoActivityClient) GetSit915Messages(startDateTime int, endDateTime int) ([]*tms.MessageActivity, error) {
	db := activityDb.dbconn.DB()
	defer activityDb.dbconn.Release(db)

	query := bson.M{
		"me.type":                       "SARSAT",
		"me.sarsat.sarsat.message_type": "SIT_915",
		"me.sarsat.sarsat.received":     true,
		"me.time.seconds": bson.M{
			"$gte": startDateTime,
			"$lt":  endDateTime,
		},
	}

	structData := NewStructData(reflect.TypeOf(DBActivity{}), NoMap)
	coder := Coder{TypeData: structData}
	activities := make([]*tms.MessageActivity, 0)
	raw := []bson.Raw{}

	err := db.C(CollectionActivity).Find(query).All(&raw)

	if err == mgo.ErrNotFound {
		return activities, nil
	}

	for _, data := range raw {
		activity := new(DBActivity)

		coder.DecodeTo(data, unsafe.Pointer(activity))

		activities = append(activities, activity.Body)
	}

	return activities, err
}
