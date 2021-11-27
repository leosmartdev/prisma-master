package mongo

import (
	"testing"
	"sync"
	"strconv"
	"context"
	"time"

	"prisma/gogroup"
	"prisma/tms"

	"github.com/globalsign/mgo"
	"github.com/json-iterator/go/require"
	"github.com/globalsign/mgo/bson"
)

func TestUpdateRegistry(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := gogroup.New(context.Background(), "test_update_registry")
	dialInfo, err := mgo.ParseURL("mongodb://localhost:8201")
	require.NoError(t, err)

	dialInfo.Timeout = 500 * time.Millisecond
	dbClient, err := NewMongoClient(ctx, dialInfo, nil)
	if err != nil && err.Error() == "no reachable servers" {
		t.Skip("run this test with mongod together")
		return
	}

	require.NoError(t, err)
	c := &MongoTrackClient{
		reg:             NewMongoRegistry(ctx, dbClient),
		ctxt:            ctx,
		condRegistry:    sync.NewCond(&sync.Mutex{}),
		registryInserts: make(map[string]struct{}),
	}
	wait := sync.WaitGroup{}
	// run several times to check race conditions
	// start with 1 to add extra time to the update time properly
	for i := 1; i < 6; i++ {
		wait.Add(2)
		go func(number int) {
			c.updateRegistry(&tms.Track{
				RegistryId: "test_registry_id",
				Metadata: []*tms.TrackMetadata{
					{Name: strconv.Itoa(number)},
				},
			})
			wait.Done()
		}(i)
		go func(number int) {
			c.updateRegistry(&tms.Track{
				RegistryId: "test_registry_id",
				Targets: []*tms.Target{
					{
						Mmsi:       strconv.Itoa(number),
						UpdateTime: tms.ToTimestamp(time.Now().Add(time.Duration(number) * time.Minute)),
					},
				},
			})
			wait.Done()
		}(i)
		wait.Wait()
		registry, err := c.reg.Get("test_registry_id")
		require.NoError(t, err)
		require.NotNil(t, registry.Target)
		require.NotNil(t, registry.Metadata)
		require.NotEmpty(t, registry.Metadata.Name)
		require.NotEmpty(t, registry.Target.Mmsi)
		dbClient.DB().C("registry").RemoveId(bson.ObjectIdHex(registry.DatabaseId))
	}
	require.Empty(t, c.registryInserts)
}
