package iridium

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMOPayload(t *testing.T) {
	var b []byte
	b = []byte{2, 0, 18, 6, 36, 241, 130, 39, 219, 28, 206, 108, 212, 0, 124, 0, 192, 0, 8, 0, 115}
	p, err := ParseMOPayload(b)
	assert.NoError(t, err)
	t.Log(p)
}
