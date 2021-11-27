package connect

import "prisma/tms/db"
import "prisma/tms/db/mongo"

// FIXME: This can be removed now
type Dao struct {
	Mongo    *mongo.MongoClient
	Tracks   db.TrackDB
	Misc     db.MiscDB
	Site     db.SiteDB
	Registry db.RegistryDB
}
