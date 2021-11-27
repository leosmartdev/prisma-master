// tdatabased is used to record track information into mongodb.
package main

import (
	"flag"
	reallog "log"
	"os"
	"prisma/gogroup"
	"prisma/tms/db"
	mdb "prisma/tms/db/mongo"
	"prisma/tms/libmain"
	"prisma/tms/log"
	"prisma/tms/tmsg"
	"strings"
	"sync"

	"github.com/globalsign/mgo"
)

type dbSchemas []string

var edbs dbSchemas = []string{"/usr/share/tms-db/mongo/schema", "/etc/trident/db", "/usr/share/tms-db/mongo/loaders"}

func init() {
	flag.Var(&edbs, "schemas", "specify js script directory path or exact path")
}

func main() {
	libmain.Main(tmsg.APP_ID_TDATABASED, realmain)
}

func realmain(ctxt gogroup.GoGroup) {
	flag.Parse()
	// MongoDb debug
	if envEnabled("MGO_DEBUG") {
		logger := reallog.New(os.Stdout, "[mgo] ", reallog.LUTC|reallog.Lshortfile)
		mgo.SetLogger(logger)
		mgo.SetDebug(true)
		logger.Println("MGO_DEBUG activated")
	}
	log.Debug("Starting tdatabased...")
	var err error
	waits := &sync.WaitGroup{}

	log.Debug("%+v", edbs)

	mdbInstance, err := mdb.NewMongoProcess(ctxt, tmsg.GClient, waits)
	if err != nil {
		log.Crit("Failed to start instance: %v", err)
	}
	mdbconn, err := mdb.NewMongoClient(ctxt, mdbInstance.DialInfo(), mdbInstance.Cred())
	if err != nil {
		log.Crit("Failed to connect to Mongo instance: %v", err)
		ctxt.Cancel(nil)
	}
	// Up schemas
	mdbconn.EnsureSetUp(edbs)

	// set dial info in context
	ctxt = gogroup.WithValue(ctxt, "mongodb", mdbInstance.DialInfo())
	activities := mdb.NewMongoActivities(ctxt, mdbconn)
	tracks := mdb.NewMongoTracks(ctxt, mdbconn)
	devices := mdb.NewMongoDeviceClient(ctxt, mdbconn)
	transmissions := mdb.NewTransmissionDb(ctxt, mdbconn)
	sites := mdb.NewSiteDb(ctxt)
	misc := mdb.NewMongoMiscData(ctxt, mdbconn)
	file := mdb.FileDb{}
	_, err = db.NewDataInserter(ctxt, tmsg.GClient, waits, tracks, activities, devices, transmissions, sites, misc, &file)
	if err != nil {
		log.Crit("Failed to start data inserter: %v", err)
		ctxt.Cancel(nil)
	}
	waits.Wait()
}

func envEnabled(name string) bool {
	switch os.Getenv(name) {
	case "1", "true", "TRUE", "t", "T":
		return true
	}
	return false
}

func (e *dbSchemas) Set(value string) error {
	for _, name := range strings.Split(value, ",") {
		*e = append(*e, name)
	}
	return nil
}

func (e dbSchemas) String() string {
	return strings.Join(e, ",")
}
