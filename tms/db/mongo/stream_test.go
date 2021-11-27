package mongo

import (
	"context"
	"fmt"
	"prisma/gogroup"
	"testing"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/stretchr/testify/assert"
)

func TestStream(t *testing.T) {
	ctx := gogroup.New(context.Background(), "stream_test")
	data, err := mgo.ParseURL("localhost:27017")
	session, err := mgo.DialWithTimeout("localhost:27017", 2*time.Second)
	if err != nil && err.Error() == "no reachable servers" {
		t.Skip("run this test with mongod together")
		return
	}
	fmt.Println("using mongodb for tests")
	session.Close()
	data.Timeout = 500 * time.Millisecond
	mClient, err := NewMongoClient(ctx, data, nil)
	assert.NoError(t, err)

	t.Run("watching", func(t *testing.T) {
		tWatch(t, ctx, mClient, "test_stream_watch")
	})
	mClient.DB().C("test_stream_watch").DropCollection()
}

func tWatch(t *testing.T, ctx gogroup.GoGroup, dbConn *MongoClient, col string) {
	s := NewStream(ctx, dbConn, col)
	count := 0
	go s.Watch(func(ctx context.Context, informer interface{}) {
		data, ok := informer.(bson.Raw)
		assert.True(t, ok)
		b := make(bson.M)
		assert.NoError(t, data.Unmarshal(b))
		assert.Contains(t, []string{"Jackson", "Travolta"}, b["name"])
		count++
	}, false, nil, nil)
	time.Sleep(2 * time.Second)
	dbConn.DB().C(col).Insert(bson.M{
		"name":   "Travolta",
		"mytime": time.Now().Add(-20 * time.Minute),
	})
	dbConn.DB().C(col).Insert(bson.M{
		"name":   "Jackson",
		"mytime": time.Now().Add(-10 * time.Minute),
	})
	<-time.After(2 * time.Second)
	assert.Equal(t, count, 2)
}
