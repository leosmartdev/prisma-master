package db

import (
	"prisma/tms/moc"
)

type Sit915DB interface {
	UpsertSit915(sit915 *moc.Sit915) error
	FindAllSit915s(params ...int) ([]*moc.Sit915, error)
	FindOneSit915(id string) (*moc.Sit915, error)
}
