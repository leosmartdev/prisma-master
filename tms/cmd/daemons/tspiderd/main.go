// tspiderd is a deamon to get information of spider tracks.
package main

import (
	"flag"
	"time"

	"prisma/gogroup"
	"prisma/tms/libmain"
	"prisma/tms/log"
	"prisma/tms/spidertracks"
	"prisma/tms/tmsg"
)

const sleepBetweenRequests = 30 * time.Second

var sysId = flag.String("system", "Orolia", "System ID")
var user = flag.String("username", "yonas.williams@orolia.com", "username")
var password = flag.String("password", "orolia", "password")
var url = flag.String("url", "https://go.spidertracks.com/api/aff/feed", "aff feed url")

func main() {
	libmain.Main(tmsg.APP_ID_TSPIDERD, realMain)
}

func realMain(ctxt gogroup.GoGroup) {
	ticker := time.Tick(sleepBetweenRequests)
	log.Info("Wait 30seconds before first request")
	<-ticker
	log.Info("Start to work")
	initTime := time.Now().Add(-spidertracks.RequestForTimeAgo * time.Minute).Format(spidertracks.TimeLayout)
	for range ticker {
		select {
		case <-ctxt.Done():
			return
		default:
		}
		simpleList, err := ReturnSimplifiedSpiderList(*sysId, *user, *password, *url, &initTime)
		if err != nil {
			log.Error("Failed to return spiderList: %s", err)
			continue
		}
		for i := range simpleList {
			nextTrack, err := PopulateTrack(simpleList[i])
			if err != nil {
				log.Error("Error when populating track due to %s", err)
				continue
			}
			if err := SendTrackToTGWAD(ctxt, nextTrack); err != nil {
				log.Error(err.Error())
			}
		}
	}
}
