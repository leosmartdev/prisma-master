package incident

import (
	"context"

	securityContext "prisma/tms/security/context"
)

// IdPrefixer get a prefix of an Id
type IdPrefixer interface {
	Prefix() string
}

// IdCreator
type IdCreator interface {
	// Returns the next reference identifier
	Next(prefixer IdPrefixer) string
}

func IdCreatorInstance(context context.Context) IdCreator {
	var creator IdCreator
	switch securityContext.SessionStoreIdFromContext(context) {
	case "mock":
		creator = mockIdCreatorInstance()
	default:
		creator = mongoIdCreatorInstance(context)
	}
	return creator
}
