package public

import (
	"context"
	"net/http"
	"strings"

	"prisma/gogroup"
	"prisma/tms/db"
	"prisma/tms/db/mongo"
	"prisma/tms/envelope"
	"prisma/tms/log"
	"prisma/tms/moc"
	"prisma/tms/rest"
	"prisma/tms/security"
	"prisma/tms/ws"

	"fmt"
	restful "github.com/orolia/go-restful" 
	"github.com/globalsign/mgo/bson"
	"github.com/go-openapi/spec"
	"github.com/golang/protobuf/proto"
)

const CLASSIDVessel = security.CLASSIDVessel
const TOPIC_VESSEL = CLASSIDVessel

var (
	SchemaVesselCreate = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Required:   []string{"name"},
			Properties: map[string]spec.Schema{},
		},
	}
	SchemaVesselUpdate = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Required:   []string{"id", "name"},
			Properties: map[string]spec.Schema{},
		},
	}
	parameterVesselId = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "vessel-id",
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
	queryFleetExists = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name: "has-fleet",
			In:   "query",
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Pattern: "^(true|false)$",
				},
			},
		},
		SimpleSchema: spec.SimpleSchema{
			Type: "boolean",
		},
	}
	queryDeviceSubscriberId = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name: "devices.networks.subscriberId",
			In:   "query",
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					MinLength: &[]int64{15}[0],
					MaxLength: &[]int64{15}[0],
					Pattern:   "[0-9]{15}",
				},
			},
		},
	}
)

type VesselRest struct {
	vesselDb  db.VesselDB
	fleetDb   db.FleetDB
	deviceDb  db.DeviceDB
	publisher *ws.Publisher
}

// NewVesselRest returns an instance for rest api of search endpoints
func NewVesselRest(group gogroup.GoGroup, client *mongo.MongoClient, publisher *ws.Publisher) *VesselRest {
	return &VesselRest{
		vesselDb:  mongo.NewMongoVesselDb(group),
		fleetDb:   mongo.NewFleetDb(group),
		deviceDb:  mongo.NewMongoDeviceDb(group, client),
		publisher: publisher,
	}
}

func (r *VesselRest) FindOne(req *restful.Request, rsp *restful.Response) {
	ACTION := moc.Vessel_READ.String()
	ctxt := req.Request.Context()
	if authorized(req, rsp, CLASSIDVessel, ACTION) {
		vesselId, errs := rest.SanitizeValidatePathParameter(req, parameterVesselId)
		if valid(errs, req, rsp, CLASSIDVessel, ACTION) {
			vessel, err := r.vesselDb.FindOne(ctxt, vesselId)
			if errorFree(err, req, rsp, CLASSIDVessel, ACTION) {
				auditVessel(ctxt, vessel, ACTION, security.SUCCESS, nil)
				rest.WriteProtoSafely(rsp, vessel)
			}
		}
	}
}

func (r *VesselRest) ReadAll(request *restful.Request,
	response *restful.Response) {
	ACTION := moc.Vessel_READ.String()
	ctxt := request.Request.Context()
	if authorized(request, response, CLASSIDVessel, ACTION) {
		var vessels []*moc.Vessel
		var err error
		fleetExists, errs := rest.SanitizeValidateQueryParameter(request, queryFleetExists)
		subscriberId, errs2 := rest.SanitizeValidateQueryParameter(request, queryDeviceSubscriberId)
		errs = append(errs, errs2...)
		if valid(errs, request, response, CLASSIDVessel, ACTION) {
			sortFields := db.SortFields{}
			// search map from request query parameters
			searchMap := make(map[string]string)
			if request.QueryParameter("devices.type") != "" {
				searchMap["devices.type"] = request.QueryParameter("devices.type")
			}
			if strings.Contains(request.SelectedRoutePath(), "/fleet") {
				sortFields = append(sortFields, db.FieldOrder{
					Field: "fleet.name",
				})
			}
			if fleetExists != "" {
				searchMap["fleet.id"] = "$exists," + fleetExists
			}
			if subscriberId != "" {
				searchMap["devices.networks.subscriberid"] = subscriberId // lowercase
			}
			pagination, ok := rest.SanitizePagination(request)
			if ok {
				pagination.Sort = "name"
				if strings.Contains(request.SelectedRoutePath(), "/fleet") {
					pagination.Sort = "fleet.name"
				}
				vessels, err = r.vesselDb.FindByMapByPagination(ctxt, searchMap, pagination)
				if !errorFree(err, request, response, CLASSIDVessel, ACTION) {
					return
				}
				if len(vessels) > 0 {
					if pagination.Anchor != "" {
						anchorFound := false
						var anchorVessels []*moc.Vessel
						for index, vessel := range vessels {
							if pagination.Anchor == vessel.Id {
								pagination.Skip = index
								anchorFound = true
							}
							if anchorFound && ((index - pagination.Skip) < pagination.Limit) {
								anchorVessels = append(anchorVessels, vessel)
							}
						}
						if anchorFound {
							vessels = anchorVessels
						}
					}
					pagination.Count = len(vessels)
					rest.AddPaginationHeaderSafely(request, response, pagination)
				}
			} else {
				sortFields = append(sortFields, db.FieldOrder{
					Field: "vessel.name",
				})
				vessels, err = r.vesselDb.FindByMap(ctxt, searchMap, sortFields)
			}
			if errorFree(err, request, response, CLASSIDVessel, ACTION) {
				security.Audit(ctxt, CLASSIDVessel, ACTION, security.SUCCESS)
				rest.WriteProtoSpliceSafely(response, toMessagesFromVessels(vessels))
			}
		}
	}
}

func (r *VesselRest) Create(request *restful.Request,
	response *restful.Response) {
	ACTION := moc.Vessel_CREATE.String()
	ctxt := request.Request.Context()
	if !authorized(request, response, CLASSIDVessel, ACTION) {
		return
	}
	vessel := &moc.Vessel{}
	errs := rest.SanitizeValidateReadEntity(request, SchemaVesselCreate, vessel)
	if !valid(errs, request, response, CLASSIDVessel, ACTION) {
		return
	}
	// setup registry id
	for _, device := range vessel.Devices {
		for _, network := range device.Networks {
			network.RegistryId = moc.GetRegistryIdByType(network.Type, network.SubscriberId)
		}
	}
	// record the devices to mongodb
	err := createDevices(ctxt, r.deviceDb, vessel.Devices)
	if err == nil {
		for _, device := range vessel.Devices {
			if device.Id == "" {
				device.Id = bson.NewObjectId().Hex()
			}
		}
		_, err = r.vesselDb.Create(ctxt, vessel)
	}
	// if we do not have errors then answer okay
	if errorFree(err, request, response, CLASSIDVessel, ACTION) {
		r.Publish(ACTION, vessel)
		auditVessel(ctxt, vessel, ACTION, security.SUCCESS, nil)
		rest.WriteHeaderAndProtoSafely(response, http.StatusCreated, vessel)
	}
}

func (r *VesselRest) Update(request *restful.Request,
	response *restful.Response) {
	ACTION := moc.Vessel_UPDATE.String()
	ctxt := request.Request.Context()
	if !authorized(request, response, CLASSIDVessel, ACTION) {
		return
	}
	vessel := &moc.Vessel{}
	errs := rest.SanitizeValidateReadProto(request, SchemaVesselUpdate, vessel)
	vesselId, errs2 := rest.SanitizeValidatePathParameter(request, parameterVesselId)
	errs = append(errs, errs2...)
	errs3 := validateIdEqual(vesselId, vessel.Id)
	errs = append(errs, errs3...)
	if !valid(errs, request, response, CLASSIDVessel, ACTION) {
		return
	}
	originalVessel, err := r.vesselDb.FindOne(ctxt, vesselId)
	if !errorFree(err, request, response, CLASSID, ACTION) {
		return
	}
	// check on duplicating devices
	if dID := duplicatingDevice(vessel); dID != "" {
		errorFree(fmt.Errorf("device %s is duplicated", dID), request, response, CLASSIDDevice, ACTION)
		return
	}
	// check if a client wants to change devices
	if err := r.deviceUpdating(originalVessel, vessel); !errorFree(err, request, response, CLASSIDVessel, ACTION) {
		return
	}
	// if fleet changed, get original fleet
	if originalVessel.Fleet != vessel.Fleet {
		if vessel.Fleet != nil && vessel.Fleet.Id != "" {
			originalFleet, err := r.fleetDb.FindOne(ctxt, vessel.Fleet.Id)
			if errorFree(err, request, response, CLASSIDVessel, ACTION) {
				vessel.Fleet.Name = originalFleet.Name
			}
		}
	}
	vessel, err = r.vesselDb.Update(ctxt, vessel)
	// upsert fleet
	if vessel.Fleet != nil {
		err = r.fleetDb.UpdateVessel(ctxt, vessel.Fleet.Id, vessel)
		if db.ErrorNotFound == err {
			err = r.fleetDb.AddVessel(ctxt, vessel.Fleet.Id, vessel)
		}
		auditFleet(ctxt, &moc.Fleet{
			Id:   vessel.Fleet.Id,
			Name: vessel.Fleet.Name,
		}, moc.Fleet_ADD_VESSEL.String(), security.SUCCESS, vessel)
	}
	if err == nil && originalVessel.Fleet != nil && (vessel.Fleet == nil || originalVessel.Fleet.Id != vessel.Fleet.Id) {
		err = r.fleetDb.RemoveVessel(ctxt, originalVessel.Fleet.Id, vessel.Id)
		auditFleet(ctxt, &moc.Fleet{
			Id:   originalVessel.Fleet.Id,
			Name: originalVessel.Fleet.Name,
		}, moc.Fleet_REMOVE_VESSEL.String(), security.SUCCESS, vessel)
	}
	if errorFree(err, request, response, CLASSIDVessel, ACTION) {
		r.Publish(ACTION, vessel)
		auditVessel(ctxt, vessel, ACTION, security.SUCCESS, nil)
		rest.WriteProtoSafely(response, vessel)
	}
}

func (r *VesselRest) Delete(request *restful.Request,
	response *restful.Response) {
	ACTION := moc.Vessel_DELETE.String()
	ctxt := request.Request.Context()
	if authorized(request, response, CLASSIDVessel, ACTION) {
		vesselId, errs := rest.SanitizeValidatePathParameter(request, parameterVesselId)
		if valid(errs, request, response, CLASSIDVessel, ACTION) {
			vessel, err := r.vesselDb.FindOne(ctxt, vesselId)
			if err == nil {
				err = r.vesselDb.Delete(ctxt, vesselId)
			}
			if errorFree(err, request, response, CLASSIDVessel, ACTION) {
				r.Publish(ACTION, &moc.Vessel{Id: vesselId})
				auditVessel(ctxt, vessel, ACTION, security.SUCCESS, nil)
				rest.WriteHeaderAndProtoSafely(response, http.StatusNoContent, nil)
			}
		}
	}
}

func (r *VesselRest) Publish(action string, vessel *moc.Vessel) {
	if r.publisher == nil {
		return
	}
	envelope := envelope.Envelope{
		Type:     TOPIC_VESSEL + "/" + action,
		Contents: &envelope.Envelope_Vessel{Vessel: vessel},
	}
	r.publisher.Publish(TOPIC_VESSEL, envelope)
}

func (r *VesselRest) deviceUpdating(prev, new *moc.Vessel) error {
	var deleteDevices []string
outLoopDeleting:
	for i := range prev.Devices {
		for j := range new.Devices {
			if new.Devices[j].Id == prev.Devices[i].Id {
				continue outLoopDeleting
			}
		}
		deleteDevices = append(deleteDevices, prev.Devices[i].Id)
	}
	if err := r.deviceDb.RemoveVesselInfoForDevices(deleteDevices); err != nil {
		log.Error("unable to remove devices: %s", err)
	}

outLoopUpdating:
	for i := range new.Devices {
		for j := range prev.Devices {
			if new.Devices[i].Id == prev.Devices[j].Id {
				continue outLoopUpdating
			}
		}
		err := r.deviceDb.UpsertVesselInfo(new.Devices[i], &moc.VesselInfo{
			Id:   new.Id,
			Type: new.Type,
		})
		if err != nil {
			err = fmt.Errorf("unable to add vessel info to a device: %s", err)
			return err
		}
	}
	return nil
}

func auditVessel(ctxt context.Context, vessel *moc.Vessel, action string, outcome string, payload interface{}) {
	if payload == nil {
		security.AuditUserObject(ctxt, CLASSIDVessel, vessel.Id, "", action, outcome)
		return
	}
	fleet, ok := payload.(*moc.Fleet)
	if ok {
		security.AuditUserObject(ctxt, CLASSIDVessel, vessel.Id, "", action, outcome, fleet.Name, fleet.Id)
		return
	}
	payloads, ok := payload.([]interface{})
	if ok {
		security.AuditUserObject(ctxt, CLASSIDVessel, vessel.Id, "", action, outcome, payloads...)
		return
	}
	security.AuditUserObject(ctxt, CLASSIDVessel, vessel.Id, "", action, outcome, payload)
}

func toMessagesFromVessels(vessels []*moc.Vessel) []proto.Message {
	var messages []proto.Message
	for _, vessel := range vessels {
		messages = append(messages, vessel)
	}
	return messages
}

func validateIdEqual(id string, id2 string) []rest.ErrorValidation {
	var errs []rest.ErrorValidation
	if id != id2 {
		errs = append(errs, rest.ErrorValidation{
			Property: "id",
			Rule:     "Equal",
			Message:  "mismatch"})
	}
	return errs
}
