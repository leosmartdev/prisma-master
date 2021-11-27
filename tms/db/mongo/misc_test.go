package mongo

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"prisma/gogroup"
	"prisma/tms"
	"prisma/tms/db"
	"prisma/tms/envelope"
	"prisma/tms/iridium"
	"prisma/tms/moc"
	"prisma/tms/omnicom"
	"prisma/tms/routing"
	"prisma/tms/tmsg"
	. "prisma/tms/tmsg/client"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	pb "github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

type MockTsiClient struct {
	t        *testing.T
	expected *omnicom.Omni
}

func TestGetMiscIncident_FindWithTrackID(t *testing.T) {
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
					Id:   "8554a82d3a9b73de1d1f75d2dbea0dac",
				},
			},
		},
	}

	_, err = misc.Upsert(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: "prisma.tms.moc.Incident",
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

	req := db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: "prisma.tms.moc.Incident",
			Obj: &db.GoObject{
				Data: &moc.Incident{
					Log: []*moc.IncidentLogEntry{
						&moc.IncidentLogEntry{
							Type: "TRACK",
							Entity: &moc.EntityRelationship{
								Type: "registry",
								Id:   "8554a82d3a9b73de1d1f75d2dbea0dac",
							},
						},
					},
				},
			},
		},
		Ctxt: ctx,
		Time: &db.TimeKeeper{},
	}

	resp, err := misc.Get(req)
	if err != nil {
		t.Error(err)
	}
	if len(resp) != 0 {
		_, ok := resp[0].Contents.Data.(*moc.Incident)
		if !ok {
			t.Errorf("Count not cast incident %+v", resp)
		}
	}
	_, err = mClient.DB().C("incidents").RemoveAll(bson.M{"me.id": incID})
	if err != nil {
		t.Error(err)
	}

}

func TestMongoMiscClient_ResolveTable(t *testing.T) {
	req := db.GoRequest{
		ObjectType: "prisma.tms.moc.Zone",
	}
	c := MongoMiscClient{}
	sd, ti, err := c.resolveTable(&req)
	_ = sd

	fmt.Printf("%+v\n", sd)
	_ = ti
	fmt.Printf("%+v\n", ti)
	assert.NoError(t, err)
}

func (client *MockTsiClient) Send(c context.Context, m *tms.TsiMessage) {
	if client.expected == nil {
		client.t.Errorf("Not expecting to send a message, but got a Send() request with %v", m)

	}
	var ir iridium.Iridium
	tmsg.UnpackTo(m.Body, &ir)
	if ir.Payload == nil || ir.Payload.Omnicom == nil {
		client.t.Errorf("Expected payload to be not-nil")
	}

	omni := ir.Payload.Omnicom

	// Can't test equality of IDs in case we are sending multiple messages  (broadcast to different beacons)
	var msgID uint32
	if omni.GetUic() != nil {
		msgID = ir.Payload.Omnicom.GetUic().ID_Msg
		ir.Payload.Omnicom.GetUic().ID_Msg = 0

	}
	if omni.GetUgp() != nil {
		msgID = ir.Payload.Omnicom.GetUgp().ID_Msg
		ir.Payload.Omnicom.GetUgp().ID_Msg = 0

	}

	if msgID <= 0 || msgID >= 4095 {
		client.t.Errorf("Expected message id to be between 1 and 4095, got %v", msgID)
	}

	if !reflect.DeepEqual(omni, client.expected) {
		client.t.Errorf("Expected the payload to be %v but got %v", client.expected, omni)
	}
}

func (client *MockTsiClient) SendNotify(c context.Context, m *tms.TsiMessage, f func(*routing.DeliveryReport)) {

}

func (client *MockTsiClient) SendTo(c context.Context, e tms.EndPoint, m pb.Message) {

}

func (client *MockTsiClient) BroadcastLocal(c context.Context, m pb.Message) {

}

func (client *MockTsiClient) SendToGateway(c context.Context, m pb.Message) {

}

func (client *MockTsiClient) Listen(c context.Context, l routing.Listener) <-chan *TMsg {
	return nil
}

func (client *MockTsiClient) RegisterHandler(msgType string, handler func(*TMsg) pb.Message) {

}

func (client *MockTsiClient) Request(c context.Context, e tms.EndPoint, m pb.Message) (pb.Message, error) {
	return nil, nil

}

func (client *MockTsiClient) Local() *tms.EndPoint {
	return nil

}
func (client *MockTsiClient) LocalRouter() *tms.EndPoint {
	return nil

}

func (client *MockTsiClient) ResolveSite(string) uint32 {
	return 0

}
func (client *MockTsiClient) ResolveApp(string) uint32 {
	return 0

}

func (client *MockTsiClient) GenerateTrackTSN() string {
	return "mocktrackid"
}

func (client *MockTsiClient) GenerateTargetTSN() *tms.TargetID {
	return nil
}

func (client *MockTsiClient) Publish(envelope envelope.Envelope) {
}

func getMongoMiscClient(t *testing.T, expected *omnicom.Omni) db.MiscDB {
	ctxt := gogroup.New(nil, "")
	mongoClient := &MongoClient{
		Ctxt: ctxt,
	}
	misc := NewMongoMiscData(ctxt, mongoClient)
	return misc
}

func TestValidate(t *testing.T) {
	point1 := &tms.Point{
		Latitude:  1.237156293090237,
		Longitude: 104.11788831089841,
	}
	point2 := &tms.Point{
		Latitude:  1.2395589910601785,
		Longitude: 104.10346875523436,
	}
	point3 := &tms.Point{
		Latitude:  1.2635858501541009,
		Longitude: 104.09488568638669,
	}
	point4 := &tms.Point{
		Latitude:  1.2646155677504396,
		Longitude: 104.13127789830075,
	}
	linestring := &tms.LineString{
		Points: []*tms.Point{point1, point2, point3, point4, point1},
	}
	zone := &moc.Zone{
		Name: "Zone 1",
		Poly: &tms.Polygon{
			Lines: []*tms.LineString{linestring},
		},
		CreateAlertOnEnter: true,
		CreateAlertOnExit:  false,
	}
	if !validate(zone) {
		t.Errorf("Expected the zone (%v) to be valid", zone)
	}

	zone = &moc.Zone{
		Name: "Zone 1",
		Poly: &tms.Polygon{
			Lines: []*tms.LineString{linestring},
		},
		CreateAlertOnEnter: true,
		CreateAlertOnExit:  true,
	}
	if validate(zone) {
		t.Errorf("Expected the zone (%v) to be invalid", zone)
	}

	point4 = &tms.Point{
		Latitude:  1.2646155677504396,
		Longitude: 184.13127789830075,
	}
	linestring = &tms.LineString{
		Points: []*tms.Point{point1, point2, point3, point4, point1},
	}
	zone = &moc.Zone{
		Name: "Zone 1",
		Poly: &tms.Polygon{
			Lines: []*tms.LineString{linestring},
		},
		CreateAlertOnEnter: true,
		CreateAlertOnExit:  false,
	}
	if validate(zone) {
		t.Errorf("Expected the zone (%v) to be invalid", zone)
	}

}

func getGoMiscRequest(trackIDs []string, uic *omnicom.Uic, ugp *omnicom.Ugp) db.GoMiscRequest {
	messageTime := time.Date(2017, 03, 07, 10, 18, 0, 0, time.UTC)

	var omni *omnicom.Omni

	if uic != nil || ugp != nil {
		omni = &omnicom.Omni{}
		if uic != nil {
			omni.Omnicom = &omnicom.Omni_Uic{uic}

		}
		if ugp != nil {
			omni.Omnicom = &omnicom.Omni_Ugp{ugp}

		}
	}
	message := &tms.OutgoingMessage{
		Time:     tms.ToTimestamp(messageTime),
		TrackIds: trackIDs,
		Omnicom:  omni,
	}
	return db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: "prisma.tms.OutgoingMessage",
			Obj: &db.GoObject{
				Data: message,
			},
		},
	}
}
