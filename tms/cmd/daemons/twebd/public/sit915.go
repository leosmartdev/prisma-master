package public

import (
	"fmt"
	"net/http"
	"prisma/gogroup"
	"prisma/tms"
	"prisma/tms/db"
	"prisma/tms/db/mongo"
	"prisma/tms/moc"
	"prisma/tms/rest"
	"prisma/tms/security"
	"strconv"
	"time"

	"github.com/go-openapi/spec"
	restful "github.com/orolia/go-restful"
)

const (
	SIT915_CLASS_ID = "Sit915"
)

var (
	// parameters with schema
	PARAMETER_SIT915_REMOTESITE_ID = spec.Parameter{
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

	PARAMETER_COMM_LINK_TYPE = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "comm-link-type",
			In:       "path",
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

	PARAMETER_SIT915_MESSAGE_ID = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "message-id",
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
	SCHEMA_SIT915 = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Required: []string{"narrative"},
			Properties: map[string]spec.Schema{
				"narrative": {
					SwaggerSchemaProps: spec.SwaggerSchemaProps{
						ReadOnly: true,
					},
				},
			},
		},
	}
)

type Sit915Response struct {
	Sit915     *moc.Sit915     `json:"sit915"`
	RemoteSite *moc.RemoteSite `json:"remoteSite"`
}

type Sit915Rest struct {
	client       *mongo.MongoClient
	group        gogroup.GoGroup
	sit915Db     db.Sit915DB
	remoteSiteDb db.RemoteSiteDB
	configDb     mongo.ConfigDb
}

func NewSit915Rest(group gogroup.GoGroup, client *mongo.MongoClient) *Sit915Rest {
	miscDb := mongo.NewMongoMiscData(group, client)

	return &Sit915Rest{
		client:       client,
		group:        group,
		sit915Db:     mongo.NewSit915Db(miscDb),
		remoteSiteDb: mongo.NewMongoRemoteSiteMiscData(miscDb),
		configDb:     mongo.ConfigDb{},
	}
}

func (sit915Rest *Sit915Rest) Create(request *restful.Request, response *restful.Response) {
	ACTION := moc.Sit915_CREATE.String()
	ctxt := request.Request.Context()
	if !security.HasPermissionForAction(ctxt, SIT915_CLASS_ID, ACTION) {
		security.Audit(ctxt, SIT915_CLASS_ID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
		return
	}

	remotesite_id, errs := rest.SanitizeValidatePathParameter(request, PARAMETER_SIT915_REMOTESITE_ID)
	if errs != nil {
		security.Audit(ctxt, SIT915_CLASS_ID, ACTION, security.FAIL_VALIDATION)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		return
	}

	comm_link_type, errs := rest.SanitizeValidatePathParameter(request, PARAMETER_COMM_LINK_TYPE)
	if errs != nil {
		security.Audit(ctxt, SIT915_CLASS_ID, ACTION, security.FAIL_VALIDATION)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		return
	}

	sit915Request := &moc.Sit915{}
	errs = rest.SanitizeValidateReadEntity(request, SCHEMA_SIT915, sit915Request)
	if errs != nil {
		security.Audit(ctxt, SIT915_CLASS_ID, ACTION, security.FAIL_VALIDATION)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		return
	}

	remoteSite, err := sit915Rest.remoteSiteDb.FindOneRemoteSite(remotesite_id, false)
	if err != nil {
		security.Audit(ctxt, REMOTESITE_CLASS_ID, ACTION, security.FAIL_ERROR)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
		return
	}

	sit915Request.Id = mongo.CreateId()
	sit915Request.RemotesiteId = remotesite_id
	sit915Request.TransmissionNum = remoteSite.CurrentMessageNum
	sit915Request.RetransmissionNum = 0
	sit915Request.Status = moc.Sit915_PENDING.String()
	sit915Request.CommLinkType = comm_link_type
	sit915Request.Timestamp = tms.Now()

	err = sit915Rest.sit915Db.UpsertSit915(sit915Request)
	if err != nil {
		security.Audit(ctxt, SIT915_CLASS_ID, ACTION, security.FAIL_ERROR)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
		return
	}

	security.Audit(ctxt, SIT915_CLASS_ID, ACTION, security.SUCCESS)
	rest.WriteHeaderAndEntitySafely(response, http.StatusCreated, sit915Request)
}

func (sit915Rest *Sit915Rest) Retry(request *restful.Request, response *restful.Response) {
	ACTION := moc.Sit915_UPDATE.String()
	ctxt := request.Request.Context()
	if !security.HasPermissionForAction(ctxt, SIT915_CLASS_ID, ACTION) {
		security.Audit(ctxt, SIT915_CLASS_ID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
		return
	}

	message_id, errs := rest.SanitizeValidatePathParameter(request, PARAMETER_SIT915_MESSAGE_ID)
	if errs != nil {
		security.Audit(ctxt, SIT915_CLASS_ID, ACTION, security.FAIL_VALIDATION)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		return
	}

	sit915, err := sit915Rest.sit915Db.FindOneSit915(message_id)
	if err != nil {
		if db.ErrorNotFound == err {
			security.Audit(ctxt, SIT915_CLASS_ID, ACTION, security.FAIL_NOTFOUND)
			response.WriteError(http.StatusNotFound, err)
		} else {
			security.Audit(ctxt, SIT915_CLASS_ID, ACTION, security.FAIL_ERROR)
			response.WriteError(http.StatusInternalServerError, err)
		}

		return
	}

	remoteSite, err := sit915Rest.remoteSiteDb.FindOneRemoteSite(sit915.RemotesiteId, false)
	if err != nil {
		security.Audit(ctxt, REMOTESITE_CLASS_ID, ACTION, security.FAIL_ERROR)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
		return
	}

	sit915.Status = moc.Sit915_PENDING.String()
	sit915.TransmissionNum = remoteSite.CurrentMessageNum
	sit915.Dismiss = false
	sit915.RetransmissionNum += 1
	if sit915.RetransmissionNum == 100000 {
		sit915.RetransmissionNum = 1
	}

	err = sit915Rest.sit915Db.UpsertSit915(sit915)
	if err != nil {
		security.Audit(ctxt, SIT915_CLASS_ID, ACTION, security.FAIL_ERROR)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
		return
	}

	sit915.MessageBody, err = GenerateSit915Message(sit915Rest.group, sit915Rest.sit915Db, sit915Rest.remoteSiteDb, sit915Rest.configDb, message_id)
	if err != nil {
		security.Audit(ctxt, SIT915_CLASS_ID, ACTION, security.FAIL_ERROR)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
		return
	}

	security.Audit(ctxt, SIT915_CLASS_ID, ACTION, security.SUCCESS)
	rest.WriteHeaderAndEntitySafely(response, http.StatusCreated, sit915)
}

func (sit915Rest *Sit915Rest) Ack(request *restful.Request, response *restful.Response) {
	ACTION := moc.Sit915_UPDATE.String()
	ctxt := request.Request.Context()
	if !security.HasPermissionForAction(ctxt, SIT915_CLASS_ID, ACTION) {
		security.Audit(ctxt, SIT915_CLASS_ID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
		return
	}

	message_id, errs := rest.SanitizeValidatePathParameter(request, PARAMETER_SIT915_MESSAGE_ID)
	if errs != nil {
		security.Audit(ctxt, SIT915_CLASS_ID, ACTION, security.FAIL_VALIDATION)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		return
	}

	sit915, err := sit915Rest.sit915Db.FindOneSit915(message_id)
	if err != nil {
		if db.ErrorNotFound == err {
			security.Audit(ctxt, SIT915_CLASS_ID, ACTION, security.FAIL_NOTFOUND)
			response.WriteError(http.StatusNotFound, err)
		} else {
			security.Audit(ctxt, SIT915_CLASS_ID, ACTION, security.FAIL_ERROR)
			response.WriteError(http.StatusInternalServerError, err)
		}

		return
	}

	sit915.Dismiss = true

	err = sit915Rest.sit915Db.UpsertSit915(sit915)
	if err != nil {
		security.Audit(ctxt, SIT915_CLASS_ID, ACTION, security.FAIL_ERROR)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
		return
	}

	security.Audit(ctxt, SIT915_CLASS_ID, ACTION, security.SUCCESS)
	rest.WriteHeaderAndEntitySafely(response, http.StatusCreated, sit915)
}

func (sit915Rest *Sit915Rest) GetAll(request *restful.Request, response *restful.Response) {
	ACTION := moc.Sit915_READ.String()
	ctxt := request.Request.Context()
	if !security.HasPermissionForAction(ctxt, SIT915_CLASS_ID, ACTION) {
		security.Audit(ctxt, SIT915_CLASS_ID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
		return
	}

	sit915s, err := sit915Rest.sit915Db.FindAllSit915s()
	if err != nil {
		security.Audit(ctxt, SIT915_CLASS_ID, ACTION, security.FAIL_ERROR)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
		return
	}

	result := make([]Sit915Response, 0)
	for _, sit915 := range sit915s {
		remoteSite, err := sit915Rest.remoteSiteDb.FindOneRemoteSite(sit915.RemotesiteId, false)
		if err != nil {
			security.Audit(ctxt, REMOTESITE_CLASS_ID, ACTION, security.FAIL_ERROR)
			response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
			return
		}

		value := Sit915Response{
			Sit915:     sit915,
			RemoteSite: remoteSite,
		}

		result = append(result, value)
	}

	security.Audit(ctxt, SIT915_CLASS_ID, ACTION, security.SUCCESS)
	rest.WriteEntitySafely(response, result)
}

func (sit915Rest *Sit915Rest) Get(request *restful.Request, response *restful.Response) {
	ACTION := moc.Sit915_READ.String()
	ctxt := request.Request.Context()
	if !security.HasPermissionForAction(ctxt, SIT915_CLASS_ID, ACTION) {
		security.Audit(ctxt, SIT915_CLASS_ID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
		return
	}

	message_id, errs := rest.SanitizeValidatePathParameter(request, PARAMETER_SIT915_MESSAGE_ID)
	if errs != nil {
		security.Audit(ctxt, SIT915_CLASS_ID, ACTION, security.FAIL_VALIDATION)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		return
	}

	sit915, err := sit915Rest.sit915Db.FindOneSit915(message_id)
	if err != nil {
		if db.ErrorNotFound == err {
			security.Audit(ctxt, SIT915_CLASS_ID, ACTION, security.FAIL_NOTFOUND)
			response.WriteError(http.StatusNotFound, err)
		} else {
			security.Audit(ctxt, SIT915_CLASS_ID, ACTION, security.FAIL_ERROR)
			response.WriteError(http.StatusInternalServerError, err)
		}

		return
	}

	remoteSite, err := sit915Rest.remoteSiteDb.FindOneRemoteSite(sit915.RemotesiteId, false)
	if err != nil {
		security.Audit(ctxt, REMOTESITE_CLASS_ID, ACTION, security.FAIL_ERROR)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
		return
	}

	result := Sit915Response{
		Sit915:     sit915,
		RemoteSite: remoteSite,
	}

	security.Audit(ctxt, SIT915_CLASS_ID, ACTION, security.SUCCESS)
	rest.WriteEntitySafely(response, result)
}

func GenerateSit915Message(ctx gogroup.GoGroup, sit915Db db.Sit915DB, remoteSiteDb db.RemoteSiteDB, configDb mongo.ConfigDb, id string) (string, error) {
	sit915, err := sit915Db.FindOneSit915(id)
	if err != nil {
		return "", err
	}

	remoteSite, err := remoteSiteDb.FindOneRemoteSite(sit915.RemotesiteId, false)
	if err != nil {
		return "", err
	}

	// Message Field # 1 - Transmission number and retransmission number
	transmissionNum := sit915.TransmissionNum
	retransmissionNum := sit915.RetransmissionNum
	mf1 := fmt.Sprintf("/%.5d %.5d", transmissionNum, retransmissionNum)

	// Message Field # 2 - Code of the reporting facility
	localSiteConfig, err := configDb.Read(ctx)
	if err != nil {
		return "", err
	}

	localCode := localSiteConfig.Site.Cscode
	mf2 := fmt.Sprintf("/%s", localCode)

	// Message Field # 3 - The date and time of the transmission
	t := time.Now()
	year := strconv.Itoa(t.Year())[2:4]
	dayOfYear := t.YearDay()
	hour := t.Hour()
	min := t.Minute()
	mf3 := fmt.Sprintf("/%s %.3d %.2d%.2d", year, dayOfYear, hour, min)

	// Message Field # 4 - SIT number
	mf4 := fmt.Sprintf("/%d", 915)

	// Message Field # 5 - Code of the final destination
	mf5 := fmt.Sprintf("/%s", remoteSite.Cscode)

	// Message Field # 41 - Narrative text
	mf41 := fmt.Sprintf("/%s\nQQQQ", sit915.Narrative)

	// Message Field # 42 - Always /LASSIT
	mf42 := "/LASSIT"

	// Message Field # 43 - Always /ENDMSG
	mf43 := "/ENDMSG"

	messageBody := fmt.Sprintf("%s%s%s\n%s%s\n%s\n%s\n%s\n", mf1, mf2, mf3, mf4, mf5, mf41, mf42, mf43)

	return messageBody, nil
}
