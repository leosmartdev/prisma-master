package main

import (
	"prisma/tms"
	log "prisma/tms/log"

	"sync"
)

type ReportDB interface {
	// Save a bunch of tracks into the database
	Save(*tms.Track)

	// Assemble a track report from the database, marking them as in a report
	Next() *tms.TrackReport

	// Mark a report as failed delivery, return tracks to DB
	Fail(*tms.TrackReport)

	// Mark a report as successful delivery, remove from database
	Delivered(*tms.TrackReport)
}

type MemReportDB struct {
	sync.Mutex

	conf ReportConf

	tracksToSend    []*tms.Track // TODO: Make this a heap which sorts by time of the track
	tracksBeingSent map[*tms.Track]struct{}
}

func NewMemReportDB(conf ReportConf) *MemReportDB {
	db := &MemReportDB{
		conf:            conf,
		tracksBeingSent: make(map[*tms.Track]struct{}),
	}
	return db
}

// Save a bunch of tracks into the database
func (db *MemReportDB) Save(track *tms.Track) {
	log.TraceMsg("Saving track to report db: %+v", track)

	db.Lock()
	defer db.Unlock()

	db.tracksToSend = append(db.tracksToSend, track)
}

// Assemble a track report from the database, marking them as in a report
func (db *MemReportDB) Next() *tms.TrackReport {
	rpt := tms.TrackReport{}

	db.Lock()
	defer db.Unlock()

	for i := uint(0); i < db.conf.ReportSize && len(db.tracksToSend) > 0; i++ {
		t := db.tracksToSend[len(db.tracksToSend)-1]
		rpt.Tracks = append(rpt.Tracks, t)
		db.tracksToSend = db.tracksToSend[0 : len(db.tracksToSend)-1]
		db.tracksBeingSent[t] = struct{}{}
	}

	if len(rpt.Tracks) > 0 {
		return &rpt
	}
	return nil
}

// Mark a report as failed delivery, return tracks to DB
func (db *MemReportDB) Fail(rpt *tms.TrackReport) {
	db.Lock()
	defer db.Unlock()

	for _, track := range rpt.Tracks {
		_, ok := db.tracksBeingSent[track]
		if !ok {
			panic("Track in failed report wasn't queued for sending!")
		}

		delete(db.tracksBeingSent, track)
		db.tracksToSend = append(db.tracksToSend, track)

		log.TraceMsg("Failing track to report db: %+v", track)
	}
}

// Mark a report as successful delivery, remove from database
func (db *MemReportDB) Delivered(rpt *tms.TrackReport) {
	db.Lock()
	defer db.Unlock()

	for _, track := range rpt.Tracks {
		_, ok := db.tracksBeingSent[track]
		if !ok {
			panic("Track in failed report wasn't queued for sending!")
		}

		delete(db.tracksBeingSent, track)

		log.TraceMsg("Removing sent track to report db: %+v", track)
	}
}
