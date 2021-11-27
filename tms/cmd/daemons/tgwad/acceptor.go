package main

import (
	"prisma/tms/log"

	"net"
	"prisma/gogroup"
	"time"
)

type NetAcceptor struct {
	port string
	rtr  *Router
	ctxt gogroup.GoGroup
}

func NewNetAcceptor(ctxt gogroup.GoGroup, rtr *Router, port string) *NetAcceptor {
	a := &NetAcceptor{
		port: port,
		rtr:  rtr,
		ctxt: ctxt,
	}
	ctxt.Go(a.accept)
	return a
}

func (a *NetAcceptor) accept() {
	addr, err := net.ResolveTCPAddr("tcp", a.port)
	if err != nil {
		log.Fatal("Could not resolve '%v' as TCP address!", a.port)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatal("Unable to listen on TCP %v", a.port)
	}
	defer l.Close()
	log.Debug("listening on TCP %v", a.port)
	for {
		select {
		case <-a.ctxt.Done():
			// Group is canceled. Time to die
			return
		default:
			l.SetDeadline(time.Now().Add(time.Duration(100) * time.Millisecond))
			conn, err := l.AcceptTCP()
			if err != nil {
				if err, ok := err.(net.Error); ok && err.Timeout() {
					continue
				}
				l.Close()
				log.Fatal("Unable to accept connection on %v: %v", a.port, err)
			}
			a.handle(conn)
		}
	}
}

func (a *NetAcceptor) handle(conn *net.TCPConn) {
	log.Debug("Accepting connection: %v", conn)
	c := NewIOChannel(conn, a.rtr)
	a.rtr.AddChannel(c)
}
