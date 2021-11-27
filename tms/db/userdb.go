package db

import (
	"context"

	"prisma/tms/security/database"
)

type UserDB interface {
	FindOne(ctx context.Context, userID string) (*database.User, error)
}
