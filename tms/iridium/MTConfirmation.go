package iridium

import "encoding/binary"

type MTConfirmation struct {
	IEI byte

	MTCL uint16

	UCM_IF int32

	IMEI string

	Auto_IDR uint32

	MT_Message_Status int16
}

func ParseMTConfirmation(data []byte) (MTConfirmation, error) {

	var MTC MTConfirmation

	MTC.IEI = data[0]

	MTC.MTCL = uint16(data[2]) | uint16(data[1])<<8

	MTC.UCM_IF = int32(binary.BigEndian.Uint32(data[3:7]))

	MTC.IMEI = ""
	for i := 7; i <= 21; i++ {

		MTC.IMEI = MTC.IMEI + string(data[i])
	}

	MTC.Auto_IDR = uint32(data[25]) | uint32(data[24])<<8 | uint32(data[23])<<16 | uint32(data[22])<<24

	MTC.MT_Message_Status = int16(binary.BigEndian.Uint16(data[26:28]))

	return MTC, nil
}
