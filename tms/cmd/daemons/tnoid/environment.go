package main

import (
	"prisma/tms"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes/any"
)

type Env struct {
	HostName string
	Radar    *tms.GeoPoint
	sync.Mutex
	sentRequests uint32
	gotErrors    uint32
	lastMsg      *any.Any
	*tms.HostInfo
	started time.Time
	*tms.TnoidConfiguration
	HostIP string
}

func (e *Env) Uptime() uint32 {
	return uint32(time.Since(e.started))
}

func (e *Env) Error() {
	e.Lock()
	e.gotErrors++
	e.Unlock()
}

func (e *Env) Errors() (n uint32) {
	e.Lock()
	n = e.gotErrors
	e.Unlock()
	return
}

func (e *Env) GetLastMessage() (msg *any.Any) {
	e.Lock()
	msg = e.lastMsg
	e.Unlock()
	return
}

func (e *Env) SetLastMessage(msg *any.Any) {
	e.Lock()
	e.sentRequests++
	e.lastMsg = msg
	e.Unlock()
}
func (e *Env) Requests() (n uint32) {
	e.Lock()
	n = e.sentRequests
	e.Unlock()
	return
}
