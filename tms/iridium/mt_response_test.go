package iridium

import (
	"encoding/binary"
	"testing"

	"github.com/json-iterator/go/assert"
)

func TestMTResponse_Encode(t *testing.T) {
	m := MTResponse{
		MTMessageStatus:       StatusOK,
		AutoIDReference:       0x02,
		UniqueClientMessageID: "qwer",
		IMEI: "234567890asdfg",
	}
	_, err := m.Encode()
	assert.Error(t, err)
	m.IMEI = "1" + m.IMEI
	b, err := m.Encode()
	assert.NoError(t, err)
	assert.Equal(t, 0x02, int(binary.BigEndian.Uint32(b[len(b)-6:])))
	assert.Len(t, b, 31)
}

func TestMTResponse_Parse(t *testing.T) {
	m := MTResponse{
		ProtocolVersion:       0x01,
		IEI:                   0x44,
		OverallLength:         28,
		Length:                25,
		MTMessageStatus:       StatusOK,
		AutoIDReference:       0x02,
		UniqueClientMessageID: "qwer",
		IMEI: "1234567890asdfg",
	}

	b, err := m.Encode()
	assert.NoError(t, err)
	a := MTResponse{}
	assert.NoError(t, a.Parse(b))
	assert.Equal(t, m, a)
}

func TestMTConfirmationParser(t *testing.T) {

	tt := []struct {
		name string
		msg  []byte
	}{
		{"MT confirmation 1", []byte{68, 0, 25, 116, 101, 115, 116, 51, 48, 48, 50, 51, 52, 48, 49, 48, 48, 51, 48, 52, 53, 49, 0, 0, 0, 0, 255, 254}},
		{"MT confirmation 2", []byte{68, 0, 25, 67, 50, 48, 49, 49, 49, 49, 50, 51, 52, 48, 49, 48, 48, 51, 48, 52, 53, 55, 0, 0, 0, 0, 255, 254}},
	}
	for _, tc := range tt {
		result, err := ParseMTConfirmation(tc.msg)
		if err != nil {
			t.Errorf("%s failed because %v", tc.name, err)
		}
		t.Logf("%s %+v", tc.name, result)
	}

}
