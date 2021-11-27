// Package clock provides a mock for time package.
package clock

import (
	"time"
)

type C interface {
	Now() time.Time
}

type Real struct{}

func (c *Real) Now() time.Time {
	return time.Now()
}

type Mock struct {
	MockNow time.Time
}

func (c *Mock) Now() time.Time {
	return c.MockNow
}
