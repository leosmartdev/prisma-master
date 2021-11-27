// Package envelope provides interface and structures to public information.
package envelope

// Subscriber is used to add an opportunity publish messages using source with contents
type Subscriber interface {
	Publish(envelope Envelope)
}
