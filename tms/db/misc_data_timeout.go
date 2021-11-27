package db

import (
	. "prisma/tms/client_api"
	"prisma/tms/log"

	"container/heap"
	"time"
)

type miscTimeout struct {
	timeoutHeap *TimeoutHeap
	timeinHeap  *TimeoutHeap
	req         GoMiscRequest
}

func MiscTimeoutStream(req GoMiscRequest, in <-chan GoGetResponse) (<-chan GoGetResponse, error) {
	out := make(chan GoGetResponse, 64)
	mst := &miscTimeout{
		timeoutHeap: NewTimeoutHeap(),
		timeinHeap:  NewTimeoutHeap(),
		req:         req,
	}

	req.Ctxt.Go(func() { mst.handle(in, out) })
	return out, nil
}

func (t *miscTimeout) handle(in <-chan GoGetResponse, out chan<- GoGetResponse) {
	done := t.req.Ctxt.Done()
	defer close(out)

	// Timer for waiting for next timeout. Default 15 minute wait is bullshit
	// value and gets canceled on the next line. The timer init just requires
	// an initial duration.
	timer := time.NewTimer(time.Duration(15) * time.Minute)
	timer.Stop()
	for {
		to := t.nextTimeoutIn()
		ti := t.nextTimeinIn()
		waitFor := to
		if ti < to {
			waitFor = ti
		}

		timer.Reset(waitFor)
		select {
		case <-done:
			timer.Stop()
			return
		case <-timer.C:
			t.sendTimeouts(out)
		case data, ok := <-in:
			if !ok {
				log.Debug("Input channel closed. Dying...")
				return
			}

			if data.Contents == nil {
				// This is strange and we can't deal with it. Pass it on
				// without processing
				out <- data
			} else {
				timer.Stop()
				sendOK := t.updateDeadlines(&data)
				if sendOK {
					out <- data
				}
			}
		}
	}
}

func (t *miscTimeout) nextTimeinIn() time.Duration {
	next, ok := t.timeinHeap.Peek()
	if !ok {
		// If there are currently no things, check again in 5 seconds. This
		// shouldn't be necessary (we should never check), but we need to
		// return a reasonable value!
		return time.Duration(5) * time.Second
	}
	curr := t.req.Time.Now()
	if next.Deadline().Before(curr) {
		// Deadline already expired. Return immediate duration
		return time.Duration(0)
	}
	//log.Printf("next deadline: %v %v %v", ok, next.deadline, next)
	return next.Deadline().Sub(curr)
}

func (t *miscTimeout) nextTimeoutIn() time.Duration {
	next, ok := t.timeoutHeap.Peek()
	if !ok {
		// If there are currently no things, check again in 5 seconds. This
		// shouldn't be necessary (we should never check), but we need to
		// return a reasonable value!
		return time.Duration(5) * time.Second
	}
	curr := t.req.Time.Now()
	if next.Deadline().Before(curr) {
		// Deadline already expired. Return immediate duration
		return time.Duration(0)
	}
	//log.Printf("next deadline: %v %v %v", ok, next.deadline, next)
	return next.Deadline().Sub(curr)
}

func (t *miscTimeout) sendTimeouts(out chan<- GoGetResponse) {
	qnow := t.req.Time.Now()

	// Send "time-in"s based on creation_time being reached
	for next, ok := t.timeinHeap.Peek(); ok; next, ok = t.timeinHeap.Peek() {

		if next.Deadline().After(qnow) {
			break
		}

		heap.Pop(t.timeinHeap)
		miscNext, ok := next.(MiscTimeoutInfo)
		if !ok {
			panic("Pulled a non-Track timeout info from heap!")
		}
		out <- GoGetResponse{
			Status:   Status_Current,
			Contents: miscNext.obj,
		}

		//log.Printf("sendTimeouts: %v %v", next.deadline, next)
	}

	// Send "time-outs" based on expiration_time being reached
	for next, ok := t.timeoutHeap.Peek(); ok && next.Deadline().Before(qnow); next, ok = t.timeoutHeap.Peek() {

		heap.Pop(t.timeoutHeap)
		miscNext, ok := next.(MiscTimeoutInfo)
		if !ok {
			panic("Pulled a non-Track timeout info from heap!")
		}
		out <- GoGetResponse{
			Status:   Status_Timeout,
			Contents: miscNext.obj,
		}

		//log.Printf("sendTimeouts: %v %v", next.deadline, next)
	}
}

func (t *miscTimeout) updateDeadlines(upd *GoGetResponse) bool {
	etime := upd.Contents.ExpirationTime
	ctime := upd.Contents.CreationTime

	ret := true
	now := t.req.Time.Now()
	if ctime.After(now) {
		// Hasn't been created yet!
		t.timeinHeap.Upsert(MiscTimeoutInfo{
			obj:      upd.Contents,
			deadline: ctime,
		})
		t.timeoutHeap.Upsert(MiscTimeoutInfo{
			obj:      upd.Contents,
			deadline: etime,
		})

		log.Debug("Inserting obj into timein heap: ctime: %v, now:%v", ctime, now)
		ret = false // Don't send since it hasn't been created yet!
	}

	eti := MiscTimeoutInfo{
		obj:      upd.Contents,
		deadline: etime,
	}
	if etime.After(now) {
		t.timeoutHeap.Upsert(eti)
	} else if t.timeoutHeap.Exists(eti) {
		// It's already expired, and we've already seen it
		t.timeoutHeap.Upsert(eti)
		ret = false // Don't send it now, we'll send the expiration separately
	} else {
		// We've never seen it and it already expired
		ret = false
	}

	return ret
}

type MiscTimeoutInfo struct {
	obj      *GoObject
	deadline time.Time
}

// Get the deadline for this object
func (t MiscTimeoutInfo) Deadline() time.Time {
	return t.deadline
}

// Get a unique identifier for this object
func (t MiscTimeoutInfo) ID() interface{} {
	return t.obj.ID
}
