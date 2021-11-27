package ws

import (
	"prisma/tms/envelope"
)

type Publisher struct {
	subscribers map[string][]envelope.Subscriber
}

func NewPublisher() *Publisher {
	return &Publisher{
		subscribers: make(map[string][]envelope.Subscriber),
	}
}

func (publisher *Publisher) Subscribe(topic string, subscriber envelope.Subscriber) {
	if nil == publisher.subscribers[topic] {
		publisher.subscribers[topic] = make([]envelope.Subscriber, 0)
	}
	publisher.subscribers[topic] = append(publisher.subscribers[topic], subscriber)
}

func (publisher *Publisher) Publish(topic string, envelope envelope.Envelope) {
	for _, subscriber := range publisher.subscribers[topic] {
		subscriber.Publish(envelope)
	}
}
