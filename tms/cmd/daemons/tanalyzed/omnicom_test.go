package main

import (
	"testing"
	"time"

	"prisma/gogroup"
	"prisma/tms"
	"prisma/tms/db"
	"prisma/tms/db/connect"
	"prisma/tms/db/mockdb"
	"prisma/tms/omnicom"
	"prisma/tms/tmsg"
	"prisma/tms/ws"

	google_protobuf1 "github.com/golang/protobuf/ptypes/wrappers"
	"github.com/gomodule/redigo/redis"
	"github.com/json-iterator/go/assert"
	"github.com/rafaeljusto/redigomock"
)

func testOmnicomImeiTrack(id string, imei string, ar *omnicom.Ar) *tms.Track {
	return &tms.Track{
		Id: id,
		Targets: []*tms.Target{
			&tms.Target{
				Omnicom: &omnicom.Omni{
					Omnicom: &omnicom.Omni_Ar{ar},
				},
				Imei: &google_protobuf1.StringValue{Value: imei},
			},
		},
	}
}

//TODO: write unit test for omnicom assistance alert

func TestOmnicomAssistanceAlert(t *testing.T) {

	ctxt := gogroup.New(nil, "test assistance alert")
	timer := time.NewTimer(time.Second * 2).C
	ch := make(chan bool)

	go func(t *testing.T, ctxt gogroup.GoGroup) {
		err := tmsg.TsiClientGlobal(ctxt, tmsg.APP_ID_UNKNOWN)
		if err != nil {
			t.Errorf("Could not create global TsiClient: %v", err)
			return
		}
		ch <- true

	}(t, ctxt)

	select {
	case <-timer:
		t.Skip("could not connect to tsi client global")
	case <-ch:
		err := tmsg.TsiClientGlobal(ctxt, tmsg.APP_ID_UNKNOWN)
		if err != nil {
			t.Errorf("Could not create global TsiClient: %v", err)
			return
		}
		conn, err := connect.GetMongoClient(ctxt, tmsg.GClient)
		if err != nil {
			t.Errorf("unable to connect to the database: %v", err)
			return
		}
		redisMock := redigomock.NewConn()
		redisPool := &redis.Pool{
			Dial: func() (redis.Conn, error) {
				return redisMock, nil
			},
		}
		n := NewNotifier(ctxt, nil, redisPool)
		n.notifydb = mockdb.NewNotifyDbStub()
		a := newOmnicomStage(n, ctxt, conn, ws.NewPublisher())
		omni := &omnicom.Omni{
			Omnicom: &omnicom.Omni_Ar{
				Ar: &omnicom.Ar{
					Assistance_Alert: &omnicom.AssistanceAlert{
						Alert_Status:                    1,
						Current_Assistance_Alert_Status: 1,
					},
				},
			},
		}

		msg := &tms.MessageActivity{
			MetaData: &tms.MessageActivity_Omni{omni},
			Imei:     &google_protobuf1.StringValue{Value: "123456789012345"},
		}

		update := db.GoGetResponse{
			Contents: &db.GoObject{
				Data: msg,
			},
		}

		hmsetCmd := redisMock.Command("HMSET")
		hmsetCmd.Expect("OK")
		a.analyzeActivity(update)

		assert.Equal(t, 1, redisMock.Stats(hmsetCmd))
	}
}

/*func TestOmnicomAssistanceAlertClear(t *testing.T) {
	ctxt := gogroup.New(nil, "")
	n := NewNotifier(ctxt, nil)
	n.notifydb = mockdb.NewNotifyDbStub()
	a := newOmnicomStage(n)

	update := api.TrackUpdate{
		Status: api.Status_Current,
		Track: testOmnicomImeiTrack("oc1", "1234", &omnicom.Ar{
			Assistance_Alert: &omnicom.AssistanceAlert{
				Alert_Status:                    1,
				Current_Assistance_Alert_Status: 1,
			},
		}),
	}
	oc1 := ident.With("event", moc.Notice_OmnicomAssistance).With("imei", "1234").Hash()
	a.analyze(update)

	update = api.TrackUpdate{
		Status: api.Status_Current,
		Track: testOmnicomImeiTrack("oc1", "1234", &omnicom.Ar{
			Assistance_Alert: &omnicom.AssistanceAlert{
				Alert_Status:                    1,
				Current_Assistance_Alert_Status: 0,
			},
		}),
	}
	a.analyze(update)

	notice, exists := n.notices[oc1]
	if !exists {
		t.Fatalf("expecting notice")
	}
	have := notice.Action
	want := moc.Notice_ACK_WAIT
	if have != want {
		t.Errorf("\n want %v \n have %v", want, have)
	}
}*/
