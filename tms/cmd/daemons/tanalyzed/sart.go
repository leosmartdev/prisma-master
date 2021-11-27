package main

import (
	"prisma/gogroup"
	api "prisma/tms/client_api"
	"prisma/tms/db/mongo"
	"prisma/tms/moc"
	"prisma/tms/util/ais"
	"prisma/tms/util/ident"
	"strings"
)

type sartStage struct {
	n Notifier
}

func newSartStage(n Notifier) *sartStage {
	return &sartStage{n: n}
}

func (s *sartStage) init(ctxt gogroup.GoGroup, client *mongo.MongoClient) error {
	return nil
}

func (s *sartStage) start() {}

func (s *sartStage) analyze(update api.TrackUpdate) error {
	if update.Track == nil || len(update.Track.Targets) == 0 {
		return nil
	}
	target := update.Track.Targets[0]
	if target.Nmea == nil || target.Nmea.Vdm == nil || target.Nmea.Vdm.M1371 == nil {
		return nil
	}
	data := target.Nmea.GetVdm().GetM1371()
	mmsi := ais.FormatMMSI(int(data.GetMmsi()))

	if !strings.HasPrefix(mmsi, "97") {
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

	noticeID := ident.
		With("event", moc.Notice_Sart).
		With("mmsi", mmsi).
		Hash()

	notice := &moc.Notice{
		NoticeId: noticeID,
		Event:    moc.Notice_Sart,
		Priority: moc.Notice_Alert,
		Target:   TargetInfoFromTrack(update.Track),
	}
	if err := s.n.Notify(notice, ruleMatch); err != nil {
		return err
	}
	return nil
}
