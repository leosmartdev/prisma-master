package main

import (
	"bytes"
	"context"
	"net/http"
	"prisma/tms"
	"prisma/tms/log"
	"prisma/tms/routing"
	"prisma/tms/tmsg"
	"time"

	"github.com/golang/protobuf/jsonpb"

	"github.com/gorilla/mux"
)

func (h Handler) getRequest(w http.ResponseWriter, r *http.Request) {

	requestID := mux.Vars(r)["id"]
	ep := tms.EndPoint{
		Site: uint32(tmsg.GClient.ResolveSite(site)),
		Aid:  uint32(tmsg.GClient.ResolveApp(app)),
	}

	route := routing.Listener{
		Destination: &ep,
		MessageType: "prisma.tms.Multicast",
	}

	ctxt, cancel := context.WithCancel(h.ctxt)
	h.requestStream = tmsg.GClient.Listen(ctxt, route)

loop:
	for {
		select {
		case <-time.After(time.Duration(timeout) * time.Second):
			break loop
		case <-ctxt.Done():
			return
		case tmsg := <-h.requestStream:

			report, ok := tmsg.Body.(*tms.Multicast)
			if !ok {
				log.Error("Got non-request message in request stream.")
			}
			if len(report.Transmissions) == 0 {
				log.Error("Got an Multicast without Transmissions %+v", report)
				continue
			}

			if report.Transmissions[0].GetMessageId() == requestID {
				marshaler := new(jsonpb.Marshaler)
				jreport, err := marshaler.MarshalToString(report)
				if err != nil {
					log.Error("%+v", err)
				}
				log.Debug("%+v", jreport)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(jreport))
				cancel()
				return
			}

		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusRequestTimeout)
	cancel()

}

func (h Handler) getRequests(w http.ResponseWriter, r *http.Request) {

	ep := tms.EndPoint{
		Site: uint32(tmsg.GClient.ResolveSite(site)),
		Aid:  uint32(tmsg.GClient.ResolveApp(app)),
	}

	route := routing.Listener{
		Destination: &ep,
		MessageType: "prisma.tms.MessageRequest",
	}

	ctxt, cancel := context.WithCancel(h.ctxt)
	h.requestStream = tmsg.GClient.Listen(ctxt, route)

	c := make(chan string, 100)
	var count int

loop:
	for {
		select {
		case <-time.After(time.Duration(timeout) * time.Second):
			break loop
		case <-ctxt.Done():
			return
		case tmsg := <-h.requestStream:

			report, ok := tmsg.Body.(*tms.Multicast)
			if !ok {
				log.Error("Got non-request message in request stream.")
			}

			if len(report.Transmissions) == 0 {
				log.Error("Got an Multicast without Transmissions %+v", report)
				continue
			}

			marshaler := new(jsonpb.Marshaler)
			jreport, err := marshaler.MarshalToString(report)
			if err != nil {
				log.Error("%+v", err)
			}
			log.Debug("%+v", jreport)
			c <- jreport
			count++

		}
	}

	if count == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusRequestTimeout)
		cancel()
		return
	}

	var report string

	for i := 0; i < count; i++ {
		report = report + <-c
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(report))
	cancel()

}
func (h Handler) postRequest(w http.ResponseWriter, r *http.Request) {

	ep := tms.EndPoint{
		Site: uint32(tmsg.GClient.ResolveSite(site)),
		Aid:  uint32(tmsg.GClient.ResolveApp(app)),
	}

	buf := new(bytes.Buffer)

	buf.ReadFrom(r.Body)

	reader := bytes.NewReader(buf.Bytes())

	unmarshaler := new(jsonpb.Unmarshaler)

	msgpb := new(tms.Multicast)

	err := unmarshaler.Unmarshal(reader, msgpb)
	if err != nil {
		log.Error("%v", err)
	}

	log.Debug("%+v", msgpb)

	timeoutCtxt, _ := context.WithTimeout(h.ctxt, 1*time.Second)

	_, err = tmsg.GClient.Request(timeoutCtxt, ep, msgpb)
	if err != nil {
		log.Warn("%+v", err)
	}
	w.WriteHeader(http.StatusOK)

}
