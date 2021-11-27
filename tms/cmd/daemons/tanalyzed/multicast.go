package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"time"
	"errors"

	"prisma/gogroup"
	"prisma/tms"
	api "prisma/tms/client_api"
	"prisma/tms/db/mongo"
	"prisma/tms/log"
	"prisma/tms/moc"
	"prisma/tms/routing"
	"prisma/tms/tmsg"
	"prisma/tms/tmsg/client"
	"prisma/tms/db"

	"github.com/globalsign/mgo/bson"
	"github.com/gomodule/redigo/redis"
)

const expiration = 30 * time.Minute
const retention = expiration * 2
const tickerTimeout = 1 + time.Minute

type multicastStage struct {
	mutex     sync.Mutex
	n         Notifier
	tsiClient client.TsiClient
	rpool     *redis.Pool
	dbClient  *mongo.MongoClient
}

func newMulticastStage(tsiClient client.TsiClient, notifier Notifier, rpool *redis.Pool) *multicastStage {
	return &multicastStage{
		tsiClient: tsiClient,
		n:         notifier,
		rpool:     rpool,
	}
}

func (s *multicastStage) init(ctx gogroup.GoGroup, dbClient *mongo.MongoClient) error {
	log.Info("multicast init")
	s.dbClient = dbClient
	transDb := mongo.NewTransmissionDb(ctx, dbClient)
	mcDb := mongo.NewMulticastDb(ctx)
	drStream := tmsg.GClient.Listen(ctx, routing.Listener{
		MessageType: "prisma.tms.routing.DeliveryReport",
		Destination: &tms.EndPoint{
			Site: tmsg.GClient.ResolveSite(""),
		},
	})
	// transmission DeliverReport stream handler
	ctx.Go(func(ctx gogroup.GoGroup) {
		log.Info("DeliveryReport stream begin")
		for {
			select {
			case <-ctx.Done():
				// Do nothing
			case msg := <-drStream:
				dr, ok := msg.Body.(*routing.DeliveryReport)
				if !ok {
					continue // channel was closed
				}
				n, err := redis.String(s.rpool.Get().Do("GET", dr.NotifyId))
				if err != nil {
					log.Debug(err.Error()+"%v", err)
					continue
				}
				retry := 0
			Retry:
				log.Info("Received Delivery Report %+v for transmission with id %v", dr, n)
				s.mutex.Lock()
				tr, err := transDb.FindByID(fmt.Sprint(n))
				if err != nil {
					log.Error(err.Error()+"%v", err)
					s.mutex.Unlock()
					continue
				}
				// check packets
				found := false
				for i := 0; i < len(tr.Packets); i++ {
					if tr.Packets[i].MessageId == fmt.Sprint(dr.NotifyId) {
						packetStateUpdate(dr, tr.Packets[i])
						if tr.Packets[i].State == tms.Transmission_Failure {
							if err := s.processFailedTransfer(ctx, dbClient, tr, i); err != nil {
								log.Error(err.Error())
							}
						}
						transDb.PacketStatus(tr.Id, tr.Packets[i])
						found = true
						break
					}

				}
				if !found {
					log.Error("not found notify_id:%v", dr.NotifyId)
					s.mutex.Unlock()
					// wait and push back on stream, loop will be averted due to redis expiration
					if retry < 2 {
						time.Sleep(200 * time.Millisecond)
						retry++
						log.Warn("retry %v %v %v", n, dr.NotifyId, retry)
						goto Retry
					}
					continue
				}
				log.Info("Transmission %v", tr)
				// if all transmission completed then update
				allSuccess := true
				for _, packet := range tr.Packets {
					if packet.State != tms.Transmission_Success {
						allSuccess = false
						break
					}
				}
				if allSuccess {
					log.Debug("Deleting transmission after all success %+v", tr.Id)
					s.rpool.Get().Do("DEL", tr.Id)
					err := transDb.StatusById(tr.Id, tms.Transmission_Success, 200)
					if err != nil {
						log.Error(err.Error()+"%v", err)
					}
					tr.State = tms.Transmission_Success // you have to make sure you are doing the right update here plese log tr
					err = mcDb.UpdateTransmission(ctx, tr)
					if err != nil {
						log.Error("%+v", err)
					}
				}
				s.mutex.Unlock()
			}
		}
		log.Info("DeliveryReport stream end")
	}, ctx)
	replay := mongo.NewReplayAll(dbClient, mongo.CollectionMulticast, bson.M{"transmissions.state": bson.M{"$in": []int{1, 4, 5}}})
	// multicast change stream handler
	d := mongo.NewStream(ctx.Child("streamer/"+mongo.CollectionMulticast), dbClient, mongo.CollectionMulticast)
	go d.Watch(s.multicastHandleStreamFunc, true, replay, nil)
	// multicast ticker handler for timeouts
	ticker := time.NewTicker(tickerTimeout)
	go func() {
		for range ticker.C {
			log.Debug("multicast ticker begin")
			replay.Do(ctx.Child("ticker/"+mongo.CollectionMulticast), s.multicastHandleStreamFunc)
			log.Debug("multicast ticker end")
		}
	}()
	return nil
}

func (s *multicastStage) multicastHandleStreamFunc(ctx context.Context, informer interface{}) {
	siteDb := mongo.NewSiteDb(ctx)
	mcDb := mongo.NewMulticastDb(ctx)
	transDb := mongo.NewTransmissionDb(ctx, s.dbClient)
	switch data := informer.(type) {
	case bson.Raw:
		mc := tms.Multicast{}
		err := data.Unmarshal(&mc)
		if err != nil {
			log.Error(err.Error()+"%v", data)
			return
		}
		for i := 0; i < len(mc.Transmissions); i++ {
			transmission := mc.Transmissions[i]
			// end state, no further processing
			if transmission.State == tms.Transmission_Failure || transmission.State == tms.Transmission_Success {
				log.Debug("end state for multicast %v transmission %v", mc.Id, transmission.Id)
				continue
			}
			// check Pending transmissions are processed (edge case)
			exists, _ := redis.Bool(s.rpool.Get().Do("EXISTS", transmission.Id))
			log.Debug("transmission %v  has state %+v, and its existance in Redis is: %v", transmission.Id, transmission.State, exists)
			// if success update multicast and clear redis
			if exists && transmission.State == tms.Transmission_Success {
				s.mutex.Lock()
				err = mcDb.Update(ctx, &mc)
				if err != nil {
					log.Error(err.Error()+"%v", mc)
				}
				s.mutex.Unlock()
				s.rpool.Get().Do("DEL", transmission.Id)
				continue
			}
			// check if timed out
			if exists || transmission.State == tms.Transmission_Partial || transmission.State == tms.Transmission_Retry {
				expired := false
				t, err := redis.String(s.rpool.Get().Do("GET", transmission.Id))
				if err != nil {
					log.Error(err.Error()+"%v. As a result transmission will be marked as expired", transmission.Id)
					expired = true
				}
				sentTime, err := time.Parse(time.RFC3339Nano, t)
				if err != nil {
					log.Error(err.Error()+"%v. As a result transmission will be marked as expired", transmission.Id)
					expired = true
				}
				log.Debug("Time elapsed since send time for %v is %v", transmission.Id, time.Since(sentTime))
				if expired || time.Since(sentTime) > expiration {
					transmission.MessageId = "" // clear to avoid collisions
					log.Debug("Transmission is going to be marked as failure")
					transmission.State = tms.Transmission_Failure
					transmission.Status = &tms.ResponseStatus{
						Code:    int32(tms.ResponseStatus_Failed),
						Message: "transmission send failure",
					}
					s.mutex.Lock()
					err = transDb.Update(transmission)
					if err != nil {
						log.Error(err.Error(), err)
					}
					err = mcDb.Update(ctx, &mc)
					if err != nil {
						log.Error(err.Error()+"%v", mc)
					}
					s.mutex.Unlock()
					s.rpool.Get().Do("DEL", transmission.Id)
					//TODO: in order to notify the UI about tranmission failure, we need the transmission
					//      to have informatio about the multicast and incident associated to it.
					//notifyTransmissionFailure(s.n, transmission)
				}
			}
			// only Pending
			if transmission.State != tms.Transmission_Pending {
				continue
			}
			// Set send time if not already set (edge case)
			if !exists {
				s.rpool.Get().Do("SETEX", transmission.Id, retention.Seconds(), time.Now().Format(time.RFC3339Nano))
			}
			// only Site
			dest := transmission.Destination
			if "prisma.tms.moc.Site" != dest.Type {
				continue
			}
			// get site
			site, err := siteDb.FindById(ctx, dest.Id)
			if err != nil {
				log.Error(err.Error()+"%v", dest)
				return
			}
			// get incident
			msg, err := tmsg.Unpack(mc.Payload)
			incident, ok := msg.(*moc.Incident)
			if !ok {
				log.Error(err.Error()+"%v", incident)
				return
			}
			// find wait group delta, wait for last log and lsat transmission update
			delta := 2
			for _, entry := range incident.Log {
				if entry.Attachment != nil {
					delta++
				}
			}
			var wg sync.WaitGroup
			wg.Add(delta)
			// create packets
			for _, entry := range incident.Log {
				if entry.Attachment != nil {
					packet := tms.Packet{
						Name:  entry.Attachment.Name,
						State: tms.Transmission_Pending,
					}
					transmission.Packets = append(transmission.Packets, &packet)
					// get file
					fileID := bson.ObjectIdHex(entry.Attachment.Id)
					db := s.dbClient.DB()
					f, err := db.GridFS("fs").OpenId(fileID)
					if err != nil {
						s.dbClient.Release(db)
						log.Error(err.Error()+"%v", entry.Attachment)
						continue
					}
					buf := bytes.NewBuffer(nil)
					io.Copy(buf, f)
					f.Close()
					s.dbClient.Release(db)
					pbFile := moc.File{
						Metadata: entry.Attachment,
						Data:     buf.Bytes(),
					}
					// pack payload
					filePayload, err := tmsg.PackFrom(&pbFile)
					fileMessage := tms.TsiMessage{
						Source: s.tsiClient.Local(),
						Destination: []*tms.EndPoint{
							{
								Site: site.SiteId,
							},
						},
						WriteTime: tms.Now(),
						SendTime:  tms.Now(),
						Body:      filePayload,
					}
					sendMessage(ctx, wg, transmission, &packet, &fileMessage, s.tsiClient, transDb, s.rpool)
				}
			}
			// incident message, log packet
			message := tms.TsiMessage{
				Source: s.tsiClient.Local(),
				Destination: []*tms.EndPoint{
					{
						Site: site.SiteId,
					},
				},
				Body: mc.Payload,
			}
			logPacket := tms.Packet{
				Name:  "log",
				State: tms.Transmission_Pending,
			}
			transmission.Packets = append(transmission.Packets, &logPacket)
			sendMessage(ctx, wg, transmission, &logPacket, &message, s.tsiClient, transDb, s.rpool)
			if transmission.Status == nil || transmission.Status.Code == int32(tms.ResponseStatus_Unknown) {
				transmission.State = tms.Transmission_Partial
				transmission.Status = &tms.ResponseStatus{
					Code:    int32(tms.ResponseStatus_Acknowledged),
					Message: "transmission sent",
				}
			}
			// update multicast with transmissions and packets
			s.mutex.Lock()
			log.Debug("Update transmission %+v", transmission)
			err = transDb.Update(transmission)
			if err != nil {
				log.Error(err.Error(), err)
			}
			err = mcDb.Update(ctx, &mc)
			if err != nil {
				log.Error(err.Error()+"%v", mc)
			}
			s.mutex.Unlock()
			// end wait group, handle callbacks
			wg.Done()
		}
	case api.Status:
		if api.Status_InitialLoadDone == data {
			log.Info("multicast load done")
			return
		}
	default:
		log.Info("data is not supported: %v", data)
	}
}

// send message, update transmission
func sendMessage(ctx context.Context, wg sync.WaitGroup, transmission *tms.Transmission, packet *tms.Packet, message *tms.TsiMessage,
	tsiClient client.TsiClient, transDb *mongo.TransmissionDb, rpool *redis.Pool) {
	transmission.State = tms.Transmission_Partial
	message.NotifySent = tsiClient.NotifyID(func(r *routing.DeliveryReport) {
		wg.Wait()
		packetStateUpdate(r, packet)
		transDb.PacketStatus(transmission.Id, packet)
	})
	rpool.Get().Do("SETEX", message.NotifySent, expiration.Seconds(), transmission.Id)
	packet.MessageId = fmt.Sprint(message.NotifySent)
	tsiClient.Send(ctx, message)
	wg.Done()
}

func packetStateUpdate(r *routing.DeliveryReport, packet *tms.Packet) {
	if packet.State == tms.Transmission_Success {
		return
	}
	switch r.Status {
	case routing.DeliveryReport_SENT:
		{
			packet.State = tms.Transmission_Partial
			packet.Status = &tms.ResponseStatus{
				Code:    int32(tms.ResponseStatus_Acknowledged),
				Message: "transmission sent",
			}
		}
	case routing.DeliveryReport_FAILED:
		{
			packet.State = tms.Transmission_Failure
			packet.Status = &tms.ResponseStatus{
				Code:    int32(tms.ResponseStatus_BadRequest),
				Message: "transmission send failure",
			}
		}
	case routing.DeliveryReport_UNKNOWN:
		{
			packet.State = tms.Transmission_Partial
			packet.Status = &tms.ResponseStatus{
				Code:    int32(tms.ResponseStatus_Unknown),
				Message: "transmission send unknown",
			}
		}
	case routing.DeliveryReport_PROCESSED:
		{
			packet.State = tms.Transmission_Success
			packet.Status = &tms.ResponseStatus{
				Code:    int32(tms.ResponseStatus_Success),
				Message: "transmission processed",
			}
		}
	}
}

// start not used
func (s *multicastStage) start() {}

// analyze not used
func (s *multicastStage) analyze(update api.TrackUpdate) error {
	return nil
}

func (s *multicastStage) processFailedTransfer(ctx gogroup.GoGroup, dbClient *mongo.MongoClient, tr *tms.Transmission, iPacket int) error {
	siteDB := mongo.NewSiteDb(ctx)
	// local site info
	localSite := moc.Site{
		SiteId: tmsg.GClient.Local().Site,
	}
	err := siteDB.FindBySiteId(ctx, &localSite)
	if err != nil {
		log.Error("error for getting info about localsite %s", err.Error())
	}
	remoteSite, err := siteDB.FindById(ctx, tr.Destination.Id)
	if err != nil {
		return fmt.Errorf("error for getting info about remote %s", err.Error())
	}
	multicastDB := mongo.NewMulticastDb(ctx)
	mc, err := multicastDB.Find(ctx, tr.ParentId)
	if err != nil {
		return err
	}
	mcMessage, err := tmsg.Unpack(mc.Payload)
	if err != nil {
		return err
	}
	incidentMC, ok := mcMessage.(*moc.Incident)
	if !ok {
		return fmt.Errorf("currepted data, expected multicast: %v", log.Spew(mcMessage))
	}
	miscDB := mongo.NewMongoMiscData(ctx, dbClient)
	resp, err := miscDB.Get(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: IncidentObjectType,
			Obj: &db.GoObject{
				ID: incidentMC.Id,
			},
		},
		Ctxt: ctx,
		Time: &db.TimeKeeper{},
	})
	if err != nil {
		return err
	}
	if len(resp) < 1 {
		return errors.New("an incident was not found")
	}
	incident, ok := resp[0].Contents.Data.(*moc.Incident)
	if !ok {
		return errors.New("corrupted data")
	}
	incident.State = moc.Incident_Open
	incident.Log = append(incident.Log, []*moc.IncidentLogEntry{
		{
			Type: "ACTION_" + moc.Incident_TRANSFER_FAIL.String(),
			Entity: &moc.EntityRelationship{
				Type: "multicast",
				Id:   tr.ParentId,
			},
			Timestamp: tms.Now(),
			Note: fmt.Sprintf("Transfer from Site %s(%d) to Site %s(%d) failed.\nReason: %s",
				localSite.Name, localSite.SiteId,
				remoteSite.Name, remoteSite.SiteId,
				tr.Packets[iPacket].Status.Message),
		},
		{
			Type:      "ACTION_" + moc.Incident_REOPEN.String(),
			Timestamp: tms.Now(),
		},
	}...)
	_, err = miscDB.Upsert(db.GoMiscRequest{
		Req: &db.GoRequest{
			Obj: &db.GoObject{
				ID:   incident.Id,
				Data: incident,
			},
		},
	})
	return err
}
