package public

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"prisma/gogroup"
	"prisma/tms"
	"prisma/tms/db"
	"prisma/tms/db/mongo"
	"prisma/tms/envelope"
	"prisma/tms/incident"
	"prisma/tms/log"
	"prisma/tms/moc"
	"prisma/tms/rest"
	"prisma/tms/security"
	secDb "prisma/tms/security/database"
	"prisma/tms/security/message"
	"prisma/tms/tmsg"
	"prisma/tms/tmsg/client"
	"prisma/tms/ws"

	wkhtmltopdf "github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/globalsign/mgo/bson"
	"github.com/go-openapi/spec"
	"github.com/golang/protobuf/ptypes"
	restful "github.com/orolia/go-restful"
	"golang.org/x/net/context"
)

const (
	CLASSID          = "Incident"
	CLASSIDIncident  = CLASSID
	OBJECT_INCIDENT  = "prisma.tms.moc." + CLASSID
	TOPIC_INCIDENT   = CLASSID
	CLASSID_LogEntry = "IncidentLogEntry"
)

var (
	// errors
	ErrorInvalidId                = errors.New("invalidId")
	ErrorLocked                   = errors.New("locked")
	ErrDuplicateEntityLogIncident = errors.New("the track is already assigned")
	// enums for schema
	incidentStateEnum = make([]interface{}, len(moc.Incident_State_name))
	incidentPhaseEnum = make([]interface{}, len(moc.IncidentPhase_name))
	// parameters with schema
	PARAMETER_INCIDENT_ID = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "incident-id",
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
	PARAMETER_LOG_ID = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "log-id",
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
	PARAMETER_USER_ID = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "user-id",
			In:       "path",
			Required: true,
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					MinLength: &[]int64{1}[0],
					MaxLength: &[]int64{128}[0],
					Pattern:   "^[a-zA-Z0-9_@.]*$",
				},
			},
		},
	}
	PARAMETER_INCIDENT_STATE = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "incident-state",
			In:       "path",
			Required: true,
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Enum: incidentStateEnum,
				},
			},
		},
	}
	PARAMETER_INCIDENT_LOGENTRY_IMPORTANT = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name: "important",
			In:   "query",
		},
		SimpleSchema: spec.SimpleSchema{
			Type:    "boolean",
			Default: "true",
		},
	}
	PARAMETER_INCIDENT_DELETED = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name: "deleted",
			In:   "query",
		},
		SimpleSchema: spec.SimpleSchema{
			Type:    "boolean",
			Default: "true",
		},
	}

	// schemas
	SCHEMA_INCIDENT_CREATE = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Required: []string{"name", "type", "phase", "commander", "assignee"},
			Properties: map[string]spec.Schema{
				"name": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{2}[0],
						MaxLength: &[]int64{255}[0],
					},
				},
				"type": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{2}[0],
						MaxLength: &[]int64{255}[0],
					},
				},
				"phase": {
					SchemaProps: spec.SchemaProps{
						Enum: incidentPhaseEnum,
					},
				},
				"commander": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{2}[0],
						MaxLength: &[]int64{255}[0],
					},
				},
				"searchObject.name": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{1}[0],
						MaxLength: &[]int64{255}[0],
					},
				},
				"searchObject.note": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{1}[0],
						MaxLength: &[]int64{2000}[0],
					},
				},
				"assignee": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{1}[0],
						MaxLength: &[]int64{128}[0],
						Pattern:   "^[a-zA-Z0-9_@.]*$",
					},
				},
			},
		},
	}
	SCHEMA_INCIDENT_UPDATE = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Required: []string{"incidentId", "name", "type", "state", "phase", "assignee"},
			Properties: map[string]spec.Schema{
				"incidentId": {
					SwaggerSchemaProps: spec.SwaggerSchemaProps{
						ReadOnly: true,
					},
				},
				"name": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{2}[0],
						MaxLength: &[]int64{255}[0],
					},
				},
				"type": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{2}[0],
						MaxLength: &[]int64{255}[0],
					},
				},
				"phase": {
					SchemaProps: spec.SchemaProps{
						Enum: incidentPhaseEnum,
					},
				},
				"commander": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{1}[0],
						MaxLength: &[]int64{128}[0],
						Pattern:   "^[a-zA-Z0-9_@.]*$",
					},
				},
				"searchObject.name": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{1}[0],
						MaxLength: &[]int64{255}[0],
					},
				},
				"searchObject.note": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{1}[0],
						MaxLength: &[]int64{2000}[0],
					},
				},
				"assignee": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{1}[0],
						MaxLength: &[]int64{128}[0],
						Pattern:   "^[a-zA-Z0-9_@.]*$",
					},
				},
				"state": {
					SchemaProps: spec.SchemaProps{
						Enum: incidentStateEnum,
					},
				},
				"outcome": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{2}[0],
						MaxLength: &[]int64{255}[0],
					},
				},
				"synopsis": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{2}[0],
						MaxLength: &[]int64{2000}[0],
					},
				},
			},
		},
	}
	SCHEMA_INCIDENT_CLOSE = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Required: []string{"incidentId", "name", "type", "state", "phase", "commander", "assignee", "outcome", "synopsis"},
			Properties: map[string]spec.Schema{
				"incidentId": {
					SwaggerSchemaProps: spec.SwaggerSchemaProps{
						ReadOnly: true,
					},
				},
				"name": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{2}[0],
						MaxLength: &[]int64{255}[0],
					},
				},
				"type": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{2}[0],
						MaxLength: &[]int64{255}[0],
					},
				},
				"phase": {
					SchemaProps: spec.SchemaProps{
						Enum: incidentPhaseEnum,
					},
				},
				"commander": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{1}[0],
						MaxLength: &[]int64{128}[0],
						Pattern:   "^[a-zA-Z0-9_@.]*$",
					},
				},
				"searchObject.name": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{1}[0],
						MaxLength: &[]int64{255}[0],
					},
				},
				"searchObject.note": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{1}[0],
						MaxLength: &[]int64{2000}[0],
					},
				},
				"assignee": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{1}[0],
						MaxLength: &[]int64{128}[0],
						Pattern:   "^[a-zA-Z0-9_@.]*$",
					},
				},
				"state": {
					SchemaProps: spec.SchemaProps{
						Enum: incidentStateEnum,
					},
				},
				"outcome": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{2}[0],
						MaxLength: &[]int64{255}[0],
					},
				},
				"synopsis": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{2}[0],
						MaxLength: &[]int64{2000}[0],
					},
				},
			},
		},
	}
	SCHEMA_INCIDENT_LOGENTRY = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Required: []string{"type"},
			Properties: map[string]spec.Schema{
				"type": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{2}[0],
						MaxLength: &[]int64{255}[0],
					},
				},
				"id": {
					SwaggerSchemaProps: spec.SwaggerSchemaProps{
						ReadOnly: true,
					},
				},
				"timestamp": {
					SwaggerSchemaProps: spec.SwaggerSchemaProps{
						ReadOnly: true,
					},
				},
				"note": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{2}[0],
						MaxLength: &[]int64{2000}[0],
					},
				},
				"attachment": {
					SwaggerSchemaProps: spec.SwaggerSchemaProps{
						ReadOnly: true,
					},
				},
				"entity.type": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{2}[0],
						MaxLength: &[]int64{64}[0],
						Pattern:   "[0-9a-fA-F]",
					},
				},
				"entity.id": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{24}[0],
						MaxLength: &[]int64{34}[0],
						Pattern:   "[0-9a-fA-F]",
					},
				},
			},
		},
	}
)

type IncidentRest struct {
	client     *mongo.MongoClient
	miscDb     db.MiscDB
	noteDb     db.NoteDB
	incidentDb db.IncidentDB
	regDb      db.RegistryDB
	trackExDb  db.TrackExDb
	group      gogroup.GoGroup
	idPrefixer incident.IdPrefixer
	publisher  *ws.Publisher
	tsiClient  client.TsiClient
}

func NewIncidentRest(group gogroup.GoGroup, client *mongo.MongoClient, idPrefixer incident.IdPrefixer, publisher *ws.Publisher) *IncidentRest {
	for iState := range moc.Incident_State_name {
		incidentStateEnum[iState] = iState
	}
	for iPhase := range moc.IncidentPhase_name {
		incidentPhaseEnum[iPhase] = iPhase
	}

	miscDb := mongo.NewMongoMiscData(group, client)

	return &IncidentRest{
		client:     client,
		miscDb:     miscDb,
		noteDb:     mongo.NewMongoNoteDb(miscDb),
		incidentDb: mongo.NewMongoIncidentMiscData(miscDb),
		regDb:      mongo.NewMongoRegistry(group, client),
		trackExDb:  mongo.NewTrackExDb(group, client),
		group:      group,
		idPrefixer: idPrefixer,
		publisher:  publisher,
		tsiClient:  tmsg.GClient,
	}
}

type incidentProcessingForm struct {
	IncidentId                  string
	Phase                       string
	Name1                       string
	Name2                       string
	InitialReportingParticulars string
	Phone1                      string
	CallSign1                   string
	NatureOfEmergency           string
	AssistanceDesired           string
	CallSign2                   string
	Size                        string
	Type                        string
	SoType                      string
	Hull                        string
	Colour1                     string
	Colour2                     string
	Fuselage                    string
	Wingtip                     string
	Rigging                     string
	NavEquipment                string
	WeatherAlongRoute           string
	WeatherAtPresentLocation    string
	ForecastAlongProposedRoute  string
	NumberPersOnBoard           string
	EPIRB                       string
	RadioFrequencies            string
	SurvivalEquipment           string
	FoodWaterDuration           string
	OwnerName                   string
	Phone2                      string
	Address                     string
	FuelOnBoard                 string
	CurrentPosition             string
	HowDetermined               string
	Time1                       string
	Time2                       string
	DepartedFrom                string
	EnRouteTo                   string
	ETA                         string
	RouteDescription            string
	PossibleRouteDeviations     string
	CraftHistory                string
	Name3                       string
	Phone3                      string
	Address3                    string
}

const incidentProcessingFormFile = "/etc/trident/incident-processing-form.html" //"/vagrant/vagrant/incident-processing-form.html" //

func (incidentRest *IncidentRest) CreateIncidentProcessingForm(req *restful.Request, rsp *restful.Response) {
	ACTION := moc.Incident_READ.String()
	ctx := req.Request.Context()
	if !authorized(req, rsp, CLASSID, ACTION) {
		return
	}
	incidentId, errs := rest.SanitizeValidatePathParameter(req, PARAMETER_INCIDENT_ID)
	if !valid(errs, req, rsp, CLASSID, ACTION) {
		return
	}
	incident, err := getIncident(incidentRest, incidentId)
	if !errorFree(err, req, rsp, CLASSID, ACTION) {
		return
	}
	var templates, _ = template.ParseFiles(incidentProcessingFormFile)
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if !errorFree(err, req, rsp, CLASSID, ACTION) {
		return
	}

	d := &incidentProcessingForm{
		IncidentId: incident.IncidentId,
		Name1:      incident.Name,
		Phase:      incident.Phase.String(),
		Type:       incident.Type,
	}

	w := &bytes.Buffer{}
	templates.ExecuteTemplate(w, "incident-processing-form.html", d)

	page := wkhtmltopdf.NewPageReader(w)
	page.EnableForms.Set(true)
	pdfg.PageSize.Set(wkhtmltopdf.PageSizeA4)
	pdfg.Orientation.Set(wkhtmltopdf.OrientationLandscape)
	pdfg.AddPage(page)
	pdfg.Dpi.Set(150)

	err = pdfg.Create()
	if !errorFree(err, req, rsp, CLASSID, ACTION) {
		return
	}

	b := pdfg.Bytes()
	security.Audit(ctx, CLASSID, ACTION, security.SUCCESS)
	rsp.Write(b)
}

func (incidentRest *IncidentRest) ReadAll(request *restful.Request, response *restful.Response) {
	ACTION := moc.Incident_READ.String()
	ctxt := request.Request.Context()
	if security.HasPermissionForAction(ctxt, CLASSID, ACTION) {
		// TODO add searchMap functionality like UserRoleRest.GetRoleUser
		incidentData, err := incidentRest.miscDb.Get(db.GoMiscRequest{
			Req: &db.GoRequest{
				ObjectType: OBJECT_INCIDENT,
			},
			Ctxt: incidentRest.group,
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
			// filter response
			for _, incident := range incidents {
				filterResponseLogEntry(incident)
			}
			security.Audit(ctxt, CLASSID, ACTION, security.SUCCESS)
			rest.WriteEntitySafely(response, incidents)
		} else {
			log.Error("unexpected error: %v", err)
			security.Audit(ctxt, CLASSID, ACTION, security.FAIL_ERROR)
			response.WriteError(http.StatusInternalServerError, err)
		}
	} else {
		security.Audit(ctxt, CLASSID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}

func (incidentRest *IncidentRest) ReadOne(request *restful.Request, response *restful.Response) {
	ACTION := moc.Incident_READ.String()
	ctxt := request.Request.Context()
	if security.HasPermissionForAction(ctxt, CLASSID, ACTION) {
		incidentId, errs := rest.SanitizeValidatePathParameter(request, PARAMETER_INCIDENT_ID)
		if errs == nil {
			withDeleted, errs := rest.SanitizeValidateQueryParameter(request, PARAMETER_INCIDENT_DELETED)
			if errs == nil {
				incident, err := getIncident(incidentRest, incidentId)
				if err == nil {
					if !(withDeleted == "true") {
						filterResponseLogEntry(incident)
					}
					security.Audit(ctxt, CLASSID, ACTION, security.SUCCESS)
					rest.WriteEntitySafely(response, incident)
				} else {
					if db.ErrorNotFound == err {
						security.Audit(ctxt, CLASSID, ACTION, security.FAIL_NOTFOUND)
						response.WriteError(http.StatusNotFound, err)
					} else {
						log.Error("unexpected error: %v", err)
						security.Audit(ctxt, CLASSID, ACTION, security.FAIL_ERROR)
						response.WriteError(http.StatusInternalServerError, err)
					}
				}
			} else {
				security.Audit(ctxt, CLASSID, ACTION, security.FAIL_VALIDATION, incidentId)
				response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
			}
		} else {
			security.Audit(ctxt, CLASSID, ACTION, security.FAIL_VALIDATION, incidentId)
			response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		}
	} else {
		security.Audit(ctxt, CLASSID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}

func (incidentRest *IncidentRest) Create(request *restful.Request, response *restful.Response) {
	ACTION := moc.Incident_CREATE.String()
	ctxt := request.Request.Context()
	if security.HasPermissionForAction(ctxt, CLASSID, ACTION) {
		incidentRequest := new(moc.Incident)
		errs := rest.SanitizeValidateReadEntity(request, SCHEMA_INCIDENT_CREATE, incidentRequest)
		if errs == nil {
			incidentRequest.IncidentId = incident.IdCreatorInstance(ctxt).Next(incidentRest.idPrefixer)
			// initial state
			incidentRequest.State = moc.Incident_Open
			// validate User Role
			errs = validateUserRole(ctxt, incidentRequest)
			if errs == nil {
				// OPEN action log entry
				logEntry := &moc.IncidentLogEntry{
					Id:        mongo.CreateId(),
					Timestamp: tms.Now(),
					Type:      "ACTION_OPEN",
					Note:      incidentRequest.IncidentId,
					Assigned:  true,
				}
				incidentRequest.Log = append([]*moc.IncidentLogEntry{logEntry}, incidentRequest.Log...)
				// log entries
				processLogEntry(incidentRequest, incidentRest, ctxt)
				upsertResponse, err := incidentRest.miscDb.Upsert(db.GoMiscRequest{
					Req: &db.GoRequest{
						ObjectType: OBJECT_INCIDENT,
						Obj: &db.GoObject{
							Data: incidentRequest,
						},
					},
					Ctxt: incidentRest.group,
					Time: &db.TimeKeeper{},
				})
				if err == nil {
					incidentRequest.Id = upsertResponse.Id
					for _, entry := range incidentRequest.Log {
						if entry.Type == "TRACK" {
							incidentRest.addToRegistry(incidentRequest.Id, entry.Entity.Id)
							log.Debug("Increasing extension count for %+v", entry.Entity.Id)
							err = incidentRest.TrackExReq(entry.Entity.Id, 1)
							if err != nil {
								log.Error("Failed to send out track ex request %+v", err)
							}
						}
					}
					AuditIncident(ctxt, incidentRequest, ACTION, security.SUCCESS, incidentRequest)
					rest.WriteHeaderAndEntitySafely(response, http.StatusCreated, incidentRequest)
					incidentRest.Publish(ACTION, incidentRequest)
				} else {
					log.Error("unexpected error: %v", err)
					security.Audit(ctxt, CLASSID, ACTION, security.FAIL_ERROR)
					response.WriteError(http.StatusInternalServerError, err)
				}
			} else {
				security.Audit(ctxt, CLASSID, ACTION, security.FAIL_VALIDATION)
				response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
			}
		} else {
			security.Audit(ctxt, CLASSID, ACTION, security.FAIL_VALIDATION)
			response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		}
	} else {
		security.Audit(ctxt, CLASSID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}

func (incidentRest *IncidentRest) UpdateLogEntry(request *restful.Request, response *restful.Response) {
	log.Info("Update log entry")
	ctxt := request.Request.Context()
	if security.HasPermissionForClass(ctxt, CLASSID) {
		// validate moc.IncidentLogEntry using schema
		logEntry := &moc.IncidentLogEntry{}
		errs := rest.SanitizeValidateReadEntity(request, SCHEMA_INCIDENT_LOGENTRY, logEntry)
		if errs == nil {
			request.SetAttribute(CLASSID_LogEntry, *logEntry)
			// temp response writer to enable chaining calls
			origResponseWriter := response.ResponseWriter
			tmpResponseWriter := &TempResponseWriter{}
			response.ResponseWriter = tmpResponseWriter
			incidentRest.DeleteLogEntry(request, response)
			response.ResponseWriter = origResponseWriter
			if http.StatusOK == response.StatusCode() {
				incidentRest.CreateLogEntry(request, response)
			} else {
				response.WriteErrorString(response.StatusCode(), tmpResponseWriter.Output)
			}
		} else {
			security.Audit(ctxt, CLASSID, "ANY", security.FAIL_VALIDATION)
			response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		}
	} else {
		security.Audit(ctxt, CLASSID, "ANY", security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}

func (incidentRest *IncidentRest) DetachLogEntry(request *restful.Request, response *restful.Response) {
	log.Info("Detach log entry")
	ACTION := moc.Incident_DETACH_NOTE.String()
	ctxt := request.Request.Context()
	if security.HasPermissionForAction(ctxt, CLASSID, ACTION) {
		incidentId, errs := rest.SanitizeValidatePathParameter(request, PARAMETER_INCIDENT_ID)
		logId, err := rest.SanitizeValidatePathParameter(request, PARAMETER_LOG_ID)
		errs = append(errs, err...)
		if errs == nil {
			clientDb := incidentRest.client.DB()
			defer incidentRest.client.Release(clientDb)

			// Delete the log entry from the incident's log list
			err := incidentRest.incidentDb.DeleteIncidentLogEntry(ctxt, incidentId, logId)
			if err == nil {
				note, err := incidentRest.noteDb.FindOneNote(logId, "false", true)

				// if a logEntry exists in Note collection then set its deleted flag as false
				// else insert new documnent
				if err == nil {
					err = incidentRest.noteDb.RestoreNote(logId)

					if err == nil {
						incident, err := incidentRest.incidentDb.FindIncidentByLogEntry(logId, true)
						if err == nil {
							AuditIncident(ctxt, incident, ACTION, security.SUCCESS, note)
							rest.WriteEntitySafely(response, incident)
						} else {
							if db.ErrorNotFound == err {
								security.Audit(ctxt, CLASSID, ACTION, security.FAIL_NOTFOUND)
								response.WriteError(http.StatusNotFound, err)
							} else {
								log.Error("unexpected error: %v", err)
								security.Audit(ctxt, CLASSID, ACTION, security.FAIL_ERROR)
								response.WriteError(http.StatusInternalServerError, err)
							}
						}
					} else {
						log.Error("unexpected error: %v", err)
						security.Audit(ctxt, CLASSID, ACTION, security.FAIL_ERROR, logId)
						response.WriteError(http.StatusInternalServerError, err)
					}
				} else {
					if db.ErrorNotFound == err {
						noteRequest, err := incidentRest.noteDb.FindOneNote(logId, "true", true)

						if err == nil {
							// set id and timestamp
							noteRequest.Id = mongo.CreateId()
							noteRequest.Timestamp = tms.Now()
							// set assigned flag as true
							noteRequest.Assigned = false
							// set deleted flag as false
							noteRequest.Deleted = false
							// if attachment then get metadata
							if nil != noteRequest.Attachment {
								err := populateFileMetadata(incidentRest.client, noteRequest.Attachment)
								if err != nil {
									security.Audit(ctxt, CLASSID, ACTION, security.FAIL_VALIDATION)
									response.WriteError(http.StatusBadRequest, err)
									return
								}
							}

							upsertResponse, err := incidentRest.noteDb.UpdateNote(logId, noteRequest)

							if err == nil {
								noteRequest.Id = upsertResponse.Id

								incident, err := incidentRest.incidentDb.FindIncidentByLogEntry(logId, true)
								if err == nil {
									AuditIncident(ctxt, incident, ACTION, security.SUCCESS, noteRequest)
									rest.WriteHeaderAndEntitySafely(response, http.StatusCreated, noteRequest)
								} else {
									if db.ErrorNotFound == err {
										security.Audit(ctxt, CLASSID, ACTION, security.FAIL_NOTFOUND)
										response.WriteError(http.StatusNotFound, err)
									} else {
										log.Error("unexpected error: %v", err)
										security.Audit(ctxt, CLASSID, ACTION, security.FAIL_ERROR)
										response.WriteError(http.StatusInternalServerError, err)
									}
								}
							} else {
								log.Error("unexpected error: %v", err)
								security.Audit(ctxt, CLASSID, ACTION, security.FAIL_ERROR)
								response.WriteError(http.StatusInternalServerError, err)
							}
						} else {
							if db.ErrorNotFound == err {
								security.Audit(ctxt, CLASSID, ACTION, security.FAIL_NOTFOUND)
								response.WriteError(http.StatusNotFound, err)
							} else {
								log.Error("unexpected error: %v", err)
								security.Audit(ctxt, CLASSID, ACTION, security.FAIL_ERROR)
								response.WriteError(http.StatusInternalServerError, err)
							}
						}
					} else {
						log.Error("unexpected error: %v", err)
						security.Audit(ctxt, CLASSID, ACTION, security.FAIL_ERROR)
						response.WriteError(http.StatusInternalServerError, err)
					}
				}
			} else {
				log.Error("unexpected error: %v", err)
				security.Audit(ctxt, CLASSID, ACTION, security.FAIL_ERROR, logId)
				response.WriteError(http.StatusInternalServerError, err)
			}
		} else {
			security.Audit(ctxt, CLASSID, ACTION, security.FAIL_VALIDATION)
			response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		}
	} else {
		security.Audit(ctxt, CLASSID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}

func (incidentRest *IncidentRest) CreateLogEntry(request *restful.Request, response *restful.Response) {
	log.Info("Create log entry")
	ACTION := moc.Incident_ADD_NOTE.String()
	ctxt := request.Request.Context()
	if security.HasPermissionForAction(ctxt, CLASSID, ACTION) {
		incidentId, errs := rest.SanitizeValidatePathParameter(request, PARAMETER_INCIDENT_ID)
		if errs == nil {
			incident, err := getIncident(incidentRest, incidentId)
			if err == nil {
				// if Closed (locked) then don't update
				if moc.Incident_Closed != incident.State {
					// validate moc.IncidentLogEntry using schema
					logEntry := &moc.IncidentLogEntry{}
					// if attribute then use else read from request
					attribute := request.Attribute(CLASSID_LogEntry)
					logEntryFromAttribute, fromAttribute := attribute.(moc.IncidentLogEntry)
					if fromAttribute {
						logEntry = &logEntryFromAttribute
					} else {
						errs = rest.SanitizeValidateReadEntity(request, SCHEMA_INCIDENT_LOGENTRY, logEntry)
					}
					if errs == nil {
						// add timestamp and id
						logEntry.Id = mongo.CreateId()
						logEntry.Timestamp = tms.Now()
						// set assigned flag as true
						logEntry.Assigned = true
						// if attachment then get metadata
						if nil != logEntry.Attachment {
							ACTION = moc.Incident_ADD_NOTE_FILE.String()
							err := populateFileMetadata(incidentRest.client, logEntry.Attachment)
							if err != nil {
								security.Audit(ctxt, CLASSID, ACTION, security.FAIL_VALIDATION)
								response.WriteError(http.StatusBadRequest, err)
								return
							}
						}
						// if entity
						if nil != logEntry.Entity {
							// check collides
							if moc.HasIncidentLogEntity(incident.Log, logEntry.Entity.Id) {
								errorFree(ErrDuplicateEntityLogIncident, request, response, CLASSID, ACTION)
								return
							}
							ACTION = moc.Incident_ADD_NOTE_ENTITY.String()
						}
						incident.Log = append(incident.Log, logEntry)
						_, err := incidentRest.miscDb.Upsert(db.GoMiscRequest{
							Req: &db.GoRequest{
								ObjectType: OBJECT_INCIDENT,
								Obj: &db.GoObject{
									ID:   incidentId,
									Data: incident,
								},
							},
							Ctxt: incidentRest.group,
						})
						if err == nil {
							incident, err = getIncident(incidentRest, incidentId)
							if logEntry.Type == "TRACK" {
								incidentRest.addToRegistry(incidentId, logEntry.Entity.Id)
								log.Debug("Increasing extention count for %+v", logEntry.Entity.Id)
								err = incidentRest.TrackExReq(logEntry.Entity.Id, 1)
								if err != nil {
									log.Error("Failed to contstruct track ex request %+v", err)
								}
							}
							filterResponseLogEntry(incident)
							AuditIncident(ctxt, incident, ACTION, security.SUCCESS, logEntry)
							rest.WriteEntitySafely(response, incident)
							incidentRest.Publish(ACTION, incident)
						} else {
							log.Error("unexpected error: %v", err)
							security.Audit(ctxt, CLASSID, ACTION, security.FAIL_ERROR, incident.Id)
							response.WriteError(http.StatusInternalServerError, errors.New("incidentRest.miscDb.Upsert "+err.Error()))
						}
					} else {
						security.Audit(ctxt, CLASSID, ACTION, security.FAIL_VALIDATION)
						response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
					}
				} else {
					security.Audit(ctxt, CLASSID, ACTION, security.FAIL_VALIDATION)
					response.WriteError(http.StatusLocked, ErrorLocked)
				}
			} else {
				if db.ErrorNotFound == err {
					security.Audit(ctxt, CLASSID, ACTION, security.FAIL_NOTFOUND)
					response.WriteError(http.StatusNotFound, err)
				} else {
					log.Error("unexpected error: %v", err)
					security.Audit(ctxt, CLASSID, ACTION, security.FAIL_ERROR)
					response.WriteError(http.StatusInternalServerError, err)
				}
			}
		} else {
			security.Audit(ctxt, CLASSID, ACTION, security.FAIL_VALIDATION, incidentId)
			response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		}
	} else {
		security.Audit(ctxt, CLASSID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}

func (incidentRest *IncidentRest) DeleteLogEntry(request *restful.Request, response *restful.Response) {
	log.Info("Delete log entry")
	ACTION := moc.Incident_DELETE_NOTE.String()
	ctxt := request.Request.Context()
	if security.HasPermissionForAction(ctxt, CLASSID, ACTION) {
		incidentId, errs := rest.SanitizeValidatePathParameter(request, PARAMETER_INCIDENT_ID)
		if errs == nil {
			logId, errs := rest.SanitizeValidatePathParameter(request, PARAMETER_LOG_ID)
			if errs == nil {
				incident, err := getIncident(incidentRest, incidentId)
				if err == nil {
					// if Closed (locked) then don't update
					if moc.Incident_Closed != incident.State {
						var found *moc.IncidentLogEntry
						for _, entry := range incident.Log {
							if entry.Id == logId {
								if entry.Type == "TRACK" {
									err := incidentRest.removeFromRegistry(incidentId, entry.Entity.Id)
									if err != nil {
										log.Error("unexpected error: %v", err)
										security.Audit(ctxt, CLASSID, ACTION, security.FAIL_ERROR)
										response.WriteError(http.StatusInternalServerError, err)
										return
									}
									log.Debug("Decreasing extention count for %+v", entry.Entity.Id)
									err = incidentRest.TrackExReq(entry.Entity.Id, -1)
									if err != nil {
										log.Error("Failed to contstruct track ex request %+v", err)
									}

								}
								found = entry
								break
							}
						}
						if found != nil {
							// mark as deleted
							found.Deleted = true
							// database set
							clientDb := incidentRest.client.DB()
							defer incidentRest.client.Release(clientDb)
							query := bson.M{
								"_id":       bson.ObjectIdHex(incidentId),
								"me.log.id": found.Id,
							}
							update := bson.M{
								"$set": bson.M{
									"me.log.$.deleted": true,
								},
							}
							err = clientDb.C("incidents").Update(query, update)
							if err == nil {
								note, err := incidentRest.noteDb.FindOneNote(logId, "true", true)

								if err == nil {
									filterResponseLogEntry(incident)
									AuditIncident(ctxt, incident, ACTION, security.SUCCESS, note)
									rest.WriteEntitySafely(response, incident)
									incidentRest.Publish(ACTION, incident)
								} else {
									security.Audit(ctxt, CLASSID, ACTION, security.FAIL_ERROR, incident.Id)
									response.WriteError(http.StatusInternalServerError, err)
								}
							} else {
								log.Error("unexpected error: %v", err)
								security.Audit(ctxt, CLASSID, ACTION, security.FAIL_ERROR, incident.Id)
								response.WriteError(http.StatusInternalServerError, err)
							}
						} else {
							security.Audit(ctxt, CLASSID, ACTION, security.FAIL_NOTFOUND, logId)
							response.WriteError(http.StatusNotFound, db.ErrorNotFound)
						}
					} else {
						// Closed (locked)
						security.Audit(ctxt, CLASSID, ACTION, security.FAIL_VALIDATION)
						response.WriteError(http.StatusLocked, ErrorLocked)
					}
				} else {
					if db.ErrorNotFound == err {
						security.Audit(ctxt, CLASSID, ACTION, security.FAIL_NOTFOUND, incidentId)
						response.WriteError(http.StatusNotFound, err)
						return
					} else {
						log.Error("unexpected error: %v", err)
						security.Audit(ctxt, CLASSID, ACTION, security.FAIL_ERROR)
						response.WriteError(http.StatusInternalServerError, err)
					}
				}
			} else {
				security.Audit(ctxt, CLASSID, ACTION, security.FAIL_VALIDATION, incidentId)
				response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
			}
		} else {
			security.Audit(ctxt, CLASSID, ACTION, security.FAIL_VALIDATION, incidentId)
			response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		}
	} else {
		security.Audit(ctxt, CLASSID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}

func (incidentRest *IncidentRest) Update(request *restful.Request, response *restful.Response) {
	log.Info("UPDATE")
	ACTION := moc.Incident_UPDATE.String()
	ctxt := request.Request.Context()
	if security.HasPermissionForAction(ctxt, CLASSID, ACTION) {
		incidentId, errs := rest.SanitizeValidatePathParameter(request, PARAMETER_INCIDENT_ID)
		if errs == nil {
			incidentRequest := &moc.Incident{}
			errs = rest.SanitizeValidateReadEntity(request, SCHEMA_INCIDENT_UPDATE, incidentRequest)
			if errs == nil {
				if moc.Incident_Closed != incidentRequest.State {
					// read existing incident
					incident, err := getIncident(incidentRest, incidentId)
					if err == nil {
						// enforce read only fields
						incidentRequest.IncidentId = incident.IncidentId
						// log entries
						processLogEntryWithOriginal(incidentRequest, incident, incidentRest, ctxt)
						// check state change
						if incident.State != incidentRequest.State {
							// assume OPEN state
							incidentRequest.Log = append(incidentRequest.Log, &moc.IncidentLogEntry{
								Id:        mongo.CreateId(),
								Timestamp: tms.Now(),
								Type:      "ACTION_" + moc.Incident_OPEN.String(),
								Note:      fmt.Sprintf("Re-opened as %v", incidentRequest.IncidentId),
							})
							AuditIncident(ctxt, incidentRequest, moc.Incident_OPEN.String(), security.SUCCESS, incidentRequest)
						}
						// check assign change
						if incident.Assignee != incidentRequest.Assignee {
							// check permissions of assignee
							if errs := validateIncidentUserRoleByString(ctxt, incidentRequest.Assignee, "assignee"); errs != nil {
								security.Audit(ctxt, CLASSID, ACTION, security.FAIL_VALIDATION)
								response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
								return
							}
							if incident.Assignee != "" {
								AuditIncident(ctxt, incident, moc.Incident_UNASSIGN.String(), security.SUCCESS, incident.Assignee)
							}
							AuditIncident(ctxt, incident, moc.Incident_ASSIGN.String(), security.SUCCESS, incidentRequest.Assignee)
						}
						if incident.Commander != incidentRequest.Commander {
							// check permissions of assignee
							if errs := validateIncidentUserRoleByString(ctxt, incidentRequest.Commander, "commander"); errs != nil {
								security.Audit(ctxt, CLASSID, ACTION, security.FAIL_VALIDATION)
								response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
								return
							}
							if incident.Commander != "" {
								AuditIncident(ctxt, incident, moc.Incident_UNASSIGN.String(), security.SUCCESS, incident.Commander)
							}
							AuditIncident(ctxt, incident, moc.Incident_ASSIGN.String(), security.SUCCESS, incidentRequest.Commander)
						}

						_, err := incidentRest.miscDb.Upsert(db.GoMiscRequest{
							Req: &db.GoRequest{
								ObjectType: OBJECT_INCIDENT,
								Obj: &db.GoObject{
									ID:   incidentId,
									Data: incidentRequest,
								},
							},
							Ctxt: incidentRest.group,
						})
						if err == nil {
							filterResponseLogEntry(incidentRequest)
							AuditIncident(ctxt, incidentRequest, ACTION, security.SUCCESS, incidentRequest)
							rest.WriteEntitySafely(response, incidentRequest)
							incidentRest.Publish(ACTION, incidentRequest)
						} else {
							log.Error("unexpected error: %v", err)
							security.Audit(ctxt, CLASSID, ACTION, security.FAIL_ERROR, incidentRequest.Id)
							response.WriteError(http.StatusInternalServerError, err)
						}
					} else {
						if db.ErrorNotFound == err {
							security.Audit(ctxt, CLASSID, ACTION, security.FAIL_NOTFOUND, incidentId)
							response.WriteError(http.StatusNotFound, err)
						} else {
							log.Error("unexpected error: %v", err)
							security.Audit(ctxt, CLASSID, ACTION, security.FAIL_ERROR)
							response.WriteError(http.StatusInternalServerError, err)
						}
					}
				} else {
					errs = rest.SanitizeValidate(incidentRequest, SCHEMA_INCIDENT_CLOSE)
					if errs == nil {
						// being closed
						request.SetAttribute(CLASSID, *incidentRequest)
						incidentRest.UpdateState(request, response)
					} else {
						security.Audit(ctxt, CLASSID, ACTION, security.FAIL_VALIDATION)
						response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
					}
				}
			} else {
				security.Audit(ctxt, CLASSID, ACTION, security.FAIL_VALIDATION)
				response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
			}
		} else {
			security.Audit(ctxt, CLASSID, ACTION, security.FAIL_VALIDATION, incidentId)
			response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		}
	} else {
		security.Audit(ctxt, CLASSID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}

func (incidentRest *IncidentRest) Assign(request *restful.Request, response *restful.Response) {
	ACTION := moc.Incident_ASSIGN.String()
	ctxt := request.Request.Context()
	if security.HasPermissionForAction(ctxt, CLASSID, ACTION) {
		incidentId, errs := rest.SanitizeValidatePathParameter(request, PARAMETER_INCIDENT_ID)
		userId, errs2 := rest.SanitizeValidatePathParameter(request, PARAMETER_USER_ID)
		errs = append(errs, errs2...)
		if errs == nil {
			incident, err := getIncident(incidentRest, incidentId)
			if err == nil {
				if incident.Assignee != userId {
					// set assignee of incident
					previousAssignee := incident.Assignee
					incident.Assignee = userId
					// validate User Role
					errs = validateUserRole(ctxt, incident)
					if errs == nil {
						// update incident
						_, err = incidentRest.miscDb.Upsert(db.GoMiscRequest{
							Req: &db.GoRequest{
								ObjectType: OBJECT_INCIDENT,
								Obj: &db.GoObject{
									ID:   incidentId,
									Data: incident,
								},
							},
							Ctxt: incidentRest.group,
						})
						if err == nil {
							incident.Id = incidentId
							filterResponseLogEntry(incident)
							AuditIncident(ctxt, incident, moc.Incident_UNASSIGN.String(), security.SUCCESS, previousAssignee)
							AuditIncident(ctxt, incident, ACTION, security.SUCCESS, incident.Assignee)
							rest.WriteEntitySafely(response, incident)
							incidentRest.Publish(ACTION, incident)
						} else {
							log.Error("unexpected error: %v", err)
							security.Audit(ctxt, CLASSID, ACTION, security.FAIL_ERROR, incidentId)
							response.WriteError(http.StatusInternalServerError, err)
						}
					} else {
						security.Audit(ctxt, CLASSID, ACTION, security.FAIL_VALIDATION)
						response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
					}
				} else {
					AuditIncident(ctxt, incident, ACTION, security.SUCCESS_NOTMODIFIED, incident)
					response.WriteHeaderAndEntity(http.StatusNotModified, incident)
				}
			} else {
				if db.ErrorNotFound == err {
					security.Audit(ctxt, CLASSID, ACTION, security.FAIL_NOTFOUND, incidentId)
					response.WriteError(http.StatusNotFound, err)
				} else {
					log.Error("unexpected error: %v", err)
					security.Audit(ctxt, CLASSID, ACTION, security.FAIL_ERROR)
					response.WriteError(http.StatusInternalServerError, err)
				}
			}
		} else {
			security.Audit(ctxt, CLASSID, ACTION, security.FAIL_VALIDATION, incidentId)
			response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		}
	} else {
		security.Audit(ctxt, CLASSID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}

func (incidentRest *IncidentRest) Delete(request *restful.Request, response *restful.Response) {
	ACTION := moc.Incident_DELETE.String()
	ctxt := request.Request.Context()
	if security.HasPermissionForAction(ctxt, CLASSID, ACTION) {
		incidentId, errs := rest.SanitizeValidatePathParameter(request, PARAMETER_INCIDENT_ID)
		if errs == nil {
			err := incidentRest.miscDb.Delete(db.GoMiscRequest{
				Req: &db.GoRequest{
					ObjectType: OBJECT_INCIDENT,
					Obj: &db.GoObject{
						ID: incidentId,
					},
				},
				Ctxt: incidentRest.group,
				Time: &db.TimeKeeper{},
			})
			if err == nil {
				security.Audit(ctxt, CLASSID, ACTION, security.SUCCESS, incidentId)
			} else {
				log.Error("unexpected error: %v", err)
				security.Audit(ctxt, CLASSID, ACTION, security.FAIL_ERROR, incidentId)
				response.WriteError(http.StatusInternalServerError, err)
			}
		} else {
			security.Audit(ctxt, CLASSID, ACTION, security.FAIL_VALIDATION, incidentId)
			response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		}
	} else {
		security.Audit(ctxt, CLASSID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}

func (incidentRest *IncidentRest) UpdateState(request *restful.Request, response *restful.Response) {
	ACTION := moc.Incident_CLOSE.String() // FIXME inputed state determines action
	ctxt := request.Request.Context()
	if security.HasPermissionForAction(ctxt, CLASSID, ACTION) {
		incidentRequest := &moc.Incident{}
		var err error
		// if attribute then use else read from request
		attribute := request.Attribute(CLASSID)
		incidentFromAttribute, fromAttribute := attribute.(moc.Incident)
		if fromAttribute {
			incidentRequest = &incidentFromAttribute
		} else {
			// read from request
			incidentId, errs := rest.SanitizeValidatePathParameter(request, PARAMETER_INCIDENT_ID)
			incidentStateString, errs2 := rest.SanitizeValidatePathParameter(request, PARAMETER_INCIDENT_STATE)
			errs = append(errs, errs2...)
			if errs == nil {
				incidentRequest.Id = incidentId
				incidentStateInt, _ := strconv.Atoi(incidentStateString)
				// set state
				incidentRequest.State = moc.Incident_State(incidentStateInt)
			} else {
				security.Audit(ctxt, CLASSID, ACTION, security.FAIL_VALIDATION, incidentId)
				response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
				return
			}
		}
		// original incident to enforce read-only portion
		incidentOriginal, err := getIncident(incidentRest, incidentRequest.Id)
		if err == nil {
			// rule: if closed then no log updates
			if moc.Incident_Closed == incidentRequest.State || moc.Incident_Archived == incidentRequest.State {
				incidentRequest.Log = incidentOriginal.Log
			}
			stateChanged := incidentOriginal.State != incidentRequest.State
			if !fromAttribute {
				incidentOriginal.State = incidentRequest.State
				incidentRequest = incidentOriginal
			}
			if stateChanged {
				// create action log entry
				logEntry := &moc.IncidentLogEntry{
					Id:        mongo.CreateId(),
					Timestamp: tms.Now(),
					Type:      "ACTION_" + moc.Incident_CLOSE.String(),
				}
				if moc.Incident_Open == incidentRequest.State {
					ACTION = moc.Incident_OPEN.String()
					logEntry.Type = "ACTION_" + moc.Incident_OPEN.String()
					//log entry check and send ext request
					for _, entry := range incidentRequest.Log {
						if entry.Type == "TRACK" {
							log.Debug("Increasing extention count for %+v", entry.Entity.Id)
							err := incidentRest.TrackExReq(entry.Entity.Id, 1)
							if err != nil {
								log.Error("Failed to contstruct track ex request %+v", err)
							}
						}
					}
				} else {
					//log entry check and send ext killer
					for _, entry := range incidentRequest.Log {
						if entry.Type == "TRACK" {
							log.Debug("Decreading extention count for %+v", entry.Entity.Id)
							//sendtotgwad here
							err := incidentRest.TrackExReq(entry.Entity.Id, -1)
							if err != nil {
								log.Error("Failed to contstruct track ex request %+v", err)
							}
						}
					}
				}
				incidentRequest.Log = append(incidentRequest.Log, logEntry)
				// update
				_, err = incidentRest.miscDb.Upsert(db.GoMiscRequest{
					Req: &db.GoRequest{
						ObjectType: OBJECT_INCIDENT,
						Obj: &db.GoObject{
							ID:   incidentRequest.Id,
							Data: incidentRequest,
						},
					},
					Ctxt: incidentRest.group,
				})
				if err == nil {
					filterResponseLogEntry(incidentRequest)
					AuditIncident(ctxt, incidentRequest, ACTION, security.SUCCESS, incidentRequest)
					rest.WriteEntitySafely(response, incidentRequest)
					incidentRest.Publish(ACTION, incidentRequest)
				} else {
					log.Error("unexpected error: %v", err)
					security.Audit(ctxt, CLASSID, ACTION, security.FAIL_ERROR, incidentRequest.Id)
					response.WriteError(http.StatusInternalServerError, err)
				}
			} else {
				AuditIncident(ctxt, incidentRequest, ACTION, security.SUCCESS_NOTMODIFIED, incidentRequest)
				response.WriteHeaderAndEntity(http.StatusNotModified, incidentRequest)
			}
		} else {
			if db.ErrorNotFound == err {
				security.Audit(ctxt, CLASSID, ACTION, security.FAIL_NOTFOUND, incidentRequest.Id)
				response.WriteError(http.StatusNotFound, err)
			} else {
				log.Error("unexpected error: %v", err)
				security.Audit(ctxt, CLASSID, ACTION, security.FAIL_ERROR)
				response.WriteError(http.StatusInternalServerError, err)
			}
		}
	} else {
		security.Audit(ctxt, CLASSID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}

func filterResponseLogEntry(incident *moc.Incident) {
	// rule: if open then filter
	if moc.Incident_Open == incident.State {
		// process log entries
		if len(incident.Log) > 0 {
			tmp := make([]*moc.IncidentLogEntry, 0)
			for _, logEntry := range incident.Log {
				if !logEntry.Deleted {
					tmp = append(tmp, logEntry)
				}
			}
			incident.Log = tmp
		}
	}
}

// process new log entries, enforce readonly log entries, if removed mark deleted
func processLogEntryWithOriginal(incident *moc.Incident, original *moc.Incident, incidentRest *IncidentRest, ctxt context.Context) {
	// rule: if closed then no log updates (locked)
	if moc.Incident_Closed == incident.State || moc.Incident_Archived == incident.State {
		return
	}
	existingLogEntryIds := make(map[string]struct{})
	logEntryTime := tms.Now()
	if len(incident.Log) > 0 {
		for _, logEntry := range incident.Log {
			// process new log entries
			if logEntry.Id == "" {
				// add timestamp and id
				logEntry.Id = mongo.CreateId()
				logEntry.Timestamp = logEntryTime
				// if attachment then get metadata
				if nil != logEntry.Attachment {
					err := populateFileMetadata(incidentRest.client, logEntry.Attachment)
					if err != nil {
						security.Audit(ctxt, CLASSID, moc.Incident_ADD_NOTE_FILE.String(), security.FAIL_ERROR)
					}
				}
				// add to original
				original.Log = append(original.Log, logEntry)
			} else {
				existingLogEntryIds[logEntry.Id] = struct{}{}
			}
		}
	}
	// if removed then mark as deleted
	if len(original.Log) > 0 {
		var has bool
		for _, logEntry := range original.Log {
			_, has = existingLogEntryIds[logEntry.Id]
			logEntry.Deleted = !has
		}
	}
	incident.Log = original.Log
}

// process only new log entries
func processLogEntry(incident *moc.Incident, incidentRest *IncidentRest, ctxt context.Context) {
	// rule: if closed then no log updates (locked)
	if moc.Incident_Closed == incident.State || moc.Incident_Archived == incident.State {
		return
	}
	logEntryTime := tms.Now()
	if len(incident.Log) > 0 {
		for _, logEntry := range incident.Log {
			// process new log entries
			if logEntry.Id == "" {
				// add timestamp and id
				logEntry.Id = mongo.CreateId()
				logEntry.Timestamp = logEntryTime
				// if attachment then get metadata
				if nil != logEntry.Attachment {
					err := populateFileMetadata(incidentRest.client, logEntry.Attachment)
					if err != nil {
						security.Audit(ctxt, CLASSID, moc.Incident_ADD_NOTE_FILE.String(), security.FAIL_ERROR)
					}
				}
			}
		}
	}
}

func (incidentRest *IncidentRest) addToRegistry(incidentID string, logID string) error {
	ok, err := incidentRest.regDb.AddToIncident(logID, incidentID)
	if !ok {
		return db.ErrorNotFound
	}
	return err
}

func (incidentRest *IncidentRest) removeFromRegistry(incidentID string, logID string) error {
	ok, err := incidentRest.regDb.RemoveFromIncident(logID, incidentID)
	if !ok {
		return db.ErrorNotFound
	}
	return err
}

// get incident, return ErrorNotFound
func getIncident(incidentRest *IncidentRest, incidentId string) (*moc.Incident, error) {
	var incident *moc.Incident
	incidentData, err := incidentRest.miscDb.Get(db.GoMiscRequest{
		Req: &db.GoRequest{
			Obj: &db.GoObject{
				ID: incidentId,
			},
			ObjectType: OBJECT_INCIDENT,
		},
		Ctxt: incidentRest.group,
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
	return incident, err
}

func (incidentRest *IncidentRest) Publish(action string, incident *moc.Incident) {
	envelope := envelope.Envelope{
		Type:     TOPIC_INCIDENT + "/" + action,
		Contents: &envelope.Envelope_Incident{Incident: incident},
	}
	incidentRest.publisher.Publish(TOPIC_INCIDENT, envelope)
}

// TrackExReq extracts last track using registry id and constructs a trackex request that is send to tanalyzed
// function returns an error if it fails
func (incidentRest *IncidentRest) TrackExReq(id string, count int32) error {
	track, err := incidentRest.trackExDb.GetTrack(bson.M{"Track.registry_id": id})
	if err != nil {
		log.Warn("Track with registry id %+v not found in trackex collections %+v", id, err)
	}
	trq := &tms.TrackExReq{
		Track:      track,
		RegistryId: id,
		Count:      count,
	}
	body, err := tmsg.PackFrom(trq)
	if err != nil {
		return err
	}
	pnow, err := ptypes.TimestampProto(time.Now())
	if err != nil {
		return err
	}
	incidentRest.tsiClient.Send(incidentRest.group, &tms.TsiMessage{
		Source: incidentRest.tsiClient.Local(),
		Destination: []*tms.EndPoint{
			{
				Site: tmsg.GClient.ResolveSite(""),
			},
			{
				Site: tmsg.TMSG_HQ_SITE,
			},
		},
		WriteTime: pnow,
		SendTime:  pnow,
		Body:      body,
	})
	return err
}

func AuditIncident(context context.Context, incident *moc.Incident, action string, outcome string, payload interface{}) {
	if payload == nil {
		security.AuditUserObject(context, CLASSID, incident.Id, "", action, outcome)
		return
	}

	note, ok := payload.(*moc.IncidentLogEntry)
	if ok {
		encodedPayload, err := json.Marshal(note)
		if err != nil {
			security.AuditUserObject(context, CLASSID, incident.Id, "", action, outcome, note.Id)
		} else {
			security.AuditUserObject(context, CLASSID, incident.Id, "", action, outcome, string(encodedPayload))
		}
		return
	}

	incidentPayload, ok := payload.(*moc.IncidentLogEntry)
	if ok {
		encodedPayload, err := json.Marshal(incidentPayload)
		if err != nil {
			security.AuditUserObject(context, CLASSID, incident.Id, "", action, outcome, incidentPayload.Id)
		} else {
			security.AuditUserObject(context, CLASSID, incident.Id, "", action, outcome, string(encodedPayload))
		}
		return
	}

	payloads, ok := payload.([]interface{})
	if ok {
		security.AuditUserObject(context, CLASSID, incident.Id, "", action, outcome, payloads...)
		return
	}

	security.AuditUserObject(context, CLASSID, incident.Id, "", action, outcome, payload)
}

func populateFileMetadata(client *mongo.MongoClient, metadata *moc.FileMetadata) error {
	if bson.IsObjectIdHex(metadata.Id) {
		fileID := mongo.GetId(metadata.Id)
		db := client.DB()
		defer client.Release(db)
		src, err := db.GridFS("fs").OpenId(fileID)
		defer src.Close()
		if err == nil {
			// update metadata
			metadata.Type = src.ContentType()
			metadata.Size = src.Size()
			metadata.Name = src.Name()
			metadata.Md5 = src.MD5()
		}
		return err
	}
	return ErrorInvalidId
}

func validateIncidentUserRoleByString(ctx context.Context, user, property string) []rest.ErrorValidation {
	var errs []rest.ErrorValidation
	// role check, validate assignee
	searchMap := make(map[string]string)
	searchMap["roles"] = message.RoleId_IncidentManager.String()
	searchMap["userId"] = user
	users, err := secDb.FindByMap(ctx, searchMap)
	if err != nil {
		return []rest.ErrorValidation{
			{
				Property: property,
				Rule:     "Rule",
				Message:  "Required user with role " + message.RoleId_IncidentManager.String(),
			},
		}
	}
	for i := range users {
		if user == string(users[i].UserId) {
			return errs
		}
	}
	return append(errs, rest.ErrorValidation{
		Property: property,
		Rule:     "Rule",
		Message:  "Required user with role " + message.RoleId_IncidentManager.String()})
}

func validateUserRole(ctxt context.Context, incidentRequest *moc.Incident) []rest.ErrorValidation {
	var errs []rest.ErrorValidation
	// role check, validate assignee
	searchMap := make(map[string]string)
	searchMap["roles"] = message.RoleId_IncidentManager.String()
	searchMap["userId"] = incidentRequest.Assignee + "," + incidentRequest.Commander
	users, err := secDb.FindByMap(ctxt, searchMap)
	if err == nil {
		found := false
		// assignee
		for _, user := range users {
			if incidentRequest.Assignee == string(user.UserId) {
				found = true
				break
			}
		}
		if !found {
			errs = append(errs, rest.ErrorValidation{
				Property: "assignee",
				Rule:     "Rule",
				Message:  "Required user with role " + message.RoleId_IncidentManager.String()})
		}
		// if no user then proceed, else check user has correct role
		found = incidentRequest.Commander == ""
		// commander
		for _, user := range users {
			if incidentRequest.Commander == string(user.UserId) {
				found = true
				break
			}
		}
		if !found {
			errs = append(errs, rest.ErrorValidation{
				Property: "commander",
				Rule:     "Rule",
				Message:  "Required user with role " + message.RoleId_IncidentManager.String()})
		}
	}
	return errs
}

type TempResponseWriter struct {
	StatusCode int
	Output     string
	header     http.Header
}

func (rw *TempResponseWriter) Header() http.Header {
	if rw.header == nil {
		rw.header = make(http.Header)
	}
	return rw.header
}

func (rw *TempResponseWriter) Write(bytes []byte) (int, error) {
	if rw.StatusCode == 0 {
		rw.WriteHeader(200)
	}
	rw.Output = rw.Output + string(bytes)
	return 0, nil
}

func (rw *TempResponseWriter) WriteHeader(i int) {
	rw.StatusCode = i
}
