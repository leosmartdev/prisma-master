package public

import (
	"net/http"
	"prisma/gogroup"
	"prisma/tms"

	api "prisma/tms/client_api"
	"prisma/tms/db"
	"prisma/tms/db/mongo"
	"prisma/tms/log"
	"prisma/tms/moc"
	"prisma/tms/rest"
	"prisma/tms/security"
	"strconv"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/go-openapi/spec"
	restful "github.com/orolia/go-restful"
)

var (
	paramRegistryID = spec.Parameter{
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
	queryTime = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name: "time",
			In:   "query",
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					MinLength: &[]int64{1}[0],
					MaxLength: &[]int64{20}[0],
					Pattern:   "[0-9]{1,20}",
				},
			},
		},
	}
)

type HistoryRest struct {
	trackdb db.TrackDB
	group   gogroup.GoGroup
}

// NewHistoryRest ...
func NewHistoryRest(client *mongo.MongoClient, group gogroup.GoGroup) *HistoryRest {
	return &HistoryRest{
		trackdb: mongo.NewMongoTracks(group, client),
		group:   group,
	}
}

// Get ...
func (r *HistoryRest) Get(request *restful.Request, response *restful.Response) {
	CLASSID := "History"
	ACTION := moc.History_GET.String()
	if !security.HasPermissionForAction(request.Request.Context(), CLASSID, ACTION) {
		log.Error("forbiden")
		security.Audit(request.Request.Context(), CLASSID, ACTION, "FAIL")
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
		return
	}
	infoResponse, errs := r.getTrack(request)
	if errs != nil {
		log.Error("fail to retreive track %+v", errs)
		security.Audit(request.Request.Context(), CLASSID, ACTION, "FAIL_ERROR")
		response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		return
	}
	security.Audit(request.Request.Context(), CLASSID, ACTION, "SUCCESS")
	rest.WriteHeaderAndEntitySafely(response, http.StatusOK, infoResponse)
}

func (r *HistoryRest) getTrack(request *restful.Request) (*InfoResponse, []rest.ErrorValidation) {
	var errs []rest.ErrorValidation
	registryID, errs := rest.SanitizeValidatePathParameter(request, paramRegistryID)
	if errs != nil {
		return nil, errs
	}
	// extract time from query
	qt, errs := rest.SanitizeValidateQueryParameter(request, queryTime)
	if errs != nil {
		return nil, errs
	}
	t, err := strconv.ParseInt(qt, 10, 64)
	if err != nil {
		return nil, []rest.ErrorValidation{
			rest.ErrorValidation{
				Property: "HISTORY",
				Rule:     "ParseTime",
				Message:  err.Error()}}
	}
	track, err := r.trackdb.GetLastTrack(bson.M{"$and": []bson.M{bson.M{"registry_id": registryID}, bson.M{"tgt": bson.M{"$exists": true}}, bson.M{"time": bson.M{"$lte": time.Unix(t, 0)}}}})
	if err != nil {
		return nil, []rest.ErrorValidation{
			rest.ErrorValidation{
				Property: "HISTORY",
				Rule:     "GetTrack",
				Message:  err.Error()}}
	}
	target := track.Targets[0]
	var metadata *tms.TrackMetadata
	if len(track.Metadata) > 0 {
		metadata = track.Metadata[0]
	}
	infoResponse := &InfoResponse{
		ID:         track.DatabaseId,
		TrackID:    track.Id,
		RegistryID: track.RegistryId,
		LookupID:   track.LookupID(),
		Target:     target,
		Metadata:   metadata,
	}

	return infoResponse, nil
}

func (r *HistoryRest) GetDatabase(request *restful.Request, response *restful.Response) {
	CLASSID := "History"
	ACTION := moc.History_GET.String()
	if !security.HasPermissionForAction(request.Request.Context(), CLASSID, ACTION) {
		security.Audit(request.Request.Context(), CLASSID, ACTION, "FAIL")
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
		return
	}

	databaseID, errs := rest.SanitizeValidatePathParameter(request, spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "database-id",
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
	})
	if errs != nil {
		security.Audit(request.Request.Context(), CLASSID, ACTION, "FAIL_ERROR")
		response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		return
	}

	result, err := r.trackdb.GetHistoricalTrack(db.GoHistoricalTrackRequest{
		Req: &api.HistoricalTrackRequest{
			DatabaseId: databaseID,
		},
		Ctxt: r.group,
	})
	if err != nil {
		security.Audit(request.Request.Context(), CLASSID, ACTION, "FAIL_ERROR")
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	track := result
	target := track.Targets[0]
	var metadata *tms.TrackMetadata
	if len(track.Metadata) > 0 {
		metadata = track.Metadata[0]
	}
	infoResponse := &InfoResponse{
		ID:         databaseID,
		TrackID:    track.Id,
		RegistryID: track.RegistryId,
		LookupID:   track.LookupID(),
		Target:     target,
		Metadata:   metadata,
	}
	security.Audit(request.Request.Context(), CLASSID, ACTION, "SUCCESS")
	rest.WriteHeaderAndEntitySafely(response, http.StatusOK, infoResponse)
}
