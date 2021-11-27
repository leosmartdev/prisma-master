package object

import (
	"fmt"
	"prisma/tms/cmd/daemons/tmccd/lib"
	"prisma/tms/log"
	"testing"

	"github.com/json-iterator/go/assert"
)

var (
	_ = fmt.Println
	_ = log.Spew
)

var seaObject = Object{}

func TestSarsat_GetPositionAlertingMessage(t *testing.T) {
	s := NewSarsat(seaObject)
	s.object.Unlocated = true
	s.object.Located = false
	s.object.Imei = "ADCC404C8400315"
	buff, err := s.GetPositionAlertingMessage()
	assert.NoError(t, err)
	assert.Equal(t, brokenMessageTpl[:10], string(buff[:10]))
	buff, err = s.GetPositionAlertingMessage()
	assert.NoError(t, err)
	_, err = lib.MccxmlParser(buff, "tcp")
	assert.NoError(t, err)
	assert.True(t, lib.XMLExp.Match(buff))

	s.object.Unlocated = false
	s.object.Located = true
	buff, err = s.GetPositionAlertingMessage()
	assert.NoError(t, err)
	_, err = lib.MccxmlParser(buff, "tcp")
	assert.NoError(t, err)
	assert.True(t, lib.XMLExp.Match(buff))

	s.object.Unlocated = false
	s.object.Located = false
	buff, err = s.GetPositionAlertingMessage()
	assert.NoError(t, err)
	_, err = lib.MccxmlParser(buff, "tcp")
	assert.NoError(t, err)

	assert.True(t, lib.XMLExp.Match(buff))
}
