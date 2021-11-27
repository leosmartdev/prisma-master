package public

import (
	"net/http"
	"prisma/gogroup"
	"prisma/tms"
	"prisma/tms/db"
	"prisma/tms/db/mongo"
	"prisma/tms/moc"
	"prisma/tms/rest"
	"prisma/tms/security"

	"github.com/go-openapi/spec"
	restful "github.com/orolia/go-restful"
)

const (
	REMOTESITE_CLASS_ID = "RemoteSite"
)

var (
	// parameters with schema
	PARAMETER_REMOTESITE_ID = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "remotesite-id",
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

	// schemas
	SCHEMA_REMOTESITE = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Properties: map[string]spec.Schema{
				"type": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{0}[0],
						MaxLength: &[]int64{255}[0],
					},
				},
				"name": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{0}[0],
						MaxLength: &[]int64{255}[0],
					},
				},
				"description": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{0}[0],
						MaxLength: &[]int64{1024}[0],
					},
				},
				"address": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{0}[0],
						MaxLength: &[]int64{255}[0],
					},
				},
				"country": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{0}[0],
						MaxLength: &[]int64{3}[0],
						Pattern:   "[A-Z]{0,3}",
					},
				},
				"cscode": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{0}[0],
						MaxLength: &[]int64{4}[0],
						Pattern:   "[0-9A-Z]{0,4}",
					},
				},
				"csname": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{0}[0],
						MaxLength: &[]int64{8}[0],
						Pattern:   "[0-9A-Z]{0,8}",
					},
				},
				"ftp_communication": {
					SwaggerSchemaProps: spec.SwaggerSchemaProps{
						ReadOnly: true,
					},
				},
			},
		},
	}
)

type RemoteSiteRest struct {
	remoteSiteDb db.RemoteSiteDB
	client       *mongo.MongoClient
	group        gogroup.GoGroup
}

func NewRemoteSiteRest(group gogroup.GoGroup, client *mongo.MongoClient) *RemoteSiteRest {
	miscDb := mongo.NewMongoMiscData(group, client)

	return &RemoteSiteRest{
		remoteSiteDb: mongo.NewMongoRemoteSiteMiscData(miscDb),
		client:       client,
		group:        group,
	}
}

func (remoteSiteRest *RemoteSiteRest) Create(request *restful.Request, response *restful.Response) {
	ACTION := moc.RemoteSite_CREATE.String()
	ctxt := request.Request.Context()
	if !security.HasPermissionForAction(ctxt, REMOTESITE_CLASS_ID, ACTION) {
		security.Audit(ctxt, REMOTESITE_CLASS_ID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
		return
	}

	remoteSiteRequest := &moc.RemoteSite{}
	errs := rest.SanitizeValidateReadEntity(request, SCHEMA_REMOTESITE, remoteSiteRequest)
	if errs != nil {
		security.Audit(ctxt, REMOTESITE_CLASS_ID, ACTION, security.FAIL_VALIDATION)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		return
	}

	// Set id & timestamp
	remoteSiteRequest.Id = mongo.CreateId()
	remoteSiteRequest.Timestamp = tms.Now()

	remoteSiteRequest.CurrentMessageNum = 1

	remoteSiteRequest.CommLinkTypes = []*moc.CommLinkType{
		{
			Name:    "FTP",
			Enabled: true,
		},
	}

	err := remoteSiteRest.remoteSiteDb.UpsertRemoteSite(remoteSiteRequest)
	if err != nil {
		security.Audit(ctxt, REMOTESITE_CLASS_ID, ACTION, security.FAIL_ERROR)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
		return
	}

	security.Audit(ctxt, REMOTESITE_CLASS_ID, ACTION, security.SUCCESS)
	rest.WriteHeaderAndEntitySafely(response, http.StatusCreated, remoteSiteRequest)
}

func (remoteSiteRest *RemoteSiteRest) GetAll(request *restful.Request, response *restful.Response) {
	ACTION := moc.RemoteSite_READ.String()
	ctxt := request.Request.Context()
	if !security.HasPermissionForAction(ctxt, REMOTESITE_CLASS_ID, ACTION) {
		security.Audit(ctxt, REMOTESITE_CLASS_ID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
		return
	}

	remoteSites, err := remoteSiteRest.remoteSiteDb.FindAllRemoteSites()
	if err != nil {
		security.Audit(ctxt, REMOTESITE_CLASS_ID, ACTION, security.FAIL_ERROR)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
		return
	}

	security.Audit(ctxt, REMOTESITE_CLASS_ID, ACTION, security.SUCCESS)
	rest.WriteEntitySafely(response, remoteSites)
}

func (remoteSiteRest *RemoteSiteRest) Get(request *restful.Request, response *restful.Response) {
	ACTION := moc.RemoteSite_READ.String()
	ctxt := request.Request.Context()
	if !security.HasPermissionForAction(ctxt, REMOTESITE_CLASS_ID, ACTION) {
		security.Audit(ctxt, REMOTESITE_CLASS_ID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
		return
	}

	id, errs := rest.SanitizeValidatePathParameter(request, PARAMETER_REMOTESITE_ID)
	if errs != nil {
		security.Audit(ctxt, REMOTESITE_CLASS_ID, ACTION, security.FAIL_VALIDATION)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		return
	}

	remoteSite, err := remoteSiteRest.remoteSiteDb.FindOneRemoteSite(id, false)
	if err != nil {
		security.Audit(ctxt, REMOTESITE_CLASS_ID, ACTION, security.FAIL_ERROR)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
		return
	}

	security.Audit(ctxt, REMOTESITE_CLASS_ID, ACTION, security.SUCCESS)
	rest.WriteEntitySafely(response, remoteSite)
}

func (remoteSiteRest *RemoteSiteRest) Update(request *restful.Request, response *restful.Response) {
	ACTION := moc.RemoteSite_UPDATE.String()
	ctxt := request.Request.Context()
	if !security.HasPermissionForAction(ctxt, REMOTESITE_CLASS_ID, ACTION) {
		security.Audit(ctxt, REMOTESITE_CLASS_ID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
		return
	}

	remoteSiteRequest := &moc.RemoteSite{}
	errs := rest.SanitizeValidateReadEntity(request, SCHEMA_REMOTESITE, remoteSiteRequest)
	if errs != nil {
		security.Audit(ctxt, REMOTESITE_CLASS_ID, ACTION, security.FAIL_VALIDATION)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		return
	}

	id, errs := rest.SanitizeValidatePathParameter(request, PARAMETER_REMOTESITE_ID)
	if errs != nil {
		security.Audit(ctxt, REMOTESITE_CLASS_ID, ACTION, security.FAIL_VALIDATION)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		return
	}

	// Set id & timestamp
	remoteSiteRequest.Id = id
	remoteSiteRequest.Timestamp = tms.Now()

	err := remoteSiteRest.remoteSiteDb.UpsertRemoteSite(remoteSiteRequest)
	if err != nil {
		security.Audit(ctxt, REMOTESITE_CLASS_ID, ACTION, security.FAIL_ERROR)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
		return
	}

	security.Audit(ctxt, REMOTESITE_CLASS_ID, ACTION, security.SUCCESS)
	rest.WriteHeaderAndEntitySafely(response, http.StatusCreated, remoteSiteRequest)
}

func (remoteSiteRest *RemoteSiteRest) Delete(request *restful.Request, response *restful.Response) {
	ACTION := moc.RemoteSite_DELETE.String()
	ctxt := request.Request.Context()
	if !security.HasPermissionForAction(ctxt, REMOTESITE_CLASS_ID, ACTION) {
		security.Audit(ctxt, REMOTESITE_CLASS_ID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
		return
	}

	id, errs := rest.SanitizeValidatePathParameter(request, PARAMETER_REMOTESITE_ID)
	if errs != nil {
		security.Audit(ctxt, REMOTESITE_CLASS_ID, ACTION, security.FAIL_VALIDATION)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		return
	}

	err := remoteSiteRest.remoteSiteDb.DeleteRemoteSite(id)
	if err != nil {
		security.Audit(ctxt, REMOTESITE_CLASS_ID, ACTION, security.FAIL_ERROR)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
		return
	}

	remoteSite, err := remoteSiteRest.remoteSiteDb.FindOneRemoteSite(id, true)
	if err != nil {
		security.Audit(ctxt, REMOTESITE_CLASS_ID, ACTION, security.FAIL_ERROR)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
		return
	}

	security.Audit(ctxt, REMOTESITE_CLASS_ID, ACTION, security.SUCCESS)
	rest.WriteHeaderAndEntitySafely(response, http.StatusCreated, remoteSite)
}
