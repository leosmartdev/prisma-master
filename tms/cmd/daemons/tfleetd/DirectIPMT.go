package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"reflect"
	"strconv"
	"sync"

	"prisma/gogroup"
	"prisma/tms"
	"prisma/tms/devices"
	. "prisma/tms/iridium"
	"prisma/tms/log"
	"prisma/tms/omnicom"
	"prisma/tms/routing"
	"prisma/tms/tmsg"
	client "prisma/tms/tmsg/client"

	"github.com/golang/protobuf/ptypes"
)

const maxMTSize = 270

var trace *log.Tracer = log.GetTracer("MobileTerminated")

//IridiumReq listens on tgwad for *Iridium messages.
type IridiumReq struct {
	tclient    client.TsiClient
	ctxt       gogroup.GoGroup
	iridiumReq <-chan *client.TMsg
}

//DirectIPSend this function listen to tgwad and handles MT data
func DirectIPSend(ctxt gogroup.GoGroup, address string, dev devices.DeviceType) {
	waits := &sync.WaitGroup{}
	//listens on tgwad for *iridium streams. I have to find how am I going to deal with *Address
	_, err := Newlistner(ctxt, tmsg.GClient, waits, address, dev)
	if err != nil {
		log.Crit("Failed to listen on tgwad streams: %+v", err)
		ctxt.Cancel(nil)
	}
	waits.Wait()
}

//Newlistner listens on tgwad and converts mo iridium messages to naf
func Newlistner(ctxt gogroup.GoGroup, client client.TsiClient, waits *sync.WaitGroup, address string, dev devices.DeviceType) (*IridiumReq, error) {
	I := &IridiumReq{
		tclient: client,
		ctxt:    ctxt,
	}
	waits.Add(1)
	ctxt.Go(func() {
		I.handle(address, dev)
		waits.Done()
	})
	return I, nil
}

func (I *IridiumReq) handle(address string, dev devices.DeviceType) {
	ctxt := I.ctxt.Child("Iridium MT request listener")
	ctxt.ErrCallback(func(err error) {
		pe, ok := err.(gogroup.PanicError)
		if ok {
			log.Error("Panic in iridium listener thread: %v\n%v", pe.Msg, pe.Stack)
		} else {
			log.Error("Error in iridium listener thread: %v", err)
		}
	})
	I.iridiumReq = I.tclient.Listen(ctxt, routing.Listener{
		MessageType: "prisma.tms.Multicast",
	})
	I.HandleMTData(address, ctxt, dev)
}

// HandleMTData handles MT iridium message on comming from tgwad and sends back MT confirmation messages to tgwad.
func (I *IridiumReq) HandleMTData(gss string, ctxt gogroup.GoGroup, dev devices.DeviceType) {
HandleMT:
	for {
		select {
		case <-I.ctxt.Done():
			return
		default:
			tmsg := <-I.iridiumReq
			report, ok := tmsg.Body.(*tms.Multicast)
			if !ok {
				log.Warn("Got non-iridium message in iridium stream. Got %v instead", reflect.TypeOf(tmsg.Body))
				continue HandleMT
			}
			if report.GetTransmissions() != nil && report.GetPayload() != nil {
				log.Debug("Got tms.Multicast %v", report)
				//make report into iridium
				if report.GetPayload().TypeUrl != "type.googleapis.com/prisma.tms.omnicom.Omni" {
					continue HandleMT
				}
				IridiumMsgs, err := packIridiumFrom(report)
				if err != nil {
					log.Error("could not pack msg request into iridium commands %v", err)
					continue HandleMT
				}
				trace.Logf("Requests received are %v", IridiumMsgs)
				for _, IridiumMsg := range IridiumMsgs {
					MTH, MTP, err := PopulateProtobufToMobileTerminated(IridiumMsg)
					if err != nil {
						log.Error("Error: %+v\n", err)
						continue HandleMT
					}
					//encode Mobile Terminated header
					Hraw, err := MTH.Encode()
					if err != nil {
						log.Error("%+v", err)
						continue HandleMT
					}
					// encode Mobile terminated payload
					Praw, err := MTP.Encode()
					if err != nil {
						log.Error("%+v", err)
						continue HandleMT
					}
					if len(Praw[3:]) > maxMTSize {
						pts, mid, err := SplitMtData(Praw[3:])
						if err != nil {
							log.Error("%+v", err)
							continue HandleMT
						}
						sendTranToTgwad(ctxt, report.Transmissions[0], len(pts), strconv.FormatUint(mid, 10))
						for _, pt := range pts {
							raw := append([]byte{0x42, byte(uint16(len(pt)) >> 8), byte(uint16(len(pt)))}, pt...)
							MTMLength := uint16(len(Hraw) + len(raw))
							MTMessage := append(append([]byte{1, byte(MTMLength >> 8), byte(MTMLength)}, Hraw...), raw...)
							result, err := sendMobileTerminated(MTMessage, gss)
							if err != nil {
								log.Error("%+v", err)
								continue HandleMT
							}
							if result[3] == 0x44 {
								//MT confirmation message IE (0x44). parse MT confirmations message
								pbMTConfirmation, err := HandleMTCData(result)
								if err != nil {
									log.Error("%+v", err)
									continue HandleMT
								}
								sendMtcToTgwad(pbMTConfirmation, report, ctxt, dev)
							}
						}
					} else {
						MTMLength := uint16(len(Hraw) + len(Praw))
						MTMessage := make([]byte, MTMLength+3)
						MTMessage = append(append([]byte{1, byte(MTMLength >> 8), byte(MTMLength)}, Hraw...), Praw...)
						trace.Logf("MT message to be sent to the GSS: %+v", MTMessage)
						result, err := sendMobileTerminated(MTMessage, gss)
						if err != nil {
							log.Error("%+v", err)
							continue HandleMT
						}
						if len(result) < 31 {
							log.Error("Mobile Terminated Confirmation message is not complete: message size %d < 31", len(result))
							continue HandleMT
						}
						trace.Logf("received result back from the gateway %+v", result)
						if result[3] == 0x44 {
							//MT confirmation message IE (0x44). parse MT confirmations message
							trace.Logf("received a Mobile terminated confirmation %+v", result[3])
							pbMTConfirmation, err := HandleMTCData(result)
							if err != nil {
								log.Error("%+v", err)
								continue HandleMT
							}
							sendMtcToTgwad(pbMTConfirmation, report, ctxt, dev)
						}
					}
				}
			} // end of condition to check if the *Iridium struct is a Mobile terminated message
		} // end of select statement
	} // end of HandleMT loop
} // end of method ReceivedMTData

func sendMobileTerminated(data []byte, addr string) ([]byte, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		return nil, err

	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	conn.Write(data)
	result, err := ioutil.ReadAll(conn)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func packIridiumFrom(report *tms.Multicast) ([]*Iridium, error) {
	var iridiumMsgs []*Iridium
	if len(report.Destinations) == 0 {
		return nil, fmt.Errorf("Multicast report with no destination %+v", report)
	}
	var omni omnicom.Omni
	err := ptypes.UnmarshalAny(report.Payload, &omni)
	if err != nil {
		return nil, fmt.Errorf("Request command of non omnicom type %v", reflect.TypeOf(report.Payload))
	}
	for _, transmission := range report.Transmissions {
		if transmission.Destination.Type == "iridium" {
			iridiumMsgs = append(iridiumMsgs, &Iridium{
				Mth: &MobileTerminatedHeader{
					MO_IEI: []byte{0x41},
					MTHL:   21,
					UniqueClientMessageID: *ClientMessageID,
					IMEI:   transmission.Destination.Id,
					MTflag: 0x0000,
				},
				Payload: &Payload{
					IEI:     []byte{0x42},
					Omnicom: &omni,
				},
			})
		}
	}
	return iridiumMsgs, nil
}

//HandleMTCData takes bytes and returns an iridium structure pointer
func HandleMTCData(data []byte) (*Iridium, error) {
	log.Debug("Raw MT confirmation received from the GSS: %+v ", data)
	if len(data) == 0 {
		return nil, fmt.Errorf("error: runtime error MT confirmation raw data length is 0")
	}
	//check for protocol version: 1-byte
	if data[0] != 1 {
		return nil, fmt.Errorf("invalid Iridium protocol version %d", data[0])
	}
	if len(data) < 31 {
		return nil, fmt.Errorf("MT confirmation message data is not complete %v", data)
	}
	var mtconfirmation MTConfirmation
	if len(data) == 31 {
		elementtype := data[3]
		elementdata := data[3:31] // same as byteindex + elementlentgh
		if elementtype == 0x44 {
			var err error
			mtconfirmation, err = ParseMTConfirmation(elementdata)
			if err != nil {
				trace.Logf("Parse MT confirmation fails in prasing MT raw confirmation")
				return nil, err
			}
			log.Debug("MT Confirmation Message: %+v\n", mtconfirmation)
		} else {
			return nil, fmt.Errorf("Unknown Iridium MT Confirmation Message Element")
		}
		MTC, err := PopulateMobileTerminatedConfirmationProtobuf(mtconfirmation)
		if err != nil {
			trace.Logf("MT confirmation fails in populating to protobug")
			return nil, err
		}
		return MTC, nil
	}
	return nil, fmt.Errorf("Something went wrong with parsing MT confimation message")
}

func packFromMTC(iridiumMtc *Iridium, report *tms.Multicast, dev devices.DeviceType) (*tms.TsiMessage, error) {
	if iridiumMtc.GetMtc() == nil {
		return nil, fmt.Errorf("iridium Mobile terminated message empty")
	}
	if len(report.Transmissions) == 0 {
		return nil, fmt.Errorf("report without Transmissions")
	}
	if report.Transmissions[0].GetMessageId() == "" {
		return nil, fmt.Errorf("report without messageId")
	}
	id := createIDforIridiumMessage(iridiumMtc.GetMtc().IMEI)
	//Default beacon registration is performed by assigning the network IMEI to registryID
	//RegistryID should be changed by tanalyzed (auto beacon registration)
	registryID := createRegistryIDforIridiumMessage(iridiumMtc.GetMtc().IMEI)
	Activity := &tms.MessageActivity{
		RegistryId: registryID,
		ActivityId: id,
		RequestId:  report.Transmissions[0].GetMessageId(),
		Time:       tms.Now(),
		MetaData:   &tms.MessageActivity_Mtc{iridiumMtc.GetMtc()},
		Type:       dev,
	}
	activitybody, errpack := tmsg.PackFrom(Activity)
	if errpack != nil {
		return nil, fmt.Errorf("Could not pack Activity structure into Any")
	}
	activityMsg := &tms.TsiMessage{
		Source: tmsg.GClient.Local(),
		Destination: []*tms.EndPoint{
			&tms.EndPoint{
				Site: tmsg.TMSG_LOCAL_SITE,
				Aid:  tmsg.APP_ID_TGWAD,
			},
		},
		WriteTime: tms.Now(),
		SendTime:  tms.Now(),
		Body:      activitybody,
	}
	return activityMsg, nil
}

func sendMtcToTgwad(iridiumMtc *Iridium, report *tms.Multicast, ctxt gogroup.GoGroup, dev devices.DeviceType) error {
	ActivityMsg, errpack := packFromMTC(iridiumMtc, report, dev)
	if errpack != nil {
		return errpack
	}
	log.Debug("sending Mobile Terminated Confirmation from %v to tgwad", report.Transmissions)
	tmsg.GClient.Send(ctxt, ActivityMsg)
	return nil
}

func sendTranToTgwad(ctxt gogroup.GoGroup, tran *tms.Transmission, npks int, mid string) error {
	tran.MessageId = mid
	for i := 0; i < npks; i++ {
		pt := &tms.Packet{
			Name:      omnicom.OmnicomConfiguration_Action_name[int32(omnicom.OmnicomConfiguration_SplitPacketMessage)],
			MessageId: mid,
			State:     tms.Transmission_Pending,
			Status:    &tms.ResponseStatus{Code: 102},
		}
		tran.Packets = append(tran.Packets, pt)
	}
	body, errpack := tmsg.PackFrom(tran)
	if errpack != nil {
		return fmt.Errorf("Could not pack tran structure into Any")
	}
	msg := &tms.TsiMessage{
		Source: tmsg.GClient.Local(),
		Destination: []*tms.EndPoint{
			&tms.EndPoint{
				Site: tmsg.TMSG_LOCAL_SITE,
				Aid:  tmsg.APP_ID_TDATABASED,
			},
		},
		WriteTime: tms.Now(),
		SendTime:  tms.Now(),
		Body:      body,
	}
	log.Debug("sending Mobile Terminated Confirmation from %v to tgwad", tran)
	tmsg.GClient.Send(ctxt, msg)
	return nil
}
