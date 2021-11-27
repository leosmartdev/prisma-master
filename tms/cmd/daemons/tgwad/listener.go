package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"

	"prisma/gogroup"
	. "prisma/tms"
	"prisma/tms/libmain"
	"prisma/tms/log"
	"prisma/tms/routing"
	"prisma/tms/tmsg"
)

const (
	ClientBuffer = 128 // Allow 128 buffered messages before they start getting dropped
)

type Channel interface {
	ListeningFor(msg *TsiMessage) bool
	GetSendQueue() (chan<- *TsiMessage, error)
	Recv() <-chan *TsiMessage
	Close(err error)
	String() string
	DebugJSON() []byte
	SiteInfo() *routing.SiteInfo
	EndPoint() *EndPoint
}

type IOChannel struct {
	sync.RWMutex

	io io.ReadWriteCloser
	r  *bufio.Reader
	w  *bufio.Writer

	rtr        *Router
	toRouter   chan *TsiMessage
	fromRouter chan *TsiMessage
	acks       chan uint64
	grp        gogroup.GoGroup

	msgsSent uint64
	msgsRecv uint64

	reg routing.Registry
}

func NewIOChannel(io io.ReadWriteCloser, rtr *Router) Channel {
	c := &IOChannel{
		io:         io,
		r:          bufio.NewReaderSize(io, tmsg.MaxMessageSize),
		w:          bufio.NewWriterSize(io, 32*1024),
		rtr:        rtr,
		toRouter:   make(chan *TsiMessage, 8),
		fromRouter: make(chan *TsiMessage, ClientBuffer),
		acks:       make(chan uint64, 8),
		grp:        libmain.TsiKillGroup.Child("iochannel"),
	}
	c.grp.Go(c.readProcess)
	c.grp.Go(c.writeProcess)
	return c
}

func (c *IOChannel) DebugJSON() []byte {
	out := map[string]interface{}{
		"name":      c.String(),
		"reg":       c.reg,
		"msgsSent":  c.msgsSent,
		"msgsRecvd": c.msgsRecv,
	}
	ret, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		log.Error("Error marshalling debug info on IOChannel: %v", err)
		return nil
	}
	return ret
}

func (c *IOChannel) readProcess() {
	for {
		if c.grp.Canceled() {
			return
		}
		msg, id, err := tmsg.ReadTsiMessageExtended(c.grp, c.r)
		if err != nil {
			if err != io.EOF {
				log.Error("Could not read tmsg from io: %v", err)
			}
			c.Close(err)
		} else {
			c.handleMessage(msg)
			c.acks <- id
		}
	}
}

func (c *IOChannel) handleMessage(msg *TsiMessage) {
	if msg.Status == TsiMessage_ACK {
		// Drop ACKs we receive on the floor. For now, we don't care about them
		return
	}
	if msg.Status == TsiMessage_KEEPALIVE {
		// Drop Keepalives on the floor also. This is the intended behavior
		return
	}
	c.msgsSent += 1
	ty := ""
	if msg.Body != nil {
		ty = msg.Body.TypeUrl
	}
	log.TraceMsg("[%v] got msg from client '%s': \n%v", c.String(), ty)

	if msg.Body != nil {
		tyUrl := strings.TrimPrefix(msg.Body.TypeUrl, "type.googleapis.com/")
		switch tyUrl {
		case "prisma.tms.routing.Registry":
			var reg routing.Registry
			tmsg.UnpackTo(msg.Body, &reg)
			c.reg = reg
			return // Don't forward to router
		}
	}

	// By default, forward message on to the router
	select {
	case <-c.grp.Done():
		return
	case c.toRouter <- msg:
		return
	}
}

func (c *IOChannel) writeProcess() {
	for {
		select {
		case <-c.grp.Done():
			return
		case id := <-c.acks:
			msg := &TsiMessage{
				Status: TsiMessage_ACK,
			}
			err := tmsg.WriteTsiMessageExtended(c.grp, c.w, msg, tmsg.Opts{ID: id})
			if err != nil {
				log.Error("Could not write tmsg to io: %v", err)
				c.grp.Cancel(err)
			} else {
				log.TraceMsg("[%v] sent message to client: %v", c.String(), msg)
			}

		case msg, ok := <-c.fromRouter:
			if !ok {
				// Channel closed by router. Time to die!
				c.grp.Cancel(io.EOF)
			}

			err := tmsg.WriteTsiMessage(c.grp, c.w, msg)
			c.msgsRecv += 1
			if err != nil {
				log.Error("Could not write tmsg to io: %v", err)
				c.grp.Cancel(err)
			} else {
				log.TraceMsg("[%v] sent message to client: %v", c.String(), msg)
			}
		}
	}
}

func (c *IOChannel) ListeningFor(msg *TsiMessage) bool {
	for _, l := range c.reg.Entries {
		if l != nil && tmsg.Matches(*l, msg, c.rtr.Local.Id) {
			log.TraceMsg("Matched: %v, %v %v", *l, msg.Type(), msg)
			return true
		} else {
			log.TraceMsg("Not Matched: %v, %v %v", *l, msg.Type(), msg)
		}
	}
	return false
}

func (c *IOChannel) GetSendQueue() (chan<- *TsiMessage, error) {
	return c.fromRouter, nil
}

func (c *IOChannel) Recv() <-chan *TsiMessage {
	return c.toRouter
}

func (c *IOChannel) Close(err error) {
	log.Debug("Closing connection: %v", c.String())

	c.grp.Cancel(err)

	c.Lock()
	defer c.Unlock()
	if c.toRouter != nil {
		close(c.toRouter)
		c.toRouter = nil
	}
	if c.io != nil {
		c.w.Flush()
		err := c.io.Close()
		if err != nil {
			log.Warn("[%v] Error closing io: %v", err)
		}
		c.io = nil
	}
}

func (c *IOChannel) String() string {
	if c.reg.SourceService != nil {
		ss := c.reg.SourceService
		if ss.Aid == tmsg.APP_ID_TGWAD {
			return fmt.Sprintf("Remote Site (%v)", ss.Site)
		}

		appName, ok := tmsg.AppIdNames[ss.Aid]
		if ok {
			return fmt.Sprintf("%s-%v", appName, ss.Eid)
		} else {
			return fmt.Sprintf("0x%x-%v", ss.Aid, ss.Eid)
		}
	}
	return fmt.Sprintf("%s", c.io) // TODO: Something better here
}

func (c *IOChannel) SiteInfo() *routing.SiteInfo {
	return nil
}

func (c *IOChannel) EndPoint() *EndPoint {
	return c.reg.GetSourceService()
}
