package main

import (
	"prisma/tms/log"

	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"prisma/gogroup"
)

type TLSNetAcceptor struct {
	port      string
	rtr       *Router
	ctxt      gogroup.GoGroup
	tlsConfig tls.Config
}

func NewTLSNetAcceptor(ctxt gogroup.GoGroup, rtr *Router, port string, caFile string, certFile string, privkeyFile string) *TLSNetAcceptor {

	ca_pem, err := ioutil.ReadFile(caFile)
	if err != nil {
		log.Fatal("Could not read CA file %v: %v", caFile, err)
	}
	pool := x509.NewCertPool()
	ok := pool.AppendCertsFromPEM(ca_pem)
	if !ok {
		log.Fatal("Could not get CAs from %v", caFile)
	}

	cert, err := tls.LoadX509KeyPair(certFile, privkeyFile)
	if err != nil {
		log.Fatal("Could not load cert/privkey pair: %v", err)
	}

	a := &TLSNetAcceptor{
		port: port,
		rtr:  rtr,
		ctxt: ctxt,
		tlsConfig: tls.Config{
			ClientAuth:   tls.RequireAndVerifyClientCert,
			Certificates: []tls.Certificate{cert},
			ClientCAs:    pool,
			Rand:         rand.Reader,
		},
	}

	ctxt.Go(a.accept)
	return a
}

func (a *TLSNetAcceptor) accept() {
	l, err := tls.Listen("tcp", a.port, &a.tlsConfig)
	if err != nil {
		log.Fatal("Unable to listen on TCP %v", a.port)
	}
	defer l.Close()

	for {
		select {
		case <-a.ctxt.Done():
			// Group is canceled. Time to die
			return
		default:
			// TODO: Figure out a way for this listener to die when the context is canceled
			//l.SetDeadline(time.Now().Add(time.Duration(100)*time.Millisecond))
			conn, err := l.Accept()
			if err != nil {
				if err, ok := err.(net.Error); ok && err.Timeout() {
					continue
				}
				l.Close()
				log.Fatal("Unable to accept connection on %v: %v", a.port, err)
			}
			tlsConn, ok := conn.(*tls.Conn)
			if ok {
				a.handle(tlsConn)
			} else {
				log.Error("Got non-TLS inbound connection. Very strange. %v", conn)
				tlsConn.Close()
			}
		}
	}
}

func (a *TLSNetAcceptor) handle(conn *tls.Conn) {
	log.Debug("Accepting connection: %v", conn)
	c := NewIOChannel(conn, a.rtr)
	a.rtr.AddChannel(c)
}
