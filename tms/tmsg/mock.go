package tmsg

import (
	"prisma/tms"
	"prisma/tms/envelope"
	"prisma/tms/routing"
	"prisma/tms/tmsg/client"

	pb "github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
)

type MockTsiClient struct {
	mock.Mock
}

func (c *MockTsiClient) NotifyID(callback func(*routing.DeliveryReport)) int32 {
	return 0
}

func (c *MockTsiClient) Send(ctxt context.Context, m *tms.TsiMessage) {
	c.Called(ctxt, m)
}

func (c *MockTsiClient) SendNotify(ctxt context.Context, m *tms.TsiMessage, report func(*routing.DeliveryReport)) {
	c.Called(ctxt, m, report)
}

func (c *MockTsiClient) SendTo(ctxt context.Context, ep tms.EndPoint, m pb.Message) {
	c.Called(ctxt, ep, m)
}

func (c *MockTsiClient) BroadcastLocal(ctxt context.Context, m pb.Message) {
	c.Called(ctxt, m)
}

func (c *MockTsiClient) SendToGateway(ctxt context.Context, m pb.Message) {
	c.Called(ctxt, m)
}

func (c *MockTsiClient) Listen(ctxt context.Context, l routing.Listener) <-chan *client.TMsg {
	args := c.Called(ctxt, l)
	return args.Get(0).(<-chan *client.TMsg)
}

func (c *MockTsiClient) RegisterHandler(msgType string, handler func(*client.TMsg) pb.Message) {
	c.Called(msgType, handler)
}

func (c *MockTsiClient) Request(ctxt context.Context, ep tms.EndPoint, m pb.Message) (pb.Message, error) {
	args := c.Called(ctxt, ep, m)
	return args.Get(0).(pb.Message), args.Error(1)
}

func (c *MockTsiClient) Local() *tms.EndPoint {
	args := c.Called()
	return args.Get(0).(*tms.EndPoint)
}

func (c *MockTsiClient) LocalRouter() *tms.EndPoint {
	args := c.Called()
	return args.Get(0).(*tms.EndPoint)
}

func (c *MockTsiClient) ResolveSite(site string) uint32 {
	args := c.Called(site)
	return args.Get(0).(uint32)
}

func (c *MockTsiClient) ResolveApp(site string) uint32 {
	args := c.Called(site)
	return args.Get(0).(uint32)
}

func (c *MockTsiClient) Publish(envelope envelope.Envelope) {}

func NewTsiClientStub() *MockTsiClient {
	c := &MockTsiClient{}
	c.On("ResolveSite", mock.Anything, mock.Anything).Return(uint32(0))
	c.On("Local").Return(&tms.EndPoint{})
	c.On("Send", mock.Anything, mock.Anything).Return()
	return c
}
