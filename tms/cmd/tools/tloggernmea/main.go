// tloggernmea listens nmea messages from tgwad and log them
package main

import (
	"prisma/tms/libmain"
	"prisma/tms/log"
	"prisma/tms/routing"
	"prisma/tms/tmsg"
	client "prisma/tms/tmsg/client"

	"fmt"
	"prisma/gogroup"

	"time"
	"prisma/tms"
	"flag"
	"os"
)

var fileOut = flag.String("fout", "stdout", "a file where will write data. stdout for stdout")

func printMsg(msg *client.TMsg, f *os.File) {
	track, ok := msg.Body.(*tms.Track)
	if !ok {
		//log.Warn("resolve an interface: %v", msg.Body)
		return
	}
	if len(track.Targets) == 0 || track.Targets[0].Nmea == nil {
		return
	}
	if _, err := fmt.Fprintf(f,"%s %d:%d %s\n", time.Now().Format(time.RFC3339),
			msg.Source.Site, msg.Source.Eid, track.Targets[0].Nmea.OriginalString); err != nil {

		log.Error(err.Error())
	}
}

func main() {
	libmain.Main(tmsg.APP_ID_TLOGGERNMEA, func(ctxt gogroup.GoGroup) {
		var (
			f   *os.File
			err error
		)
		if *fileOut != "stdout" {
			f, err = os.OpenFile(*fileOut, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0777)
			if err != nil {
				log.Fatal(err.Error())
			}
			defer f.Close()
		} else {
			f = os.Stdout
		}
		log.Debug("Registring listener for everything...")
		msgChan := tmsg.GClient.Listen(ctxt, routing.Listener{})
		log.Debug("Waiting for messages...")
		for {
			select {
			case <-ctxt.Done():
				return
			case msg := <-msgChan:
				printMsg(msg, f)
			}
		}
	})
}
