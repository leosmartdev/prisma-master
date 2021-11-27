package main

import (
	"bufio"
	"encoding/xml"
	"flag"
	"io"
	"net"
	"prisma/gogroup"
	"prisma/tms"
	"prisma/tms/adsb"
	"prisma/tms/devices"
	"prisma/tms/libmain"
	"prisma/tms/log"
	"prisma/tms/tmsg"
	"prisma/tms/util/ident"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/wrappers"
)

const (
	prog     = "tadsbd"
	frameEnd = "</MODESMESSAGE>"
)

var (
	address string
	trace   = log.GetTracer("server")
)

func init() {
	flag.StringVar(&address, "address", "", "address and port of AirNav RadarBox")
}

func main() {
	flag.Parse()
	libmain.Main(tmsg.APP_ID_TADSBD, service)
}

func service(ctxt gogroup.GoGroup) {
	for {
		log.Info("connecting to %v", address)
		conn := connect()
		handle(ctxt, conn)
		time.Sleep(5 * time.Second)
	}
}

func connect() net.Conn {
	for {
		conn, err := net.Dial("tcp", address)
		if err == nil {
			return conn
		}
		log.Error("unable to connect: %v", err)
		time.Sleep(5 * time.Second)
	}
}

func handle(ctxt gogroup.GoGroup, conn net.Conn) {
	log.Info("connection established")
	defer func() {
		conn.Close()
		log.Info("connection closed")
	}()

	me := &tms.SensorID{
		Site: tmsg.GClient.Local().Site,
		Eid:  tmsg.GClient.Local().Eid,
	}

	in := bufio.NewReader(conn)
	dec := xml.NewDecoder(in)
	for {
		var doc strings.Builder
		enc := xml.NewEncoder(&doc)
		for {
			tok, err := dec.Token()
			if tok == nil || err == io.EOF {
				log.Info("end of file")
				break
			} else if err != nil {
				log.Error("error while processing XML document: %v", err)
				return
			}
			if err := enc.EncodeToken(tok); err != nil {
				log.Error("unable to encode XML token: %v", err)
				return
			}
			if ee, ok := tok.(xml.EndElement); ok {
				if ee.Name.Local == "MODESMESSAGE" {
					break
				}
			}
		}
		if err := enc.Flush(); err != nil {
			log.Error("unable to flush XML document: %v", err)
			return
		}

		var msg adsb.ModeMessageS
		err := xml.Unmarshal([]byte(doc.String()), &msg)
		if err != nil {
			log.Error("invalid XML document: %v", err)
			return
		}
		if msg.Callsign == nil && msg.Modes == nil {
			log.Error("invalid XML document: %v", doc.String())
			return
		}
		trace.Logf("message: %+v", msg)
		track, err := populateTrack(msg, me)
		if err != nil {
			log.Info("error: %v", err)
			break
		}
		sendToTgwad(ctxt, track)
	}
}

func populateTrack(msg adsb.ModeMessageS, me *tms.SensorID) (*tms.Track, error) {
	track := &tms.Track{}
	
	if msg.Callsign != nil {
		track.Id = ident.
		With("Callsign", *msg.Callsign).
		Hash()
	track.RegistryId = ident.
		With("Callsign", *msg.Callsign).
		Hash()
	} else {
		track.Id = ident.
		With("Modes", *msg.Modes).
		Hash()
	track.RegistryId = ident.
		With("Modes", *msg.Modes).
		Hash()
	}
	
	target := &tms.Target{}
	sn := ident.TimeSerialNumber()
	target.Id = &tms.TargetID{
		Producer:     me,
		SerialNumber: &tms.TargetID_TimeSerial{&sn},
	}
	target.Type = devices.DeviceType_ADSB
	target.IngestTime = tms.Now()
	target.Time = tms.Now()
	info := &adsb.Adsb{}

	if msg.Datetime != nil {
		info.Datetime = *msg.Datetime
	}
	if msg.Modes != nil {
		info.Modes = *msg.Modes
	}
	if msg.Callsign != nil {
		info.Callsign = *msg.Callsign
	}
	if msg.Altitude != nil {
		info.Altitude = *msg.Altitude
	}
	if msg.GroundSpeed != nil {
		target.Speed = &wrappers.DoubleValue{Value: float64(*msg.GroundSpeed)}
	}
	if msg.Track != nil {
		target.Course = &wrappers.DoubleValue{Value: float64(*msg.Track)}
	}
	if msg.VRate != nil {
		info.Vrate = *msg.VRate
	}
	if msg.AirSpeed != nil {
		info.AirSpeed = *msg.AirSpeed
	}
	if msg.Latitude != nil && msg.Longitude != nil {
		target.Position = &tms.Point{
			Latitude:  float64(*msg.Latitude),
			Longitude: float64(*msg.Longitude),
		}
	}
	target.Adsb = info
	track.Targets = []*tms.Target{target}
	return track, nil
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
				&tms.EndPoint{
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
