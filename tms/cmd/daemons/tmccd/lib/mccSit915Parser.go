package lib

import (
	"prisma/tms/sar"
	"prisma/tms/sit"
)

func Sit915Parser(data []byte, protocol string) (*sar.SarsatMessage, error) {
	sitMsg, err := sit.Parse(string(data))
	if err != nil {
		return nil, err
	}
	sarsatMessage := &sar.SarsatMessage{}
	sarsatMessage.MessageBody = sitMsg.Raw
	sarsatMessage.Protocol = protocol
	sarsatMessage.Received = true
	sarsatMessage.RemoteType = sar.SarsatMessage_MCC
	sarsatMessage.MessageType = sar.SarsatMessage_SIT_915
	sarsatMessage.RemoteName = sitMsg.ReportingFacility
	sarsatMessage.NarrativeText = sitMsg.NarrativeText

	return sarsatMessage, nil
}
