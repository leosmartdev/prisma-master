package main

import (
	"fmt"
	"prisma/gogroup"
	"prisma/tms"
	api "prisma/tms/client_api"
	"prisma/tms/db"
	"prisma/tms/db/mongo"
	"prisma/tms/log"
	"prisma/tms/moc"
	"prisma/tms/sar"
	"prisma/tms/util/ident"
	"time"
)

type sarsatStage struct {
	n      Notifier
	ctxt   gogroup.GoGroup
	miscDb db.MiscDB
}

func newSarsatStage(n Notifier, ctxt gogroup.GoGroup, client *mongo.MongoClient) *sarsatStage {
	return &sarsatStage{
		n:      n,
		ctxt:   ctxt,
		miscDb: mongo.NewMongoMiscData(ctxt, client),
	}
}

func (s *sarsatStage) init(ctxt gogroup.GoGroup, client *mongo.MongoClient) error {
	go s.watchActivity()
	return nil
}

func (s *sarsatStage) start() {}

func (s *sarsatStage) analyze(update api.TrackUpdate) error {
	if update.Track.Targets == nil || len(update.Track.Targets) == 0 {
		return nil
	}
	tgt := update.Track.Targets[0]
	sarmsg := tgt.Sarmsg
	if sarmsg == nil {
		return nil
	}

	var ruleMatch bool
	switch update.Status {
	case api.Status_Current:
		ruleMatch = true
	case api.Status_Timeout:
		ruleMatch = false
	default:
		return nil
	}

	if sarmsg.SarsatAlert == nil {

		noticeID := ident.
			With("event", moc.Notice_SarsatDefaultHandling).
			With("target", tgt.Id).
			Hash()

		notice := &moc.Notice{
			NoticeId: noticeID,
			Event:    moc.Notice_SarsatDefaultHandling,
			Priority: moc.Notice_Alert,
			Target:   &moc.TargetInfo{TrackId: update.Track.Id, Type: tgt.Type.String()},
		}

		if err := s.n.Notify(notice, ruleMatch); err != nil {
			return err
		}
		return nil
	}

	beaconID := sarmsg.SarsatAlert.Beacon.HexId
	noticeID := ident.
		With("event", moc.Notice_Sarsat).
		With("beacon", beaconID).
		Hash()

	notice := &moc.Notice{
		NoticeId: noticeID,
		Event:    moc.Notice_Sarsat,
		Priority: moc.Notice_Alert,
		Target:   TargetInfoFromTrack(update.Track),
	}
	if err := s.n.Notify(notice, ruleMatch); err != nil {
		return err
	}
	return nil
}

func (s *sarsatStage) watchActivity() {
	activitydb := s.miscDb
	for {
		activityUpdates := activitydb.GetPersistentStream(db.GoMiscRequest{
			Req: &db.GoRequest{
				ObjectType: "prisma.tms.MessageActivity",
			},
			Ctxt: s.ctxt,
		}, nil, nil)
		log.Info("sarsat: watching message activity")
		done := false
		for !done {
			select {
			case update, ok := <-activityUpdates:
				if !ok {
					log.Error("sarsat: activity channel closed")
					done = true
				} else if update.Contents != nil {
					if err := s.analyzeActivity(update); err != nil {
						log.Error("sarsat: error analyzing activity update: %v", err)
					}
				}
			case <-s.ctxt.Done():
				return
			}
		}
		// Wait a bit before retrying to connect to the database
		time.Sleep(time.Second)
	}
}

func (s *sarsatStage) analyzeActivity(update db.GoGetResponse) error {
	activity, ok := update.Contents.Data.(*tms.MessageActivity)
	if !ok {
		return fmt.Errorf("Not an activity stream %+v", update.Contents)
	}
	sarmsg := activity.GetSarsat()
	if sarmsg == nil {
		return nil
	}
	if sarmsg.MessageType != sar.SarsatMessage_SIT_915 {
		return nil
	}

	notice := &moc.Notice{
		NoticeId: update.Contents.ID,
		Event:    moc.Notice_SarsatMessage,
		Priority: moc.Notice_Message,
		Source: &moc.SourceInfo{
			Name: sarmsg.RemoteName,
		},
		Target: &moc.TargetInfo{
			Message: sarmsg.NarrativeText,
		},
	}
	fmt.Printf("sending notice: %+v\n", notice)
	s.n.Notify(notice, true)
	return nil
}
