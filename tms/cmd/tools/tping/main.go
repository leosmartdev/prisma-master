// tping sends tmsg message to check live of a daemon.
package main

import (
	. "prisma/tms"
	"prisma/tms/libmain"
	"prisma/tms/log"
	. "prisma/tms/routing"
	. "prisma/tms/tmsg"

	"flag"
	"fmt"
	"prisma/gogroup"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"golang.org/x/net/context"
)

var (
	json = jsonpb.Marshaler{
		EnumsAsInts:  false,
		EmitDefaults: false,
		Indent:       "  ",
	}

	site    string
	app     string
	timeout string
)

func init() {
	flag.StringVar(&site, "site", "", "Site to ping, default: local")
	flag.StringVar(&app, "app", "tgwad", "Application to ping, default: tgwad")
	flag.StringVar(&timeout, "timeout", "5s", "Amount of time to wait for response")
}

func main() {
	flag.Set("stdlog", "true")
	flag.Set("log", "warning")
	libmain.Main(APP_ID_TPING, func(ctxt gogroup.GoGroup) {
		siteId := uint32(GClient.ResolveSite(site))
		appId := uint32(GClient.ResolveApp(app))

		timeoutDur, err := time.ParseDuration(timeout)
		if err != nil {
			log.Fatal("Error parsing timeout: %v", err)
		}

		reqTime := time.Now()
		pingReq := &Ping{
			PingSendTime: Now(),
		}
		timeoutCtxt, _ := context.WithTimeout(ctxt, timeoutDur)
		ep := EndPoint{
			Site: siteId,
			Aid:  appId,
		}
		log.Debug("Sending ping request to %v", ep)
		resp, err := GClient.Request(timeoutCtxt, ep, pingReq)
		if err != nil {
			log.Fatal("Error requesing ping: %v", err)
		}
		if resp != nil {
			respJson, err := json.MarshalToString(resp)
			if err != nil {
				fmt.Printf("Resp: %v", respJson)
			}
		}
		fmt.Printf("Ping time (measured): %v\n", time.Since(reqTime))

		ctxt.Cancel(nil)
	})
}
