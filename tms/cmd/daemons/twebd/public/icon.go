package public

import (
	"net/http"
	"prisma/gogroup"
	"prisma/tms"
	"prisma/tms/db"
	"prisma/tms/db/mongo"
	"prisma/tms/log"
	"prisma/tms/moc"
	"prisma/tms/rest"
	"prisma/tms/security"

	"github.com/go-openapi/spec"
	restful "github.com/orolia/go-restful"
)

const (
	ICON_CLASS_ID       = "Icon"
	ICON_IMAGE_CLASS_ID = "Icon Image"
)

var (
	// parameters with schema
	PARAMETER_ICON_ID = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "icon-id",
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

	PARAMETER_MAC_ADDRESS = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "mac_address",
			In:       "query",
			Required: true,
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					MinLength: &[]int64{17}[0],
					MaxLength: &[]int64{17}[0],
					Pattern:   "[0-9a-fA-F:]{17}",
				},
			},
		},
		SimpleSchema: spec.SimpleSchema{
			Type: "string",
		},
	}

	// schema
	SCHEMA_ICON = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Required: []string{"track_type", "mac_address", "metadata"},
			Properties: map[string]spec.Schema{
				"track_type": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{1}[0],
						MaxLength: &[]int64{255}[0],
					},
				},
				"track_sub_type": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{1}[0],
						MaxLength: &[]int64{255}[0],
					},
				},
				"mac_address": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{17}[0],
						MaxLength: &[]int64{17}[0],
						Pattern:   "[0-9a-fA-F:]{17}",
					},
				},
				"metadata": {
					SwaggerSchemaProps: spec.SwaggerSchemaProps{
						ReadOnly: true,
					},
				},
			},
		},
	}
	SCHEMA_ICON_IMAGE = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Required: []string{"mac_address", "metadata"},
			Properties: map[string]spec.Schema{
				"mac_address": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{17}[0],
						MaxLength: &[]int64{17}[0],
						Pattern:   "[0-9a-fA-F:]{17}",
					},
				},
				"metadata": {
					SwaggerSchemaProps: spec.SwaggerSchemaProps{
						ReadOnly: true,
					},
				},
			},
		},
	}
)

type IconRest struct {
	client *mongo.MongoClient
	group  gogroup.GoGroup
	iconDb db.IconDB
}

func NewIconRest(group gogroup.GoGroup, client *mongo.MongoClient) *IconRest {
	miscDb := mongo.NewMongoMiscData(group, client)

	return &IconRest{
		client: client,
		group:  group,
		iconDb: mongo.NewIconDb(miscDb),
	}
}

func (iconRest *IconRest) Create(request *restful.Request, response *restful.Response) {
	ACTION := moc.Icon_CREATE.String()
	ctxt := request.Request.Context()
	if security.HasPermissionForAction(ctxt, ICON_CLASS_ID, ACTION) {
		iconRequest := new(moc.Icon)

		err := rest.SanitizeValidateReadEntity(request, SCHEMA_ICON, iconRequest)
		if err == nil {
			log.Debug("create an Icon: %v", iconRequest)

			// Set id & timestamp
			iconRequest.Id = mongo.CreateId()
			iconRequest.Timestamp = tms.Now()

			err := iconRest.iconDb.UpsertIcon(iconRequest)
			if err == nil {
				security.Audit(ctxt, ICON_CLASS_ID, ACTION, security.SUCCESS)
				rest.WriteHeaderAndEntitySafely(response, http.StatusCreated, iconRequest)
			} else {
				log.Error("unexpected error: %v", err)
				security.Audit(ctxt, ICON_CLASS_ID, ACTION, security.FAIL_ERROR)
				response.WriteError(http.StatusInternalServerError, err)
			}
		} else {
			security.Audit(ctxt, ICON_CLASS_ID, ACTION, security.FAIL_VALIDATION)
			response.WriteHeaderAndEntity(http.StatusBadRequest, err)
		}
	} else {
		security.Audit(ctxt, ICON_CLASS_ID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}

func (iconRest *IconRest) Get(request *restful.Request, response *restful.Response) {
	ACTION := moc.Icon_READ.String()
	ctxt := request.Request.Context()
	if security.HasPermissionForAction(ctxt, ICON_CLASS_ID, ACTION) {
		mac_address, err := rest.SanitizeValidateQueryParameter(request, PARAMETER_MAC_ADDRESS)
		if err == nil {
			icons, err := iconRest.iconDb.FindAllIcons(mac_address, true)
			if err == nil {
				security.Audit(ctxt, ICON_CLASS_ID, ACTION, security.SUCCESS)
				rest.WriteEntitySafely(response, icons)
			} else {
				security.Audit(ctxt, ICON_CLASS_ID, ACTION, security.FAIL_ERROR)
				response.WriteError(http.StatusInternalServerError, err)
			}
		} else {
			security.Audit(ctxt, ICON_CLASS_ID, ACTION, security.FAIL_VALIDATION)
			response.WriteHeaderAndEntity(http.StatusBadRequest, err)
		}
	} else {
		security.Audit(ctxt, ICON_CLASS_ID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}

func (iconRest *IconRest) Update(request *restful.Request, response *restful.Response) {
	ACTION := moc.Icon_UPDATE.String()
	ctxt := request.Request.Context()
	if security.HasPermissionForAction(ctxt, ICON_CLASS_ID, ACTION) {
		iconRequest := new(moc.Icon)

		errs := rest.SanitizeValidateReadEntity(request, SCHEMA_ICON, iconRequest)
		iconId, err := rest.SanitizeValidatePathParameter(request, PARAMETER_ICON_ID)
		errs = append(errs, err...)
		if errs == nil {
			log.Debug("update an icon %s: %v", iconId, iconRequest)

			// Set id & timestamp
			iconRequest.Id = iconId
			iconRequest.Timestamp = tms.Now()
			// Set deleted flag as false as updating means that the icon is available
			iconRequest.Deleted = false

			err := iconRest.iconDb.UpsertIcon(iconRequest)
			if err == nil {
				icon, err := iconRest.iconDb.FindOneIcon(iconId, false)
				if err == nil {
					security.Audit(ctxt, ICON_CLASS_ID, ACTION, security.SUCCESS)
					rest.WriteHeaderAndEntitySafely(response, http.StatusCreated, icon)
				} else {
					log.Error("unexpected error: %v", err)
					security.Audit(ctxt, ICON_CLASS_ID, ACTION, security.FAIL_ERROR)
					response.WriteError(http.StatusInternalServerError, err)
				}
			} else {
				log.Error("unexpected error: %v", err)
				security.Audit(ctxt, ICON_CLASS_ID, ACTION, security.FAIL_ERROR)
				response.WriteError(http.StatusInternalServerError, err)
			}
		} else {
			security.Audit(ctxt, ICON_CLASS_ID, ACTION, security.FAIL_VALIDATION)
			response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		}
	} else {
		security.Audit(ctxt, ICON_CLASS_ID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}

func (iconRest *IconRest) Delete(request *restful.Request, response *restful.Response) {
	ACTION := moc.Icon_DELETE.String()
	ctxt := request.Request.Context()
	if security.HasPermissionForAction(ctxt, ICON_CLASS_ID, ACTION) {
		iconId, err := rest.SanitizeValidatePathParameter(request, PARAMETER_ICON_ID)
		if err == nil {
			log.Debug("delete an icon: %s", iconId)

			err := iconRest.iconDb.DeleteIcon(iconId)
			if err == nil {
				icon, err := iconRest.iconDb.FindOneIcon(iconId, true)
				if err == nil {
					security.Audit(ctxt, ICON_CLASS_ID, ACTION, security.SUCCESS)
					rest.WriteHeaderAndEntitySafely(response, http.StatusCreated, icon)
				} else {
					log.Error("unexpected error: %v", err)
					security.Audit(ctxt, ICON_CLASS_ID, ACTION, security.FAIL_ERROR)
					response.WriteError(http.StatusInternalServerError, err)
				}
			} else {
				log.Error("unexpected error: %v", err)
				security.Audit(ctxt, ICON_CLASS_ID, ACTION, security.FAIL_ERROR)
				response.WriteError(http.StatusInternalServerError, err)
			}
		} else {
			security.Audit(ctxt, ICON_CLASS_ID, ACTION, security.FAIL_VALIDATION)
			response.WriteHeaderAndEntity(http.StatusBadRequest, err)
		}
	} else {
		security.Audit(ctxt, ICON_CLASS_ID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}

func (iconRest *IconRest) CreateIconImage(request *restful.Request, response *restful.Response) {
	ACTION := moc.IconImage_CREATE.String()
	ctxt := request.Request.Context()
	if security.HasPermissionForAction(ctxt, ICON_CLASS_ID, ACTION) {
		iconImageRequest := new(moc.IconImage)

		// Set id & timestamp
		iconImageRequest.Id = mongo.CreateId()
		iconImageRequest.Timestamp = tms.Now()

		err := rest.SanitizeValidateReadEntity(request, SCHEMA_ICON_IMAGE, iconImageRequest)
		if err == nil {
			log.Debug("create an icon image: %v", iconImageRequest)

			err := iconRest.iconDb.UpsertIconImage(iconImageRequest)
			if err == nil {
				security.Audit(ctxt, ICON_IMAGE_CLASS_ID, ACTION, security.SUCCESS)
				rest.WriteEntitySafely(response, iconImageRequest)
			} else {
				log.Error("unexpected error: %v", err)
				security.Audit(ctxt, ICON_IMAGE_CLASS_ID, ACTION, security.FAIL_ERROR)
				response.WriteError(http.StatusInternalServerError, err)
			}
		} else {
			security.Audit(ctxt, ICON_IMAGE_CLASS_ID, ACTION, security.FAIL_VALIDATION)
			response.WriteHeaderAndEntity(http.StatusBadRequest, err)
		}
	} else {
		security.Audit(ctxt, ICON_IMAGE_CLASS_ID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}

func (iconRest *IconRest) GetIconImage(request *restful.Request, response *restful.Response) {
	ACTION := moc.IconImage_READ.String()
	ctxt := request.Request.Context()
	if security.HasPermissionForAction(ctxt, ICON_CLASS_ID, ACTION) {
		mac_address, err := rest.SanitizeValidateQueryParameter(request, PARAMETER_MAC_ADDRESS)
		if err == nil {
			iconImages, err := iconRest.iconDb.FindAllIconImages(mac_address)
			if err == nil {
				security.Audit(ctxt, ICON_IMAGE_CLASS_ID, ACTION, security.SUCCESS)
				rest.WriteEntitySafely(response, iconImages)
			} else {
				log.Error("unexpected error: %v", err)
				security.Audit(ctxt, ICON_IMAGE_CLASS_ID, ACTION, security.FAIL_ERROR)
				response.WriteError(http.StatusInternalServerError, err)
			}
		} else {
			security.Audit(ctxt, ICON_IMAGE_CLASS_ID, ACTION, security.FAIL_VALIDATION)
			response.WriteHeaderAndEntity(http.StatusBadRequest, err)
		}
	} else {
		security.Audit(ctxt, ICON_IMAGE_CLASS_ID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}
