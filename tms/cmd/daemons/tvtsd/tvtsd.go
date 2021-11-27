package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"prisma/gogroup"
	"prisma/tms"
	"prisma/tms/libmain"
	"prisma/tms/log"
	"prisma/tms/tmsg"
	"prisma/tms/vts"
	"time"

	"github.com/golang/protobuf/ptypes/any"
)

const (
	prog = "tvtsd"
)

var (
	url          string
	pollInterval int
	trace        = log.GetTracer("server")
)

func init() {
	flag.StringVar(&url, "url", "", "URL to the service endpoint")
	flag.IntVar(&pollInterval, "poll", 5*60, "poll interval in seconds")
}

func main() {
	flag.Parse()
	if url == "" {
		fmt.Printf("%v: error: --url is required\n", prog)
		os.Exit(1)
	}
	if pollInterval <= 0 {
		fmt.Printf("%v: error: invalid poll interval\n", prog)
		os.Exit(1)
	}
	libmain.Main(tmsg.APP_ID_TVTSD, service)
}

func service(ctxt gogroup.GoGroup) {
	for {
		err := poll(ctxt)
		if err != nil {
			log.Error("%v", err)
			time.Sleep(5 * time.Second)
		} else {
			time.Sleep(time.Duration(pollInterval) * time.Second)
		}
	}
}

func poll(ctxt gogroup.GoGroup) error {
	log.Debug("connecting to %v", url)
	data, err := fetch()
	if err != nil {
		return fmt.Errorf("unable to fetch: %v", err)
	}
	err = handle(ctxt, data)
	if err != nil {
		return fmt.Errorf("unable to process: %v", err)
	}
	return nil
}

func fetch() ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}
	return data, nil
}

func handle(ctxt gogroup.GoGroup, data []byte) error {
	me := &tms.SensorID{
		Site: tmsg.GClient.Local().Site,
		Eid:  tmsg.GClient.Local().Eid,
	}

	tracks := vts.TracksDataInfo{}
	if err := xml.Unmarshal(data, &tracks); err != nil {
		return err
	}
	log.Debug("message received, %v ais, %v radar", len(tracks.AISTracks.AISTrack), len(tracks.TrackerTracks.TrackerTrack))
	for _, vtsais := range tracks.AISTracks.AISTrack {
		pos, err := vts.PopulateTrackPositionAIS(vtsais, me)
		if err != nil {
			log.Warn("invalid AIS track: %v", err)
			continue
		}
		sendToTgwad(ctxt, pos)
		vessel, err := vts.PopulateTrackVesselAIS(vtsais, me)
		if err != nil {
			log.Warn("invalid AIS track: %v", err)
			continue
		}
		sendToTgwad(ctxt, vessel)
	}
	for _, vtsrad := range tracks.TrackerTracks.TrackerTrack {
		tmsrad, err := vts.PopulateTrackRadar(vtsrad, me)
		if err != nil {
			log.Warn("invalid radar track: %v", err)
			continue
		}
		sendToTgwad(ctxt, tmsrad)
	}
	return nil
}

func sendToTgwad(ctxt gogroup.GoGroup, track *tms.Track) {
	var body *any.Any
	body, err := tmsg.PackFrom(track)
	if err != nil {
		log.Error("%v: error in packing track: %v", prog, err)
	} else {
		infoMsg := tms.TsiMessage{
			Source: tmsg.GClient.Local(),
			Destination: []*tms.EndPoint{
				{
					Site: tmsg.GClient.ResolveSite(""),
				},
			},
			WriteTime: tms.Now(),
			SendTime:  tms.Now(),
			Body:      body,
		}
		trace.Logf("Sending to tgwad: %+v", infoMsg)
		tmsg.GClient.Send(ctxt, &infoMsg)
		trace.Log("Message sent to tgwad")
	}
}
