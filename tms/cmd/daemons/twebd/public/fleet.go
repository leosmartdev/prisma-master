package public

import (
	"context"
	"net/http"
	"strings"
	"encoding/json"

	"prisma/tms/db"
	"prisma/tms/db/mongo"
	"prisma/tms/moc"
	"prisma/tms/rest"
	"prisma/tms/security"
	"prisma/tms/envelope"
	"prisma/tms/ws"

	restful "github.com/orolia/go-restful" 
	"github.com/go-openapi/spec"
	"github.com/golang/protobuf/proto"
)

const (
	CLASSIDFleet = security.CLASSIDFleet
	TOPIC_FLEET  = CLASSIDFleet
)

var (
	SchemaFleetCreate = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Required: []string{"name"},
			Properties: map[string]spec.Schema{
				"name": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{2}[0],
						MaxLength: &[]int64{255}[0],
					},
				},
				"description": {
					SchemaProps: spec.SchemaProps{
						MaxLength: &[]int64{2000}[0],
					},
				},
			},
		},
	}
	SchemaFleetUpdate = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Required: []string{"id", "name"},
			Properties: map[string]spec.Schema{
				"name": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{2}[0],
						MaxLength: &[]int64{255}[0],
					},
				},
				"description": {
					SchemaProps: spec.SchemaProps{
						MaxLength: &[]int64{2000}[0],
					},
				},
			},
		},
	}
	parameterFleetId = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "fleet-id",
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

type FleetRest struct {
	fleetDb   db.FleetDB
	vesselDb  db.VesselDB
	ctxt      context.Context
	publisher *ws.Publisher
}

// NewFleetRest returns an instance for rest api of fleet endpoints
func NewFleetRest(ctxt context.Context, publisher *ws.Publisher) *FleetRest {
	return &FleetRest{
		fleetDb:   mongo.NewFleetDb(ctxt), //mockdb.NewFleetDb(group),
		vesselDb:  mongo.NewMongoVesselDb(ctxt),
		ctxt:      ctxt,
		publisher: publisher,
	}
}

func (r *FleetRest) ReadAll(request *restful.Request, response *restful.Response) {
	ACTION := moc.Fleet_READ.String()
	ctxt := request.Request.Context()
	if authorized(request, response, CLASSIDFleet, ACTION) {
		var fleets []*moc.Fleet
		var err error
		pagination, ok := rest.SanitizePagination(request)
		if ok {
			searchMap := make(map[string]string)
			if request.QueryParameter("type") != "" {
				searchMap["type"] = request.QueryParameter("type")
			}
			pagination.Sort = "name"
			fleets, err = r.fleetDb.FindByMapByPagination(ctxt, searchMap, pagination)
			if len(fleets) > 0 {
				pagination.Count = len(fleets)
				rest.AddPaginationHeaderSafely(request, response, pagination)
			}
		} else {
			fleets, err = r.fleetDb.FindAll(ctxt)
		}
		if strings.HasSuffix(request.SelectedRoutePath(), "/fleet") {
			for _, fleet := range fleets {
				fleet.Vessels = nil
			}
		}
		if errorFree(err, request, response, CLASSIDFleet, ACTION) {
			security.Audit(ctxt, CLASSIDFleet, ACTION, security.SUCCESS)
			rest.WriteProtoSpliceSafely(response, toMessagesFromFleets(fleets))
		}
	}
}

func (r *FleetRest) ReadOne(req *restful.Request, rsp *restful.Response) {
	ACTION := moc.Fleet_READ.String()
	ctx := req.Request.Context()
	if !authorized(req, rsp, CLASSIDFleet, ACTION) {
		return
	}

	fleetId, errs := rest.SanitizeValidatePathParameter(req, parameterFleetId)
	if !valid(errs, req, rsp, CLASSIDFleet, ACTION) {
		return
	}
	fleet, err := r.fleetDb.FindOne(ctx, fleetId)
	if !errorFree(err, req, rsp, CLASSIDFleet, ACTION) {
		return
	}

	searchMap := make(map[string]string)
	searchMap["fleet.id"] = fleetId
	sortFields := db.SortFields{}
	sortFields = append(sortFields, db.FieldOrder{
		Field: "vessel.name",
	})
	vessels, err := r.vesselDb.FindByMap(ctx, searchMap, sortFields)
	if !errorFree(err, req, rsp, CLASSIDFleet, ACTION) {
		return
	}

	fleet.Vessels = vessels
	auditFleet(ctx, fleet, ACTION, security.SUCCESS, nil)
	rest.WriteProtoSafely(rsp, fleet)
}

func (r *FleetRest) Create(request *restful.Request, response *restful.Response) {
	ACTION := moc.Fleet_CREATE.String()
	ctxt := request.Request.Context()
	if authorized(request, response, CLASSIDFleet, ACTION) {
		fleet := new(moc.Fleet)
		errs := rest.SanitizeValidateReadProto(request, SchemaFleetCreate, fleet)
		if valid(errs, request, response, CLASSIDFleet, ACTION) {
			err := r.createVessels(ctxt, fleet.Vessels)
			if err == nil {
				fleet, err = r.fleetDb.Create(ctxt, fleet)
			}
			if err == nil {
				r.Publish(ACTION, fleet)
				err = r.updateVesselsWithFleet(ctxt, fleet.Vessels, fleet)
			}
			if errorFree(err, request, response, CLASSIDFleet, ACTION) {
				auditFleet(ctxt, fleet, ACTION, security.SUCCESS, nil)
				rest.WriteHeaderAndProtoSafely(response, http.StatusCreated, fleet)
			}
		}
	}
}

func (r *FleetRest) createVessels(ctxt context.Context, vessels []*moc.Vessel) error {
	var err error
	for _, vessel := range vessels {
		if vessel.Id == "" {
			vessel, err = r.vesselDb.Create(ctxt, vessel)
		}
		vessel.Fleet = nil
	}
	return err
}

func (r *FleetRest) updateVesselsWithFleet(ctxt context.Context, vessels []*moc.Vessel, fleet *moc.Fleet) error {
	var err error
	vesselFleet := new(moc.Fleet)
	if fleet != nil {
		vesselFleet = &moc.Fleet{
			Id:   fleet.Id,
			Name: fleet.Name,
		}
	}

	for _, vessel := range vessels {
		vessel.Fleet = &moc.FleetCommonInfo{
			Id: vesselFleet.Id,
			Name: vesselFleet.Name,
		}
		if vessel.Id == "" {
			vessel, err = r.vesselDb.Create(ctxt, vessel)
		} else {
			vessel, err = r.vesselDb.Update(ctxt, vessel)
		}
	}
	return err
}

func (r *FleetRest) Update(request *restful.Request, response *restful.Response) {
	ACTION := moc.Fleet_UPDATE.String()
	ctxt := request.Request.Context()
	if authorized(request, response, CLASSIDFleet, ACTION) {
		fleet := new(moc.Fleet)
		fleetId, errs := rest.SanitizeValidatePathParameter(request, parameterFleetId)
		errs2 := rest.SanitizeValidateReadProto(request, SchemaFleetUpdate, fleet)
		errs = append(errs, errs2...)
		errs3 := validateIdEqual(fleetId, fleet.Id)
		errs = append(errs, errs3...)
		if valid(errs, request, response, CLASSIDFleet, ACTION) {
			err := r.createVessels(ctxt, fleet.Vessels)
			if err == nil {
				fleet, err = r.fleetDb.Update(ctxt, fleet)
			}
			if err == nil {
				r.Publish(ACTION, fleet)
				err = r.updateVesselsWithFleet(ctxt, fleet.Vessels, fleet)
			}
			if errorFree(err, request, response, CLASSIDFleet, ACTION) {
				auditFleet(ctxt, fleet, ACTION, security.SUCCESS, nil)
				rest.WriteProtoSafely(response, fleet)
			}
		}
	}
}

func (r *FleetRest) Delete(request *restful.Request, response *restful.Response) {
	ACTION := moc.Fleet_DELETE.String()
	ctx := request.Request.Context()
	if !authorized(request, response, CLASSIDFleet, ACTION) {
		return
	}

	fleetId, errs := rest.SanitizeValidatePathParameter(request, parameterFleetId)
	if !valid(errs, request, response, CLASSIDFleet, ACTION) {
		return
	}

	fleet, err := r.fleetDb.FindOne(ctx, fleetId)
	if err == nil {
		err = r.updateVesselsWithFleet(ctx, fleet.Vessels, nil)
	}

	err = r.fleetDb.Delete(ctx, fleetId)
	if !errorFree(err, request, response, CLASSIDFleet, ACTION) {
		return
	}

	r.Publish(ACTION, fleet)
	auditFleet(ctx, fleet, ACTION, security.SUCCESS, nil)
	rest.WriteHeaderAndProtoSafely(response, http.StatusNoContent, nil)
}

func (r *FleetRest) UpdateRemoveVessel(request *restful.Request, response *restful.Response) {
	ACTION := moc.Fleet_REMOVE_VESSEL.String()
	ctxt := request.Request.Context()
	if authorized(request, response, CLASSIDFleet, ACTION) {
		fleetId, errs := rest.SanitizeValidatePathParameter(request, parameterFleetId)
		vesselId, errs2 := rest.SanitizeValidatePathParameter(request, parameterVesselId)
		errs = append(errs, errs2...)
		if valid(errs, request, response, CLASSIDFleet, ACTION) {
			var fleet *moc.Fleet
			var err error
			vessel, err := r.vesselDb.FindOne(ctxt, vesselId)
			if err == nil {
				err = r.fleetDb.RemoveVessel(ctxt, fleetId, vesselId)
				if err == nil {
					vessel.Fleet = nil
					r.vesselDb.Update(ctxt, vessel)
					fleet, err = r.fleetDb.FindOne(ctxt, fleetId)
				}
			}
			if errorFree(err, request, response, CLASSIDFleet, ACTION) {
				r.Publish(ACTION, fleet)
				auditFleet(ctxt, fleet, ACTION, security.SUCCESS, vessel)
				rest.WriteProtoSafely(response, fleet)
			}
		}
	}
}

func (r *FleetRest) UpdateAddVessel(request *restful.Request, response *restful.Response) {
	ACTION := moc.Fleet_ADD_VESSEL.String()
	ctxt := request.Request.Context()
	if !authorized(request, response, CLASSIDFleet, ACTION) {
		return
	}
	fleetId, errs := rest.SanitizeValidatePathParameter(request, parameterFleetId)
	vesselId, errs2 := rest.SanitizeValidatePathParameter(request, parameterVesselId)
	errs = append(errs, errs2...)
	if valid(errs, request, response, CLASSIDFleet, ACTION) {
		var fleet *moc.Fleet
		var err error
		vessel, err := r.vesselDb.FindOne(ctxt, vesselId)
		if err == nil {
			fleet, err = r.fleetDb.FindOne(ctxt, fleetId)
			if err == nil {
				vessel.Fleet = &moc.FleetCommonInfo{
					Id:   fleet.Id,
					Name: fleet.Name,
				}
				err = r.fleetDb.AddVessel(ctxt, fleetId, vessel)
				if err == nil {
					r.vesselDb.Update(ctxt, vessel)
					fleet, err = r.fleetDb.FindOne(ctxt, fleetId)
				}
			}
		}
		if errorFree(err, request, response, CLASSIDFleet, ACTION) {
			r.Publish(ACTION, fleet)
			auditFleet(ctxt, fleet, ACTION, security.SUCCESS, vessel)
			rest.WriteProtoSafely(response, fleet)
		}
	}
}

func auditFleet(ctxt context.Context, fleet *moc.Fleet, action string, outcome string, payload interface{}) {
	if payload == nil {
		security.AuditUserObject(ctxt, CLASSIDFleet, fleet.Id, "", action, outcome)
		return
	}
	vessel, ok := payload.(*moc.Vessel)
	if ok {
		encodedPayload, err := json.Marshal(map[string]string{
			"vesselName": vessel.Name,
			"vesselId": vessel.Id,
			"fleetName": fleet.Name,
		})
		if err != nil {
			security.AuditUserObject(ctxt, CLASSIDFleet, fleet.Id, "", action, outcome, vessel.Name, vessel.Id)
		} else {
			security.AuditUserObject(ctxt, CLASSIDFleet, fleet.Id, "", action, outcome, string(encodedPayload))
		}
		return
	}
	payloads, ok := payload.([]interface{})
	if ok {
		security.AuditUserObject(ctxt, CLASSIDFleet, fleet.Id, "", action, outcome, payloads...)
		return
	}
	security.AuditUserObject(ctxt, CLASSIDFleet, fleet.Id, "", action, outcome, payload)
}

func toMessagesFromFleets(fleets []*moc.Fleet) []proto.Message {
	var messages []proto.Message
	for _, vessel := range fleets {
		messages = append(messages, vessel)
	}
	return messages
}

func (r *FleetRest) Publish(action string, fleet *moc.Fleet) {
	if r.publisher == nil {
		return
	}
	envelope := envelope.Envelope{
		Type:     TOPIC_FLEET + "/" + action,
		Contents: &envelope.Envelope_Fleet{Fleet: fleet},
	}
	r.publisher.Publish(TOPIC_FLEET, envelope)
}
