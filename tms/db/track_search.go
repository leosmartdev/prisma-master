package db

import (
	"prisma/gogroup"
	. "prisma/tms"
	api "prisma/tms/client_api"
	"prisma/tms/log"
	"sort"
)

type TrackLimiter struct {
	tracks []*Track
	max    int
}

func NewTrackLimiter(max int) *TrackLimiter {
	// Always store more one more than the requested amount. The last
	// value is eventually discard but kept around to hold the closest
	// track that didn't make the cut
	tl := TrackLimiter{max: max}
	tl.tracks = make([]*Track, 0, max+1)
	return &tl
}

func (tl *TrackLimiter) Add(t *Track) {
	// If track has no metadata, just ignore
	//if t.Metadata == nil || len(t.Metadata) == 0 {
	//	return
	//}

	if t.Metadata != nil && len(t.Metadata) > 0{
		// If already full...
		if len(tl.tracks) > tl.max {
			// and the given track does not sort below the last value, this
			// track doesn't make the cut and can be discarded
			if t.Metadata[0].Name > tl.tracks[tl.max].Metadata[0].Name {
				return
			}
		}
	}

	// Is this track already in the list? If so, update the track and
	// return
	newID := t.Id
	for index, prevTrack := range tl.tracks {
		prevID := prevTrack.Id
		if newID == prevID {
			tl.tracks[index] = t
			return
		}
	}

	// If already full, drop the last track and replace it with this one.
	// Otherwise append to the list. Then sort.
	if len(tl.tracks) > tl.max {
		tl.tracks[tl.max] = t
	} else {
		tl.tracks = append(tl.tracks, t)
	}

	if t.Metadata != nil && len(t.Metadata) > 0 {
		sort.Sort(api.TracksByName{tl.tracks})
	}
}

func (tl *TrackLimiter) GetTracks() []*Track {
	// If there is the extra track that didn't make the cut, remove it
	// from the results
	if len(tl.tracks) > tl.max {
		return tl.tracks[:tl.max]
	}
	return tl.tracks
}

type TrackSearcher struct {
	ctxt    gogroup.GoGroup
	req     *api.TrackRequest
	limiter *TrackLimiter
	tracks []string
}

func NewTrackSearcher(req *GoTrackRequest) (*TrackSearcher, error) {
	ts := &TrackSearcher{}
	ts.req = req.Req
	ts.ctxt = req.Ctxt
	if ts.req.Mode == api.RequestMode_Search {
		// If the limit is not specified, make it one to encourage the
		// client to set a realistc value
		limit := ts.req.Limit
		if limit == 0 {
			limit = 1
		}
		ts.limiter = NewTrackLimiter(int(limit))
		// filter
		f := req.Req.Filter
		if f != nil {
			switch x := f.(type) {
			case *api.TrackRequest_FilterSimple:
				if x.FilterSimple != nil {
					ts.tracks = x.FilterSimple.Tracks
				}

			default:
				log.Error("Could not decipher filter: %v", f, req)
				return nil, UnknownOption
			}
		}
	}
	return ts, nil
}

func (ts *TrackSearcher) Start(in <-chan api.TrackUpdate) (<-chan api.TrackUpdate, error) {
	if ts.req.Mode != api.RequestMode_Search {
		return in, nil
	}

	out := make(chan api.TrackUpdate, 64)
	ts.ctxt.Go(func() { ts.handle(in, out) })
	return out, nil
}

func (ts *TrackSearcher) handle(in <-chan api.TrackUpdate, out chan<- api.TrackUpdate) {
	done := ts.ctxt.Done()
	defer close(out)

	for {
		select {
		case <-done:
			return
		case trackUpd, ok := <-in:
			if !ok {
				ts.emit(out)
				return
			}
			status := trackUpd.Status
			track := trackUpd.Track
			if status == api.Status_Timeout || track == nil {
				continue
			}
			// in range
			for _, v := range ts.tracks {
				//log.GetTracer("lop").Logf("s(%s)	db(%s)	reg(%s)", v, track.DatabaseId, track.RegistryId)
				if v == track.RegistryId {
					ts.limiter.Add(track)
				}
			}
		}
	}
}

func (ts *TrackSearcher) emit(out chan<- api.TrackUpdate) {
	results := ts.limiter.GetTracks()
	for _, track := range results {
		out <- api.TrackUpdate{
			Status: api.Status_Current,
			Track:  track,
		}
	}
}
