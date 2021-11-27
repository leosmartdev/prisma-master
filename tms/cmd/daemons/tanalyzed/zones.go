package main

import (
	"prisma/gogroup"
	"prisma/tms"
	api "prisma/tms/client_api"
	"prisma/tms/db"
	"prisma/tms/db/mongo"
	"prisma/tms/devices"
	"prisma/tms/geojson/rtree"
	"prisma/tms/log"
	"prisma/tms/moc"
	"prisma/tms/util/ident"
	"sync"
	"time"

	"github.com/globalsign/mgo/bson"
)

var replayZoneFilters = bson.M{"utime": bson.M{"$gt": time.Now().Add(-24 * time.Hour)}}

type zoneStage struct {
	n           Notifier
	ctxt        gogroup.GoGroup
	zoneStream  <-chan db.GoGetResponse
	zones       map[string]*moc.Zone
	relAreaZone map[string]map[string]bool // it is used for looking for a zone by registry_id or track_id for O(1)
	members     map[string]map[string]*tms.Track
	mutex       sync.RWMutex
	tracer      *log.Tracer
	miscDB      db.MiscDB
	initialized bool
}

func newZoneStage(n Notifier) *zoneStage {
	return &zoneStage{
		n:           n,
		zones:       make(map[string]*moc.Zone),
		members:     make(map[string]map[string]*tms.Track),
		relAreaZone: make(map[string]map[string]bool),
		tracer:      log.GetTracer("zones"),
		initialized: false,
	}
}

func (s *zoneStage) init(ctxt gogroup.GoGroup, client *mongo.MongoClient) error {
	s.tracer.Log("initial zones loading")
	s.ctxt = ctxt
	goRequest := db.GoRequest{
		ObjectType: "prisma.tms.moc.Zone",
	}
	miscRequest := db.GoMiscRequest{
		Req:  &goRequest,
		Ctxt: ctxt,
		Time: &db.TimeKeeper{},
	}
	s.miscDB = mongo.NewMongoMiscData(ctxt, client)
	s.zoneStream = s.miscDB.GetPersistentStream(miscRequest, replayZoneFilters, nil)
	done := false
	for !done {
		select {
		case update, ok := <-s.zoneStream:
			if !ok || update.Status == api.Status_InitialLoadDone {
				done = true
				break
			}
			s.updateZone(update)
		case <-ctxt.Done():
			return nil
		}
	}
	s.initialized = true
	s.tracer.Log("initial zones done")
	return nil
}

func (s *zoneStage) start() {
	go s.watch()
}

func (s *zoneStage) updateZone(update db.GoGetResponse) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if update.Status == api.Status_InitialLoadDone {
		return
	}
	if update.Contents == nil || update.Contents.Data == nil {
		log.Warn("unexpected empty content in zone update: %+v", update)
		return
	}
	zone, ok := update.Contents.Data.(*moc.Zone)
	if !ok {
		log.Warn("unexpected data in zone update: %+v", update)
		return
	}
	id := update.Contents.ID
	if update.Status == api.Status_Timeout {
		s.removeZone(id, zone)
		return
	}
	if zone.Area != nil && zone.Area.Radius > 0 {
		if s.relAreaZone[zone.GetAreaID()] == nil {
			s.relAreaZone[zone.GetAreaID()] = make(map[string]bool)
		}
		if _, ok = s.relAreaZone[zone.GetAreaID()][id]; !ok {
			s.relAreaZone[zone.GetAreaID()][id] = true
		}
	}
	s.zones[id] = zone
	// Rebuild the member list
	members := make(map[string]*tms.Track)
	s.members[id] = members

	if !zone.CreateAlertOnEnter && !zone.CreateAlertOnExit {
		return
	}
	s.tracer.Logf("zone update starting: %v", id)
	bbox := zone.GetBBox()

	// If the zone has changed, its shape may have changed. Check each
	// member to see if they are still in the zone
	prevMembers, exists := s.members[id]
	if exists {
		for _, track := range prevMembers {
			if !zone.Intersect(track.Point()) {
				s.exitZone(id, zone, track)
			}
		}
	}

	iter := func(i rtree.Item) bool {
		track := i.(*tms.Track)
		if zone.IsExcludedTrack(track) {
			return true
		}
		if !zone.BelongAreaToTrack(track) && zone.Intersect(track.Point()) {
			s.enterZone(id, zone, track)
		}
		return true
	}
	s.n.TrackTree().Search(bbox.Min.X, bbox.Min.Y, 0, bbox.Max.X, bbox.Max.Y, 0, iter)
	s.tracer.Logf("zone update complete: %v", id)
}

func (s *zoneStage) removeZone(id string, zone *moc.Zone) {
	if zone.CreateAlertOnEnter {
		for _, member := range s.members[id] {
			notice := s.enterZoneNotice(id, zone, member)
			s.n.Notify(notice, false)
		}
	}
	delete(s.zones, id)
	delete(s.members, id)
	delete(s.relAreaZone[zone.GetAreaID()], id)
	if len(s.relAreaZone[zone.GetAreaID()]) == 0 {
		delete(s.relAreaZone, zone.GetAreaID())
	}
}

// It is used for updateing a poly for zones, which relates to a track
func (s *zoneStage) updatePoliesByTrack(track *tms.Track) {
	id := track.RegistryId
	if len(track.Targets) == 0 {
		return
	}
	if id == "" && track.Id == "" {
		return
	} else if id == "" {
		id = track.Id
	}
	if arr, ok := s.relAreaZone[id]; ok {
		for zoneId := range arr {
			if s.zones[zoneId].Area == nil {
				continue
			}
			// generate points around the target above
			s.zones[zoneId].Area.Center = &tms.Point{
				Latitude:  track.Point().Coordinates.Y,
				Longitude: track.Point().Coordinates.X,
			}
			// Since the front-end(in particular geojson) is not able to understand "circle" shape
			goreq := db.GoMiscRequest{
				Req: &db.GoRequest{
					ObjectType: "prisma.tms.moc.Zone",
					Obj: &db.GoObject{
						Data: s.zones[zoneId],
					},
				},
				Ctxt: s.ctxt,
				Time: &db.TimeKeeper{},
			}
			goreq.Req.Obj.ID = zoneId
			if s.miscDB != nil {
				_, err := s.miscDB.Upsert(goreq)
				if err != nil {
					log.Error(err.Error())
				}
			}
		}
	}
}

func (s *zoneStage) analyze(update api.TrackUpdate) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	track := update.Track
	if len(track.Targets) == 0 || s.initialized == false {
		return nil
	}
	tgt := track.Targets[0]
	if tgt.Type == devices.DeviceType_SARSAT {
		return nil
	}
	s.updatePoliesByTrack(track)
	for id, z := range s.zones {
		mems := s.members[id]
		if update.Status == api.Status_Timeout {
			delete(s.members[id], track.LookupID())
			continue
		}
		// zones have excluded vessels so see on them to not generate alerts
		if z.IsExcludedTrack(update.Track) {
			continue
		}
		_, isMember := mems[track.LookupID()]
		intersects := z.Intersect(track.Point())
		if !z.BelongAreaToTrack(track) && intersects && !isMember {
			s.enterZone(id, z, track)
		}
		if !intersects && isMember {
			s.exitZone(id, z, track)
		}
	}
	return nil
}

func (s *zoneStage) watch() {
	s.tracer.Log("watching for zone updates")
	for {
		select {
		case update, ok := <-s.zoneStream:
			if !ok {
				log.Error("A connection was closed")
				return
			}
			s.updateZone(update)
		case <-s.ctxt.Done():
			s.tracer.Log("watcher canceled")
			return
		}
	}
}

func (s *zoneStage) enterZone(id string, zone *moc.Zone, track *tms.Track) {
	s.members[id][track.LookupID()] = track
	if zone.CreateAlertOnEnter {
		notice := s.enterZoneNotice(id, zone, track)
		s.n.Notify(notice, true)
	}
}

func (s *zoneStage) exitZone(id string, zone *moc.Zone, track *tms.Track) {
	delete(s.members[id], track.LookupID())
	if zone.CreateAlertOnExit {
		notice := s.exitZoneNotice(id, zone, track)
		s.n.Notify(notice, true)
	} else if zone.CreateAlertOnEnter {
		notice := s.enterZoneNotice(id, zone, track)
		s.n.Notify(notice, false)
	}
}

func (s *zoneStage) enterZoneNotice(id string, zone *moc.Zone, track *tms.Track) *moc.Notice {
	return &moc.Notice{
		NoticeId: noticeID(id, track, moc.Notice_EnterZone),
		Event:    moc.Notice_EnterZone,
		Priority: moc.Notice_Info,
		Source: &moc.SourceInfo{
			Name:   zone.Name,
			ZoneId: id,
		},
		Target: TargetInfoFromTrack(track),
	}
}

func (s *zoneStage) exitZoneNotice(id string, zone *moc.Zone, track *tms.Track) *moc.Notice {
	return &moc.Notice{
		NoticeId: noticeID(id, track, moc.Notice_ExitZone),
		Event:    moc.Notice_ExitZone,
		Priority: moc.Notice_Info,
		Source: &moc.SourceInfo{
			Name:   zone.Name,
			ZoneId: id,
		},
		Target: TargetInfoFromTrack(track),
	}
}

func noticeID(zoneID string, track *tms.Track, event moc.Notice_Event) string {
	return ident.
		With("event", event).
		With("zone", zoneID).
		With("track", track.LookupID()).
		Hash()
}
