package mongo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	. "prisma/tms/client_api"
	"prisma/tms/db"
	"prisma/tms/log"

	"prisma/gogroup"

	"github.com/davecgh/go-spew/spew"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

var (
	BadArguments = errors.New("Arguments or parameters invalid")
	queryRequest = log.GetTracer("query-request")
	queryPlan    = log.GetTracer("query-plan")
	trackStats   = log.GetTracer("track-stats")
)

const (
	TailTimeout = 250 * time.Millisecond
)

type SimpleQuery struct {
	Filter bson.D
	Sort   []string
	Fields bson.M
}

/****
 * QueryServicer is responsible for executing a query in mongo on both data
 * currently in the database and on 'live' data being inserted. It is also
 * responsible for using multiple threads to deliver data to clients, making
 * parallelization of bson decoding easier.
 */
type QueryServicer struct {
	Name   string
	DBConn *MongoClient
	DB     *mgo.Database
	Coll   string

	// Match terms for the database to apply. There's list of them for various
	// query phases. This gets a bit complicated:
	// - If there is one Query ( len(Queries) == 1), then this match is used
	//   for the initial query and live query.
	// - If there are N > 1 Queries, QueryServicer iterates through the first
	//   (N-1) of them as initial queries then uses the Nth Query as the live
	//   query.
	Queries        []SimpleQuery
	Ctxt           gogroup.GoGroup
	ObjectCallback func(bson.Raw) bson.ObjectId
	StatusCallback func(Status)
	NumDeliverers  int // How many delivery threads should be used?

	Outstanding sync.WaitGroup // How many deliveries are out standing?
	Debug       bool           // Should we print out debugging information
	liveMinId   bson.ObjectId  // What is the latest liveId
	waitMinId   bool           // Is listMinId valid?

	currentQuery  int // Which query are we working on right now?
	backlogCursor *mgo.Iter
	liveCursor    *mgo.Iter
	tracer        *log.Tracer
}

// Start running the query
func (q *QueryServicer) Service() error {
	if q.DBConn == nil ||
		q.Coll == "" ||
		q.Ctxt == nil {
		return BadArguments
	}
	q.tracer = log.GetTracer("query-service")
	q.tracer.Logf("%v: tracing enabled", q.Name)

	q.DB = q.DBConn.DB()
	q.currentQuery = 0
	err := q.initBacklogCursor()
	if err != nil {
		return err
	}
	q.Ctxt.Go(q.funnel)

	return nil
}

// Start a backlog query
func (q *QueryServicer) initBacklogCursor() error {
	// Are we done with the backlog queries yet?
	switch len(q.Queries) {
	case 0:
		// Yes, but in a rather poor manner
		return errors.New("Must specify at least one query!")
	case 1:
		if q.currentQuery > 0 {
			// We did one and there is one. We're one
			q.backlogCursor = nil
			return nil
		}
	default:
		// The last q.Queries is the live one if there's more than 1 query
		if q.currentQuery >= len(q.Queries)-1 {
			q.backlogCursor = nil
			return nil
		}
	}

	// Which query are we working on?
	query := q.Queries[q.currentQuery]
	log.TraceMsg("Running query: %v", query)

	filt := query.Filter
	if len(filt) == 0 {
		filt = nil
	}
	dbquery := q.DB.C(q.Coll).Find(filt)
	if len(query.Sort) > 0 {
		dbquery = dbquery.Sort(query.Sort...)
	}
	if query.Fields != nil {
		dbquery = dbquery.Select(query.Fields)
	}
	bsRaw, _ := bson.Marshal(filt)
	var filtMap bson.M
	bson.Unmarshal(bsRaw, &filtMap)
	filtJSON, _ := json.MarshalIndent(filtMap, "", "  ")
	queryRequest.Logf("Running db query:\n%v Find(%v)",
		collectionName(dbquery), string(filtJSON))

	// Print the query explanation to the log if query plan debugging
	// is enabled
	if queryPlan.Enabled {
		explain := bson.M{}
		queryPlan.Log("Getting query plan...")
		err := dbquery.Explain(explain)
		if err != nil {
			queryPlan.Logf("Could not get query explanation: %v", err)
		} else {
			spewConfig := spew.ConfigState{
				Indent:   "  ",
				SortKeys: true,
				MaxDepth: 8,
			}
			queryPlan.Logf("Query explanation: %v, %v", err,
				spewConfig.Sdump(explain))
		}
	}

	q.currentQuery++
	q.backlogCursor = dbquery.Iter()
	q.liveCursor = nil
	return nil
}

// Initialize a query (listener) on the _live table
func (q *QueryServicer) initLiveCursor() error {
	log.TraceMsg("Switching to live cursor...")
	if q.liveMinId.Valid() {
		// Fast-forward to last document seen in initial query
		timeMin := bson.NewObjectIdWithTime(q.liveMinId.Time())
		log.TraceMsg("Fast-forwarding to object time %v", timeMin)
		q.liveCursor = q.DB.C(q.Coll + "_live").Find(bson.M{
			"_id": bson.M{
				"$gte": timeMin,
			},
		}).Tail(TailTimeout)
	} else {
		q.liveCursor = q.DB.C(q.Coll + "_live").Find(nil).Tail(TailTimeout)
	}
	return nil
}

// Get the row for 'id'
func (q *QueryServicer) getObjectId(id bson.ObjectId, obj *bson.Raw) (bool, error) {
	if !id.Valid() {
		return false, errors.New("Invalid ID!")
	}

	clause := bson.DocElem{
		Name:  "_id",
		Value: id,
	}
	lastQ := q.Queries[len(q.Queries)-1]
	terms := append(lastQ.Filter, clause)

	log.TraceMsg("Calling pipeline to get objects: %v", id)
	start := time.Now()
	query := q.DB.C(q.Coll).Find(terms)
	/*explain := make(bson.M)
	err := query.Explain(explain)
	if err == nil {
		expJson, _ := json.MarshalIndent(explain, "", "  ")
		log.TraceMsg("GetObjectID explanation: %v, %v", err, string(expJson))
	}*/

	iter := query.Iter()

	if iter.Next(obj) {
		log.TraceMsg("GetObjectId() db latency: %v", time.Since(start))
		log.TraceMsg("Got object: %v", id)
		return true, iter.Close()
	}
	log.TraceMsg("GetObjectId() db latency: %v", time.Since(start))
	log.TraceMsg("No results for object: %v", id)
	return false, iter.Close()
}

// Populate 'obj' with the next row listed in the _live table. Block until a
// row is available.
func (q *QueryServicer) nextLive(obj *bson.Raw) error {
	for {
		select {
		case <-q.Ctxt.Done():
			return io.EOF
		default:
			var chg DBChangeWithID
			if q.liveCursor.Next(&chg) {
				if q.waitMinId {
					nextSecond := q.liveMinId.Time()
					if chg.Id == q.liveMinId || chg.Id.Time().After(nextSecond) {
						q.waitMinId = false
					}
				} else {
					q.liveMinId = chg.Id
					got, err := q.getObjectId(chg.ObjId, obj)
					q.tracer.Logf("%v: live update: %+v %v", q.Name, chg, got)
					if got {
						return nil
					}
					if err != nil {
						return err
					}
				}
			} else {
				//q.debug("live cursor end")
				if q.liveCursor.Err() != nil {
					return q.liveCursor.Close()
				} else if q.liveCursor.Timeout() {
					log.TraceMsg("Cursor timeout... trying again!")
				} else {
					// Hmmm... No error, no timeout. Gotta restart, I think.
					// The MongoDB documentation implies that this can happen.
					q.liveCursor.Close()
					q.liveCursor = nil
					log.TraceMsg("Cursor returned nothing. Gotta restart!")
					// This can happen continuously if the _live collection is
					// empty. Prevent us from looping too frequently.
					time.Sleep(100 * time.Millisecond)
					q.initLiveCursor()
				}
			}
		}
	}
}

// Get the next row from either the backlog or a _live query
func (q *QueryServicer) next(obj *bson.Raw) error {
	//.debug("getting next")
	if q.backlogCursor == nil && q.liveCursor == nil {
		//q.debug("setting up backlog cursor")
		q.initBacklogCursor()
	}

	if q.backlogCursor != nil {
		// We are still streaming initial results
		//q.debug("still in the backlog?")
		if q.backlogCursor.Next(obj) {
			//q.debug("yes")
			return nil
		}
		//q.debug("end of the backlog")
		err := q.backlogCursor.Err()
		q.backlogCursor.Close()
		q.backlogCursor = nil
		if err != nil {
			return err
		}

		// Get the next backlog cursor
		err = q.initBacklogCursor()
		if err != nil {
			return err
		}

		if q.backlogCursor == nil && q.StatusCallback != nil {
			// Wait for all deliveries to complete before sending a signal
			// so that the signal is actually delivered _after_ all of the
			// backlog messages have been sent
			q.Outstanding.Wait()
			q.StatusCallback(Status_InitialLoadDone)
		}
	}
	return io.EOF
}

// This is a delivery thread
func (q *QueryServicer) deliver(wg *sync.WaitGroup, ch <-chan bson.Raw) {
	for raw := range ch {
		q.ObjectCallback(raw) // Deliver the raw bson
		q.Outstanding.Done()  // Decrement the outstanding counter
	}
	wg.Done() // Signal that we're exiting
}

// Initiate delivery threads, read raw bson rows, send deliveries, and status updates
func (q *QueryServicer) funnel() {
	count := uint64(0)
	total := uint64(0)
	statusTicker := time.NewTicker(db.StatusInterval)

	rawChan := make(chan bson.Raw, 128)

	// Release the database connection when we're done
	defer func() {
		defer q.DBConn.Release(q.DB)
	}()

	if q.NumDeliverers == 0 {
		// We need at least 1 decoder
		q.NumDeliverers = 1
	}

	// *** Launch delivery threads
	wg := &sync.WaitGroup{}
	for i := 0; i < q.NumDeliverers; i++ {
		wg.Add(1)
		q.Ctxt.Go(func() { q.deliver(wg, rawChan) })
	}

	// *** When we're done, close the rawChan to signal that we're done to the
	// delivery threads, wait for them to exit, then send the closing status
	defer func() {
		close(rawChan)
		wg.Wait()
		if q.StatusCallback != nil {
			// Wait for all deliveries to complete before sending a signal
			q.Outstanding.Wait()
			q.StatusCallback(Status_Closing)
		}
	}()

	// *** Send the starting status signal
	if q.StatusCallback != nil {
		q.StatusCallback(Status_Starting)
	}

	for {
		// Main loop. Read rows and send them down channel for delivery
		select {
		case <-q.Ctxt.Done():
			// We are done. Canceled, probably
			return
		case <-statusTicker.C:
			trackStats.Logf("%v: %v/sec, %v total", q.Name,
				float64(count)/db.StatusInterval.Seconds(), total)
			count = 0
		default:
			// *** Get the next bson result
			raw := bson.Raw{}
			err := q.next(&raw)
			if err == io.EOF {
				return
			} else if err != nil {
				// Ignore errors when a cursor's position in the capped collection is deleted.
				// Reads from the beginning of a capped collection are not guaranteed to succeed
				// when there are concurrent inserts that cause a truncation.
				if !strings.Contains(err.Error(), "CappedPositionLost") && !strings.Contains(err.Error(), "Invalid ID") {
					log.Error("Error from getting next object: %v", err)
					return
				}
			}
			// *** Send the bson down the rawChan to a delivery thread
			q.Outstanding.Add(1)
			select {
			case <-q.Ctxt.Done():
				q.Outstanding.Done()
				// We are done. Canceled, probably
				return
			case rawChan <- raw:
				count++
				total++
			}
		}
	}
}

func (q *QueryServicer) debug(format string, args ...interface{}) {
	if !q.Debug {
		return
	}
	msg := fmt.Sprintf(format, args...)
	log.Alert("[query-debug] %v", msg)
}
