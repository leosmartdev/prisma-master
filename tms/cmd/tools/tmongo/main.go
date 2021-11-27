package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"prisma/gogroup"
	"prisma/tms/client_api"
	"prisma/tms/db"
	"prisma/tms/db/mongo"
	"prisma/tms/moc"

	"github.com/globalsign/mgo"
)

var (
	mongoUrl string
	dialInfo *mgo.DialInfo
)

func init() {
	flag.StringVar(&mongoUrl, "url", "mongodb://:27017", "MongoDB URL")
}

func main() {
	logger := log.New(os.Stdout, "[mgo] ", log.LUTC|log.Lshortfile)
	mgo.SetLogger(logger)
	mgo.SetDebug(true)
	logger.Println("MGO_DEBUG activated")
	// workaround for ssl ParseURL bug
	ssl := strings.Contains(mongoUrl, "ssl=true")
	if ssl {
		mongoUrl = strings.Replace(mongoUrl, "ssl=true", "", -1)
	}
	// MongoDB
	dialInfo, err := mgo.ParseURL(mongoUrl)
	if err != nil {
		panic(err)
	}
	if ssl {
		tlsConfig := &tls.Config{}
		dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			conn, err := tls.Dial("tcp", addr.String(), tlsConfig)
			return conn, err
		}
	}
	dialInfo.Database = "trident"
	dialInfo.Timeout = time.Duration(2) * time.Second
	// test connection
	fmt.Println("Connecting to " + mongoUrl)
	mgo.DialWithInfo(dialInfo)
	fmt.Println("Connected to " + mongoUrl)
	// get mongo session
	sess, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		panic(err)
	}
	sess.SetCursorTimeout(time.Duration(0))
	sess.SetSocketTimeout(time.Duration(5) * time.Second)
	sess.BuildInfo()
	fmt.Println(sess.DB(dialInfo.Database).CollectionNames())
	// set context for tms code
	ctx := context.WithValue(context.Background(), "mongodb", dialInfo)
	// miscDb
	group := gogroup.New(ctx, "tmongo")
	// This is a change stream test utility we will assum that testing in the dev env is not over ssl
	// thus creds are being passed as empty to this function
	client, err := mongo.NewMongoClient(group, dialInfo, &mgo.Credential{})
	if err != nil {
		panic(err)
	}
	miscDb := mongo.NewMongoMiscData(group, client)
	// changeStream
	stream := miscDb.GetPersistentStream(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: "prisma.tms.moc.Incident",
		},
		Ctxt: group,
	}, nil, nil)
	group.Go(func() {
		for {
			select {
			case update, ok := <-stream:
				if !ok {
					continue // channel was closed
				}
				if client_api.Status_InitialLoadDone == update.Status {
					continue
				}
				if update.Contents == nil || update.Contents.Data == nil {
					fmt.Println("no content", update)
					continue
				}
				incident, ok := update.Contents.Data.(*moc.Incident)
				if !ok {
					fmt.Println("bad content", update)
					continue
				}
				fmt.Println("++++++++++++++++")
				fmt.Println(incident)
			}
		}
	})
	// insert
	r, err := miscDb.Upsert(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: "prisma.tms.moc.Incident",
			Obj: &db.GoObject{
				Data: &moc.Incident{
					IncidentId: "TEST-" + fmt.Sprint(time.Now().Unix()),
					Name:       "tmongo",
					Type:       "test",
					Phase:      1,
					Commander:  "im0",
					State:      1,
					Assignee:   "im0",
				},
			},
		},
		Ctxt: group,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(r)
	miscDb.Delete(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: "prisma.tms.moc.Incident",
			Obj: &db.GoObject{
				ID: r.Id,
			},
		},
		Ctxt: group,
	})
}
