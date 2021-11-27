// datagramdownlinkrequest sends RSM with DatagramDownlinkRequest using for testing
package main

import (
	"encoding/base64"
	"log"
	"math/rand"
	"net"
	"os"
	"prisma/gogroup"
	"prisma/tms"
	. "prisma/tms/cmd/daemons/tmsd/lib"
	"prisma/tms/ingenu"
	"prisma/tms/libmain"
	"prisma/tms/omnicom"
	"prisma/tms/tmsg"
	"time"
)

var (
	info    *log.Logger
	warning *log.Logger
)

func init() {

	info = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	warning = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
}

//This test method encodes an omnicom Request Specific Message and sends it over tgwad as tmsg with a DatagramDownlinkRequest body
func main() {

	//init()

	rand.Seed(time.Now().Unix())
	RequestSpecificMessage := omnicom.RSM{
		Header:     0x33,
		ID_Msg:     uint32(rand.Int31n(4094)),
		Date:       OmnicomDate(),
		Msg_to_Ask: 0x03,
	}
	info.Printf("This is the test for generating some omnicom request %+v\n", RequestSpecificMessage)

	raw, err := RequestSpecificMessage.Encode()
	if err != nil {
		warning.Println(err)
	} else {
		info.Printf("This is the raw data in bytes %+v\n", raw)
		base64str := base64.StdEncoding.EncodeToString(raw)
		info.Printf("This is the raw data in base64 string %+v\n", base64str)

		ddlr := &ingenu.DatagramDownlinkRequest{
			Tag:     "11112223-04d3-4a21-a8e4-148130b5484c",
			Nodeid:  "0x0000246f",
			Payload: base64str,
		}
		info.Printf("datagramdownlinkrequest proto message %+v\n", ddlr)

		body, errpack := tmsg.PackFrom(ddlr)
		if errpack != nil {
			warning.Println(errpack)
		} else {
			warning.Printf("DatagramDownLinkRequest in tmsg format %+v\n", body)
			tmsgRSM := tms.TsiMessage{
				Source: &tms.EndPoint{
					Site: tmsg.TMSG_LOCAL_SITE,
				},
				Destination: []*tms.EndPoint{
					&tms.EndPoint{
						Site: tmsg.TMSG_LOCAL_SITE,
					},
				},
				WriteTime: tms.Now(),
				SendTime:  tms.Now(),
				Body:      body,
			}
			Conf, err := ReadConfig()
			if err != nil {
				warning.Printf("%v\n", err)
			} else {
				client, err := net.Dial("unix", Conf.ControlSocket)
				if err != nil {
					warning.Printf("There is no running tmsd now. Could not connect to currently running TMSD: %v\n", err)
				} else {
					client.Close()
					var flag1 int
					var flag2 int
					for _, proc := range Conf.Processes {
						if proc.Prog == "tgwad" {
							flag1 = 1
						}
						if proc.Prog == "tfleetd" {
							flag2 = 1
						}
					}
					if flag1 == 1 && flag2 == 1 {
						libmain.Main(tmsg.APP_ID_UNKNOWN, func(ctxt gogroup.GoGroup) {
							tmsg.GClient.Send(ctxt, &tmsgRSM)
							info.Printf("This is the message we want to send to tgwad %+v\n", tmsgRSM)
							ctxt.Cancel(nil)
							info.Printf("%+v group been canceled %+v\n", ctxt, ctxt.Canceled())
						})
					}
				}
			}

		}
	}
}

//OmnicomDate reflects current date to omnicom.Date
func OmnicomDate() omnicom.Date {

	var date omnicom.Date
	date.Year = uint32(time.Now().Year() - 2000)
	date.Month = uint32(time.Now().Month())
	date.Day = uint32(time.Now().Day())
	date.Minute = uint32((uint32(time.Now().Hour()) * 60) + uint32(time.Now().Minute()))

	return date

}
