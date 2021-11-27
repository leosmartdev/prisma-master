package db

import (
	"prisma/tms/client_api"
	"prisma/tms/marker"
	"prisma/tms/moc"

	"github.com/globalsign/mgo/bson"
)

type MarkerDB interface {
	UpsertMarker(marker *marker.Marker) (*client_api.UpsertResponse, error)
	FindOneMarker(markerId string, withDeleted bool) (*marker.Marker, error)
	DeleteMarker(markerId string) error
	UpsertMarkerImage(markerImage *marker.MarkerImage) (*client_api.UpsertResponse, error)
	FindAllMarkerImages() ([]*marker.MarkerImage, error)
	GetPersistentStream(pipeline []bson.M) *MarkerStream
}

type MarkerStream struct {
	Updates chan *moc.GeoJsonFeaturePoint
}

func NewMarkerStream() *MarkerStream {
	return &MarkerStream{
		Updates: make(chan *moc.GeoJsonFeaturePoint),
	}
}
