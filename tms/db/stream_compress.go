package db

import (
	. "prisma/tms"
	"prisma/tms/log"

	"container/list"
	"prisma/gogroup"
)

type latestTrack struct {
	latest    map[string]*Track
	sendOrder *list.List
}

// Returns a channel which gives only information about the latest target. If a
// client stream backs up, the incoming stream doesn't backpressure.
func LatestTrack(ctxt gogroup.GoGroup, input <-chan *Track) <-chan *Track {
	l := &latestTrack{
		latest:    make(map[string]*Track),
		sendOrder: list.New(),
	}
	output := make(chan *Track, 128)
	ctxt.Go(func() { l.process(ctxt, input, output) })
	return output
}

func (l *latestTrack) process(ctxt gogroup.GoGroup, input <-chan *Track, output chan<- *Track) {
	done := ctxt.Done()
	defer close(output)

	var toSend *Track = nil
	for {
		if toSend == nil && l.sendOrder.Len() > 0 {
			front := l.sendOrder.Front()
			l.sendOrder.Remove(front)
			toSendID := front.Value.(string)
			toSend = l.latest[toSendID]
			delete(l.latest, toSendID)
		}

		if toSend != nil {
			// We are trying to send toSend, but also need to accept new tracks
			// from input.
			select {
			case track, ok := <-input:
				if !ok {
					log.Debug("Track stream closed. Dying...")
					return
				}
				// We got an input, but downstream is backed up. We need to
				// store it for later.
				l.appendTrack(track)
			case output <- toSend:
				// Downstream finally unblocked. Great! (Finally...)
				toSend = nil
			case <-done:
				return
			}
		} else {

			// Nothing to send. Wait until we get something
			var track *Track
			select {
			case <-done:
				return
			case trackNew, ok := <-input:
				if !ok {
					log.Debug("Track stream closed. Dying...")
					return
				}
				// Yay, a new input!
				track = trackNew
			}

			// We gotta get it outta here somehow!
			select {
			// First, try to just send it to the output. May it's not backed up.
			case output <- track:
				// Good. It sent. We are done. We got an input and were able to
				// write directly to the output. I appear to be unnecessary.
				// Good. I'm not upset that you don't need me. Not at all. It's
				// a good thing. Sure. Whatever. I don't need you anyway.
			default:
				// Output is backing up. Don't wait, store for later
				l.appendTrack(track)
			}
		}
	}
}

func (l *latestTrack) appendTrack(track *Track) {
	if track == nil {
		log.Error("Got a NIL track. That's weird!")
		return
	}
	id := track.Id
	if _, ok := l.latest[id]; !ok {
		l.sendOrder.PushBack(id)
	}
	l.latest[id] = track
}
