package public

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"prisma/gogroup"
	"prisma/tms"
	"prisma/tms/db"
	"prisma/tms/db/mongo"
	"prisma/tms/envelope"
	"prisma/tms/log"
	"prisma/tms/moc"
	"prisma/tms/multicast"
	"prisma/tms/rest"
	"prisma/tms/security"
	"prisma/tms/security/service"
	"prisma/tms/tmsg"
	"prisma/tms/tmsg/client"
	"prisma/tms/ws"

	restful "github.com/orolia/go-restful"
	restfulspec "github.com/orolia/go-restful-openapi"
	"github.com/go-openapi/spec"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
)

const CLASSIDMulticast = security.CLASSIDMulticast

var (
	parameterSiteId = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "id",
			In:       "path",
			Required: true,
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					MinLength: &[]int64{0}[0],
					MaxLength: &[]int64{24}[0],
					Pattern:   "[0-9a-fA-F]{1,24}",
				},
			},
		},
		SimpleSchema: spec.SimpleSchema{
			Type: "string",
			//Format: "hexadecimal",
		},
	}
	parameterMulticastId = spec.Parameter{
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
	schemaMulticastCreate = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Required:   []string{"payload"},
			Properties: map[string]spec.Schema{},
		},
	}
	queryCompleted = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name: "completed",
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
)

// MulticastRest is a structure to determine rest api for multicast endpoints
type MulticastRest struct {
	group     gogroup.GoGroup
	clt       *mongo.MongoClient
	tsiClient client.TsiClient
	miscDb    db.MiscDB
	siteDb    db.SiteDB
	transDb   db.TransmissionDB
	deviceDb  db.DeviceDB
	mcDb      *mongo.MulticastDb
	publisher *ws.Publisher
	routes    []string
}

// RegisterMulticastRest ...
func RegisterMulticastRest(ctx gogroup.GoGroup, mg *mongo.MongoClient, service *restful.WebService, routeAuthorizer *service.RouteAuthorizer, publisher *ws.Publisher, tsiClient client.TsiClient) {
	r := &MulticastRest{
		group:     ctx,
		clt:       mg,
		tsiClient: tsiClient,
		miscDb:    mongo.NewMongoMiscData(ctx, mg),
		siteDb:    mongo.NewSiteDb(ctx),
		transDb:   mongo.NewTransmissionDb(ctx, mg),
		mcDb:      mongo.NewMulticastDb(ctx),
		deviceDb:  mongo.NewMongoDeviceDb(ctx, mg),
		publisher: publisher,
	}
	service.Route(service.GET(r.registerRoute("/multicast/{id}")).To(r.multicastRead).
		Doc("get a multicast").
		Param(service.PathParameter("id", "multicast identifier")).
		Writes(tms.Multicast{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{"communication"}))
	service.Route(service.POST(r.registerRoute("/multicast/device/{id}")).To(r.deviceCreate).
		Doc("create a transmission to a device").
		Param(service.PathParameter("id", "device identifier")).
		ReadsWithSchema(tms.Multicast{}, schemaMulticastCreate).
		Writes(tms.Multicast{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{"communication"}))
	service.Route(service.GET(r.registerRoute("/multicast/device/{id}")).To(r.deviceRead).
		Doc("get multicasts for an device").
		Param(service.PathParameter("id", "device identifier")).
		Writes([]tms.Multicast{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{"communication"}))
	service.Route(service.POST(r.registerRoute("/multicast/site/{id}")).To(r.siteCreate).
		Doc("create a transmission to a site").
		Param(service.PathParameter("site-id", "site identifier")).
		ReadsWithSchema(tms.Multicast{}, schemaMulticastCreate).
		Writes(tms.Multicast{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{"communication"}))
	service.Route(service.GET(r.registerRoute("/multicast/incident/{incident-id}")).To(r.incidentRead).
		Doc("get multicasts for an incident").
		Param(service.PathParameter("incident-id", "incident identifier")).
		Writes([]tms.Multicast{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{"communication"}))
	routeAuthorizer.Add(r)
}

// MatchRoute RouteMatcher
func (r *MulticastRest) MatchRoute(route string) (bool, string) {
	match := false
	for _, siteRoute := range r.routes {
		match = strings.HasSuffix(route, siteRoute)
		if match {
			break
		}
	}
	return match, CLASSIDMulticast
}

func (r *MulticastRest) registerRoute(route string) string {
	if r.routes == nil {
		r.routes = make([]string, 0)
	}
	r.routes = append(r.routes, route)
	return route
}

func (r *MulticastRest) multicastRead(req *restful.Request, rsp *restful.Response) {
	action := tms.Multicast_READ.String()
	ctx := req.Request.Context()
	multicastId, errs := rest.SanitizeValidatePathParameter(req, parameterMulticastId)
	if !valid(errs, req, rsp, CLASSIDMulticast, action) {
		return
	}
	mc, err := r.mcDb.Find(ctx, multicastId)
	if !errorFree(err, req, rsp, CLASSIDMulticast, action) {
		return
	}
	security.Audit(ctx, CLASSIDMulticast, action, security.SUCCESS)
	rest.WriteHeaderAndProtoSafely(rsp, http.StatusOK, mc)
}

func (r *MulticastRest) incidentRead(req *restful.Request, rsp *restful.Response) {
	action := tms.Multicast_READ.String()
	ctx := req.Request.Context()
	var completed bool
	if !authorized(req, rsp, CLASSIDMulticast, action) {
		return
	}
	incidentId, errs := rest.SanitizeValidatePathParameter(req, PARAMETER_INCIDENT_ID)
	completedString, errs2 := rest.SanitizeValidateQueryParameter(req, queryCompleted)
	completed, err := strconv.ParseBool(completedString)
	errs = append(errs, errs2...)
	if !valid(errs, req, rsp, CLASSIDMulticast, action) {
		return
	}
	searchMap := make(map[string]string)
	searchMap["payload.typeurl"] = "type.googleapis.com/prisma.tms.moc.Incident"
	var states []tms.Transmission_State
	if !completed {
		states = append(states, tms.Transmission_Pending, tms.Transmission_Retry)
	}
	mcs, err := r.mcDb.FindByMapByState(ctx, searchMap, states)
	if !errorFree(err, req, rsp, CLASSIDMulticast, action) {
		return
	}
	var filteredMcs []*tms.Multicast
	for _, mc := range mcs {
		incident := moc.Incident{}
		if err := ptypes.UnmarshalAny(mc.Payload, &incident); err != nil {
			continue
		}
		if incident.Id != incidentId {
			continue
		}
		filteredMcs = append(filteredMcs, mc)
		// latest transmissions
		for i := 0; i < len(mc.Transmissions); i++ {
			tr, err := r.transDb.FindByID(mc.Transmissions[i].Id)
			if err != nil {
				continue
			}
			mc.Transmissions[i] = tr
		}
	}
	security.Audit(ctx, CLASSIDMulticast, action, security.SUCCESS)
	rest.WriteProtoSpliceSafely(rsp, toMessagesFromMulticasts(filteredMcs))
}

func (r *MulticastRest) siteCreate(req *restful.Request, rsp *restful.Response) {
	action := tms.Multicast_CREATE.String()
	ctx := req.Request.Context()
	if !authorized(req, rsp, CLASSIDMulticast, action) {
		return
	}
	// site
	siteID, errs := rest.SanitizeValidatePathParameter(req, parameterSiteId)
	if !valid(errs, req, rsp, CLASSIDMulticast, action) {
		return
	}
	site, err := r.siteDb.FindById(ctx, siteID)
	if !errorFree(err, req, rsp, CLASSIDMulticast, action) {
		return
	}
	// multicast
	mc := tms.Multicast{}
	errs = rest.SanitizeValidateReadProto(req, schemaMulticastCreate, &mc)
	if !valid(errs, req, rsp, CLASSIDMulticast, action) {
		return
	}
	// payload
	payload, err := tmsg.Unpack(mc.Payload)
	if !errorFree(err, req, rsp, CLASSIDMulticast, action) {
		return
	}
	var transferPayload *any.Any
	switch mc.Payload.TypeUrl {
	case "prisma.tms.moc.Incident":
		incAction := moc.Incident_TRANSFER_SEND.String()
		if !authorized(req, rsp, CLASSIDIncident, incAction) {
			return
		}
		incident, ok := payload.(*moc.Incident)
		if ok {
			// load incident
			incidentData, err := r.miscDb.Get(db.GoMiscRequest{
				Req: &db.GoRequest{
					Obj: &db.GoObject{
						ID: incident.Id,
					},
					ObjectType: OBJECT_INCIDENT,
				},
				Ctxt: r.group,
			})
			if len(incidentData) < 1 {
				err = db.ErrorNotFound
			}
			if !errorFree(err, req, rsp, CLASSIDMulticast, action) {
				return
			}
			for _, incidentDatum := range incidentData {
				if mocIncident, ok := incidentDatum.Contents.Data.(*moc.Incident); ok {
					mocIncident.Id = incidentDatum.Contents.ID
					incident = mocIncident
				}
			}
			// local site info
			localSite := moc.Site{
				SiteId: r.tsiClient.Local().Site,
			}
			err = r.siteDb.FindBySiteId(ctx, &localSite)
			if err != nil {
				log.Error(err.Error()+"%v", localSite)
			}
			// update
			incident.State = moc.Incident_Closed
			incident.Synopsis = "Incident transferred"
			incident.Outcome = "Hand Off"
			incident.Log = append(incident.Log, &moc.IncidentLogEntry{
				Id:        mongo.CreateId(),
				Timestamp: tms.Now(),
				Type:      "ACTION_" + moc.Incident_TRANSFER_SEND.String(),
				Entity: &moc.EntityRelationship{
					Type: "prisma.tms.moc.Site",
					Id:   site.Id,
				},
				Note: fmt.Sprintf("Incident %v sent to %v (%v) from %v (%v)", incident.IncidentId, site.Name, site.SiteId, localSite.Name, localSite.SiteId),
			})
			incident.Log = append(incident.Log, &moc.IncidentLogEntry{
				Id:        mongo.CreateId(),
				Timestamp: tms.Now(),
				Type:      "ACTION_" + moc.Incident_CLOSE.String(),
			})
			_, err = r.miscDb.Upsert(db.GoMiscRequest{
				Req: &db.GoRequest{
					ObjectType: OBJECT_INCIDENT,
					Obj: &db.GoObject{
						ID:   incident.Id,
						Data: incident,
					},
				},
				Ctxt: r.group,
			})
			if !errorFree(err, req, rsp, CLASSIDIncident, action) {
				return
			}
			security.AuditUserObject(ctx, CLASSIDIncident, incident.Id, "", incAction, security.SUCCESS)
			mc.Payload, err = tmsg.PackFrom(incident)
			r.publisher.Publish(TOPIC_INCIDENT, envelope.Envelope{
				Type:     TOPIC_INCIDENT + "/" + moc.Incident_CLOSE.String(),
				Contents: &envelope.Envelope_Incident{Incident: incident},
			})
			// update for transfer
			incident.State = moc.Incident_Transferring
			transferPayload, err = tmsg.PackFrom(incident)
		}
	}
	// destinations
	siteEntityRelationship := tms.EntityRelationship{
		Type: "prisma.tms.moc.Site",
		Id:   site.Id,
	}
	mc.Destinations = append(mc.Destinations, &siteEntityRelationship)
	mc.Id = mongo.CreateId()
	// transmission
	transmission := tms.Transmission{
		ParentId:    mc.Id,
		Destination: &siteEntityRelationship,
		State:       tms.Transmission_Pending,
	}
	err = r.transDb.Create(&transmission)
	if !errorFree(err, req, rsp, CLASSIDMulticast, action) {
		return
	}
	mc.Transmissions = append(mc.Transmissions, &transmission)
	security.Audit(ctx, CLASSIDMulticast, action, security.SUCCESS)
	rest.WriteHeaderAndProtoSafely(rsp, http.StatusCreated, &mc)
	// update for transfer
	mc.Payload = transferPayload
	err = r.mcDb.Create(ctx, &mc)
	if !errorFree(err, req, rsp, CLASSIDMulticast, action) {
		return
	}

}

func (r *MulticastRest) deviceRead(req *restful.Request, rsp *restful.Response) {
	action := tms.Multicast_READ.String()
	ctx := req.Request.Context()
	var completed bool
	if !authorized(req, rsp, CLASSIDMulticast, action) {
		return
	}
	deviceId, errs := rest.SanitizeValidatePathParameter(req, parameterDeviceID)
	completedString, errs2 := rest.SanitizeValidateQueryParameter(req, queryCompleted)
	completed, err := strconv.ParseBool(completedString)
	errs = append(errs, errs2...)
	if !valid(errs, req, rsp, CLASSIDMulticast, action) {
		return
	}
	searchMap := make(map[string]string)
	searchMap["destinations.id"] = deviceId
	var states []tms.Transmission_State
	if !completed {
		states = append(states, tms.Transmission_Pending, tms.Transmission_Retry)
	}
	mcs, err := r.mcDb.FindByMapByState(ctx, searchMap, states)
	if !errorFree(err, req, rsp, CLASSIDMulticast, action) {
		return
	}
	// latest transmissions
	for _, mc := range mcs {
		for i := 0; i < len(mc.Transmissions); i++ {
			tr, err := r.transDb.FindByID(mc.Transmissions[i].Id)
			if err != nil {
				continue
			}
			mc.Transmissions[i] = tr
		}
	}
	// filter states again since transmissions is not in sync
	if !completed {
		deleted := 0
		for i := range mcs {
			remove := true
			j := i - deleted
			for _, t := range mcs[j].Transmissions {
				remove = t.State == tms.Transmission_Success || t.State == tms.Transmission_Failure
				if !remove {
					break
				}
			}
			if remove {
				mcs = mcs[:j+copy(mcs[j:], mcs[j+1:])]
				deleted++
			}
		}
	}
	security.Audit(ctx, CLASSIDMulticast, action, security.SUCCESS)
	rest.WriteProtoSpliceSafely(rsp, toMessagesFromMulticasts(mcs))
}

// Post is for sending a new message to a beacon it can be different messages.
// See omnicom structures and the doc for directIP
func (r *MulticastRest) deviceCreate(req *restful.Request, rsp *restful.Response) {
	action := tms.Multicast_CREATE.String()
	ctx := req.Request.Context()
	if !authorized(req, rsp, CLASSIDMulticast, action) {
		return
	}
	// multicast
	mc := tms.Multicast{}
	errs := rest.SanitizeValidateReadProto(req, schemaMulticastCreate, &mc)
	if !valid(errs, req, rsp, CLASSIDMulticast, action) {
		return
	}
	// device
	deviceID, errs := rest.SanitizeValidatePathParameter(req, parameterDeviceID)
	if !valid(errs, req, rsp, CLASSIDMulticast, action) {
		return
	}
	device, err := r.deviceDb.FindOne(ctx, deviceID)
	if err != nil {
		// workaround since device mongo id is out-of-sync
		vesselDb := mongo.NewMongoVesselDb(ctx)
		vessel, verr := vesselDb.FindByDevice(ctx, &moc.Device{Id: deviceID})
		if vessel != nil {
			for _, vd := range vessel.Devices {
				if vd.Id == deviceID {
					device = vd
					break
				}
			}
		}
		err = verr
	}
	if !errorFree(err, req, rsp, CLASSIDMulticast, action) {
		return
	}
	// destination
	mc.Destinations = append(mc.Destinations, &tms.EntityRelationship{
		Type: device.GetType(),
		Id:   device.Id,
	})
	// iridium network (only supported)
	// TODO: needs to be changed by smart routing function or network manual selection in the future
	var net moc.Device_Network
	for _, n := range device.Networks {
		if n.ProviderId == "iridium" {
			net = *n
			break
		}
	}

	if reflect.DeepEqual(net, moc.Device_Network{}) {
		security.Audit(req.Request.Context(), CLASSIDMulticast, action, "FAIL")
		rsp.WriteError(http.StatusNotFound, fmt.Errorf("device %v has no iridium network", device.DeviceId))
	}
	if !errorFree(err, req, rsp, CLASSIDMulticast, action) {
		return
	}
	// parse payload for action, get message for tfleetd
	var bm multicast.Multicast
	switch mc.Payload.TypeUrl {
	case "type.googleapis.com/prisma.tms.moc.DeviceConfiguration":
		payload, err := tmsg.Unpack(mc.Payload)
		if !errorFree(err, req, rsp, CLASSIDMulticast, action) {
			return
		}
		dc, ok := payload.(*moc.DeviceConfiguration)
		if ok {
			dc.Id = deviceID
			// get a structure to work with the received message type
			bm, err = multicast.Parsing(dc)
			if !errorFree(err, req, rsp, CLASSIDMulticast, action) {
				return
			}
		}
	default:
		if !errorFree(fmt.Errorf("payload url type: %+v is not handled", mc.Payload.TypeUrl), req, rsp, CLASSIDMulticast, action) {
			return
		}
	}
	om, mi, pn, err := bm.GetMessage(r.miscDb)
	if !errorFree(err, req, rsp, CLASSIDMulticast, action) {
		return
	}
	// multicast.id for transmission.parentId
	mc.Id = mongo.CreateId()
	// transmission
	tran := tms.Transmission{
		ParentId:    mc.Id,
		MessageId:   mi,
		Destination: &tms.EntityRelationship{Type: net.ProviderId, Id: net.SubscriberId},
		State:       tms.Transmission_Pending,
		Packets: []*tms.Packet{
			{
				Name:  pn,
				State: tms.Transmission_Pending,
				Status: &tms.ResponseStatus{
					Code: 102,
				},
			},
		},
	}
	err = r.transDb.Create(&tran)
	if !errorFree(err, req, rsp, CLASSIDMulticast, action) {
		return
	}
	mc.Transmissions = append(mc.Transmissions, &tran)
	err = r.mcDb.Create(ctx, &mc)
	if !errorFree(err, req, rsp, CLASSIDMulticast, action) {
		return
	}
	security.Audit(ctx, CLASSIDMulticast, action, security.SUCCESS)
	rest.WriteHeaderAndProtoSafely(rsp, http.StatusCreated, &mc)
	// send tmsg for processing
	mc.Payload, err = tmsg.PackFrom(om)
	if !errorFree(err, req, rsp, CLASSIDMulticast, action) {
		return
	}
	body, err := tmsg.PackFrom(&mc)
	if !errorFree(err, req, rsp, CLASSIDMulticast, action) {
		return
	}
	// send and return an answer
	tmsg.GClient.Send(r.group, &tms.TsiMessage{
		Source: tmsg.GClient.Local(),
		Destination: []*tms.EndPoint{
			{
				Site: tmsg.TMSG_LOCAL_SITE,
			},
		},
		WriteTime: tms.Now(),
		SendTime:  tms.Now(),
		Body:      body,
	})
}

func toMessagesFromMulticasts(mcs []*tms.Multicast) []proto.Message {
	var messages []proto.Message
	for _, mc := range mcs {
		messages = append(messages, mc)
	}
	return messages
}
