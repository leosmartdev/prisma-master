// tportald provides Rest api to check multicast
package main

import (
	"flag"
	"net"
	"net/http"
	"prisma/gogroup"
	"prisma/tms/libmain"
	"prisma/tms/log"
	"prisma/tms/tmsg"
	client "prisma/tms/tmsg/client"
	"time"

	"github.com/gorilla/mux"
)

var (
	site    string
	app     string
	addr    string
	timeout int
)

func init() {
	flag.StringVar(&site, "site", "", "Site to ping, default: local")
	flag.StringVar(&app, "app", "tgwad", "Application to ping, default: tgwad")
	flag.StringVar(&addr, "addr", ":9099", "http listen and serve address")
	flag.IntVar(&timeout, "timeout", 5, "request timeout")
}

type Handler struct {
	tclient        client.TsiClient
	ctxt           gogroup.GoGroup
	trackStream    <-chan *client.TMsg
	activityStream <-chan *client.TMsg
	requestStream  <-chan *client.TMsg
	conn           net.Conn
}

func main() {
	var handler Handler
	libmain.Main(tmsg.APP_ID_TPING, func(ctxt gogroup.GoGroup) {

		router := mux.NewRouter()

		handler.ctxt = ctxt

		router.HandleFunc("/request", handler.getRequests).Methods("GET")
		router.HandleFunc("/request/{id}", handler.getRequest).Methods("GET")
		router.HandleFunc("/request", handler.postRequest).Methods("POST")
		router.HandleFunc("/activity", handler.getActivities).Methods("GET")
		router.HandleFunc("/activity/{id}", handler.getActivity).Methods("GET")
		router.HandleFunc("/activity", handler.postActivity).Methods("POST")

		srv := &http.Server{
			Addr:         addr,
			Handler:      router,
			WriteTimeout: 1 * time.Minute,
			ReadTimeout:  30 * time.Second,
		}

		err := srv.ListenAndServe()
		if err != nil {
			log.Fatal("%+v", err)
		}

		handler.ctxt.Cancel(nil)

	})
}
