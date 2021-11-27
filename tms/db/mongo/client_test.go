package mongo

import (
	"context"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"prisma/gogroup"

	"github.com/globalsign/mgo"
	"github.com/stretchr/testify/require"
)

const schemaPayload = `
let ldb = db.getSiblingDB('test');
ldb.createCollection('test_collection_migration');
ldb.test_collection_migration.insert({x: 1});
`

const schemaWithBadPayload = `
badString%%%
let ldb = db.getSiblingDB('test');
ldb.createCollection('test_collection_migration');
ldb.test_collection_migration.insert({x: 1})
`

func TestMongoClient_Migrate(t *testing.T) {
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
	mClient.Sess().DB("test").DropDatabase()
	require.NoError(t, err)
	f, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	_, err = f.WriteString(schemaPayload)
	require.NoError(t, err)
	err = mClient.Migrate(f.Name())
	require.NoError(t, err)
	c, err := mClient.Sess().DB("test").C("test_collection_migration").Count()
	require.NoError(t, err)
	require.Equal(t, 1, c)

	_, err = f.WriteString(schemaWithBadPayload)
	require.NoError(t, err)
	err = mClient.Migrate(f.Name())
	require.Error(t, err)
	require.NoError(t, mClient.Sess().DB("test").DropDatabase())
}

func TestMongoClient_EnsureSetUp(t *testing.T) {
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
	mClient.Sess().DB("test").DropDatabase()
	require.NoError(t, err)
	dir, err := ioutil.TempDir("/tmp", "db")
	require.NoError(t, err)
	f, err := ioutil.TempFile(dir, "")
	require.NoError(t, err)
	_, err = f.WriteString(schemaPayload)
	require.NoError(t, err)
	dirs := []string{dir}
	mClient.EnsureSetUp(dirs)
	//require.NoError(t, err)
	c, err := mClient.Sess().DB("test").C("test_collection_migration").Count()
	require.NoError(t, err)
	require.Equal(t, 1, c)
	require.NoError(t, mClient.Sess().DB("test").DropDatabase())
}
