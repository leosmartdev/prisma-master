package main

import (
	"crypto/tls"
	"flag"
	"prisma/gogroup"
	. "prisma/tms"
	"prisma/tms/libmain"
	"prisma/tms/log"
	"prisma/tms/nmea"
	"prisma/tms/nmea/nmealib"
	. "prisma/tms/tmsg"
)

var (
	//Global options to request orbcomm messages
	Address  = flag.String("address", "63.146.183.130:9022", "-address = IPaddress:Port The address and port to request for orbcomm messages")
	Username = flag.String("username", "McMurdo_Lss", "The username to request for orbcomm messages")
	Password = flag.String("password", "JnHpv5BT5heY", "The password to request for orbcomm messages")
)

func main() {

	libmain.Main(APP_ID_TORBCOMMD, func(ctxt gogroup.GoGroup) {

		conf := &tls.Config{
			InsecureSkipVerify: true,
		}
		conn, err := tls.Dial("tcp", *Address, conf)
		if err != nil {
			log.Error("Error dialing the orbcomm server: %v\n", err)
			return
		}
		defer conn.Close()

		n, err := conn.Write([]byte("$PMWLSS,1456512920,4," + *Username + "," + *Password + ",1*70\r\n"))
		if err != nil {
			log.Error("Error writing login message to the orbcomm server: %v\n", err)
			return
		}

		me := SensorID{
			Site: GClient.Local().Site,
			Eid:  GClient.Local().Eid,
		}

		nmeaIdentify := nmealib.NewNmeaIdentify(&me)  //Initialize a NmeaIdentify using the sensorID
		nmeaCopy := nmealib.NewNmeaCopy(nmeaIdentify) //Initialize a nmeaCopy using the nmeaIdentify

		pair := make([]string, 2)
		current := 0

		for {
			buf := make([]byte, 1024)
			n, err = conn.Read(buf)
			if err != nil {
				log.Error("Error reading data drom orbcomm server: %v\n", err)
				return
			}
			data := string(buf[:n])

			var s nmea.Sentence
			var err_parse error

			aivdm := &Aivdm{}

			DecodeInformation(data, aivdm)
			aisData := aivdm.EncapsulatedData

			if (aivdm.SentenceNumber == 1) && (aivdm.TotalSentences == 2) {
				pair[0] = aisData
				current = 1
				continue
			} else if (aivdm.SentenceNumber == 2) && (aivdm.TotalSentences == 2) {
				if current == 0 {
					continue
				}
				pair[1] = aisData
				current = 0
				s, err_parse, _ = nmea.ParseArray(pair)
			} else {
				s, err_parse = nmea.Parse(aisData)
			}

			track := &Track{
				Targets:  []*Target{},
				Metadata: []*TrackMetadata{},
			}

			if err_parse == nil {

				sen, err_proto := nmea.PopulateProtobuf(s) //Get nmea proto message using libgonmea

				if err_proto == nil {

					nmeaCopy.PopulateTrack(track, sen) //Using nmeaCopy to populate the track proto message

				} else {

					log.Error("Error when populating nmea proto message: %v \n", err_proto)

				}

			} else {

				log.Error("Error when parsing nmea message %s: %v \n", aisData, err_parse)

			}

			if (len(track.Targets) > 0) || (len(track.Metadata) > 0) {

				body, err_pack := PackFrom(track)

				if err_pack != nil {
					log.Error("Error when packing track body: %v \n", err_pack)
				}

				infoMsg := TsiMessage{
					Source: GClient.Local(),
					Destination: []*EndPoint{
						&EndPoint{
							Site: TMSG_HQ_SITE,
						},
					},
					WriteTime: Now(),
					SendTime:  Now(),
					Body:      body,
				}
				GClient.Send(ctxt, &infoMsg) //Send a message containing the track to tgwad
			}

		}
	})
}
