package iridium

import "strconv"
import "fmt"

type MOHeader struct {
	MO_IEI byte //Mo header IEI

	MOHL uint16 //MO Header Lenght

	CDR uint32 //CDR reference (ID)

	IMEI string // IMEI

	SessStatus string //Session Status

	MOMSN uint16

	MTMSN uint16

	TimeOfSession uint32
}

// ParseMOHeader is used to extract MO header data from bytes
func ParseMOHeader(data []byte) (*MOHeader, error) {
	if len(data) < 31 {
		return nil, fmt.Errorf("Not enough data to process")
	}

	var header MOHeader

	header.MO_IEI = data[0]

	header.MOHL = uint16(data[2]) | uint16(data[1])<<8

	header.CDR = uint32(data[6]) | uint32(data[5])<<8 | uint32(data[4])<<16 | uint32(data[3])<<24

	header.IMEI = ""
	for i := 7; i <= 21; i++ {

		header.IMEI = header.IMEI + string(data[i])
	}

	header.SessStatus = strconv.FormatInt(int64(data[22]), 10)

	header.MOMSN = uint16(data[24]) | uint16(data[23])<<8

	header.MTMSN = uint16(data[26]) | uint16(data[25])<<8

	header.TimeOfSession = uint32(data[30]) | uint32(data[29])<<8 | uint32(data[28])<<16 | uint32(data[27])<<24

	return &header, nil

}

func (mo MOHeader) EncodeMO() ([]byte, error) {

	var raw []byte = make([]byte, 31)

	if len(mo.IMEI) != 15 {
		return nil, fmt.Errorf("Invalid IMEI: %s", mo.IMEI)
	}

	raw[0] = mo.MO_IEI

	//encode MTHL
	raw[1] = byte(mo.MOHL >> 8)
	raw[2] = byte(mo.MOHL)

	raw[3] = byte(mo.CDR >> 24)
	raw[4] = byte(mo.CDR >> 16)
	raw[5] = byte(mo.CDR >> 8)
	raw[6] = byte(mo.CDR)

	//encode IMEI
	for i := 7; i <= 21; i++ {
		raw[i] = mo.IMEI[i-7]
	}

	raw[22] = mo.SessStatus[0]

	raw[23] = byte(mo.MOMSN >> 8)
	raw[24] = byte(mo.MOMSN)

	raw[25] = byte(mo.MTMSN >> 8)
	raw[26] = byte(mo.MTMSN)

	raw[27] = byte(mo.TimeOfSession >> 24)
	raw[28] = byte(mo.TimeOfSession >> 16)
	raw[29] = byte(mo.TimeOfSession >> 8)
	raw[30] = byte(mo.TimeOfSession)

	return raw, nil
}
