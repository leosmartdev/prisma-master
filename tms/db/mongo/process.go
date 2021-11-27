package mongo

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os/exec"
	"strings"
	"sync"
	"time"

	tmsdb "prisma/tms/db"
	"prisma/tms/log"
	tclient "prisma/tms/tmsg/client"

	"github.com/globalsign/mgo"
	pb "github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
)

// MongoDbX509Mechanism is the mechanism defines the protocol for credential negotiation is going to be X509.
const MongoDbX509Mechanism = "MONGODB-X509"

var (
	Config MongoConfig
)

type MongoConfig struct {
	DatabaseName string
	MongoUrl     string
	SslCAFile    string
	SslCertFile  string
	SslKeyFile   string
	DialInfo     *mgo.DialInfo
	Cred         *mgo.Credential
	Ssl          bool
}

func init() {
	Config.Ssl = false
	flag.StringVar(&Config.MongoUrl, "mongo-url", "mongodb://:27017", "MongoDB URL")
	flag.StringVar(&Config.DatabaseName, "mongo-db-name", "trident", "Database name")
	flag.StringVar(&Config.SslCAFile, "sslCAFile", "/etc/trident/mongoCA.crt", "Certificate Authority file for SSL")
	flag.StringVar(&Config.SslCertFile, "sslCertFile", "/etc/trident/mongo.crt", "X509 public key")
	flag.StringVar(&Config.SslKeyFile, "sslKeyfile", "/etc/trident/mongo.key", "X509 private key")
}

type MongoProcess struct {
	config  MongoConfig
	ctxt    context.Context
	exec    string
	version string
	proc    *exec.Cmd
	sess    *mgo.Session
	db      *mgo.Database
}

func NewMongoProcess(ctxt context.Context, client tclient.TsiClient, _ *sync.WaitGroup) (*MongoProcess, error) {
	r := &MongoProcess{
		config: Config,
		ctxt:   ctxt,
	}
	// workaround for ssl ParseURL bug
	ssl := strings.Contains(r.config.MongoUrl, "ssl=true")
	if ssl {
		r.config.MongoUrl = strings.Replace(r.config.MongoUrl, "ssl=true", "", -1)

	}
	// MongoDB
	dialInfo, err := mgo.ParseURL(r.config.MongoUrl)
	if err != nil {
		panic(err)
	}
	if ssl {
		tlsconfig, err := TLSConfig(Config.SslCAFile, Config.SslCertFile, Config.SslKeyFile)
		if err != nil {
			return nil, err
		}
		dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			conn, err := tls.Dial("tcp", addr.String(), tlsconfig)
			return conn, err
		}
		dialInfo.Mechanism = MongoDbX509Mechanism
		r.config.Cred = &mgo.Credential{
			Mechanism:   MongoDbX509Mechanism,
			Source:      "$external",
			Certificate: LoadX509Cert(Config.SslCertFile),
		}
	}
	// trident
	r.config.DialInfo = dialInfo
	err = r.connect()
	if err != nil {
		return nil, err
	}
	client.RegisterHandler("prisma.tms.db.DBConnectionRequest", func(_ *tclient.TMsg) pb.Message {
		var Authkey string
		// This is to make sure that DBConnectionParams does not report MongoDbX509Mechanism is ssl is false
		if ssl {
			Authkey = MongoDbX509Mechanism
		}
		return &tmsdb.DBConnectionParams{
			Engine:    tmsdb.DatabaseEngine_MongoDB,
			Addresses: []string{r.config.MongoUrl},
			Database:  r.config.DatabaseName,
			Authkey:   Authkey,
		}
	})
	return r, nil
}

// TLSConfig take --sslCAfile and --sslPEMKeyFile and return tlsconfig
func TLSConfig(caf, cf, kf string) (*tls.Config, error) {
	// --sslCAFile /etc/ssl/certs/mongodb-cert.pem
	rootCerts := x509.NewCertPool()
	ca, err := ioutil.ReadFile(caf)
	if err != nil {
		return nil, err
	}
	rootCerts.AppendCertsFromPEM(ca)

	// --sslPEMKeyFile /etc/ssl/mongodb.pem
	clientCerts := []tls.Certificate{}
	cert, err := tls.LoadX509KeyPair(cf, kf)
	if err != nil {
		return nil, err
	}
	clientCerts = append(clientCerts, cert)
	return &tls.Config{
		InsecureSkipVerify: true,
		RootCAs:            rootCerts,
		Certificates:       clientCerts,
	}, nil
}

// LoadX509Cert extracts and parses an x509 cert
func LoadX509Cert(cf string) *x509.Certificate {
	cb, err := ioutil.ReadFile(cf)
	if err != nil {
		panic("Failed to open certificate PEM file")
	}
	block, _ := pem.Decode(cb)
	if block == nil {
		panic("failed to parse certificate PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic("failed to parse certificate: " + err.Error())
	}
	return cert
}

func (r *MongoProcess) DialInfo() *mgo.DialInfo {
	return r.config.DialInfo
}

func (r *MongoProcess) Cred() *mgo.Credential {
	return r.config.Cred
}

func (r *MongoProcess) connect() error {
	log.Info("Connecting to MongoDB %v on %v", r.config.DatabaseName, r.config.MongoUrl)
	var err error
	dialInfo := r.config.DialInfo
	dialInfo.Database = r.config.DatabaseName
	dialInfo.Timeout = time.Duration(30) * time.Second
	log.Debug("DialInfo %+v", dialInfo)
	for i := 0; i < 60; i++ {
		r.sess, err = mgo.DialWithInfo(dialInfo)
		defer r.sess.Close()
		if err == nil {
			log.Info("Connected to MongoDB %v on %v", r.config.DatabaseName, r.config.MongoUrl)
			r.db = r.sess.DB(r.config.DatabaseName)
			r.sess.SetCursorTimeout(time.Duration(0))
			r.sess.SetSocketTimeout(time.Duration(5) * time.Minute)
			names, err := r.sess.DatabaseNames()
			if err != nil {
				log.Debug("Not authorized: %+v", err)
			}
			if dialInfo.Mechanism == MongoDbX509Mechanism {
				log.Debug("Authenticating with username %+v ...")
				log.Debug(fmt.Sprintf("%+v", r.config.Cred.Certificate.Subject.ToRDNSequence()))
				if r.sess.Login(r.config.Cred) != nil {
					return err
				}
				names, err = r.sess.DatabaseNames()
				if err != nil {
					log.Error("Errors in sess auth %+v", err)
				}
			}
			log.Debug("logged on %+v", names)
			return nil
		}
		time.Sleep(time.Duration(1) * time.Second)
	}
	log.Info("conn error is: %+v", err)
	return err
}
