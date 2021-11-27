// tnoid is a daemon that receives data from AIS, Radars.
package main

import (
	"os"
	"prisma/tms"
	"prisma/tms/libmain"
	"prisma/tms/log"
	"prisma/tms/nmea"
	"prisma/tms/nmea/nmealib"
	"prisma/tms/tmsg"
	"time"

	"github.com/golang/protobuf/ptypes/any"

	"flag"
	"net"
	"prisma/gogroup"
	"strconv"
	"strings"
)

// address is global option for the ip address and port to request nmea messages
var address = flag.String("address", "127.0.0.1:2001", "-address = IPaddress:Port The address and port to request for nmea messages")
var appName = flag.String("name", "tnoid0", "Name of this app.")
var reportInterval = flag.Int("ri", 60*5, "Interval to send tms info to front end")

func main() {
	flag.Parse()
	log.Debug("listening on %s\n", *address)
	libmain.Main(tmsg.APP_ID_TNOID, app)
}

func app(ctxt gogroup.GoGroup) {
	h, p, err := net.SplitHostPort(*address)
	if err != nil {
		log.Fatal("Error splitting address of %s: %v\n", *address, err)
	}

	hname, _ := os.Hostname()
	env := &Env{
		TnoidConfiguration: &tms.TnoidConfiguration{
			Name: *appName,
			Host: h,
			Port: p,
			RadarPosition: &tms.GeoPoint{
				Lat: float32(*nmealib.Conf_Radar_Latitude),
				Lng: float32(*nmealib.Conf_Radar_Longitude),
			},
		},
		started:  time.Now(),
		HostName: hname,
		HostIP:   GetLocalIP(),
	}

	go Reporter(time.Duration(*reportInterval)*time.Second, ctxt, env)

	lep := tmsg.GClient.Local()
	me := tms.SensorID{Site: lep.Site, Eid: lep.Eid}
	nmeaIdentify := nmealib.NewNmeaIdentify(&me)  // Initialize a NmeaIdentify using the sensorID
	nmeaCopy := nmealib.NewNmeaCopy(nmeaIdentify) // Initialize a nmeaCopy using the nmeaIdentify
	nmeaClient := NewNmeaClient(*address)         // Initialize a NmeaClient to receive nmea messages
	tcpAddr, err := net.ResolveTCPAddr("tcp4", *address)
	if err != nil {
		env.Error()
		log.Fatal("Error resolving tcp address of %s: %v\n", *address, err)
	}

	go nmeaClient.Read(tcpAddr, time.Second)
	ch := make(chan nmea.Sentence, 100)
	go read(ctxt, ch, nmeaCopy, env)
	go scan(nmeaClient, ch, env)
	ctxt.Wait()
}

func read(ctx gogroup.GoGroup, ch chan nmea.Sentence, nmeaCopy *nmealib.NmeaCopy, env *Env) {
	for s := range ch {
		// go send(s, nmeaCopy, ctxt)
		err := send(s, nmeaCopy, ctx, env)
		if err != nil {
			log.Error("Error when sending to gwad: %v \n", err)
			// ch <- s // retry
		}
	}
}

func scan(nmeaClient *NmeaClient, ch chan nmea.Sentence, env *Env) {
	var prev []string // A string to keep last received if last and current are together
	current := 0      // Initialize a pointer index for prev array
	cn := nmeaClient.GetResult()
	for bmsg := range cn {
		msg := string(bmsg) // Get a line from nmea client result handler
		var s nmea.Sentence
		var nmeaParseErr error
		fields := strings.Split(msg, ",")

		if len(fields) > 2 {
			indexParsed, errIndex := strconv.ParseUint(fields[2], 10, 32)
			numParsed, errNum := strconv.ParseUint(fields[1], 10, 32)
			if (errIndex == nil) && (errNum == nil) && (int(numParsed) > 1) {
				index := int(indexParsed)
				num := int(numParsed)
				if index != current+1 {
					current = 0
					continue
				} else if index == 1 {
					prev = make([]string, num)
					prev[0] = msg
					current++
					continue
				} else if index == num {
					prev[index-1] = msg
					current = 0
					s, nmeaParseErr, _ = nmea.ParseArray(prev)
				} else {
					prev[index-1] = msg
					current++
					continue
				}
			} else {
				s, nmeaParseErr = nmea.Parse(msg)
			}
		} else {
			s, nmeaParseErr = nmea.Parse(msg)
		}

		if nmeaParseErr != nil {
			env.Error()
			log.Error("Error when parsing nmea message %s: %v \n", msg, nmeaParseErr)
		} else {
			ch <- s
		}
	}
}

func send(s nmea.Sentence, nmeaCopy *nmealib.NmeaCopy, ctxt gogroup.GoGroup, env *Env) (err error) {
	mob := false
	track := &tms.Track{Targets: []*tms.Target{}, Metadata: []*tms.TrackMetadata{}}
	safetyBcast := &tms.SafetyBroadcast{}
	sen, errProto := nmea.PopulateProtobuf(s) // Get nmea proto message using libgonmea
	if errProto != nil {
		env.Error()
		log.Error("Error when populating nmea proto message: %v \n", errProto)
		return errProto
	}

	nmeaCopy.PopulateTrack(track, sen) // Using nmeaCopy to populate the track proto message
	errMob := nmeaCopy.PopulateSafetyBcast(safetyBcast, sen)
	if errMob == nil {
		safetyBcast.Nmea = sen
		safetyBcast.Type = nmeaCopy.DetermineType(sen)
		mob = true
	}

	if (len(track.Targets) > 0) || (len(track.Metadata) > 0) {
		body, errPack := tmsg.PackFrom(track)
		if errPack != nil {
			env.Error()
			log.Error("Error when packing track body: %v \n", errPack)
			return errPack
		}
		sendMessageToTGWAD(ctxt, body, env)
	} else if mob {
		body, errPack := tmsg.PackFrom(safetyBcast)
		if errPack != nil {
			env.Error()
			log.Error("Error when packing safetyBcast body: %v \n", errPack)
			return errPack
		}
		sendMessageToTGWAD(ctxt, body, env)
	}
	return
}

// Send a message containing the track to tgwad
func sendMessageToTGWAD(ctxt gogroup.GoGroup, body *any.Any, env *Env) {
	// Does body may be nil?
	if body == nil {
		return
	}

	tmsg.GClient.Send(ctxt, &tms.TsiMessage{
		// Source: tmsg.GClient.Local(), // It will be replaced.
		Destination: []*tms.EndPoint{
			&tms.EndPoint{
				Site: tmsg.GClient.ResolveSite(""),
			},
		},
		WriteTime: tms.Now(),
		SendTime:  tms.Now(),
		Body:      body,
	})

	env.SetLastMessage(body)
}
