package main

import (
	_ "expvar"
	"flag"
	"net"
	"net/http"
	"prisma/tms/log"
)

var (
	debugAddress string
)

func init() {
	flag.StringVar(&debugAddress, "debug", ":9090",
		"address:port to expose variables to, visit at /debug/vars")
}

func ServeDebug() {
	sock, err := net.Listen("tcp", debugAddress)
	if err != nil {
		log.Warn("unable to start debug service: %v", err)
		return
	}
	go func() {
		log.Debug("Exporting variables to %v/debug/vars", debugAddress)
		http.Serve(sock, nil)
	}()
}
