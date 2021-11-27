package db

import (
	"prisma/tms/moc"
)

type RemoteSiteDB interface {
	MiscDB
	UpsertRemoteSite(remoteSite *moc.RemoteSite) error
	FindAllRemoteSites() ([]*moc.RemoteSite, error)
	FindOneRemoteSite(id string, withDeleted bool) (*moc.RemoteSite, error)
	FindOneRemoteSiteByCscode(cscode string) (*moc.RemoteSite, error)
	DeleteRemoteSite(id string) error
}
