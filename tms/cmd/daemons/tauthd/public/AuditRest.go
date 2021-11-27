package public

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"prisma/gogroup"
	"prisma/tms"
	twebd "prisma/tms/cmd/daemons/twebd/public"
	"prisma/tms/db"
	"prisma/tms/db/mongo"
	"prisma/tms/devices"
	"prisma/tms/log"
	"prisma/tms/moc"
	"prisma/tms/rest"
	"prisma/tms/security"
	"prisma/tms/security/database/audit"
	"prisma/tms/util/ais"
	"strconv"

	"github.com/globalsign/mgo/bson"
	"github.com/go-openapi/spec"
	restful "github.com/orolia/go-restful"
)

var (
	PARAMETER_SESSION_ID = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "session-id",
			In:       "path",
			Required: true,
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					MinLength: &[]int64{36}[0],
					MaxLength: &[]int64{36}[0],
					Pattern:   "[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}",
				},
			},
		},
		SimpleSchema: spec.SimpleSchema{
			Type:   "string",
			Format: "RFC4122",
		},
	}
	PARAMETER_AUDIT_QUERY = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "query",
			In:       "query",
			Required: false,
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					MaxLength: &[]int64{256}[0],
				},
			},
		},
	}
)

type AuditRest struct {
	client     *mongo.MongoClient
	group      gogroup.GoGroup
	trackDb    db.TrackDB
	notifyDb   db.NotifyDb
	incidentDb db.IncidentDB
	fleetDb    db.FleetDB
	vesselDb   db.VesselDB
}

func NewAuditRest(group gogroup.GoGroup, client *mongo.MongoClient) *AuditRest {
	miscDb := mongo.NewMongoMiscData(group, client)

	return &AuditRest{
		client:     client,
		group:      group,
		trackDb:    mongo.NewMongoTracks(group, client),
		notifyDb:   mongo.NewNotifyDb(group, client),
		incidentDb: mongo.NewMongoIncidentMiscData(miscDb),
		fleetDb:    mongo.NewFleetDb(group),
		vesselDb:   mongo.NewMongoVesselDb(group),
	}
}

func (auditRest *AuditRest) Get(request *restful.Request, response *restful.Response) {
	const CLASSID = "Audit"
	const ACTION = "READ"
	ctxt := request.Request.Context()
	allowed := security.HasPermissionForAction(ctxt, CLASSID, ACTION)
	if !allowed {
		security.Audit(ctxt, CLASSID, ACTION, "FAIL")
		response.WriteErrorString(http.StatusForbidden, "")
	} else {
		records := make([]audit.Record, 0)
		var err error

		// search map from request query parameters
		searchQuery, errs := rest.SanitizeValidateQueryParameter(request, PARAMETER_AUDIT_QUERY)
		if errs != nil {
			security.Audit(ctxt, CLASSID, ACTION, "FAIL")
			response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
			return
		}

		searchMap := make(map[string]string)
		if request.QueryParameter("classId") != "" {
			searchMap["classId"] = request.QueryParameter("classId")
		}
		if request.QueryParameter("objectId") != "" {
			searchMap["objectId"] = request.QueryParameter("objectId")
		}
		if request.QueryParameter("userId") != "" {
			searchMap["userId"] = request.QueryParameter("userId")
		}
		if request.QueryParameter("action") != "" {
			searchMap["action"] = request.QueryParameter("action")
		}

		if request.QueryParameter("limit") != "" {
			limit, err := strconv.Atoi(request.QueryParameter("limit"))
			if err == nil {
				timeQuery := audit.TimeQuery{
					Limit:          limit,
					BeforeRecordId: request.QueryParameter("before"),
					AfterRecordId:  request.QueryParameter("after"),
				}
				records, err = audit.NewAuditor(ctxt).GetRecordsByMapByTimeQuery(ctxt, searchMap, timeQuery, searchQuery)
				// RFC 5988 Web Linking
				if len(records) > 0 {
					linkParam := ""

					if request.QueryParameter("classId") != "" {
						linkParam += fmt.Sprintf("&classId=%s", request.QueryParameter("classId"))
					}
					if request.QueryParameter("objectId") != "" {
						linkParam += fmt.Sprintf("&objectId=%s", request.QueryParameter("objectId"))
					}
					if request.QueryParameter("userId") != "" {
						linkParam += fmt.Sprintf("&userId=%s", request.QueryParameter("userId"))
					}
					if request.QueryParameter("action") != "" {
						linkParam += fmt.Sprintf("&action=%s", request.QueryParameter("action"))
					}

					linkValue := fmt.Sprintf("<%s?limit=%d&before=%s%s>; rel=\"previous\"", request.SelectedRoutePath(), limit, records[len(records)-1].MongoId.Hex(), linkParam)
					linkValue += fmt.Sprintf(",<%s?limit=%d&after=%s%s>; rel=\"next\"", request.SelectedRoutePath(), limit, records[0].MongoId.Hex(), linkParam)
					response.AddHeader("Link", linkValue)
				}
			}
		} else if len(searchMap) > 0 {
			records, err = audit.NewAuditor(ctxt).GetRecordsByMap(ctxt, searchMap, searchQuery)
		} else {
			records, err = audit.NewAuditor(ctxt).GetRecords(ctxt, searchQuery)
		}

		if err != nil {
			security.Audit(ctxt, CLASSID, ACTION, "FAIL_ERROR")
			response.WriteError(http.StatusInternalServerError, err)
		} else {
			records, err = auditRest.getDescription(ctxt, records)
			if err != nil {
				security.Audit(ctxt, CLASSID, ACTION, "FAIL_ERROR")
				response.WriteError(http.StatusInternalServerError, err)
				return
			}

			security.Audit(ctxt, CLASSID, ACTION, "SUCCESS")
			response.WriteEntity(records)
		}
	}
}

func (auditRest *AuditRest) getDescription(ctxt context.Context, records []audit.Record) ([]audit.Record, error) {
	var err error

	// respond with the descriptions
	for idx := range records {
		description := ""
		curRecord := records[idx]
		objectId := curRecord.ObjectId
		action := curRecord.Action
		classId := curRecord.ClassId
		payload := curRecord.Payload
		switch action {
		// actions may be duplicated
		case moc.Incident_CREATE.String(),
			moc.Incident_UPDATE.String(),
			moc.Incident_ASSIGN.String(),
			moc.Incident_UNASSIGN.String(),
			moc.Incident_CLOSE.String(),
			moc.Incident_ADD_NOTE.String(),
			moc.Incident_ADD_NOTE_FILE.String(),
			moc.Incident_ADD_NOTE_ENTITY.String(),
			moc.Incident_DELETE_NOTE.String(),
			moc.Incident_DETACH_NOTE.String(),
			moc.Fleet_READ.String(),
			moc.Fleet_CREATE.String(),
			moc.Fleet_UPDATE.String(),
			moc.Fleet_DELETE.String(),
			moc.Fleet_ADD_VESSEL.String(),
			moc.Fleet_REMOVE_VESSEL.String(),
			moc.Vessel_READ.String(),
			moc.Vessel_CREATE.String(),
			moc.Vessel_UPDATE.String(),
			moc.Vessel_DELETE.String():
			// if incident, return its IncidentId as description
			if classId == twebd.CLASSIDIncident {
				description, err = auditRest.getInfoFromIncident(objectId, action, payload)
			}
			// if fleet, return its Name as description
			if classId == twebd.CLASSIDFleet {
				description, err = auditRest.getInfoFromFleet(ctxt, objectId, action, payload)
			}
			// if vessel, return its Name as description
			if classId == twebd.CLASSIDVessel {
				description, err = auditRest.getInfoFromVessel(ctxt, objectId)
			}
		case moc.Notice_ACK.String():
			switch classId {
			case moc.Notice_Rule.String(),
				moc.Notice_Sart.String(),
				moc.Notice_Sarsat.String(),
				moc.Notice_OmnicomAssistance.String(),
				moc.Notice_SarsatDefaultHandling.String():
				description, err = auditRest.getInfoFromTrack(objectId)
			case moc.Notice_EnterZone.String(),
				moc.Notice_ExitZone.String(),
				moc.Notice_SarsatMessage.String(),
				moc.Notice_IncidentTransfer.String():
				description, err = auditRest.getInfoFromNotice(objectId, classId)
			}
		}

		log.Debug("description: %s", description)
		records[idx].Description = description
	}

	return records, err
}

// Helpers to get describable information
func (auditRest *AuditRest) getInfoFromTrack(trackId string) (string, error) {
	description := ""

	track, err := auditRest.trackDb.GetLastTrack(bson.M{"track_id": trackId})
	if err != nil {
		return description, err
	}

	if len(track.Targets) == 0 {
		return description, err
	}
	tgt := track.Targets[0]

	var md *tms.TrackMetadata
	if len(track.Metadata) != 0 {
		md = track.Metadata[0]
	}

	info := &moc.TargetInfo{
		TrackId:    track.Id,
		DatabaseId: track.DatabaseId,
		RegistryId: track.RegistryId,
		Type:       tgt.Type.String(),
	}
	if md != nil {
		info.Name = md.Name
	}

	switch tgt.Type {
	case devices.DeviceType_Radar:
		info.RadarTarget = tgt.Nmea.Ttm.Number
	case devices.DeviceType_AIS, devices.DeviceType_TV32,
		devices.DeviceType_Orb, devices.DeviceType_SART:
		info.Mmsi = ais.FormatMMSI(int(tgt.Nmea.Vdm.M1371.Mmsi))
	case devices.DeviceType_OmnicomSolar:
		if tgt.Imei != nil {
			info.Imei = tgt.Imei.Value
		}
		if tgt.Nodeid != nil {
			info.IngenuNodeId = tgt.Nodeid.Value
		}
	case devices.DeviceType_SARSAT:
		if tgt.Sarmsg.SarsatAlert != nil {
			info.SarsatBeacon = tgt.Sarmsg.SarsatAlert.Beacon
		}
	}

	if info.Imei != "" {
		description = "IMEI: " + info.Imei
	} else if info.Mmsi != "" {
		description = "MMSI: " + info.Mmsi
	} else if info.IngenuNodeId != "" {
		description = "NODE ID: " + info.IngenuNodeId
	} else if info.SarsatBeacon != nil && info.SarsatBeacon.HexId != "" {
		description = "HEX ID: " + info.SarsatBeacon.HexId
	} else if info.Name != "" {
		description = "NAME: " + info.Name
	}

	return description, nil
}

func (auditRest *AuditRest) getInfoFromNotice(id string, classId string) (string, error) {
	label := ""
	description := ""
	notice, err := auditRest.notifyDb.GetById(id)
	if err != nil {
		return description, err
	}

	if notice.Source.Name != "" {
		description = notice.Source.Name
	} else if notice.Source.IncidentId != "" {
		description = notice.Source.IncidentId
	}

	switch classId {
	case moc.Notice_EnterZone.String(),
		moc.Notice_ExitZone.String():
		label = "NAME: "
	case moc.Notice_SarsatMessage.String():
		label = "REMOTE NAME: "
	case moc.Notice_IncidentTransfer.String():
		label = "INCIDENT ID: "
	}

	description = label + description

	return description, nil
}

func (auditRest *AuditRest) getInfoFromIncident(id string, action string, payload string) (string, error) {
	description := ""

	var incident *moc.Incident
	incidentData, err := auditRest.incidentDb.Get(db.GoMiscRequest{
		Req: &db.GoRequest{
			Obj: &db.GoObject{
				ID: id,
			},
			ObjectType: twebd.OBJECT_INCIDENT,
		},
		Ctxt: auditRest.group,
		Time: &db.TimeKeeper{},
	})
	if err == nil {
		incidents := make([]*moc.Incident, 0)
		for _, incidentDatum := range incidentData {
			if mocIncident, ok := incidentDatum.Contents.Data.(*moc.Incident); ok {
				mocIncident.Id = incidentDatum.Contents.ID
				incidents = append(incidents, mocIncident)
			}
		}
		if len(incidents) > 0 {
			incident = incidents[0]
		} else {
			err = db.ErrorNotFound
		}
	}

	if err != nil {
		return description, err
	}

	description = "INCIDENT ID: " + incident.IncidentId

	switch action {
	case moc.Incident_ADD_NOTE.String(),
		moc.Incident_ADD_NOTE_FILE.String(),
		moc.Incident_ADD_NOTE_ENTITY.String(),
		moc.Incident_DELETE_NOTE.String(),
		moc.Incident_DETACH_NOTE.String():
		if payload != "" {
			log.Debug("raw incident payload: %v", payload)

			jsonMap := make(map[string]interface{})
			err = json.Unmarshal([]byte(payload), &jsonMap)
			if err != nil {
				log.Error("incident paylaod parse error: %v", err)
				return description, err
			}

			log.Debug("parsed incident payload: %v", jsonMap)

			if entity, ok := jsonMap["entity"]; ok {
				entity := entity.(map[string]interface{})

				if entityType, ok := entity["type"]; ok {
					entityType := entityType.(string)

					if entityType == "registry" || entityType == "track" {
						entityType = "Track"
					} else if entityType == "marker" {
						entityType = "Marker"
					}

					description += ", NOTE TYPE: " + entityType
				}
			} else if target, ok := jsonMap["target"]; ok {
				if target == true {
					description += ", NOTE TYPE: Search Object"
				}
			}
		}
	}

	return description, nil
}

func (auditRest *AuditRest) getInfoFromFleet(ctxt context.Context, id string, action string, payload string) (string, error) {
	description := ""

	fleet, err := auditRest.fleetDb.FindOne(ctxt, id)
	if err != nil {
		return description, err
	}

	description = "FLEET NAME: " + fleet.Name

	if (action == moc.Fleet_ADD_VESSEL.String() || action == moc.Fleet_REMOVE_VESSEL.String()) && payload != "" {
		log.Debug("raw fleet payload: %v", payload)

		jsonMap := make(map[string]string)
		err = json.Unmarshal([]byte(payload), &jsonMap)
		if err != nil {
			log.Error("fleet payload parse error: %v", err)
			return description, err
		}

		log.Debug("parsed fleet payload: %v", jsonMap)

		if vesselName, ok := jsonMap["vesselName"]; ok {
			description += ", VESSEL NAME: " + vesselName
		}
	}

	return description, nil
}

func (auditRest *AuditRest) getInfoFromVessel(ctxt context.Context, id string) (string, error) {
	description := ""
	vessel, err := auditRest.vesselDb.FindOne(ctxt, id)
	if err != nil {
		return description, err
	}

	description = "VESSEL NAME: " + vessel.Name
	return description, nil
}

func (auditRest *AuditRest) GetIncident(request *restful.Request, response *restful.Response) {
	const CLASSID = "Incident"
	const ACTION = "READ"
	ctxt := request.Request.Context()

	allowed := security.HasPermissionForAction(ctxt, CLASSID, ACTION)
	if !allowed {
		security.Audit(ctxt, CLASSID, ACTION, "FAIL")
		response.WriteErrorString(http.StatusForbidden, "")
	} else {
		records := make([]audit.Record, 0)
		var err error
		// search map from request query parameters
		searchQuery, errs := rest.SanitizeValidateQueryParameter(request, PARAMETER_AUDIT_QUERY)
		if errs != nil {
			security.Audit(ctxt, CLASSID, ACTION, "FAIL")
			response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
			return
		}

		searchMap := make(map[string]string)
		searchMap["classId"] = CLASSID
		if request.QueryParameter("objectId") != "" {
			searchMap["objectId"] = request.QueryParameter("objectId")
		}
		if request.QueryParameter("userId") != "" {
			searchMap["userId"] = request.QueryParameter("userId")
		}
		if request.QueryParameter("action") != "" {
			searchMap["action"] = request.QueryParameter("action")
		}

		if request.QueryParameter("limit") != "" {
			limit, err := strconv.Atoi(request.QueryParameter("limit"))
			if err == nil {
				timeQuery := audit.TimeQuery{
					Limit:          limit,
					BeforeRecordId: request.QueryParameter("before"),
					AfterRecordId:  request.QueryParameter("after"),
				}
				records, err = audit.NewAuditor(ctxt).GetRecordsByMapByTimeQuery(ctxt, searchMap, timeQuery, searchQuery)
				// RFC 5988 Web Linking
				if len(records) > 0 {
					linkParam := ""

					if request.QueryParameter("objectId") != "" {
						linkParam += fmt.Sprintf("&objectId=%s", request.QueryParameter("objectId"))
					}
					if request.QueryParameter("userId") != "" {
						linkParam += fmt.Sprintf("&userId=%s", request.QueryParameter("userId"))
					}
					if request.QueryParameter("action") != "" {
						linkParam += fmt.Sprintf("&action=%s", request.QueryParameter("action"))
					}

					linkValue := fmt.Sprintf("<%s?limit=%d&before=%s%s>; rel=\"previous\"", request.SelectedRoutePath(), limit, records[len(records)-1].MongoId.Hex(), linkParam)
					linkValue += fmt.Sprintf(",<%s?limit=%d&after=%s%s>; rel=\"next\"", request.SelectedRoutePath(), limit, records[0].MongoId.Hex(), linkParam)
					response.AddHeader("Link", linkValue)
				}
			}
		} else if len(searchMap) > 0 {
			records, err = audit.NewAuditor(ctxt).GetRecordsByMap(ctxt, searchMap, searchQuery)
		} else {
			records, err = audit.NewAuditor(ctxt).GetRecords(ctxt, searchQuery)
		}

		if err != nil {
			security.Audit(ctxt, CLASSID, ACTION, "FAIL_ERROR")
			response.WriteError(http.StatusInternalServerError, err)
		} else {
			security.Audit(ctxt, CLASSID, ACTION, "SUCCESS")
			response.WriteEntity(records)
		}
	}
}

func (auditRest *AuditRest) GetBySessionId(request *restful.Request, response *restful.Response) {
	const CLASSID = "Audit"
	const ACTION = "READ"
	allowed := security.HasPermissionForAction(request.Request.Context(), CLASSID, ACTION)
	if !allowed {
		security.Audit(request.Request.Context(), CLASSID, ACTION, "FAIL")
		response.WriteErrorString(http.StatusForbidden, "")
		return
	}
	sessionID, errs := rest.SanitizeValidatePathParameter(request, PARAMETER_SESSION_ID)
	if errs != nil {
		security.Audit(request.Request.Context(), CLASSID, ACTION, "FAIL_ERROR", sessionID)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		return
	}
	records, err := audit.NewAuditor(request.Request.Context()).GetRecordsBySessionId(request.Request.Context(), sessionID)
	if err != nil {
		security.Audit(request.Request.Context(), CLASSID, ACTION, "SUCCESS_ERROR")
		response.WriteError(http.StatusInternalServerError, err)
	} else {
		if len(records) > 0 {
			security.Audit(request.Request.Context(), CLASSID, ACTION, "SUCCESS")
			response.WriteEntity(records)
		} else {
			security.Audit(request.Request.Context(), CLASSID, ACTION, "RECORDS_EMPTY")
			response.WriteHeader(http.StatusNoContent)
		}
	}
}

func (auditRest *AuditRest) GetByUserId(request *restful.Request, response *restful.Response) {
	const CLASSID = "Audit"
	const ACTION = "READ"
	allowed := security.HasPermissionForAction(request.Request.Context(), CLASSID, ACTION)
	if !allowed {
		security.Audit(request.Request.Context(), CLASSID, ACTION, "FAIL")
		response.WriteErrorString(http.StatusForbidden, "")
	} else {
		userId := request.PathParameter("user-id")
		records, err := audit.NewAuditor(request.Request.Context()).GetRecordsByUserId(request.Request.Context(), userId)
		if err != nil {
			security.Audit(request.Request.Context(), CLASSID, ACTION, "FAIL_ERROR")
			response.WriteError(http.StatusInternalServerError, err)
		} else {
			security.Audit(request.Request.Context(), CLASSID, ACTION, "SUCCESS")
			response.WriteEntity(records)
		}
	}
}
