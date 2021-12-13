package db

import (
	"prisma/tms/client_api"
	"prisma/tms/moc"
)

type MapConfigDB interface {
	MiscDB
	FindAllMapConfig() ([]GoGetResponse, error)
	SaveMapConfig(mapconfig *moc.MapConfig) (*client_api.UpsertResponse, error)
}
