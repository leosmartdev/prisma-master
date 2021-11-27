// tfleetd handles beacons, MO and MT messages.
package main

import (
	"flag"
	"prisma/gogroup"
	"prisma/tms/devices"
	"prisma/tms/libmain"
	"prisma/tms/log"
	"prisma/tms/tmsg"
)

var (
	//Omnicom flag should be set to true when tfleetd is going to listen on omnicom messages
	Omnicom = flag.Bool("omnicom", true, "<true>||<false>")
	//VMS set to true when tfleetd is expecting Omnicom VMS data
	VMS = flag.Bool("vms", false, "<true>||<false>")
	//Solar set to true when tfleetd is expecting Omnicom Solar data
	Solar = flag.Bool("solar", false, "<true>||<false>")
	//IridiumClientAddress is the address where Mobile originated iridium messages will be received
	IridiumClientAddress = flag.String("iridium-client", "127.0.0.1:7777", "<IPaddress:port> where Mobile Originated messages are received")
	//GSS is the iridium gateway address
	GSS = flag.String("gss", "12.47.179.12:10800", "<IPaddress:port> the iridium gateway")
	//URLReceive where tfleetd fetches message from the INGENU platform
	URLReceive = flag.String("url", "https://glds.ingenu.com/data/v1/receive/", "<url> ingenu url to fetch data")
	// URLIngenu is a url of the ingenu platform
	URLIngenu = flag.String("url-ingenu", "https://glds.ingenu.com", "<url> ingenu url")
	//Net flags the network tfleetd needs to connect to eg (iridium, ingenu ...)
	Net = flag.String("net", "iridium", "<iridium> || <ingenu>")
	//Username for ingenu network
	Username = flag.String("username", "orolia@orolia.com", "<username> ingenu account username")
	//Password for ingenu network
	Password = flag.String("password", "0r011a_McMurd0%", "<password> ingenu account password")
	//Unique Client Message ID: tfleetd include 4-byte message ID unique within its own application.
	//This value is not used in any way by the Gateway server except to include it in the confirmation message sent back to the client.
	//This is intended to serve as a form of validation and reference for the client application. The data type is specified to be characters.
	ClientMessageID = flag.String("client-message-id", "C201", "unique client message id should be 4 chars")
)

func main() {
	flag.Parse()

	libmain.Main(tmsg.APP_ID_TFLEETD, func(ctxt gogroup.GoGroup) {

		if *Omnicom == true {

			var dev devices.DeviceType

			if (*VMS && *Solar) || !(*VMS || *Solar) {
				log.Fatal("tfleetd instance needs to parse exlusively omnicom VMS or Solar")
			} else {
				if *VMS == true {
					dev = devices.DeviceType_OmnicomVMS
				}
				if *Solar == true {
					dev = devices.DeviceType_OmnicomSolar
				}
			}

			switch *Net {
			case "iridium":

				if len(*IridiumClientAddress) == 0 {
					log.Fatal("Please provide an IP:port address where the MO server should listen")
				}

				log.Info("tfleetd started, listening on %s for Omnicom Data ...\n", *IridiumClientAddress)

				//connection where tfleetd will expect messages from iridium
				go DirectIPListen(ctxt, *IridiumClientAddress, dev)

				// routine to handle MT messages received from tgwad
				go DirectIPSend(ctxt, *GSS, dev)

			case "ingenu":
				if len(*URLReceive) == 0 {
					log.Fatal("Please provide a ingenu url")
				}

				log.Info("tfleetd started fetching omnicom data from on %s over the ingenu network ...\n", *URLReceive)
				go OmnicomReceiveIngenuRestClient(ctxt, *URLReceive, *URLIngenu, *Username, *Password, dev)

				//routine to handle ingenu messages received from tgwad
				go IngenuSend(ctxt, *URLReceive, *URLIngenu, *Username, *Password)
			default:
				// TODO: other data handlers should go here: blueforce, Vlink ...
				log.Fatal("Network %s not implemented", *Net)
			}

		}
	})
}
