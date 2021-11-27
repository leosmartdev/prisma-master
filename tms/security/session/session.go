// Package session publishes actual information about current session.
package session

import (
	"errors"

	"prisma/tms/security/message"
)

var (
	ErrorInvalidId = errors.New("invalidId")
	ErrorNotFound  = errors.New("notFound")
	ErrorExpired   = errors.New("expired")
)

type InternalSession interface {
	Id() string
	GetRoles() []string
	GetOwner() string
	GetState() message.Session_State
}

type Publisher interface {
	Publish(action message.Session_Action, session InternalSession)
}
