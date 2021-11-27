// Package connect provides functions and structure to get mongo db connection.
package connect

import (
	"crypto/tls"
	"net"
	"strings"
	"time"

	"prisma/gogroup"
	tmsmain "prisma/tms"
	"prisma/tms/db"
	"prisma/tms/db/mongo"
	"prisma/tms/log"
	"prisma/tms/tmsg"
	"prisma/tms/tmsg/client"

	"github.com/globalsign/mgo"
	"golang.org/x/net/context"
)

func GetMongoClient(pctxt gogroup.GoGroup, cli client.TsiClient) (*mongo.MongoClient, error) {
	ctxt, _ := context.WithTimeout(pctxt, time.Duration(1)*time.Second)
	var mongoUrl, mongoSsl string

	resp, err := cli.Request(ctxt,
		tmsmain.EndPoint{
			Site: tmsg.TMSG_LOCAL_SITE,
			Aid:  tmsg.APP_ID_TDATABASED,
		},
		&db.DBConnectionRequest{})
	if err != nil {
		log.Debug("Connection err, using flag: %v", err)
		mongoUrl = mongo.Config.MongoUrl
	} else {
		cparams, ok := resp.(*db.DBConnectionParams)
		if !ok {
			log.Error("Got %v instead of DBConnectionParams", resp)
			return nil, client.BadMessageType
		}
		mongoUrl = cparams.Addresses[0]
		mongoSsl = cparams.Authkey

	}
	log.Debug("Connection info: %v", mongoUrl)
	// workaround for ssl ParseURL bug
	ssl := (strings.Contains(mongoUrl, "ssl=true") || mongoSsl == mongo.MongoDbX509Mechanism)
	if ssl {
		mongoUrl = strings.Replace(mongoUrl, "ssl=true", "", -1)
	}
	// MongoDB
	dialInfo, err := mgo.ParseURL(mongoUrl)
	if err != nil {
		panic(err)
	}
	var cred *mgo.Credential
	if ssl {
		tlsConfig, err := mongo.TLSConfig(mongo.Config.SslCAFile, mongo.Config.SslCertFile, mongo.Config.SslKeyFile)
		if err != nil {
			return nil, err
		}
		dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			conn, err := tls.Dial("tcp", addr.String(), tlsConfig)
			return conn, err
		}
		dialInfo.Mechanism = mongo.MongoDbX509Mechanism
		cred = &mgo.Credential{
			Mechanism:   mongo.MongoDbX509Mechanism,
			Source:      "$external",
			Certificate: mongo.LoadX509Cert(mongo.Config.SslCertFile),
		}
	}
	log.Debug("Starting mongo db connection")
	mconn, err := mongo.NewMongoClient(pctxt, dialInfo, cred)
	if err != nil {
		return nil, err
	}
	return mconn, nil
}

func DBConnect(pctxt gogroup.GoGroup, cli client.TsiClient) (db.TrackDB, db.MiscDB, db.SiteDB, db.RegistryDB, db.DeviceDB, *mongo.MongoClient, error) {
	mconn, err := GetMongoClient(pctxt, cli)
	misc := mongo.NewMongoMiscData(pctxt, mconn)
	reg := mongo.NewMongoRegistry(pctxt, mconn)
	tracks := mongo.NewMongoTracks(pctxt, mconn)
	sites := mongo.NewSiteDb(pctxt)
	devices := mongo.NewMongoDeviceDb(pctxt, mconn)
	if err != nil {
		return nil, nil, nil, nil, nil, mconn, err
	}
	return tracks, misc, sites, reg, devices, mconn, nil
}
