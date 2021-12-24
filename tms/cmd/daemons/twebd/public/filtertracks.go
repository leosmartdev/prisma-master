package public

import (
	"net/http"

	"prisma/gogroup"
	"prisma/tms/db"
	"prisma/tms/db/mongo"
	"prisma/tms/log"
	"prisma/tms/moc"
	"prisma/tms/rest"
	"prisma/tms/security"
	"prisma/tms/tmsg"
	"prisma/tms/tmsg/client"

	"github.com/go-openapi/spec"
	restful "github.com/orolia/go-restful"
)

const (
	FILTERTRACKS_CLASSID = "FilterTracks"
	// OBJECT_NOTE         = "prisma.tms.moc." + FILTERTRACKS_CLASSID
)

var (
	// interface for schema
	// schema
	SCHEMA_FILTER_TRACKS = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Required: []string{"filterTracks"},
			Properties: map[string]spec.Schema{
				"filterTracks": {
					SwaggerSchemaProps: spec.SwaggerSchemaProps{
						ReadOnly: true,
					},
				},
			},
		},
	}
)

type FilterTracksRest struct {
	client         *mongo.MongoClient
	filtertracksDb db.FilterTracksDB
	group          gogroup.GoGroup
	tsiClient      client.TsiClient
}

func NewFilterTracksRest(group gogroup.GoGroup, client *mongo.MongoClient) *FilterTracksRest {
	miscDb := mongo.NewMongoMiscData(group, client)

	return &FilterTracksRest{
		client:         client,
		filtertracksDb: mongo.NewMongoFilterTracksDb(miscDb),
		group:          group,
		tsiClient:      tmsg.GClient,
	}
}

func (FilterTracksRest *FilterTracksRest) GetFilterTracks(request *restful.Request, response *restful.Response) {
	ACTION := moc.FilterTracks_GET.String()
	ctxt := request.Request.Context()
	if security.HasPermissionForAction(ctxt, FILTERTRACKS_CLASSID, ACTION) {
		userId, err := rest.SanitizeValidatePathParameter(request, PARAMETER_USER_ID)
		if err == nil {
			filterTracks, err := FilterTracksRest.filtertracksDb.GetFilterTracks(userId)
			if err == nil {
				security.Audit(ctxt, FILTERTRACKS_CLASSID, ACTION, security.SUCCESS)
				rest.WriteEntitySafely(response, filterTracks)
			} else {
				log.Error("unexpected error: %v", err)
				security.Audit(ctxt, FILTERTRACKS_CLASSID, ACTION, security.FAIL_ERROR)
				response.WriteError(http.StatusInternalServerError, err)
			}
		} else {
			security.Audit(ctxt, FILTERTRACKS_CLASSID, ACTION, security.FAIL_VALIDATION, userId)
			response.WriteHeaderAndEntity(http.StatusBadRequest, err)
		}
	} else {
		security.Audit(ctxt, FILTERTRACKS_CLASSID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}

func (FilterTracksRest *FilterTracksRest) SaveFilterTracks(request *restful.Request, response *restful.Response) {
	ACTION := moc.FilterTracks_UPDATE.String()
	ctxt := request.Request.Context()
	if security.HasPermissionForAction(ctxt, FILTERTRACKS_CLASSID, ACTION) {
		filterTrackRequest := new(moc.FilterTrackSet)

		err := rest.SanitizeValidateReadEntity(request, SCHEMA_FILTER_TRACKS, filterTrackRequest)
		userId, err2 := rest.SanitizeValidatePathParameter(request, PARAMETER_USER_ID)
		err = append(err, err2...)
		if err == nil {
			filterTracks := make([]*moc.FilterTracks, 0)
			for _, item := range filterTrackRequest.FilterTracks {
				filterTrack := new(moc.FilterTracks)
				filterTrack.Id = mongo.CreateId()
				filterTrack.User = userId
				filterTrack.Show = item.Show
				filterTrack.Timeout = item.Timeout
				filterTrack.Type = item.Type
				upsertResponse, err := FilterTracksRest.filtertracksDb.SaveFilterTrack(filterTrack)
				if err == nil {
					filterTrack.Id = upsertResponse.Id
					filterTracks = append(filterTracks, filterTrack)
				}
			}
			rest.WriteHeaderAndEntitySafely(response, http.StatusCreated, filterTracks)
		} else {
			security.Audit(ctxt, FILTERTRACKS_CLASSID, ACTION, security.FAIL_VALIDATION)
			response.WriteHeaderAndEntity(http.StatusBadRequest, err)
		}
	} else {
		security.Audit(ctxt, FILTERTRACKS_CLASSID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}
