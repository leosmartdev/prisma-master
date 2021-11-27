package session

import (
	"context"

	securityContext "prisma/tms/security/context"
)

type Store interface {
	Create(owner string, roles []string) (InternalSession, error)
	Get(id string) (InternalSession, error)
	Delete(id string) error
}

func GetStore(ctxt context.Context) Store {
	store := Store(nil)
	switch securityContext.SessionStoreIdFromContext(ctxt) {
	case "mock":
		store = mockStoreInstance()
	default:
		store = mongoStoreInstance(ctxt)
	}
	return store
}

func SetPublisher(ctxt context.Context, publisher Publisher) error {
	switch securityContext.SessionStoreIdFromContext(ctxt) {
	case "mock":
		// ignore
	default:
		// init
		mongoStoreInstance(ctxt)
		return mongoSetPublisher(ctxt, publisher)
	}
	return nil
}
