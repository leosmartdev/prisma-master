package db

import (
	"context"

	"prisma/tms/moc"
)

type DeviceDB interface {
	Update(ctx context.Context, device *moc.Device) error
	Delete(ctx context.Context, devicelId string) error
	FindOne(ctx context.Context, deviceID string) (*moc.Device, error)
	FindAll(ctx context.Context, sortFields SortFields) ([]*moc.Device, error)
	// FindByMapByPagination(ctx context.Context, searchMap map[string]string, pagination *rest.PaginationQuery) ([]*moc.Vessel, error)
	// FindByMap(ctx context.Context, searchMap map[string]string, sortFields SortFields) ([]*moc.Vessel, error)
	//Insert is used by tdabased in order to insert new devices coming from tfleetd
	Insert(*moc.Device) error
	// RemoveVesselInfoForDevices is used to remove vessel information for several devices for one query
	RemoveVesselInfoForDevices(devices []string) error
	// FindNet is used to find records with the same network id
    FindNet(netID string) (*moc.Device, error)
	// FindByDevice uses DeviceId and Type
	//FIXME: I should take type and deviceId and return moc.Device or just take
	FindByDevice(device *moc.Device) (*moc.Device, error)
	UpsertDeviceConfig(DeviceId, Type string, conf *moc.DeviceConfiguration) error
	// UpsertVesselInfo adds vessel information into a device
	UpsertVesselInfo(device *moc.Device, vesselinfo *moc.VesselInfo) error
}
