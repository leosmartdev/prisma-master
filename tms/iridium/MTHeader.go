package iridium

import (

	"fmt"
	"errors"
	"encoding/binary"
	"strconv"
)

type MTHeader struct {
	IEI                   byte
	MTHL                  uint16
	UniqueClientMessageID string
	IMEI                  string
	MTflag                uint16
}

var ErrInvalidIMEI = errors.New("IMEI is not numeric")

func ParseMTHeader(data []byte) (*MTHeader, error) {
	if len(data) < 24 {
		return nil, errors.New("Not enough data to process")
	}
	header := new(MTHeader)
	header.IEI = data[0]
	header.MTHL = binary.BigEndian.Uint16(data[1:])
	header.UniqueClientMessageID = string(data[3:7])
	header.IMEI = string(data[7:22])
	if _, err := strconv.ParseFloat(header.IMEI, 64); err != nil {
		return nil, ErrInvalidIMEI
	}
	header.MTflag = binary.BigEndian.Uint16(data[22:])
	return header, nil

}


func (mt MTHeader) Encode() ([]byte, error) {

	var raw []byte = make([]byte, 24)

	if len(mt.IMEI) != 15 {
		return nil, fmt.Errorf("Invalid IMEI: %s", mt.IMEI)
	}

	raw[0] = mt.IEI
	//encode UniqueClientMessageID
	for i := 3; i <= 6; i++ {
		raw[i] = mt.UniqueClientMessageID[i-3]
	}
	//encode IMEI
	for i := 7; i <= 21; i++ {
		raw[i] = mt.IMEI[i-7]
	}
	//encode MTflag
	raw[22] = byte(mt.MTflag >> 8)
	raw[23] = byte(mt.MTflag)
	binary.BigEndian.PutUint16(raw[1:], uint16(len(raw[3:])))
	return raw, nil
}
