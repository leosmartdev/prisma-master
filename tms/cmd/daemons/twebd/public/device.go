package public

import (
	"context"
	"net/http"

	"prisma/tms/db"
	"prisma/tms/db/mongo"
	"prisma/tms/moc"
	"prisma/tms/rest"
	"prisma/tms/security"
	"prisma/tms/validate"

	restful "github.com/orolia/go-restful" 
	"github.com/go-openapi/spec"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"fmt"
)

const (
	CLASSIDDevice = security.CLASSIDDevice
)

var (
	SchemaDeviceCreate = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Required: []string{"type"},
			Properties: map[string]spec.Schema{
				"deviceId": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{1}[0],
						MaxLength: &[]int64{255}[0],
					},
				},
				"type": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{1}[0],
						MaxLength: &[]int64{255}[0],
					},
				},
			},
		},
	}
	SchemaDeviceUpdate = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Required: []string{"type", "networks"},
			Properties: map[string]spec.Schema{
				"deviceId": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{1}[0],
						MaxLength: &[]int64{255}[0],
					},
				},
				"type": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{1}[0],
						MaxLength: &[]int64{255}[0],
					},
				},
			},
		},
	}
	parameterDeviceID = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "id",
			In:       "path",
			Required: true,
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					MinLength: &[]int64{24}[0],
					MaxLength: &[]int64{24}[0],
					Pattern:   "[0-9a-fA-F]{24}",
				},
			},
		},
		SimpleSchema: spec.SimpleSchema{
			Type:   "string",
			Format: "hexadecimal",
		},
	}
)

type DeviceRest struct {
	vesselDb db.VesselDB
	deviceDb db.DeviceDB
	ctxt     context.Context
}

func NewDeviceRest(ctx context.Context, client *mongo.MongoClient) *DeviceRest {
	return &DeviceRest{
		vesselDb: mongo.NewMongoVesselDb(ctx),
		deviceDb: mongo.NewMongoDeviceDb(ctx, client),
		ctxt:     ctx,
	}
}

func (r *DeviceRest) CreateWithVessel(req *restful.Request, rsp *restful.Response) {
	ACTION := moc.Device_CREATE.String()
	ctx := req.Request.Context()
	if !authorized(req, rsp, CLASSIDDevice, ACTION) {
		return
	}

	var err error
	vesselID, errs := rest.SanitizeValidatePathParameter(req, parameterVesselId)
	if !valid(errs, req, rsp, CLASSIDDevice, ACTION) {
		return
	}

	vessel, err := r.vesselDb.FindOne(ctx, vesselID)
	if !errorFree(err, req, rsp, CLASSIDDevice, ACTION) {
		return
	}
	device := new(moc.Device)
	errs = rest.SanitizeValidateReadProto(req, SchemaDeviceCreate, device)
	if !valid(errs, req, rsp, CLASSIDDevice, ACTION) {
		return
	}
	savedDevice, err := r.deviceDb.FindByDevice(device)
	if err != nil {
		if device.VesselInfo == nil {
			device.VesselInfo = &moc.VesselInfo{
				Id: vesselID,
				Type: vessel.Type,
			}
		}
		savedDevice = device
		err = r.deviceDb.Insert(device)
	} else if savedDevice.VesselInfo == nil {
		err = r.deviceDb.UpsertVesselInfo(device,  &moc.VesselInfo{
			Id: vesselID,
			Type: vessel.Type,
		})
	} else if savedDevice.VesselInfo != nil && savedDevice.VesselInfo.Id != vessel.Id {
		err = errors.New("this devices is already assigned")
	}
	if !errorFree(err, req, rsp, CLASSIDDevice, ACTION) {
		return
	}
	vessel.Devices = append(vessel.Devices, savedDevice)
	// check on duplicating devices
	if dID := duplicatingDevice(vessel); dID != "" {
		errorFree(fmt.Errorf("device %s is duplicated", dID), req, rsp, CLASSIDDevice, ACTION)
		return
	}
	vessel, err = r.vesselDb.Update(ctx, vessel)

	security.Audit(ctx, CLASSIDDevice, ACTION, security.SUCCESS)
	rest.WriteHeaderAndEntitySafely(rsp, http.StatusCreated, savedDevice)
}

func (r *DeviceRest) Create(req *restful.Request, rsp *restful.Response) {
	ACTION := moc.Device_CREATE.String()
	ctx := req.Request.Context()
	if !authorized(req, rsp, CLASSIDDevice, ACTION) {
		return
	}

	device := new(moc.Device)
	errs := rest.SanitizeValidateReadProto(req, SchemaDeviceCreate, device)
	if !valid(errs, req, rsp, CLASSIDDevice, ACTION) {
		return
	}

	var errsDeviceType []rest.ErrorValidation
	switch device.Type {
	case "email":
		for _, v := range device.Networks {
			errsDeviceType = append(errsDeviceType, validate.Email(v.SubscriberId)...)
		}
	case "ais", "mob-ais", "sart-ais":
		for _, v := range device.Networks {
			errsDeviceType = append(errsDeviceType, validate.MMSI(v.SubscriberId)...)
		}
	}
	if !valid(errsDeviceType, req, rsp, CLASSIDDevice, ACTION) {
		return
	}

	err := r.deviceDb.Insert(device)
	if !errorFree(err, req, rsp, CLASSIDDevice, ACTION) {
		return
	}

	security.Audit(ctx, CLASSIDDevice, ACTION, security.SUCCESS)
	rest.WriteHeaderAndEntitySafely(rsp, http.StatusCreated, device)
}

func (r *DeviceRest) Update(req *restful.Request, rsp *restful.Response) {
	ACTION := moc.Device_UPDATE.String()
	ctx := req.Request.Context()

	if !authorized(req, rsp, CLASSIDDevice, ACTION) {
		return
	}

	device := &moc.Device{}
	deviceId, errs := rest.SanitizeValidatePathParameter(req, parameterDeviceID)
	errs2 := rest.SanitizeValidateReadProto(req, SchemaDeviceUpdate, device)
	errs = append(errs, errs2...)
	errs3 := validateIdEqual(deviceId, device.Id)
	errs = append(errs, errs3...)
	if !valid(errs, req, rsp, CLASSIDDevice, ACTION) {
		return
	}

	err := r.deviceDb.Update(ctx, device)
	if !errorFree(err, req, rsp, CLASSIDDevice, ACTION) {
		return
	}

	security.Audit(ctx, CLASSIDDevice, ACTION, security.SUCCESS)
	rest.WriteProtoSafely(rsp, device)
}

func (r *DeviceRest) Delete(req *restful.Request, rsp *restful.Response) {
	ACTION := moc.Device_DELETE.String()
	ctx := req.Request.Context()
	if !authorized(req, rsp, CLASSIDDevice, ACTION) {
		return
	}

	deviceId, errs := rest.SanitizeValidatePathParameter(req, parameterDeviceID)
	if !valid(errs, req, rsp, CLASSIDDevice, ACTION) {
		return
	}

	err := r.deviceDb.Delete(ctx, deviceId)
	if !errorFree(err, req, rsp, CLASSIDDevice, ACTION) {
		return
	}

	security.Audit(ctx, CLASSIDDevice, ACTION, security.SUCCESS, deviceId)
	rest.WriteHeaderAndProtoSafely(rsp, http.StatusNoContent, nil)
}

func (r *DeviceRest) GetAllDevices(req *restful.Request, rsp *restful.Response) {
	ACTION := moc.Device_READ.String()
	ctx := req.Request.Context()
	if !authorized(req, rsp, CLASSIDDevice, ACTION) {
		return
	}

	sortFields := db.SortFields{}
	sortFields = append(sortFields, db.FieldOrder{
		Field: "deviceId",
	})
	arr, err := r.deviceDb.FindAll(ctx, sortFields)
	if !errorFree(err, req, rsp, CLASSIDDevice, ACTION) {
		return
	}

	security.Audit(ctx, CLASSIDDevice, ACTION, security.SUCCESS)
	rsp.WriteEntity(arr)
	//rest.WriteProtoSpliceSafely(rsp, toMessagesFromDevices(arr))
}

func (r *DeviceRest) GetDeviceByVesselId(req *restful.Request, rsp *restful.Response) {
	ACTION := moc.Device_READ.String()
	ctx := req.Request.Context()

	if !authorized(req, rsp, CLASSIDDevice, ACTION) {
		return
	}

	vesselID, errs := rest.SanitizeValidatePathParameter(req, parameterVesselId)
	if !valid(errs, req, rsp, CLASSIDDevice, ACTION) {
		return
	}

	vessel, err := r.vesselDb.FindOne(ctx, vesselID)
	if !errorFree(err, req, rsp, CLASSIDDevice, ACTION) {
		return
	}

	security.Audit(ctx, CLASSIDDevice, ACTION, security.SUCCESS)
	rest.WriteHeaderAndEntitySafely(rsp, http.StatusOK, vessel.Devices)
}

func (r *DeviceRest) GetDeviceById(req *restful.Request, rsp *restful.Response) {
	ACTION := moc.Device_READ.String()
	ctx := req.Request.Context()

	if !authorized(req, rsp, CLASSIDDevice, ACTION) {
		return
	}

	deviceID, errs := rest.SanitizeValidatePathParameter(req, parameterDeviceID)
	if !valid(errs, req, rsp, CLASSIDDevice, ACTION) {
		return
	}

	v, err := r.deviceDb.FindOne(ctx, deviceID)
	if !errorFree(err, req, rsp, CLASSIDDevice, ACTION) {
		return
	}

	security.Audit(ctx, CLASSIDDevice, ACTION, security.SUCCESS)
	rest.WriteProtoSafely(rsp, v)
}

func createDevices(ctxt context.Context, deviceDb db.DeviceDB, devices []*moc.Device) error {
	var err error
	for i := range devices {
		if devices[i].Id == "" {
			foundDevice, err := deviceDb.FindByDevice(devices[i])
			if err == nil {
				devices[i] = foundDevice
				continue
			}
			err = deviceDb.Insert(devices[i])
		} else {
			foundDevice, err := deviceDb.FindOne(ctxt, devices[i].Id)
			if err == nil {
				devices[i] = foundDevice
			}
		}
		if err != nil {
			break
		}
	}
	return err
}

func toMessagesFromDevices(mcs []*moc.Device) []proto.Message {
	var messages []proto.Message
	for _, mc := range mcs {
		messages = append(messages, mc)
	}
	return messages
}
