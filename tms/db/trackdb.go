package db

import (
	. "prisma/tms"
	. "prisma/tms/client_api"
	"prisma/tms/debug"

	"time"

	"prisma/gogroup"

	"github.com/globalsign/mgo/bson"
)

// Interface to a database which contains tracks
type TrackDB interface {
	GetTracks(req GoTrackRequest) (*Tracks, error)
	GetTrackStream(req GoTrackRequest) (<-chan TrackUpdate, error)
	GetPipeline(stages []TrackPipelineStage) (<-chan TrackUpdate, error)
	// GetLastTrack is used to get the last track by specific filter
	GetLastTrack(filter bson.M) (*Track, error)
	GetFirstTrack(filter bson.M) (*Track, error)
	Get(req GoTrackRequest) (<-chan TrackUpdate, error)
	Insert(*Track) error
	GetHistoricalTrack(req GoHistoricalTrackRequest) (*Track, error)
}

// A track request supplemented with stuff only useful to go code or needed by
// various track pipeline stages.
type GoTrackRequest struct {
	Req  *TrackRequest
	Ctxt gogroup.GoGroup
	Time *TimeKeeper

	MaxHistory      time.Duration
	Stream          bool
	DisableMerge    bool
	DisableTimeouts bool
	DebugQuery      bool
}

type GoHistoricalTrackRequest struct {
	Req  *HistoricalTrackRequest
	Ctxt gogroup.GoGroup
}

// Create a new request based on raw client request params
func NewTrackRequest(req *TrackRequest, ctxt gogroup.GoGroup) *GoTrackRequest {
	greq := GoTrackRequest{
		Req:        req,
		Ctxt:       ctxt,
		MaxHistory: 0,
		Time:       &TimeKeeper{},
	}
	if req.History != nil {
		greq.MaxHistory = time.Duration(req.History.Seconds) * time.Second
	}

	tk := greq.Time
	if req.ReplayTime != nil {
		tk.Replay = true
		tk.ReplayTime = FromTimestamp(req.ReplayTime)
		tk.StartTime = time.Now()
		tk.Speed = req.ReplaySpeed
		if tk.Speed == 0.0 {
			tk.Speed = 1.0
		}
	}
	return &greq
}

// Set up a GoTrackRequest which already has some stuff in it
func PopulateTracksRequest(greq *GoTrackRequest) {
	req := greq.Req

	if greq.Time == nil {
		greq.Time = &TimeKeeper{}
		tk := greq.Time
		if req.ReplayTime != nil {
			tk.Replay = true
			tk.ReplayTime = FromTimestamp(req.ReplayTime)
			tk.StartTime = time.Now()
			tk.Speed = req.ReplaySpeed
			if tk.Speed == 0.0 {
				tk.Speed = 1.0
			}
		}
	}

	if req.History != nil {
		if req.History.Seconds == 0 {
			req.History = nil
		} else {
			h := time.Duration(req.History.Seconds) * time.Second
			if h > greq.MaxHistory {
				greq.MaxHistory = h
			}
		}
	}

	if greq.MaxHistory == 0 {
		greq.MaxHistory = time.Duration(15) * time.Minute
		if debug.FastTimers {
			greq.MaxHistory = time.Duration(1) * time.Minute
		}
	}

}
