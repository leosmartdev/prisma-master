package public

import (
	"net/http"
	"prisma/gogroup"
	"prisma/tms/db"
	"prisma/tms/db/mongo"
	"prisma/tms/moc"
	"prisma/tms/rest"
	"prisma/tms/security"
	"strconv"

	restful "github.com/orolia/go-restful" 
	"github.com/go-openapi/spec"
)

type RegistryRest struct {
	regdb db.RegistryDB
	group gogroup.GoGroup
}

var (
	PARAMETER_REGISTRY_ID = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "registry-id",
			In:       "path",
			Required: true,
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					MinLength: &[]int64{32}[0],
					MaxLength: &[]int64{32}[0],
					Pattern:   "[0-9a-fA-F]{32}",
				},
			},
		},
	}
	PARAMETER_REGISTRY_QUERY = spec.Parameter{
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
	PARAMETER_REGISTRY_LIMIT = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "limit",
			In:       "query",
			Required: false,
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					MaxLength: &[]int64{3}[0],
					Pattern:   "[0-9]{1,3}",
				},
			},
		},
	}
)

func NewRegistryRest(client *mongo.MongoClient, group gogroup.GoGroup) *RegistryRest {
	return &RegistryRest{
		regdb: mongo.NewMongoRegistry(group, client),
		group: group,
	}
}

func (r *RegistryRest) Get(request *restful.Request, response *restful.Response) {
	CLASSID := "Registry"
	ACTION := moc.Registry_GET.String()
	if !security.HasPermissionForAction(request.Request.Context(), CLASSID, ACTION) {
		security.Audit(request.Request.Context(), CLASSID, ACTION, "FAIL")
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
		return
	}
	registryID, errs := rest.SanitizeValidatePathParameter(request, PARAMETER_REGISTRY_ID)
	if errs != nil {
		security.Audit(request.Request.Context(), CLASSID, ACTION, "FAIL_ERROR")
		response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		return
	}

	entry, err := r.regdb.Get(registryID)
	if !errorFree(err, request, response, CLASSID, ACTION) {
		return
	}
	incidents := []string{}
	if entry.Incidents != nil {
		incidents = entry.Incidents
	}
	infoResponse := &InfoResponse{
		ID:         registryID,
		RegistryID: registryID,
		LookupID:   registryID,
		Target:     entry.Target,
		Metadata:   entry.Metadata,
		Registry: &Registry{
			Incidents: incidents,
		},
	}
	security.Audit(request.Request.Context(), CLASSID, ACTION, "SUCCESS")
	rest.WriteHeaderAndEntitySafely(response, http.StatusOK, infoResponse)
}

func (r *RegistryRest) Search(request *restful.Request, response *restful.Response) {
	CLASSID := "Registry"
	ACTION := moc.Registry_GET.String()
	if !security.HasPermissionForAction(request.Request.Context(), CLASSID, ACTION) {
		security.Audit(request.Request.Context(), CLASSID, ACTION, "FAIL")
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
		return
	}
	query, errs := rest.SanitizeValidateQueryParameter(request, PARAMETER_REGISTRY_QUERY)
	if errs != nil {
		security.Audit(request.Request.Context(), CLASSID, ACTION, "FAIL_ERROR")
		response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		return
	}
	slimit, errs := rest.SanitizeValidateQueryParameter(request, PARAMETER_REGISTRY_LIMIT)
	if errs != nil {
		security.Audit(request.Request.Context(), CLASSID, ACTION, "FAIL_ERROR")
		response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		return
	}
	limit, _ := strconv.Atoi(slimit)
	if limit == 0 {
		limit = 1000
	}
	results, err := r.regdb.Search(db.RegistrySearchRequest{
		Query: query,
		Limit: limit,
	})
	if err != nil {
		security.Audit(request.Request.Context(), CLASSID, ACTION, "FAIL_ERROR")
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	security.Audit(request.Request.Context(), CLASSID, ACTION, "SUCCESS")
	rest.WriteHeaderAndEntitySafely(response, http.StatusOK, results)
}
