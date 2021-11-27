// Package mongo provides implementations of db abstractions, additional function to reach mongo features.
package mongo

import (
	"expvar"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"prisma/gogroup"
	"prisma/tms/log"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const (
	DATABASE        = "trident"
	RetryAttempts   = 5
	averageSessions = 16
)

var (
	stats = expvar.NewMap("database")
	trace = log.GetTracer("db-client")
)

func init() {
	stats.Add("free", 0)
	stats.Add("used", 0)
}

// NewMongoClient create a client mongo session using connect() function
func NewMongoClient(ctxt gogroup.GoGroup, dialInfo *mgo.DialInfo, cred *mgo.Credential) (*MongoClient, error) {
	c := &MongoClient{
		DialInfo: dialInfo,
		free:     make([]*mgo.Session, 0, 64),
		used:     make(map[*mgo.Session]struct{}),
		Ctxt:     ctxt,
		Cred:     cred,
	}
	log.Debug("Mongo client %+v", c.DialInfo)
	err := c.connect()
	if err != nil {
		return nil, err
	}
	ctxt.GoRestart(c.printStats)
	return c, nil
}

/****
 * A single MongoDB connection enforces a bunch of operation ordering
 * semantics, most of which we only want per-operation not overall in the
 * daemon. What we typically want instead is one database connection per
 * operation, wherein operation is something like a tracks.Get() or
 * misc_data.Upsert(). This struct maintains a pool of database connections so
 * we don't have to re-connect for each operation.
 */
type MongoClient struct {
	DialInfo *mgo.DialInfo
	lock     sync.Mutex
	free     []*mgo.Session
	used     map[*mgo.Session]struct{} // Set of in use connections
	Cred     *mgo.Credential

	Ctxt gogroup.GoGroup
}

// Create a connection to mongo, append it to the 'free' list
func (c *MongoClient) connect() error {
	mgo.SetStats(true)
	var (
		err  error
		sess *mgo.Session
	)
	for i := 0; i < 10; i++ {
		trace.Logf("Attempting to connect to MongoDB: %v", c.DialInfo)
		sess, err = mgo.DialWithInfo(c.DialInfo)
		if err == nil {
			trace.Logf("Connected to MongoDB")
			sess.SetBatch(5000)
			sess.SetPrefetch(0.75)
			sess.SetCursorTimeout(time.Duration(0))
			sess.SetSocketTimeout(time.Duration(5) * time.Minute)
			//sess.SetPoolLimit(5)
			names, err := sess.DatabaseNames()
			if err != nil {
				log.Debug("Not authorized: %+v", err)
			}
			if c.DialInfo.Mechanism == MongoDbX509Mechanism {
				log.Debug("Authenticating...")
				if sess.Login(c.Cred) != nil {
					return err
				}
				names, _ = sess.DatabaseNames()
				log.Debug("logged to %+v", names)
			}
			sess.SetMode(mgo.SecondaryPreferred, true)
			c.free = append(c.free, sess)
			stats.Add("free", 1)
			return nil
		}
		log.Warn("Could not connect to MongoDB: %v. (Maybe it's still starting.)", err)
		time.Sleep(time.Duration(1) * time.Second)
	}
	return err
}

func (c *MongoClient) printStats() {
	tckr := time.NewTicker(15 * time.Second)
	defer tckr.Stop()
	for {
		select {
		case <-c.Ctxt.Done():
			return
		case <-tckr.C:
			c.lock.Lock()
			trace.Logf("Mongo connection stats used: %v, free: %v, driver: %+v",
				len(c.used), len(c.free), mgo.GetStats())
			c.lock.Unlock()
		}
	}
}

func (c *MongoClient) DB() *mgo.Database {
	return c.Sess().DB(DATABASE)
}

func (c *MongoClient) Sess() *mgo.Session {
	unlocked := true
	defer func() {
		if !unlocked {
			c.lock.Unlock()
		}
	}()

	attempts := len(c.free) + RetryAttempts
	for i := 0; i < attempts; i++ {
		trace.Logf("getting session")
		c.lock.Lock()
		unlocked = false
		if len(c.free) == 0 {
			trace.Logf("creating new session, none available")
			err := c.connect()

			if err != nil {
				panic(err)
			}
		}
		sess := c.free[len(c.free)-1]
		c.free = c.free[0 : len(c.free)-1]
		stats.Add("free", -1)
		if _, ok := c.used[sess]; !ok {
			stats.Add("used", 1)
		}
		c.used[sess] = struct{}{}
		c.lock.Unlock()
		unlocked = true

		err := sess.Ping()
		if err == nil {
			trace.Logf("%p: got session, %v used, %v free", sess,
				len(c.used), len(c.free))
			return sess
		}
		c.err(sess, err)
	}

	log.Error("Could not get connection to Mongo!")
	return nil
}

// Migrate is used to up schema for mongodb
func (c *MongoClient) Migrate(schemaFile string) error {
	f, err := os.OpenFile(schemaFile, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	return c.DB().Run(bson.M{"eval": string(b)}, nil)
}

// EnsureSetUp executes schemas under a given directory
func (c *MongoClient) EnsureSetUp(dirs []string) {
	for _, dir := range dirs {
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			log.Error("Failed to ensure mongodb schemas set up: %+v", err)
		}
		for _, file := range files {
			if err := c.Migrate(dir + "/" + file.Name()); err != nil {
				log.Error("%s failed to execute: %+v", file.Name(), err)
			}
			log.Info("%s was executed.", file.Name())
		}
	}
	return
}

func (c *MongoClient) err(sess *mgo.Session, err error) {
	c.lock.Lock()
	if _, ok := c.used[sess]; ok {
		delete(c.used, sess)
		stats.Add("used", -1)
	}
	c.lock.Unlock()
	log.Error("Error reported in mongo connection: %v", err)
	trace.Logf("%p: session no longer valid: %v", sess, err)
}

func (c *MongoClient) release(sess *mgo.Session) {
	trace.Logf("%p: releasing session", sess)
	err := sess.Ping()
	if err != nil {
		c.err(sess, err)
	} else {
		c.lock.Lock()
		defer c.lock.Unlock()
		if _, ok := c.used[sess]; !ok {
			trace.Logf("%p: session was not in use", sess)
			return
		}
		delete(c.used, sess)
		stats.Add("used", -1)
		if len(c.free) > averageSessions {
			sess.Close()
		} else {
			c.free = append(c.free, sess)
			stats.Add("free", 1)
		}
		trace.Logf("%p: session released, %v used, %v free", sess, len(c.used),
			len(c.free))
	}
}

func (c *MongoClient) Release(i interface{}) {
	if i == nil {
		return
	}
	if db, ok := i.(*mgo.Database); ok {
		c.Release(db.Session)
		return
	}
	if sess, ok := i.(*mgo.Session); ok {
		c.release(sess)
		return
	}
	panic("Trying to release unknown type!")
}
