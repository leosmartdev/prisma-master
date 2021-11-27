package db

import (
	"prisma/gogroup"
	"prisma/tms"
	api "prisma/tms/client_api"
	"prisma/tms/debug"
	"prisma/tms/devices"
	"prisma/tms/log"

	"container/heap"
	"time"
)

var (
	standardTimeout = time.Duration(15) * time.Minute
	fastTimeout     = time.Duration(1) * time.Minute
	maxTimeout      = time.Duration(24*365) * time.Hour
)

type TimeoutOptions struct {
	Disable bool
	Time    *TimeKeeper
}

func DefaultTimeoutOptions() TimeoutOptions {
	return TimeoutOptions{
		Time: &TimeKeeper{},
	}
}

// Pipeline stage which times out tracks based on how long it's been since they
// last saw an update.
type StaticTimeout struct {
	ctxt    gogroup.GoGroup
	opts    TimeoutOptions
	req     *GoTrackRequest
	Timeout time.Duration

	timeoutHeap *TimeoutHeap
}

// Construct a timeout stage using the default global timeouts
func DefaultStaticTimeouts(ctxt gogroup.GoGroup, opts TimeoutOptions) TrackPipelineStage {
	timeout := standardTimeout
	if debug.FastTimers {
		log.Info("using fast timeout")
		timeout = fastTimeout
	}
	ret := &StaticTimeout{
		ctxt:        ctxt,
		opts:        opts,
		Timeout:     timeout,
		timeoutHeap: NewTimeoutHeap(),
	}
	return ret
}

// DefaultMaxTimeouts constructs a timeout stage using the maximum allowed time period (the client can request data up
// to 1 year old).
func DefaultMaxTimeouts(ctxt gogroup.GoGroup, opts TimeoutOptions) TrackPipelineStage {
	ret := &StaticTimeout{
		ctxt:        ctxt,
		opts:        opts,
		timeoutHeap: NewTimeoutHeap(),
		Timeout:     maxTimeout,
	}
	return ret
}

// What's the timeout for this track type?
func (t *StaticTimeout) GetTimeout(devType devices.DeviceType) time.Duration {
	return t.Timeout
}

// Start a timeout processor
func (t *StaticTimeout) Start(in <-chan api.TrackUpdate) (<-chan api.TrackUpdate, error) {
	if t.opts.Disable {
		return in, nil
	}

	out := make(chan api.TrackUpdate, 64)
	t.ctxt.Go(func() { t.handle(in, out) })
	return out, nil
}

// This is the timeout processing thread
func (t *StaticTimeout) handle(in <-chan api.TrackUpdate, out chan<- api.TrackUpdate) {
	done := t.ctxt.Done()
	defer close(out)

	// Timer for waiting for next timeout. Default 15 minute wait is bullshit
	// value and gets canceled on the next line. The timer init just requires
	// an initial duration.
	timer := time.NewTimer(time.Duration(15) * time.Minute)
	timer.Stop()
	for {
		waitFor := t.nextTimeoutIn()
		timer.Reset(waitFor)
		select {
		case <-done:
			timer.Stop()
			return
		case <-timer.C:
			// The next timeout needs to get processed!
			t.sendTimeouts(out)
		case trackUpd, ok := <-in:
			// Update our database, send along if the track isn't already timed
			// out
			if !ok {
				log.Debug("Input channel closed. Dying...")
				return
			}

			track := trackUpd.Track
			if track == nil {
				out <- trackUpd
				continue
			}
			timer.Stop()
			status := t.updateDeadlines(track)
			switch status {
			default:
				select {
				case <-done:
					timer.Stop()
					return
				case out <- api.TrackUpdate{
					Status: trackUpd.Status,
					Track:  track,
				}:
				}
			case api.Status_Timeout:
				select {
				case <-done:
					timer.Stop()
					return
				case out <- api.TrackUpdate{
					Status: api.Status_Timeout,
					Track:  track,
				}:
				}
			}
		}
	}
}

// When is the next timeout?
func (t *StaticTimeout) nextTimeoutIn() time.Duration {
	next, ok := t.timeoutHeap.Peek()
	if !ok {
		// If there are currently no things, check again in 5 seconds. This
		// shouldn't be necessary (we should never check), but we need to
		// return a reasonable value!
		return time.Duration(5) * time.Second
	}
	curr := t.opts.Time.Now()
	if next.Deadline().Before(curr) {
		// Deadline already expired. Return immediate duration
		return time.Duration(0)
	}
	//log.Printf("next deadline: %v %v %v", ok, next.deadline, next)
	return next.Deadline().Sub(curr)
}

// Send all the current timeouts, pop them off the heap
func (t *StaticTimeout) sendTimeouts(out chan<- api.TrackUpdate) {
	for next, ok := t.timeoutHeap.Peek(); ok && next.Deadline().Before(t.opts.Time.Now()); next, ok = t.timeoutHeap.Peek() {
		heap.Pop(t.timeoutHeap)
		trackNext, ok := next.(TrackTimeoutInfo)
		if !ok {
			panic("Pulled a non-Track timeout info from heap!")
		}
		update := api.TrackUpdate{
			Status: api.Status_Timeout,
			Track:  trackNext.track,
		}
		out <- update
		//log.Printf("sendTimeouts: %v %v", next.deadline, next)
	}
}

// Update the timeouts for this track
func (t *StaticTimeout) updateDeadlines(track *tms.Track) api.Status {
	if track == nil {
		log.Error("Got a NIL track. That's weird")
		return api.Status_Unknown
	}

	deadline := time.Unix(0, 0)

	// Find the timeout deadline furthest in the future
	for _, target := range track.Targets {
		if target == nil {
			log.Error("Got a NIL target. Very strange")
			continue
		}
		if target.Type == devices.DeviceType_Manual && target.Manual != nil && target.Manual.IssueTimeout {
			return api.Status_Timeout
		}
		if target.Nmea != nil && target.Nmea.Ttm != nil && target.Nmea.Ttm.Status == "L" {
			return api.Status_Timeout
		}

		if target.UpdateTime != nil {
			tutu := time.Unix(target.UpdateTime.Seconds, int64(target.UpdateTime.Nanos))
			ttu := time.Unix(target.Time.Seconds, int64(target.Time.Nanos))
			if target.Repeat == false &&
				tutu.Equal(ttu) == false &&
				ttu.Add(t.GetTimeout(target.Type)).Before(t.opts.Time.Now()) == true {
				return api.Status_Timeout
			}
			ttime := target.UpdateTime
			tm := time.Unix(ttime.Seconds, int64(ttime.Nanos)).Add(t.GetTimeout(target.Type))
			if tm.After(deadline) {
				if target.Repeat == true || (target.Repeat == false && tutu.Equal(ttu) == true) {
					deadline = tm
				} else if target.Repeat == false && tutu.After(ttu) == true {
					deadline = tm.Add(-tutu.Sub(ttu))
				}
			}
		}
	}

	if deadline.Before(t.opts.Time.Now()) {
		return api.Status_Timeout
	}

	info := TrackTimeoutInfo{
		id:       track.Id,
		track:    track,
		deadline: deadline,
	}

	//log.Printf("Deadline: %v -> %v", info.id, info.deadline)

	t.timeoutHeap.Upsert(info)
	return api.Status_Current
}

type TrackTimeoutInfo struct {
	id       string
	track    *tms.Track
	deadline time.Time
}

// Get the deadline for this object
func (t TrackTimeoutInfo) Deadline() time.Time {
	return t.deadline
}

// Get a unique identifier for this object
func (t TrackTimeoutInfo) ID() interface{} {
	return t.id
}
