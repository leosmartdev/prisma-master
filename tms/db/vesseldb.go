package db

import (
	"context"

	"prisma/tms/moc"
	"prisma/tms/rest"
)

type SortFields []FieldOrder

type FieldOrder struct {
	Desc  bool
	Field string
}

type VesselDB interface {
	Create(ctxt context.Context, vessel *moc.Vessel) (*moc.Vessel, error)
	Update(ctxt context.Context, vessel *moc.Vessel) (*moc.Vessel, error)
	Delete(ctxt context.Context, vesselId string) error
	FindOne(ctxt context.Context, vesselId string) (*moc.Vessel, error)
	FindAll(ctxt context.Context, sortFields SortFields) ([]*moc.Vessel, error)
	FindByMapByPagination(ctxt context.Context, searchMap map[string]string, pagination *rest.PaginationQuery) ([]*moc.Vessel, error)
	FindByMap(ctxt context.Context, searchMap map[string]string, sortFields SortFields) ([]*moc.Vessel, error)
	FindByDevice(ctxt context.Context, device *moc.Device) (*moc.Vessel, error)
}
