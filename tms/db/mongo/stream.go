package mongo

import (
	"context"
	"net"
	"strings"
	"time"

	"prisma/gogroup"
	"prisma/tms/log"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// HandleStreamFunc is an interface for functions to handle data from streaming
// It is a function that is able to parse raw data and send to a channel
// It should use ctx to avoid leakage
type HandleStreamFunc func(ctx context.Context, informer interface{})

// Stream implements features for streaming data from mongodb
type Stream struct {
	ctx      gogroup.GoGroup
	dbClient *MongoClient
	col      string
}

// NewStream returns an instance for streaming the collection
func NewStream(ctx gogroup.GoGroup, dbClient *MongoClient, col string) *Stream {
	return &Stream{
		ctx:      ctx,
		dbClient: dbClient,
		col:      col,
	}
}

const sleepAfterError = 1 * time.Second

// Watch create pipeline for watching a collection for any changes
// It sends data using HandlerStreamFunc
// It can keep permanent connection if permanent is true
// replay can be passed as nil, it means you do not want to replay anything
func (s *Stream) Watch(funcToSend HandleStreamFunc, permanent bool, replay *Replay, pipeline []bson.M) {
	// if we need permanent streaming we will recreate watchers permanently
	for {
		// Of course, we want to avoid exceeded processing
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		if err := s.watchLoopProcessing(funcToSend, replay, pipeline, s.col); err != nil {
			if strings.HasPrefix(err.Error(),
				"cannot open $changeStream for non-existent database:") {
				log.Crit("got a critical error from change streaming: %s - %s", s.col, err.Error())
			} else {
				log.Error("got an error from change streaming: %s - %s", s.col, err.Error())
			}
		}

		if !permanent {
			break
		}
		time.Sleep(sleepAfterError)
	}
}

func (s *Stream) watchLoopProcessing(funcToSend HandleStreamFunc, replay *Replay, pipeline []bson.M, name string) error {
	w := NewWatcher(s.ctx.Child("watcher"), name)
	db := s.dbClient.DB()
	defer func() {
		if _, ok := w.GetLastError().(*net.OpError); ok {
			return
		}
		// the session is closed when error is queryError
		if _, ok := w.GetLastError().(*mgo.QueryError); ok {
			return
		}
		if w.GetLastError() == ErrTimeoutConnection {
			return
		}
		s.dbClient.Release(db)
	}()
	w.Start(db.C(s.col), pipeline)
	if replay != nil {
		// we need to replay changes from mongodb
		if err := replay.Do(s.ctx, funcToSend); err != nil {
			log.Error("replay error: %s", err.Error())
		}
	}
	// watch changes and send to the channel
	for data := range w.GetChannel() {
		funcToSend(s.ctx, data)
	}
	w.Stop()
	return w.GetLastError()
}
