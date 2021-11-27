package mongo

import (
	"time"

	"prisma/gogroup"
	"prisma/tms/client_api"

	"github.com/globalsign/mgo/bson"
)

// Replay is used to replay data from mongodb
// It is a good way to pass data to stream stuff and reduce parameters in interfaces
type Replay struct {
	dbClient        *MongoClient
	timeSince       time.Duration
	col             string
	updateTimeField string
	filters         bson.M
}

// NewReplay it creates an instance to replay data from specific collection
func NewReplay(dbClient *MongoClient, timeSince time.Duration,
	col string, updateTimeField string,
	filters bson.M) *Replay {
	return &Replay{
		dbClient:        dbClient,
		timeSince:       timeSince,
		col:             col,
		updateTimeField: updateTimeField,
		filters:         filters,
	}
}

// NewReplay it creates an instance to replay data from specific collection based on filter only
func NewReplayAll(dbClient *MongoClient, col string, filters bson.M) *Replay {
	return &Replay{
		dbClient: dbClient,
		col:      col,
		filters:  filters,
	}
}

// Do function is for replaying data from mongodb
// It sends data using HandleStreamFunc
// updateTimeField is the name of the field where is time of last updating
func (r *Replay) Do(ctx gogroup.GoGroup, funcToSend HandleStreamFunc) error {
	db := r.dbClient.DB()
	defer r.dbClient.Release(db)
	if r.filters == nil {
		r.filters = make(bson.M)
	}
	if r.updateTimeField != "" {
		if _, ok := r.filters[r.updateTimeField]; !ok {
			r.filters[r.updateTimeField] = bson.M{
				"$gt": time.Now().Add(-r.timeSince),
			}
		}
	}
	iter := db.C(r.col).Find(r.filters).Iter()
	var data bson.Raw
	for iter.Next(&data) {
		funcToSend(ctx, data)
	}
	funcToSend(ctx, client_api.Status_InitialLoadDone)
	return iter.Close()
}
