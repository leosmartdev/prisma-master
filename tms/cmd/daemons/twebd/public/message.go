package public

import (
	"net/http"
	"prisma/gogroup"
	"prisma/tms/db"
	"prisma/tms/db/mongo"
	"prisma/tms/moc"
	"prisma/tms/rest"
	"prisma/tms/sar"
	"prisma/tms/security"
	"sort"
	"strconv"

	"github.com/go-openapi/spec"
	"github.com/golang/protobuf/ptypes/timestamp"
	restful "github.com/orolia/go-restful"
)

const (
	MESSAGE_CLASS_ID = "Message"
)

var (
	// parameters with schema
	PARAMETER_MESSAGE_SIT_NUMBER = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "sit-number",
			In:       "query",
			Required: true,
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					MinLength: &[]int64{1}[0],
					Pattern:   "[0-9]{1,3}",
				},
			},
		},
		SimpleSchema: spec.SimpleSchema{
			Type: "string",
		},
	}

	PARAMETER_MESSAGE_START_DATETIME = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "start-datetime",
			In:       "query",
			Required: true,
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					MinLength: &[]int64{1}[0],
				},
			},
		},
		SimpleSchema: spec.SimpleSchema{
			Type: "string",
		},
	}

	PARAMETER_MESSAGE_END_DATETIME = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "end-datetime",
			In:       "query",
			Required: true,
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					MinLength: &[]int64{1}[0],
				},
			},
		},
		SimpleSchema: spec.SimpleSchema{
			Type: "string",
		},
	}

	PARAMETER_MESSAGE_DIRECTION = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "direction",
			In:       "query",
			Required: true,
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					MinLength: &[]int64{1}[0],
				},
			},
		},
		SimpleSchema: spec.SimpleSchema{
			Type: "string",
		},
	}
)

type MessageResponse struct {
	Id           string
	SitNumber    int
	MessageBody  string
	Direction    int
	CommLinkType string
	Cscode       string
	Csname       string
	ErrorDetail  string
	Dismiss      bool
	Time         *timestamp.Timestamp
}

type MessageRest struct {
	client       *mongo.MongoClient
	group        gogroup.GoGroup
	activityDb   db.ActivityDB
	registryDb   db.RegistryDB
	sit915Db     db.Sit915DB
	remoteSiteDb db.RemoteSiteDB
	configDb     mongo.ConfigDb
}

func NewMessageRest(group gogroup.GoGroup, client *mongo.MongoClient) *MessageRest {
	miscDb := mongo.NewMongoMiscData(group, client)

	return &MessageRest{
		client:       client,
		group:        group,
		activityDb:   mongo.NewMongoActivities(group, client),
		registryDb:   mongo.NewMongoRegistry(group, client),
		sit915Db:     mongo.NewSit915Db(miscDb),
		remoteSiteDb: mongo.NewMongoRemoteSiteMiscData(miscDb),
		configDb:     mongo.ConfigDb{},
	}
}

func (messageRest *MessageRest) GetAll(request *restful.Request, response *restful.Response) {
	ACTION := sar.SarsatMessage_READ.String()
	ctxt := request.Request.Context()
	if !security.HasPermissionForAction(ctxt, MESSAGE_CLASS_ID, ACTION) {
		security.Audit(ctxt, MESSAGE_CLASS_ID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
		return
	}

	paramSitNumber, errs := rest.SanitizeValidateQueryParameter(request, PARAMETER_MESSAGE_SIT_NUMBER)

	paramStartDateTime, errs := rest.SanitizeValidateQueryParameter(request, PARAMETER_MESSAGE_START_DATETIME)
	errs = append(errs, errs...)

	paramEndDateTime, errs := rest.SanitizeValidateQueryParameter(request, PARAMETER_MESSAGE_END_DATETIME)
	errs = append(errs, errs...)

	paramDirection, errs := rest.SanitizeValidateQueryParameter(request, PARAMETER_MESSAGE_DIRECTION)
	errs = append(errs, errs...)

	if errs != nil {
		security.Audit(ctxt, MESSAGE_CLASS_ID, ACTION, security.FAIL_VALIDATION)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		return
	}

	sitNumber, err := strconv.Atoi(paramSitNumber)
	if err != nil {
		security.Audit(ctxt, MESSAGE_CLASS_ID, ACTION, security.FAIL_ERROR)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
		return
	}

	startDateTime, err := strconv.Atoi(paramStartDateTime)
	if err != nil {
		security.Audit(ctxt, MESSAGE_CLASS_ID, ACTION, security.FAIL_ERROR)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
		return
	}

	endDateTime, err := strconv.Atoi(paramEndDateTime)
	if err != nil {
		security.Audit(ctxt, MESSAGE_CLASS_ID, ACTION, security.FAIL_ERROR)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
		return
	}

	direction, err := strconv.Atoi(paramDirection)
	if err != nil {
		security.Audit(ctxt, MESSAGE_CLASS_ID, ACTION, security.FAIL_ERROR)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
		return
	}

	result := make([]MessageResponse, 0)

	// SIT NUMBER
	// 0	:	ALL
	// 185:	SIT 185
	// 915:	SIT 915
	if (sitNumber == 0 || sitNumber == 185) && (direction == 0 || direction == 2) {
		registries, err := messageRest.registryDb.GetSit185Messages(startDateTime, endDateTime)
		if err != nil {
			security.Audit(ctxt, MESSAGE_CLASS_ID, ACTION, security.FAIL_ERROR)
			response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
			return
		}

		for _, registry := range registries {
			target := registry.Target

			cscode := target.Sarmsg.RemoteName
			csname, err := messageRest.findCsnameByCode(cscode)
			if err != nil && err != db.ErrorNotFound {
				security.Audit(ctxt, MESSAGE_CLASS_ID, ACTION, security.FAIL_ERROR)
				response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
				return
			}

			message := MessageResponse{
				Id:           registry.DatabaseId,
				SitNumber:    185,
				MessageBody:  target.Sarmsg.MessageBody,
				Direction:    2,
				CommLinkType: target.Sarmsg.Protocol,
				Cscode:       cscode,
				Csname:       csname,
				Dismiss:      false,
				Time:         target.Time,
			}

			result = append(result, message)
		}
	}

	if sitNumber == 0 || sitNumber == 915 {
		// DIRECION
		// 0	:	ALL
		// 1	:	SENT
		// 2	:	RECEIVED
		// 3  : PENDING
		// 4  : FAILED
		if direction == 0 || direction == 1 || direction == 3 || direction == 4 {
			sit915s, err := messageRest.sit915Db.FindAllSit915s(startDateTime, endDateTime, direction)
			if err != nil {
				security.Audit(ctxt, MESSAGE_CLASS_ID, ACTION, security.FAIL_ERROR)
				response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
				return
			}

			for _, sit915 := range sit915s {
				remoteSite, err := messageRest.remoteSiteDb.FindOneRemoteSite(sit915.RemotesiteId, false)
				if err != nil && err != db.ErrorNotFound {
					security.Audit(ctxt, MESSAGE_CLASS_ID, ACTION, security.FAIL_ERROR)
					response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
					return
				}

				cscode := ""
				csname := ""
				if err != db.ErrorNotFound {
					cscode = remoteSite.Cscode
					csname = remoteSite.Csname
				}

				var messageDirection int
				if sit915.Status == moc.Sit915_SENT.String() {
					messageDirection = 1
				} else if sit915.Status == moc.Sit915_PENDING.String() {
					messageDirection = 3
				} else if sit915.Status == moc.Sit915_FAILED.String() {
					messageDirection = 4
				}

				messageBody := sit915.MessageBody
				if messageDirection == 3 || messageDirection == 4 {
					messageBody, err = GenerateSit915Message(messageRest.group, messageRest.sit915Db, messageRest.remoteSiteDb, messageRest.configDb, sit915.Id)
					if err != nil {
						security.Audit(ctxt, MESSAGE_CLASS_ID, ACTION, security.FAIL_ERROR)
						response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
						return
					}
				}

				message := MessageResponse{
					Id:           sit915.Id,
					SitNumber:    915,
					MessageBody:  messageBody,
					Direction:    messageDirection,
					CommLinkType: sit915.CommLinkType,
					Cscode:       cscode,
					Csname:       csname,
					Dismiss:      sit915.Dismiss,
					ErrorDetail:  sit915.ErrorDetail,
					Time:         sit915.Timestamp,
				}

				result = append(result, message)
			}
		}

		if direction == 0 || direction == 2 {
			activities, err := messageRest.activityDb.GetSit915Messages(startDateTime, endDateTime)
			if err != nil {
				security.Audit(ctxt, MESSAGE_CLASS_ID, ACTION, security.FAIL_ERROR)
				response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
				return
			}

			for _, activity := range activities {
				cscode := activity.GetSarsat().RemoteName
				csname, err := messageRest.findCsnameByCode(cscode)
				if err != nil && err != db.ErrorNotFound {
					security.Audit(ctxt, MESSAGE_CLASS_ID, ACTION, security.FAIL_ERROR)
					response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
					return
				}

				message := MessageResponse{
					Id:           activity.DatabaseId,
					SitNumber:    915,
					MessageBody:  activity.GetSarsat().MessageBody,
					Direction:    2,
					CommLinkType: activity.GetSarsat().Protocol,
					Cscode:       cscode,
					Csname:       csname,
					Dismiss:      false,
					Time:         activity.Time,
				}

				result = append(result, message)
			}
		}
	}

	// Sort the list of messages by date
	sort.Slice(result, func(i, j int) bool {
		return result[i].Time.Seconds > result[j].Time.Seconds
	})

	security.Audit(ctxt, MESSAGE_CLASS_ID, ACTION, security.SUCCESS)
	rest.WriteEntitySafely(response, result)
}

func (messageRest *MessageRest) findCsnameByCode(cscode string) (string, error) {
	remoteSite, err := messageRest.remoteSiteDb.FindOneRemoteSiteByCscode(cscode)
	if err != nil {
		return "", err
	}

	return remoteSite.Csname, nil
}
