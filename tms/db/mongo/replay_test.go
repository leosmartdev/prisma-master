package mongo

import (
	"context"
	"prisma/gogroup"
	"prisma/tms/client_api"
	"testing"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/stretchr/testify/assert"
)

func TestReplay(t *testing.T) {
	ctx := gogroup.New(context.Background(), "replay_test")
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

	t.Run("replaying", func(t *testing.T) {
		tReplay(t, ctx, mClient, "test_replay")
	})
	mClient.DB().C("test_replay").DropCollection()
}

func tReplay(t *testing.T, ctx gogroup.GoGroup, dbConn *MongoClient, col string) {

	dbConn.DB().C(col).Insert(bson.M{
		"name":   "Travolta",
		"mytime": time.Now().Add(-20 * time.Minute),
	})
	dbConn.DB().C(col).Insert(bson.M{
		"name":   "Jackson",
		"mytime": time.Now().Add(-10 * time.Minute),
	})
	r := NewReplay(dbConn, 15*time.Minute, col, "mytime", nil)
	r.Do(ctx, func(ctx context.Context, informer interface{}) {
		switch informer.(type) {
		case bson.Raw:
			data, ok := informer.(bson.Raw)
			assert.True(t, ok)
			b := make(bson.M)
			assert.NoError(t, data.Unmarshal(b))
			assert.Equal(t, "Jackson", b["name"])
		case client_api.Status:
			data, ok := informer.(client_api.Status)
			assert.True(t, ok)
			assert.Equal(t, client_api.Status_InitialLoadDone, data)
		default:
			t.Log("unknown type")
		}
	})
}
