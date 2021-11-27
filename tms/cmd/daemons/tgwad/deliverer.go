package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"

	"prisma/gogroup"

	. "prisma/tms"
	"prisma/tms/log"
	routing "prisma/tms/routing"
	"prisma/tms/tmsg"
)

type Deliverer interface {
	Deliver(*TsiMessage) bool
	SetStatusCallback(func(routing.Route_Status))
}

func NewDeliverer(r *routing.Route, local EndPoint, cfg RemoteConfig, ctxt gogroup.GoGroup) (Deliverer, error) {
	if r == nil {
		return nil, errors.New("Nil route specified!")
	}

	parts := strings.Split(r.Url, ":")
	if len(parts) > 0 {
		ty := strings.ToLower(parts[0])
		switch ty {
		case "tcp", "tcp4", "tcp6":
			nd, err := NewNetDeliverer(ty, strings.Join(parts[1:], ":"), r, local, ctxt)
			if err != nil {
				return nd, err
			}
			nd.start()
			return nd, err

		case "tcps":
			nd, err := NewNetDeliverer("tcp", strings.Join(parts[1:], ":"), r, local, ctxt)
			if err != nil {
				return nd, err
			}
			nd.setupTLS(cfg)
			nd.start()
			return nd, err

		default:
			log.Fatal("Unknown connection method '%v'", ty)
		}
	}

	return nil, errors.New("Could not determine delivery type for: " + r.Url)
}

type NetDeliverer struct {
	Type   string //Connection type (e.g. "tcp")
	Descr  string // Connection string "host:port"
	status func(routing.Route_Status)

	sync.Mutex
	local EndPoint
	ctxt  gogroup.GoGroup
	ctr   uint64
	conn  net.Conn
	buf   *bufio.ReadWriter

	lastDeliveryLatency time.Duration

	tryConnect func()

	tlsConfig tls.Config
}

func NewNetDeliverer(ty string, addrHost string, r *routing.Route, local EndPoint, ctxt gogroup.GoGroup) (*NetDeliverer, error) {
	d := &NetDeliverer{
		Type:  ty,
		Descr: addrHost,

		local: local,
		ctxt:  ctxt,
		ctr:   uint64(rand.Int31()),
	}
	d.tryConnect = d.tryConnectTCP
	d.SetStatusCallback(nil)
	return d, nil
}

func (d *NetDeliverer) start() {
	d.ctxt.Go(d.connect)
}

func (d *NetDeliverer) SetStatusCallback(f func(routing.Route_Status)) {
	if f == nil {
		d.status = func(_ routing.Route_Status) {}
	} else {
		d.status = f
	}
}

func (d *NetDeliverer) connect() {
	tckr := time.NewTicker(time.Duration(15) * time.Second)
	defer tckr.Stop()
	d.tryConnect()
	for {
		// Loop forever either pinging or reconnecting
		select {
		case <-d.ctxt.Done():
			return
		case <-tckr.C:
			if d.conn == nil {
				d.tryConnect()
			} else {
				d.ping()
			}
		}
	}
}

func (d *NetDeliverer) ping() {
	sent := d.Deliver(&TsiMessage{
		Status: TsiMessage_KEEPALIVE,
	})
	if sent {
		log.Debug("Ping time: %v", d.lastDeliveryLatency)
	}
}

func (d *NetDeliverer) sendHello() {
	reg := routing.Registry{
		SourceService: &d.local,
	}
	helloBody, err := tmsg.PackFrom(&reg)
	if err != nil {
		log.Error("Error packing registry: %v", err)
	}

	hello := TsiMessage{
		Source: &d.local,
		Body:   helloBody,
	}

	d.Deliver(&hello)
}

func (d *NetDeliverer) kill() {
	d.conn = nil
	d.status(routing.Route_DOWN)
}

func (d *NetDeliverer) Deliver(msg *TsiMessage) bool {
	if msg == nil {
		return true
	}

	d.Lock()
	defer d.Unlock()
	if d.conn == nil {
		return false
	}
	buf := d.buf
	id := d.ctr
	d.ctr += 1

	start := time.Now()
	defer func() {
		end := time.Now()
		d.lastDeliveryLatency = end.Sub(start)
	}()

	err := tmsg.WriteTsiMessageExtended(d.ctxt, buf.Writer, msg, tmsg.Opts{ID: id, Compress: true})
	if err != nil {
		log.Warn("Error writing to remote connection: %v", err)
		d.kill()
		return false
	}

	for {
		resp, ackid, err := tmsg.ReadTsiMessageExtended(d.ctxt, buf.Reader)
		if err != nil {
			log.Warn("Error getting ack message: %v", err)
			d.kill()
			return false
		}
		if resp.Body != nil {
			continue
		}

		if resp.Status != TsiMessage_ACK || ackid != id {
			log.Warn("Expected ACK message for ID %v, got something else: %v", id, *msg)
			d.kill()
			return false
		}

		return true
	}
}

func (d *NetDeliverer) tryConnectTCP() {
	c, err := net.DialTimeout(d.Type, d.Descr, time.Duration(5)*time.Second)
	if err != nil {
		log.Warn("Could not connect to remote site: %v", err)
		d.conn = nil
		d.status(routing.Route_DOWN)
	} else {
		tcp, ok := c.(*net.TCPConn)
		if ok {
			tcp.SetKeepAlive(true)
			tcp.SetKeepAlivePeriod(time.Duration(2) * time.Second)
		}
		d.Lock()
		d.status(routing.Route_UP)
		d.buf = bufio.NewReadWriter(
			bufio.NewReader(c),
			bufio.NewWriter(c))
		d.conn = c
		d.Unlock()

		d.sendHello()
	}
}

func (d *NetDeliverer) tryConnectTLS() {
	c, err := net.DialTimeout(d.Type, d.Descr, time.Duration(5)*time.Second)
	if err != nil {
		log.Warn("Could not connect to remote site: %v", err)
		d.conn = nil
		d.status(routing.Route_DOWN)
	} else {
		tcp, ok := c.(*net.TCPConn)
		if ok {
			tcp.SetKeepAlive(true)
			tcp.SetKeepAlivePeriod(time.Duration(2) * time.Second)
		}

		tls := tls.Client(tcp, &d.tlsConfig)
		log.Debug("TLS info: %v", tls.ConnectionState())
		err := tls.Handshake()
		if err != nil {
			log.Warn("Could not connect to remote site: %v", err)
			log.Warn("Extra info: %v", tls.ConnectionState())
		} else {

			d.Lock()
			d.status(routing.Route_UP)
			d.buf = bufio.NewReadWriter(
				bufio.NewReader(tls),
				bufio.NewWriter(tls))
			d.conn = tls
			d.Unlock()

			d.sendHello()
		}
	}
}

func (d *NetDeliverer) setupTLS(cfg RemoteConfig) {
	if cfg.CA == "" || cfg.PrivKey == "" || cfg.Cert == "" {
		log.Fatal("If TCPS is specified CA, cert, and privkeys must also be")
	}
	ca_pem, err := ioutil.ReadFile(cfg.CA)
	if err != nil {
		log.Fatal("Could not read CA file %v: %v", cfg.CA, err)
	}
	pool := x509.NewCertPool()
	ok := pool.AppendCertsFromPEM(ca_pem)
	if !ok {
		log.Fatal("Could not get CAs from %v", cfg.CA)
	}

	cert, err := tls.LoadX509KeyPair(cfg.Cert, cfg.PrivKey)
	if err != nil {
		log.Fatal("Could not load cert/privkey pair: %v", err)
	}

	d.tlsConfig = tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
		ServerName:   "tgwad",
	}
	d.tryConnect = d.tryConnectTLS
}
