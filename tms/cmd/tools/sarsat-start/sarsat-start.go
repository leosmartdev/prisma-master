// sarsat-start sends alert messages to specific server.
package main

import (
	"flag"
	"log"
	"net"
	mcc "prisma/tms/cmd/daemons/tmccd/lib"
	"time"
)

var (
	Address = flag.String("address", "127.0.0.1:9999", "-address = IPaddress:Port")
)

func main() {
	flag.Parse()

	tcpAddr, err := net.ResolveTCPAddr("tcp", *Address)
	if err != nil {
		log.Fatal(err)
	}

	sarmsgs := []string{
		mcc.UnlocatedAlertMessageSample,
		mcc.LocatedAlertMessageSample,
		mcc.LocatedAlertMessageMeoElementSample,
		mcc.ConfirmedAlertMessageSample,
	}

	for {
		for _, msg := range sarmsgs {

			conn, err := net.DialTCP("tcp", nil, tcpAddr)

			if err != nil {
				log.Fatal(err)
			}
			if err == nil {
				_, err = conn.Write([]byte(msg))
				if err != nil {
					log.Fatal(err)
				}
			}
			conn.Close()
			time.Sleep(1 * time.Second)
		}
		time.Sleep(10 * time.Second)
	}
}
