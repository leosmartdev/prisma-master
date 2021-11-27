package public

import (
	"net/http"
	"prisma/gogroup"
	"prisma/tms"
	"prisma/tms/db"
	"prisma/tms/db/mongo"
	"prisma/tms/devices"
	"prisma/tms/log"
	"prisma/tms/marker"
	"prisma/tms/rest"
	"prisma/tms/security"

	"github.com/go-openapi/spec"
	restful "github.com/orolia/go-restful"
)

const (
	MARKER_CLASS_ID       = "Marker"
	MARKER_IMAGE_CLASS_ID = "Marker Image"
)

var (
	// parameters with schema
	PARAMETER_MARKER_ID = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "marker-id",
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

	// schema
	SCHEMA_MARKER = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Required: []string{"type", "position"},
			Properties: map[string]spec.Schema{
				"id": {
					SwaggerSchemaProps: spec.SwaggerSchemaProps{
						ReadOnly: true,
					},
				},
				"type": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{2}[0],
						MaxLength: &[]int64{255}[0],
					},
				},
				"shape": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{2}[0],
						MaxLength: &[]int64{255}[0],
					},
				},
				"position": {
					SwaggerSchemaProps: spec.SwaggerSchemaProps{
						ReadOnly: true,
					},
				},
				"color": {
					SwaggerSchemaProps: spec.SwaggerSchemaProps{
						ReadOnly: true,
					},
				},
				"image_metadata": {
					SwaggerSchemaProps: spec.SwaggerSchemaProps{
						ReadOnly: true,
					},
				},
				"timestamp": {
					SwaggerSchemaProps: spec.SwaggerSchemaProps{
						ReadOnly: true,
					},
				},
				"description": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{0}[0],
						MaxLength: &[]int64{255}[0],
					},
				},
			},
		},
	}

	SCHEMA_MARKER_IMAGE = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Required: []string{"metadata"},
			Properties: map[string]spec.Schema{
				"metadata": {
					SwaggerSchemaProps: spec.SwaggerSchemaProps{
						ReadOnly: true,
					},
				},
			},
		},
	}
)

type MarkerRest struct {
	client     *mongo.MongoClient
	markerDb   db.MarkerDB
	noteDb     db.NoteDB
	incidentDB db.IncidentDB
	group      gogroup.GoGroup
}

func NewMarkerRest(group gogroup.GoGroup, client *mongo.MongoClient) *MarkerRest {
	miscDb := mongo.NewMongoMiscData(group, client)

	return &MarkerRest{
		client:     client,
		markerDb:   mongo.NewMarkerDb(group, client),
		noteDb:     mongo.NewMongoNoteDb(miscDb),
		incidentDB: mongo.NewMongoIncidentMiscData(miscDb),
		group:      group,
	}
}

func (markerRest *MarkerRest) Create(request *restful.Request, response *restful.Response) {
	ACTION := marker.Marker_CREATE.String()
	ctxt := request.Request.Context()
	if security.HasPermissionForAction(ctxt, MARKER_CLASS_ID, ACTION) {
		markerRequest := new(marker.Marker)

		err := rest.SanitizeValidateReadEntity(request, SCHEMA_MARKER, markerRequest)
		if err == nil {
			// set id and timestamp
			markerRequest.Id = mongo.CreateId()
			markerRequest.Timestamp = tms.Now()

			log.Debug("Creating marker: %+v", markerRequest)
			_, err := markerRest.markerDb.UpsertMarker(markerRequest)

			if err == nil {
				security.Audit(ctxt, MARKER_CLASS_ID, ACTION, security.SUCCESS)
				rest.WriteHeaderAndEntitySafely(response, http.StatusCreated, markerRequest)
			} else {
				log.Error("unexpected error: %v", err)
				security.Audit(ctxt, MARKER_CLASS_ID, ACTION, security.FAIL_ERROR)
				response.WriteError(http.StatusInternalServerError, err)
			}
		} else {
			security.Audit(ctxt, MARKER_CLASS_ID, ACTION, security.FAIL_VALIDATION)
			response.WriteHeaderAndEntity(http.StatusBadRequest, err)
		}
	} else {
		security.Audit(ctxt, MARKER_CLASS_ID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}

func (markerRest *MarkerRest) Get(request *restful.Request, response *restful.Response) {
	ACTION := marker.Marker_READ.String()
	ctxt := request.Request.Context()
	if security.HasPermissionForAction(ctxt, MARKER_CLASS_ID, ACTION) {
		markerId, errs := rest.SanitizeValidatePathParameter(request, PARAMETER_MARKER_ID)
		if errs == nil {
			marker, err := markerRest.markerDb.FindOneMarker(markerId, false)
			if err == nil {
				incidents, err := markerRest.incidentDB.GetIncidentWithMarkerID(markerId)
				if err == nil {
					incidentIds := make([]string, 0)

					for _, incident := range incidents {
						incidentIds = append(incidentIds, incident.Id)
					}

					infoResponse := &InfoResponse{
						ID:       marker.Id,
						MarkerID: marker.Id,
						LookupID: marker.Id,
						Target: &tms.Target{
							Marker:   marker,
							Position: (*tms.Point)(marker.Position),
							Type:     devices.DeviceType_Marker,
						},
						Registry: &Registry{
							Incidents: incidentIds,
						},
					}

					security.Audit(ctxt, MARKER_CLASS_ID, ACTION, security.SUCCESS)
					rest.WriteEntitySafely(response, infoResponse)
				} else {
					security.Audit(ctxt, MARKER_CLASS_ID, ACTION, security.FAIL_ERROR)
					response.WriteError(http.StatusInternalServerError, err)
				}
			} else {
				if db.ErrorNotFound == err {
					security.Audit(ctxt, MARKER_CLASS_ID, ACTION, security.FAIL_NOTFOUND)
					response.WriteError(http.StatusNotFound, err)
				} else {
					log.Error("unexpected error: %v", err)
					security.Audit(ctxt, MARKER_CLASS_ID, ACTION, security.FAIL_ERROR)
					response.WriteError(http.StatusInternalServerError, err)
				}
			}
		} else {
			security.Audit(ctxt, MARKER_CLASS_ID, ACTION, security.FAIL_VALIDATION, markerId)
			response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		}
	} else {
		security.Audit(ctxt, MARKER_CLASS_ID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}

func (markerRest *MarkerRest) Update(request *restful.Request, response *restful.Response) {
	ACTION := marker.Marker_UPDATE.String()
	ctxt := request.Request.Context()
	if security.HasPermissionForAction(ctxt, MARKER_CLASS_ID, ACTION) {
		markerRequest := new(marker.Marker)

		errs := rest.SanitizeValidateReadEntity(request, SCHEMA_MARKER, markerRequest)
		markerId, err := rest.SanitizeValidatePathParameter(request, PARAMETER_MARKER_ID)
		errs = append(errs, err...)
		if errs == nil {
			// set id and timestamp
			markerRequest.Id = markerId
			markerRequest.Timestamp = tms.Now()

			log.Debug("Updating marker: %+v", markerRequest)
			_, err := markerRest.markerDb.UpsertMarker(markerRequest)

			if err == nil {
				security.Audit(ctxt, MARKER_CLASS_ID, ACTION, security.SUCCESS)
				rest.WriteHeaderAndEntitySafely(response, http.StatusCreated, markerRequest)
			} else {
				log.Error("unexpected error: %v", err)
				security.Audit(ctxt, MARKER_CLASS_ID, ACTION, security.FAIL_ERROR)
				response.WriteError(http.StatusInternalServerError, err)
			}
		} else {
			security.Audit(ctxt, MARKER_CLASS_ID, ACTION, security.FAIL_VALIDATION)
			response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		}
	} else {
		security.Audit(ctxt, MARKER_CLASS_ID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}

func (markerRest *MarkerRest) Delete(request *restful.Request, response *restful.Response) {
	ACTION := marker.Marker_DELETE.String()
	ctxt := request.Request.Context()
	if security.HasPermissionForAction(ctxt, MARKER_CLASS_ID, ACTION) {
		markerId, errs := rest.SanitizeValidatePathParameter(request, PARAMETER_MARKER_ID)
		if errs == nil {
			err := markerRest.markerDb.DeleteMarker(markerId)
			if err == nil {
				marker, err := markerRest.markerDb.FindOneMarker(markerId, true)
				if err == nil {
					security.Audit(ctxt, MARKER_CLASS_ID, ACTION, security.SUCCESS)
					rest.WriteHeaderAndEntitySafely(response, http.StatusCreated, marker)
				} else {
					if db.ErrorNotFound == err {
						security.Audit(ctxt, MARKER_CLASS_ID, ACTION, security.FAIL_NOTFOUND)
						response.WriteError(http.StatusNotFound, err)
					} else {
						log.Error("unexpected error: %v", err)
						security.Audit(ctxt, MARKER_CLASS_ID, ACTION, security.FAIL_ERROR)
						response.WriteError(http.StatusInternalServerError, err)
					}
				}
			} else {
				if db.ErrorNotFound == err {
					security.Audit(ctxt, MARKER_CLASS_ID, ACTION, security.FAIL_NOTFOUND)
					response.WriteError(http.StatusNotFound, err)
				} else {
					log.Error("unexpected error: %v", err)
					security.Audit(ctxt, MARKER_CLASS_ID, ACTION, security.FAIL_ERROR)
					response.WriteError(http.StatusInternalServerError, err)
				}
			}
		} else {
			security.Audit(ctxt, MARKER_CLASS_ID, ACTION, security.FAIL_VALIDATION, markerId)
			response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		}
	} else {
		security.Audit(ctxt, MARKER_CLASS_ID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}

func (markerRest *MarkerRest) CreateMarkerImage(request *restful.Request, response *restful.Response) {
	ACTION := marker.MarkerImage_CREATE.String()
	ctxt := request.Request.Context()
	if security.HasPermissionForAction(ctxt, MARKER_CLASS_ID, ACTION) {
		markerImageRequest := new(marker.MarkerImage)

		err := rest.SanitizeValidateReadEntity(request, SCHEMA_MARKER_IMAGE, markerImageRequest)
		if err == nil {
			// set id and timestamp
			markerImageRequest.Id = mongo.CreateId()
			markerImageRequest.Timestamp = tms.Now()

			log.Debug("Creating marker image: %+v", markerImageRequest)
			_, err := markerRest.markerDb.UpsertMarkerImage(markerImageRequest)

			if err == nil {
				security.Audit(ctxt, MARKER_IMAGE_CLASS_ID, ACTION, security.SUCCESS)
				rest.WriteHeaderAndEntitySafely(response, http.StatusCreated, markerImageRequest)
			} else {
				log.Error("unexpected error: %v", err)
				security.Audit(ctxt, MARKER_IMAGE_CLASS_ID, ACTION, security.FAIL_ERROR)
				response.WriteError(http.StatusInternalServerError, err)
			}
		} else {
			security.Audit(ctxt, MARKER_IMAGE_CLASS_ID, ACTION, security.FAIL_VALIDATION)
			response.WriteHeaderAndEntity(http.StatusBadRequest, err)
		}
	} else {
		security.Audit(ctxt, MARKER_IMAGE_CLASS_ID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}

func (markerRest *MarkerRest) ReadAllMarkerImage(request *restful.Request, response *restful.Response) {
	ACTION := marker.MarkerImage_READ.String()
	ctxt := request.Request.Context()
	if security.HasPermissionForAction(ctxt, MARKER_CLASS_ID, ACTION) {
		res, err := markerRest.markerDb.FindAllMarkerImages()

		if err == nil {
			security.Audit(ctxt, MARKER_IMAGE_CLASS_ID, ACTION, security.SUCCESS)
			rest.WriteEntitySafely(response, res)
		} else {
			log.Error("unexpected error: %v", err)
			security.Audit(ctxt, MARKER_IMAGE_CLASS_ID, ACTION, security.FAIL_ERROR)
			response.WriteError(http.StatusInternalServerError, err)
		}
	} else {
		security.Audit(ctxt, MARKER_IMAGE_CLASS_ID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}
