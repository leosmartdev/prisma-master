package main

import (
	"fmt"
	"strconv"
	"strings"

	"prisma/gogroup"
	"prisma/tms"
	api "prisma/tms/client_api"
	"prisma/tms/db"
	"prisma/tms/db/mongo"
	inc "prisma/tms/incident"
	"prisma/tms/log"
	"prisma/tms/moc"
	"prisma/tms/routing"
	"prisma/tms/tmsg"
	client "prisma/tms/tmsg/client"
	"prisma/tms/util/ident"
)

const (
	IncidentObjectType = "prisma.tms.moc.Incident"
)

type incidentStage struct {
	n         Notifier
	tsiClient client.TsiClient
	config    *mongo.Configuration
}

func newIncidentStage(tsiClient client.TsiClient, notifier Notifier) *incidentStage {
	return &incidentStage{
		tsiClient: tsiClient,
		n:         notifier,
	}
}

func (s *incidentStage) init(ctx gogroup.GoGroup, dbClient *mongo.MongoClient) error {
	log.Info("incident init")
	configDb := mongo.ConfigDb{}
	config, err := configDb.Read(ctx)
	siteDb := mongo.NewSiteDb(ctx)
	if err != nil {
		log.Error("config %v", err)
	}
	s.config = config
	miscDB := mongo.NewMongoMiscData(ctx, dbClient)
	stream := miscDB.GetPersistentStream(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: IncidentObjectType,
		},
		Ctxt: ctx,
		Time: &db.TimeKeeper{},
	}, nil, nil)
	ctx.Go(func() {
		for {
			select {
			case update, ok := <-stream:
				if !ok {
					continue // channel was closed
				}
				if api.Status_InitialLoadDone == update.Status {
					log.Info("incident initial load done")
					continue
				}
				if update.Contents == nil || update.Contents.Data == nil {
					log.Error("no content: %v", update)
					continue
				}
				incident, ok := update.Contents.Data.(*moc.Incident)
				if !ok {
					log.Error("bad content: %v", update)
					continue
				}
				if incident.State != moc.Incident_Transferring {
					continue
				}
				// get tmsg notify id, set in tms/db/inserter.go
				msgPart := strings.Fields(incident.Log[len(incident.Log)-1].Note)
				var notifyId int32
				var siteId uint32
				var err error
				if len(msgPart) == 2 {
					sId, err := strconv.ParseInt(msgPart[0], 10, 32)
					if err == nil {
						siteId = uint32(sId)
					}
					nId, err := strconv.ParseInt(msgPart[1], 10, 32)
					if err == nil {
						notifyId = int32(nId)
					}
				}
				// update log entry with incident.IncidentId
				prevIncidentId := incident.IncidentId
				incident.Log[len(incident.Log)-1].Note = prevIncidentId
				incident.IncidentId = inc.IdCreatorInstance(ctx).Next(s)
				incident.State = moc.Incident_Transferred
				incident.Assignee = ""
				incident.Commander = ""
				incident.Outcome = ""
				incident.Synopsis = ""
				// site info
				site := moc.Site{
					SiteId: siteId,
				}
				err = siteDb.FindBySiteId(ctx, &site)
				if err != nil {
					log.Error(err.Error()+"%v", site)
				}
				// local site info
				localSite := moc.Site{
					SiteId: s.tsiClient.Local().Site,
				}
				err = siteDb.FindBySiteId(ctx, &localSite)
				if err != nil {
					log.Error(err.Error()+"%v", localSite)
				}
				incident.Log = append(incident.Log, &moc.IncidentLogEntry{
					Id:        mongo.CreateId(),
					Timestamp: tms.Now(),
					Type:      "ACTION_" + moc.Incident_TRANSFER_RECEIVE.String(),
					Entity: &moc.EntityRelationship{
						Type: "prisma.tms.moc.Site",
						Id:   site.Id,
					},
					Note: fmt.Sprintf("Incident %v received at %v (%v) from %v (%v)", prevIncidentId, localSite.Name, localSite.SiteId, site.Name, site.SiteId),
				})
				incident.Id = update.Contents.ID
				// update
				_, err = miscDB.Upsert(db.GoMiscRequest{
					Req: &db.GoRequest{
						ObjectType: IncidentObjectType,
						Obj: &db.GoObject{
							ID:   incident.Id,
							Data: incident,
						},
					},
					Ctxt: ctx,
				})
				if err != nil {
					log.Error(err.Error()+"%v", err)
					// send fail
					drAny, _ := tmsg.PackFrom(&routing.DeliveryReport{
						NotifyId: notifyId,
						Status:   routing.DeliveryReport_FAILED,
					})
					s.tsiClient.Send(ctx, &tms.TsiMessage{
						Destination: []*tms.EndPoint{
							{
								Site: siteId,
							},
						},
						Body: drAny,
					})
					continue
				}
				// send success
				drAny, _ := tmsg.PackFrom(&routing.DeliveryReport{
					NotifyId: notifyId,
					Status:   routing.DeliveryReport_PROCESSED,
				})
				s.tsiClient.Send(ctx, &tms.TsiMessage{
					Destination: []*tms.EndPoint{
						{
							Site: siteId,
						},
					},
					Body: drAny,
				})
				log.Info("incident processed %v", incident.IncidentId)
				// notify
				s.n.Notify(&moc.Notice{
					NoticeId: ident.With("incidentId", incident.IncidentId).Hash(),
					Event:    moc.Notice_IncidentTransfer,
					Priority: moc.Notice_Info,
					Source: &moc.SourceInfo{
						IncidentId: incident.Id,
						Name:       incident.IncidentId,
					},
				}, true)
			case <-ctx.Done():
				return
			}
		}
	})
	log.Info("incident init done")
	return nil
}

// start not used
func (s *incidentStage) start() {}

// analyze not used
func (s *incidentStage) analyze(update api.TrackUpdate) error {
	return nil
}

func (s *incidentStage) Prefix() string {
	var prefix string
	if nil != s.config && nil != s.config.Site {
		prefix = s.config.Site.IncidentIdPrefix
	}
	return prefix
}
