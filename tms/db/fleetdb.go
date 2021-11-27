package db

import (
	"context"

	"prisma/tms/moc"
	"prisma/tms/rest"
)

type FleetDB interface {
	Create(ctxt context.Context, fleet *moc.Fleet) (*moc.Fleet, error)
	Update(ctxt context.Context, fleet *moc.Fleet) (*moc.Fleet, error)
	Delete(ctxt context.Context, fleetId string) error
	FindOne(ctxt context.Context, fleetId string) (*moc.Fleet, error)
	FindAll(ctxt context.Context) ([]*moc.Fleet, error)
	FindByMapByPagination(ctxt context.Context, searchMap map[string]string, pagination *rest.PaginationQuery) ([]*moc.Fleet, error)
	FindByMap(ctxt context.Context, searchMap map[string]string) ([]*moc.Fleet, error)
	RemoveVessel(ctxt context.Context, fleetId string, vesselId string) error
	AddVessel(ctxt context.Context, fleetId string, vessel *moc.Vessel) error
	UpdateVessel(ctxt context.Context, fleetId string, vessel *moc.Vessel) error
}
