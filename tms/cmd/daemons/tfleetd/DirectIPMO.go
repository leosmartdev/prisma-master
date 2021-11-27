package main

import (
	"fmt"
	"io/ioutil"
	"net"

	"prisma/gogroup"
	. "prisma/tms"
	"prisma/tms/devices"
	. "prisma/tms/iridium"
	"prisma/tms/log"
	"prisma/tms/tmsg"

	"github.com/golang/protobuf/ptypes/wrappers"
)

//DirectIPListen listens on a port over tcp
func DirectIPListen(ctxt gogroup.GoGroup, addr string, dev devices.DeviceType) {

	tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		log.Warn("%+v", err)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Warn("%+v", err)
	}

	for {

		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("%+v", err)
		}

		handlerequest(conn, ctxt, dev)

	}

}

func handlerequest(conn net.Conn, ctxt gogroup.GoGroup, dev devices.DeviceType) {
	defer func() {
		log.Debug("omnicomd: connection %+v closed", conn)
		conn.Close()
	}()

	result, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Error("error: %s", err.Error())
	}

	log.Debug("Received %v", result)

	go HandleMOData(result, ctxt, dev)

}

// HandleMOData listens on a port then decodes iridium headers and iridium payload
func HandleMOData(data []byte, ctxt gogroup.GoGroup, dev devices.DeviceType) {
	if len(data) == 0 {
		log.Warn("MO raw data size is 0")
		return
	}
	if data[0] != 1 {
		log.Warn("invalid Iridium protocol version %d", data[0])
		return
	}
	messageLength := Length(data, 1) + 3
	if int(messageLength) > len(data) {
		log.Warn("MO raw data message is not complete %v", data)
		return
	}
	byteindex := 3
	var err error
	var header *MOHeader
	var payload MPayload
	var location MOLocationInformation
	for byteindex < int(messageLength) {
		elementtype := data[byteindex]
		elementlength := Length(data, byteindex+1) + 3
		if len(data) < (byteindex + int(elementlength)) {
			log.Error("invalid element length or mobile originated raw message incomplete")
			return
		}
		elementdata := data[byteindex : byteindex+int(elementlength)]
		switch elementtype {
		case 1:
			log.Debug("MO raw header data: %+v \n", elementdata)
			header, err = ParseMOHeader(elementdata)
			if err != nil {
				log.Error(err.Error())
			}
			log.Debug("MO header message%+v\n", header)
		case 2:
			log.Debug("MO raw payload Data: %+v\n", elementdata)
			payload, err = ParseMOPayload(elementdata)
			if err != nil {
				log.Warn("%+v", err)
			}
			log.Debug("MO payload message %+v \n", payload)
		case 3:
			log.Debug("MO raw Location Information data:", elementdata)
			location = ParseMOlocationInformation(elementdata)
			log.Debug("MO Location Information message %+v\n", location)
		default:
			log.Warn("Unknown Iridium Information Element type %d ", elementtype)
		}
		byteindex = byteindex + int(elementlength)
	}
	if payload.IEI == 0x02 {
		MOMessage, err := PopulateMOProtobuf(*header, payload)
		if err != nil {
			log.Warn("%+v", err)
		}
		//TODO: I am not sure we should keep this code, it seems like it will never happen.
		if err == nil && MOMessage.Payload.Omnicom.GetSpm() != nil &&
			len(MOMessage.Payload.Omnicom.GetSpm().Header) > 0 && MOMessage.Payload.Omnicom.GetSpm().Header[0] == 64 {
			SR, err := SplitPacketMessageHandler(MOMessage.Payload.Omnicom.GetSpm())
			if err != nil {
				log.Warn("%+v", err)
			}
			if len(SR) != 0 && err == nil {
				MOPayloadSPM := make([]byte, len(SR)+3)
				MOPayloadSPM[0] = 0x02
				MOPayloadSPM[1] = byte(uint16(len(SR)) >> 8)
				MOPayloadSPM[2] = byte(uint16(len(SR)))
				for bit := 3; bit < len(MOPayloadSPM); bit++ {
					MOPayloadSPM[bit] = SR[bit-3]
				}
				payloadSR, err := ParseMOPayload(MOPayloadSPM)
				if err != nil {
					log.Warn("%+v", err)
				} else {
					MOMessage, err = PopulateMOProtobuf(*header, payloadSR)
					if err != nil {
						log.Warn("%+v", err)
					}
				}
			}

		} // end of if spm condition

		log.Debug("%+v", MOMessage)
		SendToTgwad(MOMessage, ctxt, dev)
	} // end of condition for 0x02 message

	if location.IEI == 0x03 {
		MOMessage, err := PopulateLocationProtobuf(*header, location)
		if err != nil {
			log.Warn("%+v", err)
		} else {
			log.Debug("%+v", MOMessage)
			SendToTgwad(MOMessage, ctxt, dev)
		}
	}
}

// Dial messages using tcp
func Dial(network string, address string, data []byte) error {

	tcpAddr, err := net.ResolveTCPAddr(network, address) // sending to the webpage
	if err != nil {
		return err
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return err
	}

	defer conn.Close()

	if err == nil {
		_, err = conn.Write(data) // don't care about return value
		if err != nil {
			return err
		}
	}

	return nil

} // end of method Dial

func SendToTgwad(Handler *Iridium, ctxt gogroup.GoGroup, dev devices.DeviceType) {

	trackMsgs, err := packfromIridium(Handler, dev)
	if err != nil {
		log.Error("error: %+v", err)
	} else {
		for _, trackMsg := range trackMsgs {
			log.Debug("sending Mobile originated message to tgwad")
			tmsg.GClient.Send(ctxt, trackMsg)
		}
	}
}

func packfromIridium(Handler *Iridium, dev devices.DeviceType) ([]*TsiMessage, error) {

	var Msgs []*TsiMessage

	if Handler.GetMoh() == nil {
		return nil, fmt.Errorf("Error when processing Iridium message: message empty")
	}

	//Moli stands for MO Location Information The location values included in this IE provide an estimate of the originating IMEI’s location.
	//The inclusion of this information in an MO message delivery is optional.
	//Whether or not it is included is established when the IMEI is provisioned and may be changed at any time via SPNet.
	//The CEP radius provides the radius around the center point within which the unit is located.
	//While the resolution of the reported position is given to 1/1000th of a minute.
	// it is only accurate to within 10Km 80% of the time.
	// We populate Moli into activity without any intention to exploit this data in UI.
	// This might be used in the future to double check beacon localisation or other applications.
	if Handler.Moli != nil {

		id := createIDforIridiumMessage(Handler.Moh.IMEI)

		//Default beacon registration is performed by assigning the network IMEI to registryID
		//RegistryID should be changed by tanalyzed (auto beacon registration)
		registryID := createRegistryIDforIridiumMessage(Handler.Moh.IMEI)
		Activity := &MessageActivity{
			RegistryId: registryID,
			ActivityId: id,
			Time:       Now(),
			MetaData:   &MessageActivity_Moli{Handler.Moli},
			Type:       dev,
			Imei: &wrappers.StringValue{
				Value:                Handler.Moh.IMEI,
				XXX_NoUnkeyedLiteral: struct{}{},
				XXX_unrecognized:     nil,
				XXX_sizecache:        0,
			},
		}

		activitybody, errpack := tmsg.PackFrom(Activity)
		if errpack != nil {
			return nil, fmt.Errorf("Could not pack Activity structure into Any")
		}
		activityMsg := &TsiMessage{
			Destination: []*EndPoint{
				{
					Site: tmsg.GClient.ResolveSite(""),
				},
			},
			WriteTime: Now(),
			SendTime:  Now(),
			Body:      activitybody,
		}
		return []*TsiMessage{activityMsg}, nil
	}

	if Handler.Payload.GetOmnicom() != nil {

		id := createIDforIridiumMessage(Handler.Moh.IMEI)

		//Default beacon registration is performed by assigning the network IMEI to registryID
		//RegistryID should be changed by tanalyzed (auto beacon registration)
		registryID := createRegistryIDforIridiumMessage(Handler.Moh.IMEI)
		tracks, activity, err := ProcessOmnicom(Handler.Payload.Omnicom, dev)
		if err != nil {
			return nil, fmt.Errorf("error: %+v", err)
		}

		if activity != nil {
			activity.ActivityId = id
			activity.RegistryId = registryID
			activity.Imei = &wrappers.StringValue{
				Value:                Handler.Moh.IMEI,
				XXX_NoUnkeyedLiteral: struct{}{},
				XXX_unrecognized:     nil,
				XXX_sizecache:        0,
			}
			activitybody, errpack := tmsg.PackFrom(activity)
			if errpack != nil {
				log.Error("Error when packing activity message body: %v", errpack)
			} else {
				infoMsg := &TsiMessage{
					Destination: []*EndPoint{
						{
							Site: tmsg.GClient.ResolveSite(""),
						},
					},
					WriteTime: Now(),
					SendTime:  Now(),
					Body:      activitybody,
				}
				Msgs = append(Msgs, infoMsg)
			}
		}
		for _, track := range tracks {
			track.Id = id
			track.RegistryId = registryID
			for _, target := range track.Targets {
				target.Imei = &wrappers.StringValue{
					Value:                Handler.Moh.IMEI,
					XXX_NoUnkeyedLiteral: struct{}{},
					XXX_unrecognized:     nil,
					XXX_sizecache:        0,
				}
			}
		}

		Msgs = append(Msgs, PackfromOmnicom(tracks)...)

		log.Debug("This is the message we are sending from tfleetd %+v", Msgs)

		return Msgs, nil
	}

	return nil, fmt.Errorf("Error when processing Iridium message payload to create a track: %+v", Handler)

}

//Length extract the size of an iridium stream
func Length(data []byte, i int) uint16 {

	return uint16(data[i+1]) | uint16(data[i])<<8

}
