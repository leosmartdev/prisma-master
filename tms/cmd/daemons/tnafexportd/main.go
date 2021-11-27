package main

import (
	"flag"
	"prisma/gogroup"
	"prisma/tms/libmain"
	"prisma/tms/log"
	"prisma/tms/tmsg"
	"sync"
)

//Address for tcp or sftp conn
var Address = flag.String("address", ":9977", "-address <ip address> address to forward naf messages to")

//Username for sftp conn
var Username = flag.String("username", "", "sftp username")

//Password for sftp conn
var Password = flag.String("password", "", "sftp password")

//Conn type supported are sftp or tcp
var Conn = flag.String("conn", "tcp", "conn type <tcp || sftp>")

//Dir where to sftp the data
var Dir = flag.String("dir", "", "sftp directory")

func main() {
	if *Conn != "tcp" && *Conn != "sftp" {
		log.Fatal("support connections are tcp and sftp")
	}
	libmain.Main(tmsg.APP_ID_TNAFEXPORTD, run)
}

func run(ctxt gogroup.GoGroup) {

	waits := &sync.WaitGroup{}
	//listens on tgwad for *iridium streams
	_, err := Newlistner(ctxt, tmsg.GClient, waits, *Address)
	if err != nil {
		log.Crit("Failed to linten on tgwad streams: %v", err)
		ctxt.Cancel(err)
	}

	waits.Wait()

}
