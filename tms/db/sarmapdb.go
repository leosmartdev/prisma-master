package db

import (
	"context"
	"prisma/tms/moc"
)

// SarmapDB interface ...
type SarmapDB interface {
	// FindAll recovers all the targets assigned to an open incident
	FindAll(ctx context.Context) ([]*moc.GeoJsonFeaturePoint, error)
}
