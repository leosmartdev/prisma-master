// Package db container abstraction to work with different resources.
package db

import (
	"prisma/tms"
)

//interface to a database that contains activities
type ActivityDB interface {
	Insert(*tms.MessageActivity) error
	GetSit915Messages(startDateTime int, endDateTime int) ([]*tms.MessageActivity, error)
}
