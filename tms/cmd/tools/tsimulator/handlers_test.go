package main

import (
	"testing"
	"context"
	"prisma/tms/cmd/tools/tsimulator/object"
	"net"
	"github.com/json-iterator/go/assert"
	"time"
	"io/ioutil"
	"prisma/tms/omnicom"
	"prisma/tms/iridium"
)

func Test_handleBeacons(t *testing.T) {
	tFleetdTCPAddr, _ = net.ResolveTCPAddr("tcp", ":65531")
	listener, err := net.ListenTCP("tcp", tFleetdTCPAddr)
	assert.NoError(t, err)
	defer listener.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	go func() {
		c, err := listener.Accept()
		assert.NoError(t, err)
		defer c.Close()
		data, err := ioutil.ReadAll(c)
		assert.NoError(t, err)
		head, err := iridium.ParseMOHeader(data[3:])
		assert.NoError(t, err)
		omni, err := iridium.ParseMOPayload(data[head.MOHL+6:])

		assert.NoError(t, err)
		ar, ok := omni.Omn.(*omnicom.AR)
		assert.True(t, ok)
		assert.Equal(t, omnicom.Test_Mode{1, 1}, ar.Test_Mode)
	}()

	chObject := make(chan object.Object, 1)
	go handleBeacons(ctx, chObject)()
	so := object.Object{
		Device: "omnicom",
	}
	so.Init()
	chObject <- so
	<-ctx.Done()
}
