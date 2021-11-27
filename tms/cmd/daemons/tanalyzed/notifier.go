package main

import (
	"fmt"
	"prisma/gogroup"
	"prisma/tms"
	api "prisma/tms/client_api"
	"prisma/tms/db"
	"prisma/tms/db/mongo"
	"prisma/tms/geojson/rtree"
	"prisma/tms/log"
	"prisma/tms/moc"
	"prisma/tms/util/clock"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/gomodule/redigo/redis"
)

const (
	redisNotifierName      = "notifier"
	updateMongoDBPerRecord = 100
	redisReloadSleepSecond = 5 * time.Second
)

// Notifier receives the evaluation results of the rules engine and determines
// if a notice should be dispatched. Call Update with the notice that may be
// sent and the result of the rule evaluation.
//
// The Action field in the notice should be left blank. Set the NoticeID to a
// hashed value that can be used to lookup this notice. This identifier is used
// on subsquent calls to see if this notice has already been sent. For example,
// a man overboard device might hash the event type "SART" and the MMSI
// of the vessel.
//
// A notice can be in one of three states:
//
// - NEW: When Update is called the first time with a true value, a notice is
// emitted and it enters the NEW state. Further calls to Update will adjust the
// UpdatedTime field if the rule evaulates to true.
// - ACK: When the notice is acknowledged by an end-user, the notice enters the
// ACK state. A call to Update with true will adjust the UpdatedTime field
// while a call with false will change the state to CLEAR.
// - CLEAR: Notifier stops tracking a notice when it enters this state.
// If Update is called again with a true value, a new notice is created. Any
// acknowledged notice that has not been updated in an hour cleared by
// a reaper goroutine.
//
// Notifier keeps redis hash-map of all notices it is actively tracking.
// This hash-map is repopulated on startup from the database.
type Notifier interface {
	Notify(note *moc.Notice, ruleMatch bool) error
	TrackMap() map[string]*tms.Track
	TrackTree() *rtree.RTree
}

type Notices struct {
	trackMap  map[string]*tms.Track
	trackTree *rtree.RTree

	ctxt       gogroup.GoGroup
	notifydb   db.NotifyDb
	mutex      sync.Mutex
	clk        clock.C
	stream     *db.NoticeStream
	tracer     *log.Tracer
	ackTimeout time.Duration
	reapTicker *time.Ticker
	redisPool  *redis.Pool

	activeReload bool
	muReload     sync.Mutex
}

func NewNotifier(ctxt gogroup.GoGroup, client *mongo.MongoClient, redisPool *redis.Pool) *Notices {
	n := &Notices{
		trackMap:   make(map[string]*tms.Track),
		trackTree:  rtree.New(),
		ctxt:       ctxt,
		notifydb:   mongo.NewNotifyDb(ctxt, client),
		clk:        &clock.Real{},
		tracer:     log.GetTracer("notifier"),
		ackTimeout: 1 * time.Hour,
		reapTicker: time.NewTicker(10 * time.Minute),
		redisPool:  redisPool,
	}
	return n
}

func (n *Notices) Init() error {
	if err := n.notifydb.Startup(); err != nil {
		return err
	}
	stream := n.notifydb.GetPersistentStream(nil)
	n.loadInitial(stream)
	n.stream = stream
	return nil
}

func (n *Notices) Start() {
	go n.watch(n.stream)
	go n.reaper()
}

// when redis will crash we need to reload data from mongodb
func (n *Notices) reloadRedis() {
	n.muReload.Lock()
	if n.activeReload {
		n.muReload.Unlock()
		return
	}
	n.activeReload = true
	n.muReload.Unlock()

	for {
		log.Debug("reload redis")
		select {
		case <-n.ctxt.Done():
			return
		default:
		}
		redisConn := n.redisPool.Get()
		if redisConn.Err() == nil {
			ctx := n.ctxt.Child("reload-redis")
			stream := n.notifydb.GetPersistentStreamWithGroupContext(ctx, nil)
			n.loadInitial(stream)
			ctx.Cancel(nil)
			n.muReload.Lock()
			n.activeReload = false
			n.muReload.Unlock()
			return

		}
		redisConn.Close()
		time.Sleep(redisReloadSleepSecond)
	}
}

func (n *Notices) safeCloseRedis(conn redis.Conn) {
	if conn.Err() != nil {
		n.ctxt.Go(n.reloadRedis)
	}
	conn.Close()
}

func (n *Notices) Notify(note *moc.Notice, ruleMatch bool) error {
	if note.NoticeId == "" {
		return fmt.Errorf("notice id cannot be blank: %+v", note)
	}
	prev := new(moc.Notice)
	redisConn := n.redisPool.Get()
	defer n.safeCloseRedis(redisConn)

	data, err := redis.String(redisConn.Do("HGET", getRedisKeyByNote(note), "data"))
	exists := data != "" && err == nil
	if err != nil && err != redis.ErrNil {
		log.Error(err.Error())
		// may be redis was closed, so try to take a look at mongodb to avoid creating a new notice
		if redisConn.Err() != nil {
			prev, err = n.notifydb.GetByNoticeId(note.NoticeId)
			exists = err == nil
			if err != nil && err != mongo.ErrNotFound {
				log.Error(err.Error())
			}
		}
	} else {
		if err := proto.UnmarshalText(data, prev); err != nil {
			log.Error(err.Error())
			exists = false
		}
	}
	if exists {
		defer func() {
			// TODO(Aleksandr Rassanov): AckMiss - del a note do we need to call this func?
			if _, err = redisConn.Do("HSET", getRedisKeyByNote(note), "data", prev.String()); err != nil {
				log.Error(err.Error())
			}
		}()
	}
	switch {
	case !exists:
		return n.fromClearState(note, ruleMatch)
	case prev.Action == moc.Notice_NEW:
		return n.fromNewState(prev, ruleMatch)
	case prev.Action == moc.Notice_ACK:
		return n.fromAckState(prev, ruleMatch)
	case prev.Action == moc.Notice_ACK_WAIT:
		return n.fromAckWaitState(prev, ruleMatch)
	}
	return err
}

func (n *Notices) updateTrack(update api.TrackUpdate) {
	switch update.Status {
	case api.Status_Current:
		prev, exists := n.trackMap[update.Track.Id]
		if exists && prev.HasPosition() {
			n.trackTree.Remove(prev)
		}
		n.trackMap[update.Track.Id] = update.Track
		if update.Track.HasPosition() {
			n.trackTree.Insert(update.Track)
		}
	case api.Status_Timeout:
		prev, exists := n.trackMap[update.Track.Id]
		if exists && prev.HasPosition() {
			n.trackTree.Remove(prev)
		}
		delete(n.trackMap, update.Track.Id)
	}
}

// Get all active notices from the database
func (n *Notices) loadInitial(stream *db.NoticeStream) {
	for {
		select {
		case update := <-stream.Updates:
			redisConn := n.redisPool.Get()
			n.tracer.Logf("previous: %+v", notestr(update))
			if _, err := redisConn.Do("HMSET", getRedisKeyByNote(update),
				"data", update.String(), "count_updates", 0); err != nil {
				log.Error(err.Error())
			}
			redisConn.Close()
		case <-stream.InitialLoadDone:
			n.tracer.Logf("end of previous notices")
			return
		case <-n.ctxt.Done():
			return
		}
	}
}

// Watch for notice acknowledgements
func (n *Notices) watch(stream *db.NoticeStream) {
	for {
		select {
		case update := <-stream.Updates:
			if update.Action != moc.Notice_ACK {
				continue
			}
			n.tracer.Logf("changed: %+v", notestr(update))
			n.receivedAck(update)
		case <-n.ctxt.Done():
			n.tracer.Logf("watcher canceled")
			return
		}
	}
}

func (n *Notices) receivedAck(note *moc.Notice) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	prev := new(moc.Notice)

	redisConn := n.redisPool.Get()
	defer n.safeCloseRedis(redisConn)

	redisNote, err := redis.String(redisConn.Do("HGET", getRedisKeyByNote(note), "data"))
	exists := redisNote != "" && err == nil
	if err := proto.UnmarshalText(redisNote, prev); err != nil {
		log.Error(err.Error())
		exists = false
	}
	if !exists {
		return fmt.Errorf("no such notice to ack: %v", note.NoticeId)
	}
	if prev.Action == moc.Notice_CLEAR {
		return fmt.Errorf("notice already cleared: %v", note.NoticeId)
	}
	if prev.Action == moc.Notice_ACK_WAIT {
		n.toClearState(note)
		return nil
	}
	_, err = redisConn.Do("HSET", getRedisKeyByNote(note), "data", note.String())
	return err
}

func (n *Notices) fromClearState(update *moc.Notice, ruleMatch bool) error {
	switch {
	case ruleMatch:
		return n.toNewState(update)
	case !ruleMatch:
		// nothing
		return nil
	}
	return nil
}

func (n *Notices) fromAckWaitState(update *moc.Notice, ruleMatch bool) error {
	switch {
	case ruleMatch:
		return n.toNewStateViaRenew(update)
	case !ruleMatch:
		// nothing
		return nil
	}
	return nil
}

func (n *Notices) fromNewState(prev *moc.Notice, ruleMatch bool) error {
	switch {
	case ruleMatch:
		return n.updateTime(prev)
	case !ruleMatch:
		return n.toAckWaitState(prev)
	}
	return nil
}

func (n *Notices) fromAckState(prev *moc.Notice, ruleMatch bool) error {
	switch {
	case ruleMatch:
		return n.updateTime(prev)
	case !ruleMatch:
		return n.toClearState(prev)
	}
	return nil
}

func (n *Notices) toNewState(note *moc.Notice) error {
	now, err := ptypes.TimestampProto(n.clk.Now())
	if err != nil {
		return err
	}
	note.Action = moc.Notice_NEW
	note.CreatedTime = now
	note.UpdatedTime = now
	n.tracer.Logf("created: %+v", notestr(note))
	if err := n.notifydb.Create(note); err != nil {
		return err
	}

	redisConn := n.redisPool.Get()
	defer n.safeCloseRedis(redisConn)

	_, err = redisConn.Do("HMSET", getRedisKeyByNote(note), "data", note.String(), "count_updates", 0)
	return err
}

func (n *Notices) toNewStateViaRenew(note *moc.Notice) error {
	now, err := ptypes.TimestampProto(n.clk.Now())
	if err != nil {
		return err
	}
	note.Action = moc.Notice_NEW
	note.UpdatedTime = now
	n.tracer.Logf("renewed: %+v", notestr(note))
	if err := n.notifydb.Renew(note.DatabaseId, now); err != nil {
		return err
	}

	redisConn := n.redisPool.Get()
	defer n.safeCloseRedis(redisConn)

	_, err = redisConn.Do("HMSET", getRedisKeyByNote(note), "data", note.String(), "count_updates", 0)
	return err
}

func (n *Notices) toAckWaitState(note *moc.Notice) error {
	now, err := ptypes.TimestampProto(n.clk.Now())
	if err != nil {
		return err
	}
	note.Action = moc.Notice_ACK_WAIT
	note.UpdatedTime = now
	n.tracer.Logf("changed: %+v", notestr(note))
	if err := n.notifydb.AckWait(note.DatabaseId, now); err != nil {
		return err
	}

	redisConn := n.redisPool.Get()
	defer n.safeCloseRedis(redisConn)

	if _, err := redisConn.Do("HSET", getRedisKeyByNote(note), "data", note.String()); err != nil {
		log.Error(err.Error())
	}
	return nil
}

func (n *Notices) toClearState(note *moc.Notice) error {
	now, err := ptypes.TimestampProto(n.clk.Now())
	if err != nil {
		return err
	}
	note.Action = moc.Notice_CLEAR
	note.ClearedTime = now
	n.tracer.Logf("changed: %+v", notestr(note))
	if err := n.notifydb.Clear(note.DatabaseId, now); err != nil {
		return err
	}

	redisConn := n.redisPool.Get()
	defer n.safeCloseRedis(redisConn)

	_, err = redisConn.Do("DEL", getRedisKeyByNote(note))
	return err
}

func (n *Notices) updateTime(note *moc.Notice) error {
	n.tracer.Logf("updated: %v", notestr(note))
	now, err := ptypes.TimestampProto(n.clk.Now())
	if err != nil {
		return err
	}

	redisConn := n.redisPool.Get()
	defer n.safeCloseRedis(redisConn)

	isDBChange := false
	redisNote, err := redis.Bool(redisConn.Do("EXISTS", getRedisKeyByNote(note)))
	isDBChange = !redisNote || err != nil
	note.UpdatedTime = now
	if _, err := redisConn.Do("HSET", getRedisKeyByNote(note), "data", note.String()); err != nil {
		log.Error(err.Error())
	}
	countUpdates, err := redis.Uint64(redisConn.Do("HINCRBY", getRedisKeyByNote(note), "count_updates", 1))
	if err != nil {
		log.Error(err.Error())
		isDBChange = true
		if _, err := redisConn.Do("HSET", getRedisKeyByNote(note), "count_updates", 1); err != nil {
			log.Error(err.Error())
		}
	}
	if isDBChange || countUpdates%updateMongoDBPerRecord == 0 {
		if err := n.notifydb.UpdateTime(note.DatabaseId, now); err != nil {
			return err
		}
	}
	return nil
}

func (n *Notices) reaper() {
	// Run the reaper once when starting up since there could be
	// a bit of old stuff
	n.reap()
	for {
		select {
		case <-n.reapTicker.C:
			n.reap()
		case <-n.ctxt.Done():
			n.tracer.Logf("reaper canceled")
			return
		}
	}
}

func (n *Notices) reap() {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	redisConn := n.redisPool.Get()
	defer n.safeCloseRedis(redisConn)

	olderThan := n.clk.Now().Add(-n.ackTimeout)
	if redisConn.Err() != nil {
		log.Error(redisConn.Err().Error())
		pOlderThan, err := ptypes.TimestampProto(olderThan)
		if err != nil {
			log.Error(err.Error())
			return
		}
		n.notifydb.Timeout(pOlderThan)
		return
	}
	var idToDelete []string
	var redisDeleted int
	var mongoUpdated int
	var iter int
	var err error
	var keys []string
	var loopStarted bool
	redisNote := new(moc.Notice)
	for iter != 0 || !loopStarted {
		loopStarted = true
		arr, err := redis.Values(redisConn.Do("SCAN", iter, "MATCH", redisNotifierName+":*"))
		if len(arr) != 2 {
			log.Error("an odd response: %v", log.Spew(arr))
			return
		}
		if iter, err = redis.Int(arr[0], err); err != nil {
			log.Error(err.Error())
			return
		}
		if keys, err = redis.Strings(arr[1], err); err != nil {
			log.Error(err.Error())
			return
		}
		for _, key := range keys {
			dataStruct, err := redis.String(redisConn.Do("HGET", key, "data"))
			if err != nil {
				log.Error(err.Error())
				redisConn.Do("DEL", key)
				continue
			}
			if err := proto.UnmarshalText(dataStruct, redisNote); err != nil {
				log.Error(err.Error())
				redisConn.Do("DEL", key)
				continue
			}
			if redisNote.Action != moc.Notice_ACK {
				continue
			}
			t, err := ptypes.Timestamp(redisNote.UpdatedTime)
			if err != nil {
				log.Error(err.Error())
				continue
			}
			if t.Before(olderThan) {
				idToDelete = append(idToDelete, redisNote.DatabaseId)
				redisConn.Do("DEL", key)
				redisDeleted++
			}
		}
	}
	if idToDelete != nil {
		mongoUpdated, err = n.notifydb.TimeoutBySliceId(idToDelete)
		if err != nil {
			log.Error(err.Error())
		}
	}
	log.Info("reaping completed, stale db %v, mem %v", mongoUpdated, redisDeleted)
}

func (n *Notices) TrackMap() map[string]*tms.Track {
	return n.trackMap
}

func (n *Notices) TrackTree() *rtree.RTree {
	return n.trackTree
}

func notestr(note *moc.Notice) string {
	return fmt.Sprintf("%v %v %v", note.Action, note.Source, note.Target)
}

func getRedisKeyByNote(note *moc.Notice) string {
	return fmt.Sprintf("%s:%s", redisNotifierName, note.NoticeId)
}
