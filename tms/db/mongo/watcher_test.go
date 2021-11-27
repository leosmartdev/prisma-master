package mongo

import (
	"fmt"
	"testing"
	"time"

	"prisma/gogroup"
	"prisma/tms/test/context"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/stretchr/testify/assert"
)

func TestWatch(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	const dbName = "test_watcher_MY_TEST"
	ctx := gogroup.New(context.Background(), dbName)
	session, err := mgo.Dial("localhost:8201")
	if err != nil && err.Error() == "no reachable servers" {
		t.Skip("run this test with mongod together")
		return
	}
	fmt.Println("using mongodb for tests")
	defer session.Close()
	assert.NoError(t, err)
	colStart := session.DB(dbName).C("test_watch_start")

	t.Run("replaying", func(t *testing.T) {
		tStart(t, ctx, colStart)
	})
	assert.NoError(t, session.DB(dbName).DropDatabase())
}

func tStart(t *testing.T, ctx gogroup.GoGroup, col *mgo.Collection) {
	col.Insert(bson.M{
		"name": "Travolta",
	})
	w := NewWatcher(ctx, "test")
	w.Start(col, nil)

	<-time.After(3 * time.Second)
	col.Insert(bson.M{
		"name": "Travolta",
	})
	select {
	case data, ok := <-w.GetChannel():
		assert.True(t, ok)
		var bm bson.M
		assert.NoError(t, data.Unmarshal(&bm))
		assert.Equal(t, "Travolta", bm["name"])
	case <-time.After(2 * time.Second):
		t.Fail()
	}
}
