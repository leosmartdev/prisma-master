package main

import (
	"container/list"
	"net/http"
	"sync"
	"time"

	"prisma/gogroup"
	"prisma/tms/moc"

	"github.com/golang/protobuf/jsonpb"

	"prisma/tms"
	"prisma/tms/log"
	"prisma/tms/routing"
	"prisma/tms/tmsg"
)

const (
	HistorySize = 50
)

var (
	ReportInterval = time.Duration(5) * time.Second
)

type Router struct {
	sync.Mutex
	ctxt       gogroup.GoGroup
	channels   map[Channel]struct{}
	inbox      chan *tms.TsiMessage
	isFullFill map[Channel]bool
	Local      routing.SiteInfo

	history *list.List
	*http.ServeMux

	totalMessages           uint64
	messagesSinceLastReport uint64
}

func NewRouter(ctxt gogroup.GoGroup, localSite uint32, localName string) *Router {
	r := &Router{
		ctxt:     ctxt,
		channels: make(map[Channel]struct{}),
		inbox:    make(chan *tms.TsiMessage, 32),
		Local: routing.SiteInfo{
			Id:    localSite,
			Name:  localName,
			Local: true,
		},
		isFullFill: make(map[Channel]bool),
		history:    list.New(),
		ServeMux:   http.NewServeMux(),
	}
	r.regDebug()
	ctxt.Go(r.processor)
	return r
}

func (r *Router) regDebug() {
	r.HandleFunc("/history", func(w http.ResponseWriter, req *http.Request) {
		for e := r.history.Front(); e != nil; e = e.Next() {
			msg, ok := e.Value.(*tms.TsiMessage)
			if !ok {
				log.Error("Got unexpected type in history list")
				continue
			}
			m := jsonpb.Marshaler{
				EnumsAsInts:  false,
				EmitDefaults: false,
				Indent:       "  ",
				OrigName:     false,
			}
			err := m.Marshal(w, msg)
			if err != nil {
				log.Error("Error encoding history message to JSON: %v", err)
			} else {
				w.Write([]byte("\n"))
			}
		}
	})

	r.HandleFunc("/listeners", func(w http.ResponseWriter, req *http.Request) {
		r.Lock()
		for l, _ := range r.channels {
			w.Write(l.DebugJSON())
			w.Write([]byte("\n"))
		}
		r.Unlock()
	})
}

// Process incoming messages in the inbox, route them to any listeners which
// may be listening
func (r *Router) processor() {
	tckr := time.NewTicker(ReportInterval)
	defer tckr.Stop()
	for {
		select {
		case <-r.ctxt.Done():
			return
		case <-tckr.C:
			rate := float64(r.messagesSinceLastReport) / ReportInterval.Seconds()
			log.Debug("Total routed: %v, current rate: %v/s", r.totalMessages, rate)
			r.messagesSinceLastReport = 0
		case msg := <-r.inbox:
			r.handle(msg)
		}
	}
}

func (r *Router) handle(msg *tms.TsiMessage) {
	r.totalMessages += 1
	r.messagesSinceLastReport += 1

	dests := 0
	queues := 0
	r.Lock()
	r.history.PushBack(msg)
	for r.history.Len() > HistorySize {
		r.history.Remove(r.history.Front())
	}
	for ch, _ := range r.channels {
		sq, err := ch.GetSendQueue()
		if err != nil {
			log.Warn("Error getting send queue for '%v': %v.", ch.String(), err)
		}
		// If the channel is subscribed to this messages, then send this msg to the channel.
		if ch.ListeningFor(msg) {
			dests += 1
			select {
			case sq <- msg:
				// Nothing. Good send
				queues += 1
				r.isFullFill[ch] = false
			default:
				if isFull, ok := r.isFullFill[ch]; !isFull && ok {
					log.Warn("Send queue for '%v' is full."+
						"Dropping message which we are trying to send", ch.String())
				}
				r.isFullFill[ch] = true
			}
		}
	}
	r.Unlock()
	if dests == 0 || queues == 0 {
		r.NotifyMessage(msg, routing.DeliveryReport_FAILED)
		log.Debug("Message dropped -- no listeners got it. %v matched listeners, %v queues hit. Type: %v", dests, queues, msg.Type())
	}
}

func (r *Router) NotifyMessage(msg *tms.TsiMessage, status routing.DeliveryReport_Status) {
	if msg.NotifySent != 0 && msg.Source != nil && msg.Source.Site == r.Local.Id {
		dr := &routing.DeliveryReport{
			NotifyId: msg.NotifySent,
			Status:   status,
		}
		drAny, err := tmsg.PackFrom(dr)
		if err != nil {
			log.Fatal("Could not pack message: %v", err)
			return
		}

		drMsg := &tms.TsiMessage{
			Destination: []*tms.EndPoint{
				msg.Source,
			},
			Body: drAny,
		}

		r.inbox <- drMsg
	}
}

func (r *Router) AddChannel(ch Channel) {
	// Send service info
	sq, err := ch.GetSendQueue()
	if err != nil {
		log.Error("Could not get send channel for new channel %v: %v", ch.String(), err)
		return
	}
	serviceInfo := routing.ServiceInfo{
		Sites: []*routing.SiteInfo{
			&r.Local,
		},
	}
	for l, _ := range r.channels {
		si := l.SiteInfo()
		if si != nil {
			serviceInfo.Sites = append(serviceInfo.Sites, si)
		}
	}

	body, err := tmsg.PackFrom(&serviceInfo)
	if err != nil {
		panic(err)
	}
	siMsg := &tms.TsiMessage{
		Body: body,
	}
	sq <- siMsg
	// Add channel
	r.Lock()
	r.channels[ch] = struct{}{}
	r.Unlock()
	r.ctxt.Go(r.listen, ch)
}

func (r *Router) listen(l Channel) {
	log.Debug("Opening channel:%v", l.String())
	rcv := l.Recv()
	stop := false
	// Ok Site.connectionStatus
	if l.EndPoint() != nil {
		body, err := tmsg.PackFrom(&moc.Site{
			SiteId:           l.EndPoint().Site,
			ConnectionStatus: moc.Site_Ok,
		})
		if err == nil {
			r.handle(&tms.TsiMessage{
				Destination: []*tms.EndPoint{
					{
						Site: r.Local.Id,
					},
				},
				Body: body,
			})
		}
	}
	// listening
	for !stop {
		select {
		case <-r.ctxt.Done():
			stop = true
		case msg, ok := <-rcv:
			if !ok {
				log.Debug("Closing channel:%v", l.String())
				stop = true
			} else {
				r.Recv(msg)
			}
		}
	}
	// Bad Site.connectionStatus
	if l.EndPoint() != nil {
		body, err := tmsg.PackFrom(&moc.Site{
			SiteId:           l.EndPoint().Site,
			ConnectionStatus: moc.Site_Bad,
		})
		if err == nil {
			r.handle(&tms.TsiMessage{
				Destination: []*tms.EndPoint{
					{
						Site: r.Local.Id,
					},
				},
				Body: body,
			})
		}
	}
	l.Close(nil)
	r.Lock()
	delete(r.channels, l)
	r.Unlock()
}

func (r *Router) Recv(msg *tms.TsiMessage) {
	select {
	case <-r.ctxt.Done():
		return
	case r.inbox <- msg:
		return
	}
}
