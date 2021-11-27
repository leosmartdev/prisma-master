package mongo

import (
	"prisma/gogroup"
	"prisma/tms/db"
	"prisma/tms/moc"
	"prisma/tms/test/context"
	"testing"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/stretchr/testify/assert"
)

func TestMisc_GetIncidentWithTrackID(t *testing.T) {
	ctx := gogroup.New(context.Background(), "incident_test")
	data, err := mgo.ParseURL("localhost:27017")
	session, err := mgo.DialWithTimeout("localhost:27017", 2*time.Second)
	if err != nil && err.Error() == "no reachable servers" {
		t.Skip("run this test when mongod is up")
		return
	}
	session.Close()
	data.Timeout = 1 * time.Second
	mClient, err := NewMongoClient(ctx, data, nil)
	assert.NoError(t, err)
	misc := NewMongoMiscData(ctx, mClient)

	incID := bson.NewObjectId().Hex()

	registryID := bson.NewObjectId().Hex()

	incident := &moc.Incident{
		Id:         incID,
		IncidentId: "2019",
		Name:       "Incident_test",
		Type:       "Unlawful",
		Phase:      moc.IncidentPhase_nonphase,
		Commander:  "El MR",
		State:      moc.Incident_Open,
		Assignee:   "admin",
		Log: []*moc.IncidentLogEntry{
			&moc.IncidentLogEntry{
				Type: "TRACK",
				Entity: &moc.EntityRelationship{
					Type: "registry",
					Id:   registryID,
				},
			},
		},
	}

	_, err = misc.Upsert(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: IncidentObjectType,
			Obj: &db.GoObject{
				Data: incident,
			},
		},
		Ctxt: ctx,
		Time: &db.TimeKeeper{},
	})
	if err != nil {
		t.Error(err)
	}

	imisc := NewMongoIncidentMiscData(misc)

	incidents, err := imisc.GetIncidentWithTrackID(registryID)
	if err != nil {
		t.Error(err)
	}

	t.Log(incidents)

	_, err = mClient.DB().C(CollectionIncident).RemoveAll(bson.M{"me.id": incID})
	if err != nil {
		t.Error(err)
	}

}
