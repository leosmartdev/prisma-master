package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"prisma/gogroup"
	"prisma/tms"
	"prisma/tms/goejdb"
	"prisma/tms/log"
	"prisma/tms/routing"
	"prisma/tms/tmsg"

	"github.com/globalsign/mgo/bson"
)

const (
	MsgBuffer      = 8
	NumConnections = 5 // TODO make this per-site variable
)

type RemoteConfig struct {
	Router  *Router
	DB      *goejdb.EjColl
	Info    routing.SiteInfo
	CA      string
	Cert    string
	PrivKey string
}

type DBMessage struct {
	ID       bson.ObjectId   `bson:"_id,omitempty"`
	SiteDest uint32          `bson:"dest,omitempty"`
	SendTime time.Time       `bson:"time,omitempty"`
	Msg      *tms.TsiMessage `bson:"msg,omitempty"`
}

type RemoteSite struct {
	siteInfo routing.SiteInfo

	toRemote     chan *tms.TsiMessage
	fromRemote   chan *tms.TsiMessage
	internalSend chan *tms.TsiMessage

	ctxt   gogroup.GoGroup
	rtr    *Router
	db     *goejdb.EjColl
	dblock sync.Mutex

	senders []*SendWorker

	*http.ServeMux
}

type SendWorker struct {
	*RemoteSite
	deliverers []Deliverer
	routes     []routing.Route
	numSent    uint64
	numSaved   uint64
	numNotSent uint64
}

func NewSendWorker(r *RemoteSite, local tms.EndPoint, cfg RemoteConfig) (*SendWorker, error) {
	s := &SendWorker{
		RemoteSite: r,
	}

	for i, route := range cfg.Info.Routes {
		del, err := NewDeliverer(route, local, cfg, s.ctxt)
		if err != nil {
			return nil, err
		}
		pos := i
		s.deliverers = append(s.deliverers, del)
		s.routes = append(s.routes, *route)
		del.SetStatusCallback(func(status routing.Route_Status) {
			s.routes[pos].Status = status
		})
	}

	return s, nil
}

func NewRemoteSite(ctxt gogroup.GoGroup, cfg RemoteConfig) (*RemoteSite, error) {
	r := &RemoteSite{
		siteInfo:     cfg.Info,
		toRemote:     make(chan *tms.TsiMessage, 32),
		fromRemote:   make(chan *tms.TsiMessage, 32),
		internalSend: make(chan *tms.TsiMessage, MsgBuffer),
		ctxt:         ctxt,
		rtr:          cfg.Router,
		db:           cfg.DB,

		ServeMux: http.NewServeMux(),
	}

	local := tms.EndPoint{
		Site: cfg.Router.Local.Id,
		Aid:  tmsg.APP_ID_TGWAD,
	}

	r.regDebug()
	cfg.Router.AddChannel(r)
	ctxt.GoRestart(r.recieveMsgs)

	for i := 0; i < NumConnections; i++ {
		s, err := NewSendWorker(r, local, cfg)
		if err != nil {
			return nil, err
		}
		r.senders = append(r.senders, s)
	}

	for _, s := range r.senders {
		ctxt.GoRestart(s.sendMsgs)
	}
	log.Debug("Adding remote site: %v", r.String())
	return r, nil
}

func (r *RemoteSite) regDebug() {
	r.HandleFunc("/queue", func(w http.ResponseWriter, req *http.Request) {
		r.dblock.Lock()
		defer r.dblock.Unlock()
		res, jberr := r.db.Find(fmt.Sprintf(`{
			"dest": %v,
		}`, r.siteInfo.Id))
		if jberr != nil {
			log.Error("Error finding messages in db: %v", jberr)
		}

		for _, bsmsg := range res {
			//var dbmsg DBMessage
			var dbmsg map[string]interface{}
			err := bson.Unmarshal(bsmsg, &dbmsg)
			if err != nil {
				log.Error("Error unmarshaling message from database: %v", err)
			} else {
				ret, err := json.MarshalIndent(dbmsg, "", "  ")
				if err != nil {
					log.Error("Error marshalling debug info on IOChannel: %v", err)
					return
				}
				w.Write(ret)
				w.Write([]byte("\n"))
			}
		}
	})

	r.HandleFunc("/status", func(w http.ResponseWriter, req *http.Request) {
		w.Write(r.DebugJSON())
	})
}

func (r *RemoteSite) recieveMsgs() {
	for {
		select {
		case <-r.ctxt.Done():
			return
		case msg := <-r.toRemote:
			select {
			case <-r.ctxt.Done():
				return
			case r.internalSend <- msg:
				// Do nothing. Successful send
			default:
				r.storeMsg(msg)
			}
		}
	}
}

func (r *RemoteSite) storeMsg(msg *tms.TsiMessage) {
	if msg.RealTime {
		// If msg is marked as realtime, don't store (and potentially send
		// delivery failure report)
		r.rtr.NotifyMessage(msg, routing.DeliveryReport_FAILED)
		return
	}

	obj := DBMessage{
		SiteDest: r.siteInfo.Id,
		SendTime: time.Now(),
		Msg:      msg,
	}
	b, err := bson.Marshal(obj)
	if err != nil {
		log.Error("Error marshaling message to bson: %v", err)
		return
	}

	r.dblock.Lock()
	_, jberr := r.db.SaveBson(b)
	if jberr != nil {
		log.Error("Error saving message to db: %v", jberr)
	}
	_, jberr = r.db.Sync()
	if jberr != nil {
		log.Error("Error sync'ing message db: %v", jberr)
	}
	r.dblock.Unlock()
}

func (s *SendWorker) sendStoredMsgs() {
	s.dblock.Lock()

	// TODO sort by age, send oldest first
	res, jberr := s.db.Find(fmt.Sprintf(`{
		"dest": %v,
	}`, s.siteInfo.Id))
	if jberr != nil {
		log.Error("Error finding messages in db: %v", jberr)
	}

	if len(res) == 0 {
		s.dblock.Unlock()
		return
	}
	bsmsg := res[0]
	var dbmsg DBMessage
	err := bson.Unmarshal(bsmsg, &dbmsg)
	// TODO instead of removing, change status
	s.db.RmBson(dbmsg.ID.Hex())
	s.dblock.Unlock()

	if err != nil {
		log.Error("Error unmarshaling message from database: %v", err)
	} else {
		if !s.sendMsg(dbmsg.Msg) {
			s.numSaved += 1
			s.storeMsg(dbmsg.Msg)
		}
	}
}

func (s *SendWorker) sendMsgs() {
	tckr := time.NewTicker(time.Duration(1) * time.Second)
	defer tckr.Stop()
	for {
		s.sendStoredMsgs()
		select {
		case <-s.ctxt.Done():
			return
		case <-tckr.C:
			// Break select every second or so to check for saved messages
		case msg := <-s.internalSend:
			if !s.sendMsg(msg) {
				s.numSaved += 1
				s.storeMsg(msg)
			}
		}
	}
}

func (s *SendWorker) sendMsg(msg *tms.TsiMessage) bool {
	for _, del := range s.deliverers {
		if del.Deliver(msg) {
			s.rtr.NotifyMessage(msg, routing.DeliveryReport_SENT)
			s.numSent += 1
			return true
		}
	}
	s.numNotSent += 1
	return false
}

func (r *RemoteSite) ListeningFor(msg *tms.TsiMessage) bool {
	if msg == nil {
		return false
	}
	for _, endPoint := range msg.Destination {
		if endPoint != nil && endPoint.Site == r.siteInfo.Id {
			return true
		}
	}
	return false
}

func (r *RemoteSite) GetSendQueue() (chan<- *tms.TsiMessage, error) {
	return r.toRemote, nil
}

func (r *RemoteSite) Recv() <-chan *tms.TsiMessage {
	return r.fromRemote
}

func (r *RemoteSite) Close(err error) {
	if err != nil {
		log.Warn("Closing remote site with error: %s", err)
	}

	// This happens at daemon close
	close(r.toRemote)
	close(r.fromRemote)
}

func (r *RemoteSite) String() string {
	return fmt.Sprintf("Site:%v", r.siteInfo.Name)
}

func (r *RemoteSite) DebugJSON() []byte {
	var workers []interface{}
	for _, worker := range r.senders {
		workers = append(workers, map[string]interface{}{
			"routes":      worker.routes,
			"msgsSent":    worker.numSent,
			"msgsSaved":   worker.numSaved,
			"msgsNotSent": worker.numNotSent,
		})
	}

	out := map[string]interface{}{
		"remoteSiteName": r.siteInfo.Name,
		"siteInfo":       r.siteInfo,
		"sendWorkers":    workers,
	}
	ret, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		log.Error("Error marshalling debug info on IOChannel: %v", err)
		return nil
	}
	return ret
}

func (r *RemoteSite) SiteInfo() *routing.SiteInfo {
	return &r.siteInfo
}

func (r *RemoteSite) EndPoint() *tms.EndPoint {
	return nil
}
