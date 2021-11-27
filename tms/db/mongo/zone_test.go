package mongo

import (
	"prisma/gogroup"
	"prisma/tms"
	"prisma/tms/db"
	"prisma/tms/moc"
	"prisma/tms/test/context"
	"testing"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/stretchr/testify/assert"
)

func TestGetOne(t *testing.T) {
	ctx := gogroup.New(context.Background(), "zone_test")
	data, err := mgo.ParseURL("localhost:27017")
	assert.NoError(t, err)
	session, err := mgo.DialWithTimeout("localhost:27017", 2*time.Second)
	if err != nil && err.Error() == "no reachable servers" {
		t.Skip("run this test with mongod together")
		return
	}
	session.Close()
	data.Timeout = 500 * time.Millisecond
	mClient, err := NewMongoClient(ctx, data, nil)
	assert.NoError(t, err)

	t.Run("Zones", func(t *testing.T) {
		read(t, ctx, mClient)
	})
}

func read(t *testing.T, ctx gogroup.GoGroup, dbconn *MongoClient) {
	mocZone := &moc.Zone{
		DatabaseId: bson.NewObjectId().Hex(),
		ZoneId:     3726019210,
		Name:       "Omnicom Geofence 1",
		Poly: &tms.Polygon{
			Lines: []*tms.LineString{
				&tms.LineString{
					Points: []*tms.Point{
						&tms.Point{
							Latitude:  1.237156293090237,
							Longitude: 104.11788831089841,
						},
						&tms.Point{
							Latitude:  1.2395589910601785,
							Longitude: 104.10346875523436,
						},
						&tms.Point{
							Latitude:  1.2635858501541009,
							Longitude: 104.09488568638669,
						},
						&tms.Point{
							Latitude:  1.237156293090237,
							Longitude: 104.11788831089841,
						},
					},
				},
			},
		},
	}
	goreq := db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: ZoneObjectType,
			Obj: &db.GoObject{
				Data: mocZone,
				ID:   mocZone.DatabaseId,
			},
		},
		Ctxt: ctx,
		Time: &db.TimeKeeper{},
	}
	misc := NewMongoMiscData(ctx, dbconn)
	client := NewMongoZoneMiscData(misc)
	_, err := client.Upsert(goreq)
	assert.NoError(t, err)
	zone, err := client.GetOne(3726019210)
	assert.NoError(t, err)
	assert.Equal(t, zone, mocZone, "Should be equal")
	_, err = dbconn.DB().C("zones").RemoveAll(bson.M{"_id": bson.ObjectIdHex(mocZone.DatabaseId)})
	assert.NoError(t, err)
}
