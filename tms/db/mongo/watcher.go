package mongo

import (
	"errors"
	"sync"

	"prisma/gogroup"
	"prisma/tms/log"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// Watcher is an implementation for change streaming api of mongodb
type Watcher struct {
	mu sync.Mutex
	// a context to control for watcher
	ctx gogroup.GoGroup
	// a channel for receiving updates from a collection
	ch chan bson.Raw
	// last error that was issued from watching
	lastErr error
	// name for logging
	name string
}

// ErrTimeoutConnection is used for using timeout as an error
var ErrTimeoutConnection = errors.New("timeout")

type updateDesc struct {
	UpdatedFields map[string]interface{} `bson:"updatedFields"`
	RemovedFields []string               `bson:"removedFields"`
}

type evNamespace struct {
	DB   string `bson:"db"`
	Coll string `bson:"coll"`
}

type changeEvent struct {
	ID                bson.M      `bson:"_id"`
	OperationType     string      `bson:"operationType"`
	FullDocument      *bson.Raw   `bson:"fullDocument,omitempty"`
	Ns                evNamespace `bson:"ns"`
	DocumentKey       bson.M      `bson:"documentKey"`
	UpdateDescription *updateDesc `bson:"updateDescription,omitempty"`
}

// NewWatcher is a function to create an instance of watcher
func NewWatcher(ctx gogroup.GoGroup, name string) *Watcher {
	return &Watcher{
		ctx:  ctx,
		ch:   make(chan bson.Raw, 100),
		name: name,
	}
}

// GetChannel returns a channel for receiving updates
func (w Watcher) GetChannel() <-chan bson.Raw {
	return w.ch
}

// Start starts watching for the collection
// see mongodb change streaming documentation for pipeline
func (w *Watcher) Start(col *mgo.Collection, pipeline []bson.M) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.ctx.Go(func() {
		w.watch(col, pipeline)
	})
}

// Stop stops all processing
func (w *Watcher) Stop() {
	w.ctx.Cancel(nil)
}

// GetLastError returns last error that was issues by watcher methods
func (w *Watcher) GetLastError() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.lastErr
}

func (w *Watcher) setLastError(err error) {
	w.mu.Lock()
	w.lastErr = err
	w.mu.Unlock()
}

func (w *Watcher) watch(col *mgo.Collection, pipeline []bson.M) {
	defer close(w.ch)

	options := mgo.ChangeStreamOptions{
		FullDocument: mgo.UpdateLookup,
	}

	cs, err := col.Watch(pipeline, options)
	if err != nil {
		w.setLastError(err)
		return
	}
	defer func() {
		if err := cs.Close(); err != nil {
			w.setLastError(err)
		}
		if cs.Timeout() {
			w.setLastError(ErrTimeoutConnection)
		}
	}()
	change := new(changeEvent)
	// actually we need to handle "not found" error and continue to watch
	// but in practice we lose data after upping mongodb, so it would be better to recreate the watcher
	for cs.Next(change) || cs.Err() == nil {
		select {
		case <-w.ctx.Done():
			log.Info("watcher %v was canceled: %v", w.name, w.ctx.Err())
			return
		default:
		}
		if change.OperationType == "" {
			continue
		}
		// some operation types are like invalidate can return empty FullDocument
		if change.FullDocument == nil {
			if change.OperationType != "delete" && change.OperationType != "invalidate" {
				log.Error("undefined behavior. Expected fulldocument field is not empty. %v",
					log.Spew(change))
			}
			continue
		}
		select {
		case w.ch <- *change.FullDocument:
		case <-w.ctx.Done():
			log.Info("watcher %v was canceled: %v", w.name, w.ctx.Err())
			return
		}
		change.OperationType = ""
	}
}
