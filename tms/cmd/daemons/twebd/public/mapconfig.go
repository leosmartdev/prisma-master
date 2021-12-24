package public

import (
	"net/http"
	"time"

	"prisma/gogroup"
	"prisma/tms"
	"prisma/tms/db"
	"prisma/tms/db/mongo"
	"prisma/tms/log"
	"prisma/tms/moc"
	"prisma/tms/rest"
	"prisma/tms/security"
	"prisma/tms/tmsg"
	"prisma/tms/tmsg/client"

	"github.com/go-openapi/spec"
	"github.com/golang/protobuf/ptypes"
	restful "github.com/orolia/go-restful"
)

const (
	MAPCONFIG_CLASSID = "MapConfig"
	// OBJECT_NOTE         = "prisma.tms.moc." + FILTERTRACK_CLASSID
)

var (
	// interface for schema
	// schema
	SCHEMA_MAP_CONFIG = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Required: []string{"key"},
			Properties: map[string]spec.Schema{
				"key": {
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
				"value": {
					SwaggerSchemaProps: spec.SwaggerSchemaProps{
						ReadOnly: true,
					},
				},
			},
		},
	}
)

type MapConfigRest struct {
	client      *mongo.MongoClient
	mapconfigDb db.MapConfigDB
	group       gogroup.GoGroup
	tsiClient   client.TsiClient
}

func NewMapConfigRest(group gogroup.GoGroup, client *mongo.MongoClient) *MapConfigRest {
	miscDb := mongo.NewMongoMiscData(group, client)

	return &MapConfigRest{
		client:      client,
		mapconfigDb: mongo.NewMongoMapConfigDb(miscDb),
		group:       group,
		tsiClient:   tmsg.GClient,
	}
}

func (mapconfigRest *MapConfigRest) ReadAll(request *restful.Request, response *restful.Response) {
	ACTION := moc.MapConfig_GET.String()
	ctxt := request.Request.Context()
	if security.HasPermissionForAction(ctxt, MAPCONFIG_CLASSID, ACTION) {
		configData, err := mapconfigRest.mapconfigDb.FindAllMapConfig()
		if err == nil {
			mapconfig := make([]*moc.MapConfig, 0)
			for _, configDatum := range configData {
				if mocMapconfig, ok := configDatum.Contents.Data.(*moc.MapConfig); ok {
					mocMapconfig.Id = configDatum.Contents.ID
					mapconfig = append(mapconfig, mocMapconfig)
				}
			}
			security.Audit(ctxt, MAPCONFIG_CLASSID, ACTION, security.SUCCESS)
			rest.WriteEntitySafely(response, mapconfig)
		} else {
			log.Error("unexpected error: %v", err)
			security.Audit(ctxt, MAPCONFIG_CLASSID, ACTION, security.FAIL_ERROR)
			response.WriteError(http.StatusInternalServerError, err)
		}
	} else {
		security.Audit(ctxt, MAPCONFIG_CLASSID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}

func (mapconfigRest *MapConfigRest) SetSetting(request *restful.Request, response *restful.Response) {
	ACTION := moc.MapConfig_UPDATE.String()
	ctxt := request.Request.Context()
	if security.HasPermissionForAction(ctxt, MAPCONFIG_CLASSID, ACTION) {
		mapconfigRequest := new(moc.MapConfig)

		err := rest.SanitizeValidateReadEntity(request, SCHEMA_MAP_CONFIG, mapconfigRequest)
		if err == nil {
			// set id
			mapconfigRequest.Id = mongo.CreateId()

			upsertResponse, err := mapconfigRest.mapconfigDb.SaveMapConfig(mapconfigRequest)

			if err == nil {
				mapconfigRequest.Id = upsertResponse.Id
				if mapconfigRequest.Key == "track_timeouts" {
					err = mapconfigRest.TrackTimeoutReq(mapconfigRequest.Value)
					if err != nil {
						log.Error("Failed to send out track timeout request %+v", err)
					}
				}
				rest.WriteHeaderAndEntitySafely(response, http.StatusCreated, mapconfigRequest)
			} else {
				log.Error("unexpected error: %v", err)
				security.Audit(ctxt, MAPCONFIG_CLASSID, ACTION, security.FAIL_ERROR)
				response.WriteError(http.StatusInternalServerError, err)
			}
		} else {
			security.Audit(ctxt, MAPCONFIG_CLASSID, ACTION, security.FAIL_VALIDATION)
			response.WriteHeaderAndEntity(http.StatusBadRequest, err)
		}
	} else {
		security.Audit(ctxt, MAPCONFIG_CLASSID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}

// constructs a track timeout request that is send to tanalyzed
// function returns an error if it fails
func (mapconfigRest *MapConfigRest) TrackTimeoutReq(trackTimeout *moc.TrackTimeout) error {
	body, err := tmsg.PackFrom(trackTimeout)
	if err != nil {
		return err
	}
	pnow, err := ptypes.TimestampProto(time.Now())
	if err != nil {
		return err
	}
	mapconfigRest.tsiClient.Send(mapconfigRest.group, &tms.TsiMessage{
		Source: mapconfigRest.tsiClient.Local(),
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
