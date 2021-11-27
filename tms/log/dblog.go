package log

import (
	"encoding/json"
	reallog "log"
	log "log/syslog"
	"runtime"
	"time"

	"prisma/tms"
	"prisma/tms/goejdb"
	. "prisma/tms/tmsg/client"

	"github.com/globalsign/mgo/bson"
)

type DBLogger interface {
	Write(prio log.Priority, file string, line int, msg string, objs []interface{})
}

func NewDBLogger(client TsiClient, fn string) (DBLogger, error) {
	ejdb, err := goejdb.Open(fn, goejdb.JBOCREAT|goejdb.JBOWRITER|goejdb.JBOREADER)
	if err != nil {
		return nil, err
	}
	coll, err := ejdb.CreateColl("logmsgs", &goejdb.EjCollOpts{
		Large:         true,
		Compressed:    true,
		Records:       128000,
		CachedRecords: 0,
	})
	if err != nil {
		return nil, err
	}
	ret := &dbLoggerImpl{
		db:     ejdb,
		coll:   coll,
		client: client,
	}
	go ret.syncProcess()
	return ret, nil
}

type dbLoggerImpl struct {
	db     *goejdb.Ejdb
	coll   *goejdb.EjColl
	client TsiClient
}

type codeLoc struct {
	File string
	Line int
	Pc   uintptr
}

type record struct {
	Endpoint tms.EndPoint
	Time     time.Time
	Prio     log.Priority
	File     string
	Line     int
	Msg      string
	Objs     string
	Stack    string
}

func (l *dbLoggerImpl) Write(prio log.Priority, file string, line int, msg string, objs []interface{}) {
	ep := l.client.Local()
	if ep == nil {
		ep = &tms.EndPoint{}
	}
	objStr, err := json.Marshal(objs)
	if err != nil {
		reallog.Printf("Error marshalling objs to json: %v", err)
		objStr = []byte{}
	}
	r := record{
		Endpoint: *ep,
		Time:     time.Now(),
		Prio:     prio,
		File:     file,
		Line:     line,
		Msg:      msg,
		Objs:     string(objStr),
	}

	if prio >= LOG_TRACE {
		skip := 1
		stack := make([]codeLoc, 0, 16)
		for {
			pc, file, line, ok := runtime.Caller(skip)
			if !ok {
				break
			}

			stack = append(stack, codeLoc{
				File: file,
				Line: line,
				Pc:   pc,
			})
			skip++
		}

		stackStr, err := json.Marshal(stack)
		if err != nil {
			reallog.Printf("Error marshalling stack to json: %v", err)
			stackStr = []byte{}
		}
		r.Stack = string(stackStr)
	}
	l.WriteRecord(&r)
}

func (l *dbLoggerImpl) WriteRecord(r *record) {
	bsondata, err := bson.Marshal(*r)
	if err != nil {
		reallog.Printf("[dblog] Error marshalling record: %v", err)
	} else {
		_, err := l.coll.SaveBson(bsondata)
		if err != nil {
			reallog.Printf("[dblog] Error saving record bson: %v", err)
		}
	}
}

func (l *dbLoggerImpl) syncProcess() {
	for {
		time.Sleep(time.Duration(1) * time.Second)
		_, err := l.db.Sync()
		if err != nil {
			reallog.Printf("[dblog] Error saving record bson: %v", err)
		}
	}
}
