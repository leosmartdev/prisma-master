package iridium

import (
	"encoding/binary"
	"fmt"
)

// SessionStatus is used to issue a status to a server
type SessionStatus int16

// Statuses for handlers of mt messages
const (
	StatusOK                SessionStatus = 0
	StatusInvalidIMEI       SessionStatus = -1
	StatusUnknownIMEI       SessionStatus = -2
	StatusViolationProtocol SessionStatus = -7
	StatusQueueFull         SessionStatus = -5
)

// MTResponse should be used for sending a response on an mt request
type MTResponse struct {
	ProtocolVersion       byte
	OverallLength         uint16
	IEI                   byte
	Length                uint16
	UniqueClientMessageID string // [4]byte
	IMEI                  string // [15]byte
	AutoIDReference       uint32
	MTMessageStatus       SessionStatus
}

func (m *MTResponse) Parse(b []byte) error {
	if len(b) < 31 {
		return fmt.Errorf("buffer is too short")
	}
	m.ProtocolVersion = b[0]
	m.OverallLength = binary.BigEndian.Uint16(b[1:])
	m.IEI = b[3]
	m.Length = binary.BigEndian.Uint16(b[4:])
	m.UniqueClientMessageID = string(b[6:10])
	m.IMEI = string(b[10:25])
	m.AutoIDReference = binary.BigEndian.Uint32(b[25:29])
	m.MTMessageStatus = SessionStatus(binary.BigEndian.Uint16(b[29:]))
	return nil
}

func (m *MTResponse) Encode() ([]byte, error) {
	if len(m.IMEI) != 15 {
		return nil, fmt.Errorf("Invalid IMEI: %s", m.IMEI)
	}
	if len(m.UniqueClientMessageID) != 4 {
		return nil, fmt.Errorf("Invalid UniqueClientMessageID: %s", m.UniqueClientMessageID)
	}
	var (
		praw []byte
		bi   = make([]byte, 4)
	)
	praw = append(praw, []byte(m.UniqueClientMessageID)...)
	praw = append(praw, []byte(m.IMEI)...)

	binary.BigEndian.PutUint32(bi, m.AutoIDReference)
	praw = append(praw, bi...)

	binary.BigEndian.PutUint16(bi, uint16(m.MTMessageStatus))
	praw = append(praw, bi[:2]...) // cause 2 bytes

	var res []byte
	binary.BigEndian.PutUint32(bi, uint32(len(praw)+3))
	res = append(res, 0x01)
	res = append(res, bi[2:]...)
	binary.BigEndian.PutUint32(bi, uint32(len(praw)))
	res = append(res, 0x44)
	res = append(res, bi[2:]...)
	res = append(res, praw...)

	return res, nil
}
