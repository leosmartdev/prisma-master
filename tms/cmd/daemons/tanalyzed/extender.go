package main

import (
	"errors"
	"prisma/gogroup"
	"prisma/tms"
	api "prisma/tms/client_api"
	"prisma/tms/db"
	"prisma/tms/db/mongo"
	"prisma/tms/debug"
	"prisma/tms/devices"
	"prisma/tms/log"
	"prisma/tms/routing"
	"prisma/tms/tmsg"
	client "prisma/tms/tmsg/client"
	"prisma/tms/util/clock"
	"sync"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/golang/protobuf/ptypes"
)

// The current system only sends position reports for targets that have been
// updated within the past 15 minutes. When a target is being actively tracked
// and this 15 minute threshold is exceeded, a timeout message is issued to
// remove it from the map. This value was selected based on receiving regular
// position reports from AIS and radar.
//
// This is not appropriate for SARSAT beacon messages and Blueforce messages
// which have a much larger reporting interval. According to John K, a SARSAT
// position should not be removed until after 12 hours of inactivity.
// Blueforce positions should never timeout and always post
// the "last-known" position.
//
// When a C2 client connects to the system, the current state of the map is
// constructed by issuing a database query for all position reports in the
// past 15 minutes. Anything older which is excluded by the query is assumed
// to have timed out. Simply changing the 15 minute "timeout-window" does not
// work since constructing Blueforce state would require a full table scan
// which is not practical.
//
// This analyzer stage handles this case by extending the validity of a
// track by inserting a repeater track. This is simply a repeat of the
// previous track report that is inserted if a normal update isn't recieved
// in 10 minutes. It will contine to extend the track until the desired
// timeout is recieved.
//
// All tracks eligible to be extended are monitored by this stage and
// stored in the database. On startup, the database is first cleaned of
// expired tracks and then the initial state is loaded.

type trackExtenderStage struct {
	ctxt      gogroup.GoGroup
	db        db.TrackExDb
	trackDb   db.TrackDB
	watching  map[string]tms.TrackExtension
	mutex     sync.RWMutex
	tracer    *log.Tracer
	clk       clock.C
	tsiClient client.TsiClient

	extensionAt time.Duration
	timeouts    map[devices.DeviceType]time.Duration
	watchTicker *time.Ticker

	trackExReqStream <-chan *client.TMsg

	// initilized is bool variable that gets set to true when the init function is done
	// only when initilized is true, we can analyze the stage
	initialized bool
}

func newTrackExtenderStage() *trackExtenderStage {
	s := &trackExtenderStage{
		watching:    make(map[string]tms.TrackExtension),
		tracer:      log.GetTracer("extender"),
		clk:         &clock.Real{},
		watchTicker: time.NewTicker(1 * time.Minute),
		tsiClient:   tmsg.GClient,

		extensionAt: 10 * time.Minute,
		timeouts: map[devices.DeviceType]time.Duration{
			devices.DeviceType_SARSAT:       12 * time.Hour,
			devices.DeviceType_OmnicomSolar: 12 * time.Hour,
			devices.DeviceType_OmnicomVMS:   12 * time.Hour,
			devices.DeviceType_Manual:       100 * 365 * 24 * time.Hour, // never timeout
		},
		initialized: false,
	}
	if debug.FastTimers {
		s.watchTicker = time.NewTicker(1 * time.Second)
		s.extensionAt = 20 * time.Second
		for k := range s.timeouts {
			s.timeouts[k] = 1 * time.Minute
		}
	}
	return s
}

func (s *trackExtenderStage) init(ctxt gogroup.GoGroup, client *mongo.MongoClient) error {
	s.ctxt = ctxt
	s.db = mongo.NewTrackExDb(ctxt, client)
	s.trackDb = mongo.NewMongoTracks(ctxt, client)
	removed, err := s.db.Startup()
	if err != nil {
		return err
	}
	log.Info("%+v trackex documents removed", removed)
	prev, err := s.db.Get()
	if err != nil {
		return err
	}
	for _, ex := range prev {
		s.watching[ex.Track.Id] = ex
		err := s.extend(ex)
		if err != nil {
			log.Error("init cannot not extend tracks %+v", err)
		}
	}
	if len(s.watching) > 0 {
		log.Info("%v previous extensions still active", len(s.watching))
	}
	ctx := s.ctxt.Child("TrackExReq stream")
	s.trackExReqStream = s.tsiClient.Listen(ctx, routing.Listener{
		MessageType: "prisma.tms.TrackExReq",
	})
	s.initialized = true
	return nil
}

func (s *trackExtenderStage) start() {
	go s.watch()
	go s.extendReq()
}

func (s *trackExtenderStage) extendReq() {
	for {
		select {
		case <-s.ctxt.Done():
			return
		default:
			tmsg := <-s.trackExReqStream
			report, ok := tmsg.Body.(*tms.TrackExReq)
			if !ok {
				log.Error("Problem with TrackExReq Stream")
				continue
			}
			log.Info("track extention request received %+v", report)
			if report.Track == nil {
				track, err := s.trackDb.GetLastTrack(bson.M{"registry_id": report.RegistryId})
				if err != nil {
					log.Error("can not resolve track with registry id %+v", report.RegistryId)
					continue
				}
				report.Track = track
			}
			s.mutex.Lock()
			tex, ok := s.watching[report.Track.Id]
			s.mutex.Unlock()
			if !ok {
				err := s.startWatching(report.Track, report.Count)
				if err != nil {
					log.Error("trackex failed with err: %+v", err)
				}
			} else {
				if tex.Count+report.Count > 0 || s.clk.Now().Before(tex.Expires) {
					if err := s.update(report.Track, report.Count); err != nil {
						log.Error("trackex failed with err: %+v", err)
					}
				} else {
					if err := s.timeout(report.Track); err != nil {
						log.Error("trackex stop watching failed with err: %+v", err)
					}
				}
			}
		}

	}
}

func (s *trackExtenderStage) analyze(update api.TrackUpdate) error {

	if s.initialized == false || update.Track == nil || update.Track.Targets == nil || len(update.Track.Targets) == 0 {
		return nil
	}
	tgt := update.Track.Targets[0]
	if _, ok := s.timeouts[tgt.Type]; !ok {
		return nil
	}

	if update.Status == api.Status_Timeout {
		s.tracer.Logf("timeout update status: %v", update.Track.Id)
		// you can send out a time out when it's a timeout notice, you have to just stop watching.
		return s.stopWatching(update.Track)
	} else if _, exists := s.watching[update.Track.Id]; exists {
		if tgt.Repeat {
			return nil
		}
		s.tracer.Logf("update: %v", update)
		return s.update(update.Track, 0)
	}
	s.tracer.Logf("start: %v", update.Track.Id)
	return s.startWatching(update.Track, 0)
}

func (s *trackExtenderStage) startWatching(t *tms.Track, count int32) error {
	now := s.clk.Now()
	ex := tms.TrackExtension{
		Track:   t,
		Updated: now,
		Next:    now.Add(s.extensionAt),
		Expires: now.Add(s.timeouts[t.Targets[0].Type]),
		Count:   count,
	}

	s.mutex.Lock()
	s.watching[t.Id] = ex
	s.mutex.Unlock()

	return s.db.Insert(ex)
}

func (s *trackExtenderStage) update(t *tms.Track, count int32) error {
	now := s.clk.Now()
	ex := tms.TrackExtension{}
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if count != 0 {
		ex = s.watching[t.Id]
		ex.Count = ex.Count + count
	} else {
		ex = tms.TrackExtension{
			Track:   t,
			Updated: now,
			Next:    now.Add(s.extensionAt),
			Expires: now.Add(s.timeouts[t.Targets[0].Type]),
			Count:   s.watching[t.Id].Count,
		}
	}

	s.watching[t.Id] = ex

	return s.db.Update(ex)
}

func (s *trackExtenderStage) timeout(t *tms.Track) error {
	s.tracer.Logf("timeout: %v", t.Id)

	if err := s.stopWatching(t); err != nil {
		return err
	}
	t.Targets[0].Repeat = false

	body, err := tmsg.PackFrom(t)
	if err != nil {
		return err
	}
	now := s.clk.Now()
	pnow, err := ptypes.TimestampProto(now)
	if err != nil {
		return err
	}
	s.tsiClient.Send(s.ctxt, &tms.TsiMessage{
		Source: s.tsiClient.Local(),
		Destination: []*tms.EndPoint{
			{
				Site: tmsg.GClient.ResolveSite(""),
			},
			{
				Site: tmsg.TMSG_HQ_SITE,
			},
		},
		WriteTime: pnow,
		SendTime:  pnow,
		Body:      body,
	})
	return err
}

func (s *trackExtenderStage) stopWatching(t *tms.Track) error {
	s.mutex.Lock()
	delete(s.watching, t.Id)
	s.mutex.Unlock()
	s.tracer.Logf("Delete: %v", t.Id)

	if len(t.Targets) == 0 {
		// the record doesn't have any targets,
		// so it is not useful and we will not continue handle it
		return errors.New("target are not found")
	}
	log.Debug("Deleting watched track")
	return s.db.Remove(t.Id)
}

func (s *trackExtenderStage) watch() {
	for {
		select {
		case <-s.watchTicker.C:
			s.check()
		case <-s.ctxt.Done():
			return
		}
	}
}

func (s *trackExtenderStage) check() {
	s.mutex.RLock()
	watchingCopy := make(map[string]tms.TrackExtension)
	for k, v := range s.watching {
		watchingCopy[k] = v
	}
	s.mutex.RUnlock()
	now := s.clk.Now()

	for _, ex := range watchingCopy {
		if now.After(ex.Expires) && ex.Count == 0 {
			s.tracer.Logf("EXPIRE: %v", ex.Track.Id)
			s.timeout(ex.Track)
		} else if now.After(ex.Next) || ex.Count > 0 {
			s.tracer.Logf("EXTEND: %v", ex.Track.Id)
			s.extend(ex)
		}
	}
}

func (s *trackExtenderStage) extend(ex tms.TrackExtension) error {
	now := s.clk.Now()
	pnow, err := ptypes.TimestampProto(now)
	if err != nil {
		return err
	}
	t := ex.Track
	if len(t.Targets) == 0 {
		// the record doesn't have any targets,
		// so it is not useful and we will not continue handle it
		return errors.New("target are not found")
	}
	t.Targets[0].Repeat = true

	body, err := tmsg.PackFrom(t)
	if err != nil {
		return err
	}

	update := tms.TrackExtension{
		Track:   t,
		Updated: ex.Updated,
		Next:    now.Add(s.extensionAt),
		Expires: ex.Expires,
		Count:   ex.Count,
	}
	s.mutex.Lock()
	s.watching[t.Id] = update
	s.mutex.Unlock()

	err = s.db.Update(update)
	if err != nil {
		return err
	}

	m := &tms.TsiMessage{
		Source: s.tsiClient.Local(),
		Destination: []*tms.EndPoint{
			{
				Site: tmsg.GClient.ResolveSite(""),
			},
			{
				Site: tmsg.TMSG_HQ_SITE,
			},
		},
		WriteTime: pnow,
		SendTime:  pnow,
		Body:      body,
	}
	s.tsiClient.Send(s.ctxt, m)
	return err
}
