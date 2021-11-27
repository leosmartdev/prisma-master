package mongo

import (
	"bytes"
	"context"
	"io"
	"time"

	"prisma/gogroup"
	"prisma/tms"
	api "prisma/tms/client_api"
	"prisma/tms/db"
	"prisma/tms/envelope"
	"prisma/tms/log"
	"prisma/tms/moc"
	"prisma/tms/routing"
	"prisma/tms/tmsg"
	clt "prisma/tms/tmsg/client"
	"prisma/tms/ws"

	"github.com/globalsign/mgo/bson"
	"github.com/golang/protobuf/jsonpb"
)

var replayZoneFilters = bson.M{"utime": bson.M{"$gt": time.Now().Add(-24 * time.Hour)}}

type mongoStreamer struct {
	dbc    *MongoClient
	tracer *log.Tracer
}

func NewStreamer(dbc *MongoClient) ws.Streamer {
	return &mongoStreamer{
		dbc:    dbc,
		tracer: log.GetTracer("streamer"),
	}
}

func (s *mongoStreamer) ToClient(ctxt gogroup.GoGroup, client *ws.Client) {
	go s.StreamNotices(ctxt.Child("notices"), client.Send)
	go s.StreamZones(ctxt.Child("zones"), client.Send)
	go s.StreamTransmissions(ctxt.Child("transmissions"), client.Send)
	go s.StreamDevices(ctxt.Child("devices"), client.Send)
	go s.StreamMarkers(ctxt.Child("markers"), client.Send)
}

func (s *mongoStreamer) StreamDevices(ctx gogroup.GoGroup, sendq chan string) {

	type listner struct {
		tclient   clt.TsiClient
		ctxt      gogroup.GoGroup
		devupdate <-chan *clt.TMsg
	}

	l := listner{
		tclient: tmsg.GClient,
		ctxt:    ctx,
	}

	child := ctx.Child("update devices streamer")

	child.Go(func(l listner, sendq chan string) {
		lctx := l.ctxt.Child("device update listner")
		lctx.ErrCallback(func(err error) {
			pe, ok := err.(gogroup.PanicError)
			if ok {
				log.Error("Panic in device update listener thread: %v\n%v", pe.Msg, pe.Stack)
			} else {
				log.Error("Error in device update listener thread: %v", err)
			}
		})

		l.devupdate = l.tclient.Listen(lctx, routing.Listener{
			MessageType: "prisma.tms.moc.Device",
		})

	hdev:
		for {
			select {
			case <-l.ctxt.Done():
				return
			case message, ok := <-l.devupdate:
				if !ok {
					log.Error("closed channel")
					return
				}
				dev, ok := message.Body.(*moc.Device)
				if !ok {
					log.Warn("Got non-device message in device stream")
					continue hdev
				}
				e := envelope.Envelope{
					Type: "Device/UPDATE",
					Contents: &envelope.Envelope_Device{
						Device: dev,
					},
				}
				marshaller := jsonpb.Marshaler{}
				payload := bytes.Buffer{}
				err := marshaller.Marshal(&payload, &e)
				if err != nil {
					log.Error("unable to marshal message: %v", err)
					continue hdev
				}
				sendq <- payload.String()
			}
		}
	}, l, sendq)
}

func (s *mongoStreamer) StreamNotices(ctxt gogroup.GoGroup, sendq chan string) {
	defer ctxt.Cancel(io.EOF)
	notifydb := NewNotifyDb(ctxt, s.dbc)
	pipelines := []bson.M{
		{
			"$match": bson.M{
				"updateDescription.updatedFields.me.action": bson.M{
					"$ne": "",
				},
			},
		},
	}
	stream := notifydb.GetPersistentStream(pipelines)
	for {
		select {
		case update, ok := <-stream.Updates:
			if !ok {
				log.Error("closed channel")
				return
			}
			// Don't send ACK messages
			if update.Action == moc.Notice_ACK {
				continue
			}
			s.tracer.Logf("sending notice: %v", update)
			e := envelope.Envelope{
				Type: "Notice/" + update.Action.String(),
				Contents: &envelope.Envelope_Notice{
					Notice: update,
				},
			}
			marshaller := jsonpb.Marshaler{}
			payload := bytes.Buffer{}
			err := marshaller.Marshal(&payload, &e)
			if err != nil {
				log.Error("unable to marshal message: %v", err)
				ctxt.Cancel(err)
				return
			}
			sendq <- payload.String()
		case <-ctxt.Done():
			return
		}
	}
}

func (s *mongoStreamer) StreamZones(ctxt gogroup.GoGroup, sendq chan string) {
	defer ctxt.Cancel(io.EOF)
	miscdb := NewMongoMiscData(ctxt, s.dbc)
	stream := miscdb.GetPersistentStream(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: "prisma.tms.moc.Zone",
		},
		Ctxt: ctxt,
	}, replayZoneFilters, nil)
	for {
		select {
		case update, ok := <-stream:
			if !ok {
				log.Error("closed channel")
				return
			}
			if update.Status == api.Status_InitialLoadDone {
				continue
			}
			s.tracer.Logf("sending zone: %v", update)
			action := moc.Zone_UPDATE.String()
			if update.Status == api.Status_Timeout {
				action = moc.Zone_DELETE.String()
			}
			object := update.Contents
			if object == nil {
				log.Error("wrong record %v", log.Spew(update))
				continue
			}
			zone, ok := object.Data.(*moc.Zone)
			if !ok {
				log.Error("Expecting zone, got: %+v", object.Data)
				continue
			}
			zone.DatabaseId = object.ID
			e := envelope.Envelope{
				Type: "Zone/" + action,
				Contents: &envelope.Envelope_Zone{
					Zone: zone,
				},
			}
			marshaller := jsonpb.Marshaler{}
			payload := bytes.Buffer{}
			err := marshaller.Marshal(&payload, &e)
			if err != nil {
				log.Error("unable to marshal message: %v", err)
				ctxt.Cancel(err)
				return
			}
			sendq <- payload.String()
		case <-ctxt.Done():
			return
		}
	}
}

func (s *mongoStreamer) StreamTransmissions(ctx gogroup.GoGroup, sendq chan string) {
	log.Info("transmission stream init")
	defer ctx.Cancel(io.EOF)
	d := NewStream(ctx.Child("ws/"+CollectionTransmission), s.dbc, CollectionTransmission)
	d.Watch(func(_ context.Context, informer interface{}) {
		switch data := informer.(type) {
		case bson.Raw:
			transmission := tms.Transmission{}
			err := data.Unmarshal(&transmission)
			if err != nil {
				log.Error(err.Error(), err)
				return
			}
			e := envelope.Envelope{
				Type: "Transmission/UPDATE",
				Contents: &envelope.Envelope_Transmission{
					Transmission: &transmission,
				},
			}
			marshaller := jsonpb.Marshaler{}
			payload := bytes.Buffer{}
			err = marshaller.Marshal(&payload, &e)
			if err != nil {
				log.Error(err.Error(), err)
				return
			}
			sendq <- payload.String()
		default:
			log.Info("data is not supported: %v", data)
		}
	}, false, nil, nil)
}

func (s *mongoStreamer) StreamMarkers(ctxt gogroup.GoGroup, sendq chan string) {
	defer ctxt.Cancel(io.EOF)
	markerdb := NewMarkerDb(ctxt, s.dbc)
	stream := markerdb.GetPersistentStream(nil)
	for {
		select {
		case update, ok := <-stream.Updates:
			if !ok {
				log.Error("closed channel")
				return
			}
			s.tracer.Logf("sending marker: %v", update)
			e := envelope.Envelope{
				Type: "Marker/Update",
				Contents: &envelope.Envelope_Marker{
					Marker: update,
				},
			}
			marshaller := jsonpb.Marshaler{}
			payload := bytes.Buffer{}
			err := marshaller.Marshal(&payload, &e)
			if err != nil {
				log.Error("unable to marshal message: %v", err)
				ctxt.Cancel(err)
				return
			}
			sendq <- payload.String()
		case <-ctxt.Done():
			return
		}
	}
}
