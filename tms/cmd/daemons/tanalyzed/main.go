// tanalyzed is a daemon, which allows to analyze and process any updates as a streamer.
package main

import (
	"flag"
	glog "log"
	"os"
	"reflect"
	"sync"
	"time"

	api "prisma/tms/client_api"
	"prisma/tms/db"
	"prisma/tms/db/connect"
	"prisma/tms/db/mongo"
	"prisma/tms/libmain"
	"prisma/tms/log"
	"prisma/tms/tmsg"
	"prisma/tms/ws"

	"prisma/gogroup"

	"github.com/globalsign/mgo"
	"github.com/gomodule/redigo/redis"
)

var (
	redisDB       = 0
	redisHost     = ""
	redisPassword = ""
)

func init() {
	flag.IntVar(&redisDB, "redis-db", 0, "redis database to keep data there")
	flag.StringVar(&redisHost, "redis-host", "localhost:6379", "redis host:port to connect")
	flag.StringVar(&redisPassword, "redis-password", "", "redis password to connect to host")
}

type stage interface {
	init(gogroup.GoGroup, *mongo.MongoClient) error
	start()
	analyze(update api.TrackUpdate) error
}

var (
	consulServer = flag.String("consul-server", "127.0.0.1", "Consul server for communicating by using API")
	dc           = flag.String("datacenter", "dc1", "Current data center")
)

func main() {
	// MongoDb debug
	if envEnabled("MGO_DEBUG") {
		logger := glog.New(os.Stdout, "[mgo] ", glog.LUTC|glog.Lshortfile)
		mgo.SetLogger(logger)
		mgo.SetDebug(true)
		logger.Println("MGO_DEBUG activated")
	}
	flag.Parse()

	libmain.Main(tmsg.APP_ID_TANALYZED, func(ctxt gogroup.GoGroup) {
		log.Info("starting analyzer")

		var conn *mongo.MongoClient
		var err error
		log.Debug("Conn mongo %+v", conn)
		for conn == nil {
			conn, err = connect.GetMongoClient(ctxt, tmsg.GClient)
			if err != nil {
				log.Error("unable to connect to the database: %v", err)
				time.Sleep(time.Second)
			}
		}
		log.Info("database connections established")
		// set dial info in context
		ctxt = gogroup.WithValue(ctxt, "mongodb", conn.DialInfo)
		// set cred in context
		ctxt = gogroup.WithValue(ctxt, "mongodb-cred", conn.Cred)
		redisPool := &redis.Pool{
			Dial: func() (redis.Conn, error) {
				return redis.Dial("tcp", redisHost, redis.DialDatabase(redisDB), redis.DialPassword(redisPassword))
			},
		}
		defer redisPool.Close()
		// Publisher-Subscriber
		publisher := ws.NewPublisher()
		publisher.Subscribe("Vessel", tmsg.GClient)
		publisher.Subscribe("Sit915", tmsg.GClient)
		notifier := NewNotifier(ctxt, conn, redisPool)
		if err := notifier.Init(); err != nil {
			log.Fatal("unable to init: %v", err)
		}
		stages := []stage{
			newTrackExtenderStage(),
			newSarsatStage(notifier, ctxt, conn),
			newSartStage(notifier),
			newOmnicomStage(notifier, ctxt, conn, publisher),
			newZoneStage(notifier),
			newRuleStage(notifier),
			newMulticastStage(tmsg.GClient, notifier, redisPool),
			newIncidentStage(tmsg.GClient, notifier),
			newSit915Stage(tmsg.GClient, notifier, publisher),
		}

		log.Info("initializing stages")
		waitInit := sync.WaitGroup{}
		for _, st := range stages {

			child := ctxt.Child(reflect.TypeOf(st).String())
			waitInit.Add(1)
			child.Go(func(ctxt gogroup.GoGroup, st stage, conn *mongo.MongoClient) {
				if err := st.init(ctxt, conn); err != nil {
					log.Error("unable to initialilze: %v", err)
				}
				st.start()
				waitInit.Done()
			}, child, st, conn)

		}
		waitInit.Wait()

		log.Info("initial tracks loading")
		trackdb := mongo.NewMongoTracks(ctxt, conn)
		updates, err := trackdb.GetTrackStream(db.GoTrackRequest{
			Req:  &api.TrackRequest{},
			Ctxt: ctxt,
		})
		if err != nil {
			log.Fatal("unable to get track stream: %v", err)
		}

		log.Info("initialization complete")
		go RuleChangeNotifier(ctxt)
		notifier.Start()

		done := ctxt.Done()
		for {
			select {
			case update := <-updates:
				// Filter out status updates that have no track information
				if update.Track == nil {
					continue
				}
				notifier.updateTrack(update)
				for _, stage := range stages {
					if err := stage.analyze(update); err != nil {
						log.Error("error analyzing update: %+v, %+v", err)
					}
				}
			case <-done:
				log.Info("stopping analyzer")
				os.Exit(0)
			}
		}
	})
}

func envEnabled(name string) bool {
	switch os.Getenv(name) {
	case "1", "true", "TRUE", "t", "T":
		return true
	}
	return false
}
