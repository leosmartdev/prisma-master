// rsm sends RSM MT message and get MO message back.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"prisma/tms/iridium"
	"prisma/tms/omnicom"
	d "prisma/tms/util/omnicom"
	"time"
)

var (
	Address = flag.String("gss", "12.47.179.12:10800", "<IPaddress:port> the iridium gateway (default: 12.47.179.12:10800)")
	Imei    = flag.String("imei", "300234010031990", "!5 digit imei")
	Msg     = flag.Int("msg_to_ask", 0x03, "0x00 || 0x01 || 0x02 || 0x03 || 0x04. read documentation for more info (default: 0x03)")
	root    = rand.New(rand.NewSource(time.Now().UnixNano()))
	mtflag  = flag.Int("mt_flag", 0x0002, "0x0001 || 0x0002. This flag is used to clear the MT queue (default: 0x0002)")
)

func main() {
	flag.Parse()

	rsm := &omnicom.RSM{
		Header: 0x33,
		//FIXME: Have to determine a way to avoid collision
		ID_Msg: uint32(root.Intn(4095)),
		Date:   d.CreateOmnicomDate(time.Now().UTC()),
		// 0x00: send an alert report: message 'Alert report". Response with “Alert Report(0x02)”
		// 0x01: send the last position recorded with the message "History position report(0x01)".
		// 0x02: make a new position acquisition and send it with the message "History position report(0x01)".
		// 0x03: send the global parameters setting.Response with Global parameters(0x03)
		// 0x04: send the parameters URL. Response with API url parameters(0x08)
		// 0x10: test the 3G. The Dome send a “Single position report(0x06) and a History report(0x01) in a 3g message.
		Msg_to_Ask: uint32(*Msg),
	}

	mth := &iridium.MTHeader{
		IMEI: *Imei,
		IEI:  0x41,
		MTHL: 21, // IMEI15 + UNIQ4 + MTFLAG2
		UniqueClientMessageID: "asdf",
		MTflag:                uint16(*mtflag),
	}

	mtp := &iridium.MPayload{
		IEI: 0x42,
		Omn: rsm,
	}

	fmt.Printf("%+v %+v sending ...", mth, mtp.Omn)

	Hraw, err := mth.Encode()
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}
	Praw, err := mtp.Encode()
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}

	MTMLength := uint16(len(Hraw) + len(Praw))
	MTMessage := make([]byte, MTMLength+3)
	MTMessage = append(append([]byte{1, byte(MTMLength >> 8), byte(MTMLength)}, Hraw...), Praw...)

	tcpAddr, err := net.ResolveTCPAddr("tcp4", *Address)
	if err != nil {
		fmt.Printf("%+v", err)
		os.Exit(1)
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Printf("%+v", err)
		return
	}
	defer conn.Close()
	conn.Write(MTMessage)
	result, err := ioutil.ReadAll(conn)
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}

	if len(result) < 4 {
		fmt.Printf("received incomplete server response %+v", result)
		return
	}
	if result[3] == 0x44 {
		//MT confirmation message IE (0x44). parse MT confirmations message
		fmt.Printf("received a Mobile terminated confirmation %+v", result[3])
		pbMTConfirmation, err := iridium.ParseMTConfirmation(result[3:])
		if err != nil {
			fmt.Printf("%+v", err)
			return
		}
		fmt.Printf("%+v", pbMTConfirmation)
	}

	return
}
