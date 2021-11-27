// Package task contains structures to work with tasks for beacons.
package task

import "errors"

// ErrTaskQueueFull is used to determine that a queue for tasks is full
// and you cannot add more tasks yet
var ErrTaskQueueFull = errors.New("queue is full")

// Result will contain data for sending and where to (App constants)
type Result struct {
	Data []byte
	To   uint32
}
