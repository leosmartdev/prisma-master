package public

import (
	"net/http/httptest"
	"prisma/gogroup"
	"prisma/tms"
	"prisma/tms/db/mongo"
	"prisma/tms/devices"
	"testing"
	"time"

	"github.com/globalsign/mgo"
	restful "github.com/orolia/go-restful"
)

func TestgetTrack(t *testing.T) {
	reqr := httptest.NewRequest("GET", "/history/4b6a6510ec550c47e20ae2f66ed365f5?time=1662006489", nil)
	respw := httptest.NewRecorder()
	req := restful.NewRequest(reqr)
	req.PathParameters()["registry-id"] = "4b6a6510ec550c47e20ae2f66ed365f5"
	resp := restful.NewResponse(respw)
	ctxt := gogroup.New(nil, "testhistory")

	data, _ := mgo.ParseURL("localhost:27017")
	session, err := mgo.DialWithTimeout("localhost:27017", 2*time.Second)
	if err != nil && err.Error() == "no reachable servers" {
		t.Skip("run this test with mongod together")
		return
	}
	session.Close()
	mClient, _ := mongo.NewMongoClient(ctxt, data, nil)
	hr := NewHistoryRest(mClient, ctxt)

	err = hr.trackdb.Insert(&tms.Track{
		RegistryId: "4b6a6510ec550c47e20ae2f66ed365f5",
		Id:         "4a1a6510ec550c47e20ae2f66ed365f4",

		Targets: []*tms.Target{
			&tms.Target{
				Type: devices.DeviceType_Manual,
				Time: tms.Now(),
			},
		},
	})
	if err != nil {
		t.Errorf("Can not insert track %+v", err)
	}
	respinfo, errs := hr.getTrack(req)
	if errs != nil {
		t.Errorf("Can not find track %+v", errs)
	}
	if respinfo.LookupID != req.PathParameter("registry-id") {
		t.Error("lookupID and registry-id are not equal")
	}
	if resp.StatusCode() != 200 {
		t.Errorf("bad status code: %+v", resp.StatusCode())
	}
}
