package iridium

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestParseMTHeader(t *testing.T) {
	m := MTHeader{
		IMEI:                  "123456789012345",
		IEI:                   0x01,
		MTHL:                  21, // IMEI15 + UNIQ4 + MTFLAG2
		UniqueClientMessageID: "asdf",
		MTflag:                0x05,
	}
	data, err := m.Encode()
	assert.NoError(t, err)
	mParsed, err := ParseMTHeader(data)
	assert.NoError(t, err)
	assert.Equal(t, &m, mParsed)
	mParsed, err = ParseMTHeader(data[:10])
	assert.Error(t, err)
	assert.Nil(t, mParsed)
}
