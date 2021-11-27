package db

import (
	"context"

	"prisma/tms/moc"
)

type FileDB interface {
	Create(ctx context.Context, f *moc.File) error
}
