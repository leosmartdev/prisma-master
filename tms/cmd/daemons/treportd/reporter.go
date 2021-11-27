package main

import (
	. "prisma/tms"
	"prisma/tms/log"
	. "prisma/tms/routing"
	"prisma/tms/tmsg"
	. "prisma/tms/tmsg/client"

	"prisma/gogroup"

	"sync"
	"time"
)

type Reporter struct {
	ctxt    gogroup.GoGroup
	conf    ReportConf
	client  TsiClient
	storage ReportDB

	tgtStream <-chan *TMsg
	trkStream <-chan *TMsg
	mdStream  <-chan *TMsg

	messageReports <-chan *DeliveryReport
}

func NewReporter(ctxt gogroup.GoGroup, conf ReportConf, client TsiClient) *Reporter {
	db := NewMemReportDB(conf) // TODO: Make config option to select db
	r := &Reporter{
		ctxt:    ctxt,
		conf:    conf,
		client:  client,
		storage: db,

		messageReports: make(chan *DeliveryReport),
	}
	return r
}

func (r *Reporter) Start() {
	// Start send process
	r.ctxt.Go(r.sendProcess)

	// Listen for tracks from tgwad
	r.trkStream = r.client.Listen(r.ctxt, Listener{
		MessageType: "prisma.tms.Track",
	})

	for {
		select {
		case <-r.ctxt.Done():
			return
		case msg := <-r.trkStream:
			r.process(msg)
		}
	}
}

func (r *Reporter) process(msg *TMsg) {
	switch x := msg.Body.(type) {
	case *Track:
		r.processTrack(x)
	}
}

func (r *Reporter) processTrack(track *Track) {
	r.storage.Save(track)
}

func (r *Reporter) sendProcess() {
	var lock sync.Mutex // lock to protect messagesEnRoute
	messagesEnRoute := make(map[*TrackReport]struct{})
	wakeMe := make(chan struct{})

	tckr := time.NewTicker(r.conf.QueueTime)
	defer tckr.Stop()

	check := false // When we've just been woken up, should we assemble a report?
	for {

		// Wait for ticker to pop!
		select {
		case <-r.ctxt.Done():
			// We are being asked to die
			return

		case <-tckr.C:
			// Break select to check for tracks to report
			check = true

		case <-wakeMe:
			// Wake up on successful send
		}

		if !check {
			continue
		}

		lock.Lock()
		enRoute := uint(len(messagesEnRoute))
		lock.Unlock()
		for enRoute < r.conf.ConcurrentReports {
			rpt := r.storage.Next()
			if rpt == nil {
				// No tracks to send!
				check = false
				break
			}

			log.Debug("Sending report with %v entries to %v", len(rpt.Tracks), r.conf.Destination)

			body, err := tmsg.PackFrom(rpt)
			if err != nil {
				panic(err)
			}

			dstSiteNum := r.client.ResolveSite(r.conf.Destination)
			if dstSiteNum == tmsg.TMSG_UNKNOWN_SITE {
				log.Fatal("Could not determine site number for '%v'!", r.conf.Destination)
			}

			msg := &TsiMessage{
				Destination: []*EndPoint{
					&EndPoint{
						Site: dstSiteNum,
					},
				},
				RealTime: true,
				Body:     body,
			}

			lock.Lock()
			messagesEnRoute[rpt] = struct{}{}
			lock.Unlock()

			r.client.SendNotify(r.ctxt, msg, func(deliv *DeliveryReport) {
				log.Debug("Got delivery report: %v", deliv)

				lock.Lock()
				delete(messagesEnRoute, rpt)
				lock.Unlock()

				switch deliv.Status {
				case DeliveryReport_SENT:
					r.storage.Delivered(rpt)
					wakeMe <- struct{}{}
				case DeliveryReport_FAILED:
					r.storage.Fail(rpt)
				}
			})
		}
	}
}
