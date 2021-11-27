package main

import (
	"testing"
	"prisma/gogroup"
	"context"
	"prisma/tms"
	"github.com/stretchr/testify/assert"
	"bytes"
	"prisma/tms/routing"
	"github.com/golang/protobuf/ptypes/any"
)

type mockRWC struct {
	*bytes.Buffer
}

func (*mockRWC) Close() error {
	return nil
}

func TestRouterHandle(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r := NewRouter(gogroup.New(ctx, "test"), 1, "test")
	buff := mockRWC{}
	buff.String()
	channel := &IOChannel{
		fromRouter: make(chan *tms.TsiMessage, 1),
		rtr:        r,
		reg: routing.Registry{
			Entries: []*routing.Listener{
				{
					MessageType: "test",
				},
			},
		},
	}
	_ = r
	_ = channel
	msg := &tms.TsiMessage{
		Status: tms.TsiMessage_REQUEST,
		Source: &tms.EndPoint{
			Eid: 1,
		},
		Destination: []*tms.EndPoint{
			{
				Eid: 1,
			},
		},
		Body: &any.Any{
			TypeUrl: "test",
			Value:   []byte{0x01},
		},
	}
	r.AddChannel(channel)
	<-channel.fromRouter
	r.handle(msg)
	assert.Equal(t, msg, <-channel.fromRouter)
	r.handle(msg)
	assert.False(t, r.isFullFill[channel])
	msg.Body.Value = []byte{0x02}
	r.handle(msg)
	assert.True(t, r.isFullFill[channel])
}
