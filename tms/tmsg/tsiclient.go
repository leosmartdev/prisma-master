// Package tmsg provides API to communicate via tgwad daemon.
package tmsg

import (
	"bufio"
	"errors"
	"flag"
	"math/rand"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"prisma/gogroup"
	"prisma/tms"
	"prisma/tms/envelope"
	"prisma/tms/log"
	"prisma/tms/routing"
	"prisma/tms/tmsg/client"

	pb "github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/wrappers"
	"golang.org/x/net/context"
)

var (
	tgwadAddr     = "localhost:31228"
	entry     int = int(ENTRY_ID_UNKNOWN)
	GClient   client.TsiClient
)

func init() {
	flag.StringVar(&tgwadAddr, "host", "localhost:31228", "address:port of tgwad")
	flag.IntVar(&entry, "entryid", int(ENTRY_ID_UNKNOWN), "Entry ID for process")
}

func TsiClientGlobal(ctxt gogroup.GoGroup, appid uint32) error {
	conn, err := ConnectTsiClient(ctxt, tgwadAddr, appid)
	GClient = conn
	return err
}

func ConnectTsiClient(ctxt gogroup.GoGroup, address string, appid uint32) (*TsiClientTcp, error) {
	c := &TsiClientTcp{
		address:         address,
		outbox:          make(chan *tms.TsiMessage, 32),
		conn:            nil,
		writer:          nil,
		reader:          nil,
		site:            TMSG_LOCAL_SITE,
		app:             appid,
		entry:           uint32(entry),
		si:              nil,
		sites:           make(map[string]*routing.SiteInfo),
		listeners:       make(map[chan *client.TMsg]InternalListener),
		handlers:        make(map[string]func(*client.TMsg) pb.Message),
		cmdSeq:          1,
		notifyCallbacks: make(map[int32]func(*routing.DeliveryReport)),
	}

	if c.entry == ENTRY_ID_UNKNOWN {
		// Default to AppId * 1000
		c.entry = c.app * 1000
	}
	// Listen for my AppID on the local site
	c.Lock()
	c.registry.Entries = append(c.registry.Entries, &routing.Listener{
		Destination: &tms.EndPoint{
			Site: uint32(TMSG_LOCAL_SITE),
			Aid:  uint32(appid),
		},
	})
	c.registry.Entries = append(c.registry.Entries, &routing.Listener{
		Destination: c.Local(),
	})
	c.Unlock()
	siChan := c.internalListen(ctxt, routing.Listener{
		MessageType: "prisma.tms.routing.ServiceInfo",
	})
	ctxt.Go(func() { c.siProcess(ctxt, siChan) })
	ctxt.Go(func() { c.ioProcess(ctxt) })

	// Wait for initial SI
	log.Debug("Waiting for service info from tgwad...")
	for c.si == nil {
		time.Sleep(time.Duration(100) * time.Millisecond)
	}

	c.RegisterHandler("prisma.tms.routing.Ping", c.handlePing)
	return c, nil
}

type TsiClientTcp struct {
	envelope.Subscriber
	// This lock protects everything here. Only used for fields which have
	// multi-gorouting writes/accesses
	sync.Mutex

	address string
	outbox  chan *tms.TsiMessage
	conn    net.Conn
	writer  *bufio.Writer
	reader  *bufio.Reader
	site    uint32
	app     uint32
	entry   uint32
	si      *routing.ServiceInfo
	sites   map[string]*routing.SiteInfo

	registry  routing.Registry
	listeners map[chan *client.TMsg]InternalListener
	handlers  map[string]func(*client.TMsg) pb.Message
	cmdSeq    uint32
	// Notification IDs currently in use
	notifyCallbacks map[int32]func(*routing.DeliveryReport)

	// For GenerateTSN
	lastTSN tms.TimeSerialNumber
}

type InternalListener struct {
	ctxt     context.Context
	listener routing.Listener
}

func (c *TsiClientTcp) Publish(envelope envelope.Envelope) {
	timeNow, _ := ptypes.TimestampProto(time.Now())
	body, err := PackFrom(&envelope)
	if err != nil {
		log.Error("envelope pack error", envelope)
	}
	m := &tms.TsiMessage{
		Source: c.Local(),
		Destination: []*tms.EndPoint{
			{
				Site: TMSG_HQ_SITE,
			},
		},
		WriteTime: timeNow,
		SendTime:  timeNow,
		Body:      body,
	}
	c.Send(context.Background(), m)
}

func (c *TsiClientTcp) ioProcess(ctxt gogroup.GoGroup) {
	done := ctxt.Done()
	group := ctxt.Child("io")
	for {
		select {
		case <-done:
			return
		default:
			if c.conn == nil {
				group.Cancel(nil)
				c.writer = nil
				c.reader = nil
				conn, err := net.Dial("tcp", c.address)
				if err != nil {
					log.Warn("Could not connect to tgwad: %v", err)
				} else {
					group = ctxt.Child("io")
					c.reader = bufio.NewReaderSize(conn, MaxMessageSize)
					c.writer = bufio.NewWriterSize(conn, MaxMessageSize)
					c.conn = conn
					group.Go(c.receive)
					group.Go(c.send)
					log.Debug("Connected to tgwad")
					c.sendRegistry(group)
				}
			}
			time.Sleep(time.Duration(1) * time.Second)
		}
	}
}

func (c *TsiClientTcp) siProcess(ctxt context.Context, sichan <-chan *client.TMsg) {
	for {
		select {
		case <-ctxt.Done():
			return
		case msg := <-sichan:
			si, ok := msg.Body.(*routing.ServiceInfo)
			if !ok {
				log.Error("Service Info process got non-SI: %v", msg)
				continue
			}
			c.updateSI(ctxt, si)
		}
	}
}

func (c *TsiClientTcp) updateSI(ctxt context.Context, si *routing.ServiceInfo) {
	log.Debug("Got service info from tgwad: %v", si)
	c.si = si
	oldSite := c.site
	for _, site := range si.Sites {
		if site.Local {
			c.site = site.Id
		}

		c.sites[site.Name] = site
	}

	if oldSite != c.site {
		// Re-add registry entries w/ new site id
		c.Lock()
		c.registry.Entries = append(c.registry.Entries, &routing.Listener{
			Destination: &tms.EndPoint{
				Site: c.site,
				Aid:  c.app,
			},
		})
		c.registry.Entries = append(c.registry.Entries, &routing.Listener{
			Destination: c.Local(),
		})
		c.Unlock()
		c.sendRegistry(ctxt)
	}

}

func (c *TsiClientTcp) send(ctxt gogroup.GoGroup) {
	done := ctxt.Done()
	for {
		select {
		case <-done:
			return
		case out := <-c.outbox:
			out.SendTime = tms.Now()
			err := WriteTsiMessage(ctxt, c.writer, out)
			log.TraceMsg("Sent message: %v, %v", out, err)
			if err != nil {
				c.conn = nil
				return
			}
		}
	}
}

func (c *TsiClientTcp) receive(ctxt gogroup.GoGroup) {
	for {
		select {
		case <-ctxt.Done():
			return
		default:
		}
		msg, err := ReadTsiMessage(ctxt, c.reader)
		if err != nil {
			if err != RetryError {
				log.Error("Error reading message: %v", err)
				c.conn = nil
				return
			}
		} else if msg != nil {
			c.handle(ctxt, msg)
		}
	}
}

func (c *TsiClientTcp) handle(ctxt gogroup.GoGroup, msg *tms.TsiMessage) {
	if msg.Status == tms.TsiMessage_ACK {
		// Clients don't use ACK signals. They are for inter-site only, but
		// tgwad always responds with them. So we can just drop them on the floor here.
		return
	}

	// Unpack the message and create a TMsg
	body, err := Unpack(msg.Body)
	if err != nil {
		log.Error("Error unpacking body: %v", err)
		return
	}
	tmsg := &client.TMsg{
		*msg, body,
	}
	ty := MessageType(&tmsg.TsiMessage)

	log.Trace(tmsg)

	// Find all the listeners which match this message
	handlers := make([]chan *client.TMsg, 0, 8)
	c.Lock()
	for ch, il := range c.listeners {
		if Matches(il.listener, &tmsg.TsiMessage, c.site) {
			handlers = append(handlers, ch)
		}
	}
	c.Unlock()

	// Was a handler of some sort found?
	foundHandler := false

	if ty == "prisma.tms.routing.DeliveryReport" {
		rpt, ok := body.(*routing.DeliveryReport)
		if !ok {
			log.Warn("Got odd unpacked type for DeliveryReport: %v", body)
		} else {
			c.Lock()
			defer c.Unlock()
			callback, ok := c.notifyCallbacks[rpt.NotifyId]
			if ok {
				// Notify callback exists!
				delete(c.notifyCallbacks, rpt.NotifyId)
				ctxt.Child("notify_handler").Go(callback, rpt)
				foundHandler = true
			}
		}
	}

	// Clear body after finding listeners
	msg.Body = nil

	if msg.Status == tms.TsiMessage_REQUEST {
		if handler, ok := c.handlers[ty]; ok {
			foundHandler = true
			ctxt.Go(func() {
				resp := handler(tmsg)
				if resp != nil {
					body, err := PackFrom(resp)
					if err != nil {
						log.Error("Packing into Any: %v", err, resp, handler, c)
						return
					}
					respMsg := tms.TsiMessage{
						Source: c.Local(),
						Destination: []*tms.EndPoint{
							msg.Source,
						},
						Status: tms.TsiMessage_REPLY,
						Body:   body,
					}
					if msg.CommandSequence != nil {
						respMsg.CommandSequence = &wrappers.UInt32Value{
							Value: msg.CommandSequence.Value,
						}
					}
					c.Send(ctxt, &respMsg)
				}
			})
		}
	}

	if len(handlers) == 0 && !foundHandler {
		log.Warn("Warning: found no handlers for %v, %v", ty, tmsg)
	}

	// Send the TMsg to each of the listeners
	done := ctxt.Done()
	for _, h := range handlers {
		select {
		case <-done:
			return
		case h <- tmsg:
			// Do nothing
		}
	}
}

// Sending functions (of various sorts)
func (c *TsiClientTcp) Send(ctxt context.Context, msg *tms.TsiMessage) {
	msg.Source = c.Local()
	select {
	case <-ctxt.Done():
		return
	case c.outbox <- msg:
		return
	}
}

func (c *TsiClientTcp) NotifyID(callback func(*routing.DeliveryReport)) int32 {
	c.Lock()
	defer c.Unlock()

	// Assume that the notifyIDs is quite sparse and do something simple to
	// generate the IDs
	for true {
		id := rand.Int31()
		_, exists := c.notifyCallbacks[id]
		if !exists {
			c.notifyCallbacks[id] = callback
			return id
		}
	}
	panic("Internal error 0x7823498")
}

func (c *TsiClientTcp) SendNotify(ctxt context.Context, tmsg *tms.TsiMessage, callback func(*routing.DeliveryReport)) {
	tmsg.NotifySent = c.NotifyID(callback)
	c.Send(ctxt, tmsg)
}

func (c *TsiClientTcp) SendTo(ctxt context.Context, ep tms.EndPoint, unpackedBody pb.Message) {
	body, err := PackFrom(unpackedBody)
	if err != nil {
		log.Error("Packing into Any: %v", err, unpackedBody, c)
		return
	}

	msg := tms.TsiMessage{
		Source: c.Local(),
		Destination: []*tms.EndPoint{
			&ep,
		},
		WriteTime:   tms.Now(),
		RequestTime: tms.Now(),

		Body: body,
	}
	c.Send(ctxt, &msg)
}

func (c *TsiClientTcp) BroadcastLocal(ctxt context.Context, unpackedBody pb.Message) {
	body, err := PackFrom(unpackedBody)
	if err != nil {
		log.Error("Packing into Any: %v", err, unpackedBody, c)
		return
	}

	msg := tms.TsiMessage{
		Source: c.Local(),
		Destination: []*tms.EndPoint{
			{
				Site: c.site,
			},
		},
		WriteTime:   tms.Now(),
		RequestTime: tms.Now(),

		Body: body,
	}
	c.Send(ctxt, &msg)
}

func (c *TsiClientTcp) SendToGateway(ctxt context.Context, unpackedBody pb.Message) {
	// TODO
	panic("Unimplemented")
}

func (c *TsiClientTcp) RegisterHandler(msgType string, handler func(*client.TMsg) pb.Message) {
	c.Lock()
	if handler == nil {
		delete(c.handlers, msgType)
	} else {
		c.handlers[msgType] = handler
	}
	c.Unlock()
}

func (c *TsiClientTcp) Request(ctxt context.Context, ep tms.EndPoint, msg pb.Message) (pb.Message, error) {

	body, err := PackFrom(msg)
	if err != nil {
		log.Error("Packing into Any: %v", err, msg, c)
		return nil, err
	}

	// TODO: Using cmdSeq like this, we will get duplicate command series every
	// 2^32 RPC calls. That's probably OK unless they are a ton of RPC calls
	// and some a very long lived. Ideally, we'd keep track of which cmdSeq IDs
	// are already in use and not re-use them. There's no great mechanism for
	// this right now and it's probably not necessary.
	id := atomic.AddUint32(&c.cmdSeq, 1)
	reqMsg := tms.TsiMessage{
		Source: c.Local(),
		Destination: []*tms.EndPoint{
			&ep,
		},
		WriteTime:   tms.Now(),
		RequestTime: tms.Now(),
		RealTime:    true,
		Status:      tms.TsiMessage_REQUEST,
		CommandSequence: &wrappers.UInt32Value{
			Value: id,
		},

		Body: body,
	}

	lclCtxt, cancel := context.WithCancel(ctxt)
	defer cancel()

	ch := make(chan *client.TMsg, 1)
	il := InternalListener{
		ctxt: lclCtxt,
		listener: routing.Listener{
			Source: &ep,
		},
	}

	c.Lock()
	c.listeners[ch] = il
	c.Unlock()
	defer func() {
		c.Lock()
		delete(c.listeners, ch)
		c.Unlock()
	}()

	failchan := make(chan struct{})
	c.SendNotify(lclCtxt, &reqMsg, func(rpt *routing.DeliveryReport) {
		if rpt.Status != routing.DeliveryReport_SENT {
			failchan <- struct{}{}
		}
	})

	for {
		select {
		case <-ctxt.Done():
			return nil, ctxt.Err()
		case <-failchan:
			return nil, errors.New("Message failed sending")
		case msg := <-ch:
			if msg.Status == tms.TsiMessage_REPLY &&
				msg.CommandSequence != nil &&
				msg.CommandSequence.Value == id {
				return msg.Body, nil
			} else {
				log.TraceMsg("RPC Request rejected response: %v", msg)
			}
		}
	}
}

func (c *TsiClientTcp) sendRegistry(ctxt context.Context) {
	c.Lock()
	c.registry.SourceService = c.Local()
	body, err := PackFrom(&c.registry)
	if err != nil {
		log.Error("Error packing Listener: %v", err)
	}
	c.Unlock()

	req := tms.TsiMessage{
		Source:      c.Local(),
		Destination: []*tms.EndPoint{c.LocalRouter()},
		Body:        body,
	}
	c.Send(ctxt, &req)
}

// Listen for messages
func (c *TsiClientTcp) internalListen(ctxt context.Context, l routing.Listener) <-chan *client.TMsg {
	ch := make(chan *client.TMsg, 32)
	il := InternalListener{
		ctxt:     ctxt,
		listener: l,
	}

	c.Lock()
	c.listeners[ch] = il
	c.Unlock()

	return ch
}

func (c *TsiClientTcp) Listen(ctxt context.Context, l routing.Listener) <-chan *client.TMsg {
	ch := c.internalListen(ctxt, l)

	c.Lock()
	c.registry.Entries = append(c.registry.Entries, &l)
	c.Unlock()
	c.sendRegistry(ctxt)
	return ch
}

// General info
func (c *TsiClientTcp) Local() *tms.EndPoint {
	return &tms.EndPoint{
		Site: uint32(c.site),
		Aid:  uint32(c.app),
		Eid:  uint32(c.entry),
		Pid:  uint32(os.Getpid()),
	}
}

func (c *TsiClientTcp) LocalRouter() *tms.EndPoint {
	return &tms.EndPoint{
		Site: uint32(c.site),
		Aid:  uint32(APP_ID_TGWAD),
	}
}

// Resolution funcs

func (c *TsiClientTcp) ResolveSite(sitename string) uint32 {
	if sitename == "" {
		return c.site
	}
	if sitename == "local" {
		return TMSG_LOCAL_SITE
	}

	si, ok := c.sites[sitename]
	if !ok {
		return TMSG_UNKNOWN_SITE
	}

	return si.Id
}

func (c *TsiClientTcp) ResolveApp(sitename string) uint32 {
	for id, name := range AppIdNames {
		if name == sitename {
			return id
		}
	}
	return APP_ID_UNKNOWN
}

func (c *TsiClientTcp) handlePing(msg *client.TMsg) pb.Message {
	ping, ok := msg.Body.(*routing.Ping)
	if !ok {
		log.Warn("handlePing got non-ping: %v", msg)
		return nil
	}
	log.Debug("Got ping, responding...")
	ping.PongSendTime = tms.Now()
	return ping
}

func (c *TsiClientTcp) GenerateTSN() tms.TimeSerialNumber {
	now := time.Now()

	c.Lock()
	if c.lastTSN.Seconds != now.Unix() {
		c.lastTSN = tms.TimeSerialNumber{
			Seconds: now.Unix(),
			Counter: 0,
		}
	}
	c.lastTSN.Counter++
	tsn := c.lastTSN
	c.Unlock()
	return tsn
}

// Generate Target TimeSerialNumber
func (c *TsiClientTcp) GenerateTargetTSN() *tms.TargetID {
	tsn := c.GenerateTSN()
	return &tms.TargetID{
		Producer: &tms.SensorID{
			Site: c.site,
			Eid:  c.entry,
		},
		SerialNumber: &tms.TargetID_TimeSerial{
			TimeSerial: &tsn,
		},
	}
}
