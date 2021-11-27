package db

/* FIXME: Remove once verified to be dead code
import (
	. "prisma/tms"
	. "prisma/tms/client_api"
	"prisma/tms/log"
	. "prisma/tms/routing"
	. "prisma/tms/tmsg/client"

	"errors"
	"prisma/gogroup"
	"sync"
	"time"
)

type LiveDB struct {
	client    TsiClient
	ctxt      gogroup.GoGroup
	counter   uint64
	lastPrint time.Time

	lock    sync.Mutex
	streams map[chan<- *Track]bool
	reg     chan chan *Track
	unreg   chan chan *Track

	tracks map[string]*Track
}

func NewLiveTrackDB(ctxt gogroup.GoGroup, client TsiClient) *LiveDB {
	db := &LiveDB{
		ctxt:      ctxt,
		client:    client,
		counter:   0,
		lastPrint: time.Now(),

		streams: make(map[chan<- *Track]bool),
		reg:     make(chan chan *Track, 0),
		unreg:   make(chan chan *Track, 1024),

		tracks: make(map[string]*Track),
	}
	go db.process()
	return db
}

func (db *LiveDB) getTracks() []*Track {
	db.lock.Lock()
	defer db.lock.Unlock()
	ret := make([]*Track, 0, len(db.tracks))
	for _, track := range db.tracks {
		ret = append(ret, track)
	}
	return ret
}

func (db *LiveDB) process() {
	log.Debug("LiveDB recieving tracks...")

	targetStream := db.client.Listen(db.ctxt, Listener{
		MessageType: "prisma.tms.Target",
	})
	trackStream := db.client.Listen(db.ctxt, Listener{
		MessageType: "prisma.tms.Track",
	})

	done := db.ctxt.Done()
	for {
		select {
		case <-done:
			return
		case reg := <-db.reg:
			db.streams[reg] = true
			go sendInit(reg, db.getTracks())
		case unreg := <-db.unreg:
			delete(db.streams, unreg)
			unreg <- nil

		case target := <-targetStream:
			db.processTarget(target)

		case track := <-trackStream:
			db.processTrack(track)
		}
	}
}

func sendInit(ch chan *Track, tracks []*Track) {
	for _, track := range tracks {
		ch <- track
	}
}

func (db *LiveDB) processTrack(tmsg *TMsg) {
	track, isTrack := tmsg.Body.(*Track)
	if !isTrack {
		log.Warn("LiveDB got something that ain't a track!")
		return
	}

	for _, meta := range track.Metadata {
		if track.Id != "" {
			db.processMetadata(track.Id, meta)
		}
	}

	for _, target := range track.Targets {
		db.processTarget(&TMsg{
			Body: target,
		})
	}
}

func (db *LiveDB) processMetadata(tid string, meta *TrackMetadata) {
	db.lock.Lock()
	defer db.lock.Unlock()
	track, ok := db.tracks[tid]
	if !ok {
		track = &Track{
			Id: tid,
		}
		db.tracks[tid] = track
	}
	track.Metadata = []*TrackMetadata{meta}
}

func (db *LiveDB) processTarget(tmsg *TMsg) {
	target, isTarget := tmsg.Body.(*Target)
	if !isTarget {
		log.Warn("LiveDB got something that ain't a target!")
		return
	}

	id := target.TrackId
	db.lock.Lock()
	track, ok := db.tracks[id]
	if !ok {
		track = &Track{
			Id: target.TrackId,
		}
		db.tracks[id] = track
	}
	db.lock.Unlock()
	track.Targets = []*Target{target}

	db.counter++
	since := time.Since(db.lastPrint)
	if since.Seconds() > 5 {
		log.Debug("Updates per second: %v", float64(db.counter)/since.Seconds())
		db.lastPrint = time.Now()
		db.counter = 0
	}

	done := db.ctxt.Done()
	for ch, _ := range db.streams {
		select {
		case <-done:
			return
		case ch <- track:
			// Do nothing!
		}
	}
}

func (db *LiveDB) Get(req GoTrackRequest) (<-chan TrackUpdate, error) {
	if req.Stream == false {
		//TODO #dirtyhack
		return nil, errors.New("GetTracks() unsupported")
	}
	return db.GetTrackStream(req)
}

func (db *LiveDB) GetTracks(
	req GoTrackRequest) (
	*Tracks, error) {

	tracks := Tracks{
		Tracks: db.getTracks(),
	}

	return &tracks, nil
}

func (db *LiveDB) GetTrackStream(
	req GoTrackRequest) (
	<-chan TrackUpdate, error) {

	out := make(chan TrackUpdate)

	go func() {
		done := req.Ctxt.Done()
		defer close(out)

		dbStream := make(chan *Track, 128)
		db.reg <- dbStream
		stopped := false
		for track := range dbStream {
			if !stopped {
				select {
				case <-done:
					db.unreg <- dbStream
					stopped = true
				case out <- TrackUpdate{
					Status: Status_Current,
					Track:  track,
				}:
					// pass
				}
			}
		}
	}()

	return out, nil
}

func (db *LiveDB) Insert(*Track) error {
	return errors.New("Unsupported")
}

func (db *LiveDB) GetHistoricalTrack(req GoHistoricalTrackRequest) (*Track, error) {
	return nil, errors.New("Unsupported")
}

func (db *LiveDB) GetPipeline(stages []TrackPipelineStage) (<-chan TrackUpdate, error) {
	return nil, errors.New("Unsupported")
}
*/
