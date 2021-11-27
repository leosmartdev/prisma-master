package client

import (
	"errors"

	"prisma/tms"
	"prisma/tms/routing"
	"prisma/tms/envelope"

	pb "github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
)

var (
	BadMessageType = errors.New("Unexpected message type recieved")
)

type TsiClient interface {
	envelope.Subscriber
	// ******* Sending functions (of various sorts)
	Send(context.Context, *tms.TsiMessage)
	// NotifyID gives you ID before sending, must call Send() afterwards
	NotifyID(callback func(*routing.DeliveryReport)) int32
	// Send message, call callback when a delivery report is returned
	SendNotify(context.Context, *tms.TsiMessage, func(*routing.DeliveryReport))
	SendTo(context.Context, tms.EndPoint, pb.Message)
	BroadcastLocal(context.Context, pb.Message)
	SendToGateway(context.Context, pb.Message)
	// Listen for messages
	Listen(context.Context, routing.Listener) <-chan *TMsg
	// RPC functions
	RegisterHandler(msgType string, handler func(*TMsg) pb.Message)
	Request(context.Context, tms.EndPoint, pb.Message) (pb.Message, error)
	// General info
	Local() *tms.EndPoint
	LocalRouter() *tms.EndPoint
	// Resolution functions
	ResolveSite(string) uint32
	ResolveApp(string) uint32
}

type TMsg struct {
	tms.TsiMessage
	Body pb.Message
}

func (msg *TMsg) Type() string {
	if msg.Body == nil {
		return ""
	}
	return pb.MessageName(msg.Body)
}
