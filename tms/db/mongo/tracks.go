package mongo

import (
	"errors"
	"flag"
	"fmt"
	"unsafe"

	"prisma/tms"
	api "prisma/tms/client_api"
	tmsdb "prisma/tms/db"
	"prisma/tms/log"

	"prisma/gogroup"

	"context"
	"sync"

	"github.com/globalsign/mgo/bson"
)

var (
	DecodeThreads = 8
	//targetIDEncoder     = NewStructData(reflect.TypeOf(tms.TargetID{}), NoMap)
	trackStreamChanSize = 512
)

func init() {
	flag.IntVar(&DecodeThreads, "decode-threads", 8, "Maximum number of threads to use for bson decoding, per client request, per database stream")
}

func NewMongoTracks(ctxt gogroup.GoGroup, dbconn *MongoClient) tmsdb.TrackDB {
	c := &MongoTrackClient{
		dbconn:          dbconn,
		reg:             NewMongoRegistry(ctxt, dbconn),
		ctxt:            ctxt,
		condRegistry:    sync.NewCond(&sync.Mutex{}),
		registryInserts: make(map[string]struct{}),
	}
	return c
}

type MongoTrackClient struct {
	dbconn *MongoClient     // Connection to DB
	reg    tmsdb.RegistryDB // Connection to registry
	ctxt   gogroup.GoGroup  // Execution context

	muRegistryInserts sync.Mutex
	registryInserts   map[string]struct{} // is used to sync inserts
	condRegistry      *sync.Cond
}

func (c *MongoTrackClient) getOneTrack(filter bson.M, first bool) (*tms.Track, error) {
	ti := MongoTables.TableFromType(DBTrack{})
	if ti.Name == "" {
		return nil, errors.New("table was not found")
	}
	result := new(bson.Raw)
	var sort string
	if first {
		sort = "update_time"
	} else {
		sort = "-update_time"
	}
	if err := c.dbconn.DB().C(ti.Name).Find(filter).Sort(sort).Limit(1).One(&result); err != nil {
		return nil, err
	}
	coder := Coder{TypeData: DBTrackSD}
	track := new(DBTrack)
	coder.DecodeTo(*result, unsafe.Pointer(track))
	return track.ToTrack(), nil
}

func (c *MongoTrackClient) GetLastTrack(filter bson.M) (*tms.Track, error) {
	return c.getOneTrack(filter, false)
}

func (c *MongoTrackClient) GetFirstTrack(filter bson.M) (*tms.Track, error) {
	return c.getOneTrack(filter, true)
}

// Get a list of tracks, no streaming
func (c *MongoTrackClient) GetTracks(req tmsdb.GoTrackRequest) (*api.Tracks, error) {
	req.Stream = false
	trackStream, err := c.Get(req)
	if err != nil {
		return nil, err
	}

	trackMap := make(map[string]*tms.Track)
	for trackUpd := range trackStream {
		track := trackUpd.Track
		if track != nil {
			trackMap[track.Id] = track
		}
	}

	tracks := api.Tracks{
		Tracks: make([]*tms.Track, 0, 1024),
	}
	for _, track := range trackMap {
		tracks.Tracks = append(tracks.Tracks, track)
	}

	return &tracks, nil
}

func (c *MongoTrackClient) Get(req tmsdb.GoTrackRequest) (<-chan api.TrackUpdate, error) {
	return c.get(req)
}

func (c *MongoTrackClient) GetTrackStream(req tmsdb.GoTrackRequest) (<-chan api.TrackUpdate, error) {
	req.Stream = true
	return c.get(req)
}

func (c *MongoTrackClient) GetPipeline(stages []tmsdb.TrackPipelineStage) (<-chan api.TrackUpdate, error) {
	tr := &api.TrackRequest{}
	req := tmsdb.NewTrackRequest(tr, c.ctxt)
	tmsdb.PopulateTracksRequest(req)
	req.Stream = true

	pipeline := tmsdb.NewTrackPipeline(c, c.reg)
	for _, stage := range stages {
		pipeline.Append(stage)
	}
	log.Debug("Getting raw track stream")
	tracks, err := c.watch(*req, true, false, req.Stream)
	if err != nil {
		return nil, err
	}

	return pipeline.Start(tracks)
}

func (c *MongoTrackClient) get(req tmsdb.GoTrackRequest) (<-chan api.TrackUpdate, error) {
	tmsdb.PopulateTracksRequest(&req)

	//**** Build the tracks pipeline
	pipeline := tmsdb.NewTrackPipeline(c, c.reg)
	pipeline.Append(tmsdb.NewLogStage(req.Ctxt))

	if !req.DisableMerge {
		pipeline.Append(tmsdb.NewTrackMerger(req.Ctxt, tmsdb.MergeOptions{
			Mode:       api.MergeMode_None, // FIXME TrackID,
			Historical: req.Req.History != nil,
		}))
	}

	gf, err := tmsdb.NewGeoFilter(&req)
	if err != nil {
		return nil, fmt.Errorf("unable to create geo filter: %v", err)
	}
	pipeline.Append(gf)

	pipeline.Append(tmsdb.DefaultStaticTimeouts(req.Ctxt, tmsdb.TimeoutOptions{
		Disable: req.DisableTimeouts,
		Time:    req.Time,
	}))

	ts, err := tmsdb.NewTrackSearcher(&req)
	if err != nil {
		return nil, fmt.Errorf("unable to create track searcher: %v", err)
	}
	pipeline.Append(ts)

	// *** Get the raw track stream
	log.Debug("Getting raw track stream")
	tracks, err := c.watch(req, true, false, req.Stream)
	if err != nil {
		return nil, err
	}

	// *** Start the pipeline with the raw input stream
	return pipeline.Start(tracks)
}

func (c *MongoTrackClient) sendRawGetResponse(ctx context.Context, informer interface{}, s *Stream, sd *StructData, ch chan<- api.TrackUpdate, stream bool) {
	switch data := informer.(type) {
	case bson.Raw:
		coder := Coder{TypeData: sd}
		track := new(DBTrack)
		coder.DecodeTo(data, unsafe.Pointer(track))
		resp := api.TrackUpdate{
			Status: api.Status_Current,
			Track:  track.ToTrack(),
		}
		select {
		case ch <- resp:
		case <-ctx.Done():
			return
		}
	case api.Status:
		if !stream && (data == api.Status_InitialLoadDone) {
			s.ctx.Cancel(nil)
			return
		}
	default:
		log.Info("data is not supported: %v", data)
	}
}

func (c *MongoTrackClient) watch(replayInfo tmsdb.GoTrackRequest, replay, permanent, stream bool) (<-chan api.TrackUpdate, error) {
	ti := MongoTables.TableFromType(DBTrack{})
	if ti.Name == "" {
		return nil, errors.New("table was not found")
	}
	var replaySt *Replay
	if replay {
		var filters bson.M
		if structFilter := replayInfo.Req.GetFilterSimple(); structFilter != nil {
			filters = make(bson.M)
			filters["registry_id"] = bson.M{"$in": structFilter.Registries}
		}
		replaySt = NewReplay(c.dbconn, replayInfo.MaxHistory, ti.Name,
			"update_time", filters)
	}
	ch := make(chan api.TrackUpdate, trackStreamChanSize)
	c.ctxt.Go(func() {
		defer close(ch)
		s := NewStream(c.ctxt.Child("streamer/misc/"+ti.Name), c.dbconn, ti.Name)
		s.Watch(func(ctx context.Context, informer interface{}) {
			c.sendRawGetResponse(ctx, informer, s, DBTrackSD, ch, stream)
		}, permanent, replaySt, nil)
	})
	return ch, nil
}

func (c *MongoTrackClient) GetHistoricalTrack(req tmsdb.GoHistoricalTrackRequest) (*tms.Track, error) {
	db := c.dbconn.DB()
	defer c.dbconn.Release(db)

	coder := Coder{TypeData: DBTrackSD}
	t := new(DBTrack)
	raw := &bson.Raw{}
	err := db.C("tracks").FindId(bson.ObjectIdHex(req.Req.DatabaseId)).One(raw)
	coder.DecodeTo(*raw, unsafe.Pointer(t))

	return t.ToTrack(), err
}
