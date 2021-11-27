// Package lib has tools to work with mcc formats.
package lib

import (
	"prisma/tms/sar"
)

//DefaultParser populates SarsatMessage, this function is used when xml, and sit185 parsers fail
func DefaultParser(data []byte, protocol string) *sar.SarsatMessage {
	sarsatMessage := &sar.SarsatMessage{}

	sarsatMessage.MessageBody = string(data)
	sarsatMessage.Protocol = protocol
	sarsatMessage.Received = true
	sarsatMessage.RemoteType = sar.SarsatMessage_MCC
	sarsatMessage.MessageType = sar.SarsatMessage_UNKNOWN
	return sarsatMessage
}
