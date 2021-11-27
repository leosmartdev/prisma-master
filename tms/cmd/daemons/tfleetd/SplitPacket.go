package main

import (
	"fmt"
	"math/rand"
	"prisma/tms/omnicom"
	"time"
)

// mtstep is the maximum that split package message payload size after we substracting the over head
// Max MT size 270 byte
// - 1 byte Header
// - 8 bytes packet ID
// - 1 byte index of packet
// - 1 byte total number of packets
// - 9 bits lenght of payload
// - 7 bits padding
// - 1 byte CRC
// = 256 byte
const mtstep int = 256

// SplitMtData is a function that takes an omnicom message that > 270 byte
/* 256 bytes is the max size of the message data part
   Split Packet message header field size is 1 byte
   Split Packet message ID field size is 8 bytes
   Split Packet message index of message data part field size is 1 byte
   Split Packet message total number of message data parts is 1 byte
   Split Packet message data part size is 9 bits, and 7bits (2bytes) are for padding
   As every omnicom message the last 1 byte is for CRC field
   Max size 270 = 256 + 1 + 8 + 1 + 1 + 2 + 1 */
func SplitMtData(mt []byte) ([][]byte, uint64, error) {
	nop := int(len(mt) / mtstep)
	pts := make([][]byte, 0)
	var key int
	for i := 0; i < nop; i++ {
		pts = append(pts, mt[key:key+mtstep])
		key = +mtstep
	}
	if rdr := len(mt) % mtstep; rdr != 0 {
		pts = append(pts, mt[key:key+rdr])
	}
	if len(pts) == 0 {
		return nil, 0, fmt.Errorf("Could not construct raw MT payloads")
	}
	ran := rand.New(rand.NewSource(time.Now().Unix()))
	mts := make([][]byte, len(pts))
	id := uint64(ran.Int63())
	spm := omnicom.SPM{
		Header:              0x40,
		Split_Msg_ID:        id,
		Packets_Total_Count: uint32(len(pts)),
	}
	for k, pt := range pts {
		spm.Packet_Number = uint32(k + 1)
		spm.Length_Msg_Data_Part_in_Byte = uint32(len(pt))
		spm.Msg_Data_Part = pt
		bts, err := spm.Encode()
		if err != nil {
			return nil, 0, err
		}
		mts[k] = bts
	}
	return mts, id, nil
}

type SplitMessage struct {
	MessageID        uint64
	PacketNumbers    []uint32
	PacketTotalCount uint32
	MessageDataParts [][]byte
}

var (
	SplitMessages    []SplitMessage = make([]SplitMessage, 0)
	SplitMessageStep                = SplitMessage{}
)

func contains(SplitMessages []SplitMessage, e uint64) (bool, int) {
	for i := 0; i < len(SplitMessages); i++ {
		if SplitMessages[i].MessageID == e {
			return true, i
		}
	}
	return false, -1
} // end of method contains

func SplitPacketMessageHandler(SPM *omnicom.Spm) ([]byte, error) {
	//check is we have this split msg id already
	contain, index := contains(SplitMessages, SPM.Split_Msg_ID)
	if contain == true {
		if SPM.Packet_Number <= SplitMessages[index].PacketTotalCount {
			SplitMessages[index].PacketNumbers[SPM.Packet_Number-1] = SPM.Packet_Number
			SplitMessages[index].MessageDataParts[SPM.Packet_Number-1] = SPM.Msg_Data_Part
		}

		// if we received the last packet number of a series of SPM sentences then the condition will check
		if SPM.Packet_Number == SplitMessages[index].PacketTotalCount {
			var barray []byte
			var barrayindex int = 0
			for i := 0; i < len(SplitMessages[index].MessageDataParts); i++ {
				for j := 0; j < len(SplitMessages[index].MessageDataParts[i]); j++ {
					barray[barrayindex] = SplitMessages[index].MessageDataParts[i][j]
					barrayindex = barrayindex + 1
				} //end of iner loop
			} //enf of outer loop

			// delete the fully constructed message
			for i := index; i < len(SplitMessages)-1; i++ {
				SplitMessages[i] = SplitMessages[i+1]
			}
			SplitMessages = SplitMessages[:len(SplitMessages)-1]

			return barray, nil
		}

	} else { //new SPM ID
		if len([]byte(SPM.Msg_Data_Part)) != int(SPM.Length_Msg_Data_PartIn_Byte) {
			return []byte{}, fmt.Errorf("Length of message data part and value of field Length mismatch "+
				"for message ID %d, packet number %d", SPM.Split_Msg_ID, SPM.Packet_Number)
		}
		if SPM.Packet_Number != 1 {
			return []byte{}, fmt.Errorf("Packet number for message ID %d should be 1", SPM.Split_Msg_ID)
		}
		SplitMessages = append(SplitMessages, SplitMessageStep)

		SplitMessages[len(SplitMessages)-1].MessageID = SPM.Split_Msg_ID

		SplitMessages[len(SplitMessages)-1].PacketNumbers = make([]uint32, SPM.Packets_Total_Count)
		SplitMessages[len(SplitMessages)-1].PacketNumbers[0] = SPM.Packet_Number

		SplitMessages[len(SplitMessages)-1].PacketTotalCount = SPM.Packets_Total_Count

		SplitMessages[len(SplitMessages)-1].MessageDataParts = make([][]byte, SPM.Packets_Total_Count)
		SplitMessages[len(SplitMessages)-1].MessageDataParts[0] = SPM.Msg_Data_Part

		if SPM.Packet_Number == SplitMessages[len(SplitMessages)-1].PacketTotalCount {
			return SplitMessages[len(SplitMessages)-1].MessageDataParts[0], nil
		}

		return []byte{}, nil

	}

	return []byte{}, fmt.Errorf("message with ID= %d is not a correct split message", SPM.Split_Msg_ID)

}
