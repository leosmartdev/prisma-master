package db

import (
	"prisma/tms/client_api"
	"prisma/tms/moc"
)

type FilterTracksDB interface {
	MiscDB
	GetFilterTracks(userId string) ([]*moc.FilterTracks, error)
	SaveFilterTrack(filtertracks *moc.FilterTracks) (*client_api.UpsertResponse, error)
}
