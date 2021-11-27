package public

import (
	"net/http"
	"time"

	"prisma/gogroup"
	"prisma/tms/db"
	"prisma/tms/db/mongo"
	"prisma/tms/moc"
	"prisma/tms/rest"
	"prisma/tms/security"

	"github.com/globalsign/mgo/bson"

	"github.com/go-openapi/spec"
	"github.com/golang/protobuf/ptypes"

	"github.com/globalsign/mgo"
	restful "github.com/orolia/go-restful"
)

const (
	NOTICE_CLASSID = "Notice"
	TOPIC_NOTICE   = CLASSID
)

var (
	PARAMETER_NOTICE_DATABASE_ID = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "id",
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
	SCHEMA_NOTICE_TIMEOUT = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Required: []string{"olderThan"},
			Properties: map[string]spec.Schema{
				"olderThan": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{20}[0],
						MaxLength: &[]int64{25}[0],
						Pattern:   "^[0-9:\\-TZ+]*$",
					},
				},
			},
		},
	}
)

type NoticeRest struct {
	notifydb db.NotifyDb
	ctxt     gogroup.GoGroup
}

type AckResponsePublic struct {
	ID string `json:"ID"`
}

type AckAllResponsePublic struct {
	Updated int `json:"updated"`
}

type TimeoutRequestPublic struct {
	OlderThan string `json:"olderThan"`
}

type TimeoutResponsePublic struct {
	Updated int `json:"updated"`
}

func NewNoticeRest(client *mongo.MongoClient, ctxt gogroup.GoGroup) *NoticeRest {
	return &NoticeRest{
		notifydb: mongo.NewNotifyDb(ctxt, client),
		ctxt:     ctxt,
	}
}

func (r *NoticeRest) GetAllNewNotices(request *restful.Request, response *restful.Response) {
	ACTION := moc.Notice_GET.String()
	ctxt := request.Request.Context()
	allowed := security.HasPermissionForAction(ctxt, NOTICE_CLASSID, ACTION)
	if !allowed {
		security.Audit(ctxt, NOTICE_CLASSID, ACTION, "FAIL")
		response.WriteErrorString(http.StatusForbidden, "")
		return
	}
	pagination, ok := rest.SanitizePagination(request)
	if !ok {
		security.Audit(ctxt, NOTICE_CLASSID, ACTION, "FAIL")
		response.WriteErrorString(http.StatusBadRequest, "")
		return
	}
	query := bson.M{"me.action": "NEW"}
	pagination.Sort = "ctime"
	results, err := r.notifydb.GetHistory(query, pagination)
	if err != nil {
		security.Audit(ctxt, NOTICE_CLASSID, ACTION, "FAIL")
		response.WriteErrorString(http.StatusInternalServerError, "")
		return
	}
	if len(results) > 0 {
		rest.AddPaginationHeaderSafely(request, response, pagination)
	}
	security.Audit(ctxt, NOTICE_CLASSID, ACTION, "SUCCESS")
	rest.WriteHeaderAndEntitySafely(response, http.StatusOK, results)
}

func (r *NoticeRest) GetHistory(request *restful.Request, response *restful.Response) {
	ACTION := moc.Notice_GET.String()
	ctxt := request.Request.Context()
	allowed := security.HasPermissionForAction(ctxt, NOTICE_CLASSID, ACTION)
	if !allowed {
		security.Audit(ctxt, NOTICE_CLASSID, ACTION, "FAIL")
		response.WriteErrorString(http.StatusForbidden, "")
		return
	}
	pagination, ok := rest.SanitizePagination(request)
	if !ok {
		security.Audit(ctxt, NOTICE_CLASSID, ACTION, "FAIL")
		response.WriteErrorString(http.StatusBadRequest, "")
		return
	}
	query := bson.M{
		"$or": []bson.M{
			bson.M{"me.action": "CLEAR"},
			bson.M{"me.action": "ACK"},
		},
	}
	pagination.Sort = "ctime"
	results, err := r.notifydb.GetHistory(query, pagination)
	if err != nil {
		security.Audit(ctxt, NOTICE_CLASSID, ACTION, "FAIL")
		response.WriteErrorString(http.StatusInternalServerError, "")
		return
	}
	if len(results) > 0 {
		rest.AddPaginationHeaderSafely(request, response, pagination)
	}
	security.Audit(ctxt, NOTICE_CLASSID, ACTION, "SUCCESS")
	rest.WriteHeaderAndEntitySafely(response, http.StatusOK, results)
}

func (r *NoticeRest) Ack(request *restful.Request, response *restful.Response) {
	ACTION := moc.Notice_ACK.String()
	if !security.HasPermissionForAction(request.Request.Context(), NOTICE_CLASSID, ACTION) {
		security.Audit(request.Request.Context(), NOTICE_CLASSID, ACTION, "FAIL")
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
		return
	}
	id, errs := rest.SanitizeValidatePathParameter(request, PARAMETER_NOTICE_DATABASE_ID)
	if errs != nil {
		security.Audit(request.Request.Context(), NOTICE_CLASSID, ACTION, "FAIL_ERROR")
		response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		return
	}
	now, err := ptypes.TimestampProto(time.Now())
	if err != nil {
		security.Audit(request.Request.Context(), NOTICE_CLASSID, ACTION, "FAIL_ERROR")
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	err = r.notifydb.Ack(id, now)
	if err == mgo.ErrNotFound {
		security.Audit(request.Request.Context(), NOTICE_CLASSID, ACTION, "FAIL_ERROR")
		response.WriteError(http.StatusNotFound, err)
		return
	}
	if err != nil {
		security.Audit(request.Request.Context(), NOTICE_CLASSID, ACTION, "FAIL_ERROR")
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	notice, err := r.notifydb.GetById(id)
	if err != nil {
		security.Audit(request.Request.Context(), NOTICE_CLASSID, ACTION, "FAIL_ERROR")
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	var classId = notice.Event.String()
	objectId := ""
	switch notice.Event {
	case moc.Notice_Rule:
		objectId = notice.Target.TrackId
	case moc.Notice_EnterZone:
		objectId = id
	case moc.Notice_ExitZone:
		objectId = id
	case moc.Notice_Sart:
		objectId = notice.Target.TrackId
	case moc.Notice_Sarsat:
		objectId = notice.Target.TrackId
	case moc.Notice_OmnicomAssistance:
		objectId = notice.Target.TrackId
	case moc.Notice_SarsatDefaultHandling:
		objectId = notice.Target.TrackId
	case moc.Notice_IncidentTransfer:
		objectId = id
	case moc.Notice_SarsatMessage:
		objectId = id
	}

	security.AuditUserObject(request.Request.Context(), classId, objectId, "", ACTION, "SUCCESS")
	payload := AckResponsePublic{ID: id}
	rest.WriteHeaderAndEntitySafely(response, http.StatusOK, payload)
}

func (r *NoticeRest) AckAll(request *restful.Request, response *restful.Response) {
	ACTION := moc.Notice_ACK_ALL.String()
	if !security.HasPermissionForAction(request.Request.Context(), NOTICE_CLASSID, ACTION) {
		security.Audit(request.Request.Context(), NOTICE_CLASSID, ACTION, "FAIL")
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
		return
	}
	now, err := ptypes.TimestampProto(time.Now())
	if err != nil {
		security.Audit(request.Request.Context(), NOTICE_CLASSID, ACTION, "FAIL_ERROR")
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	updated, err := r.notifydb.AckAll(now)
	if err != nil {
		security.Audit(request.Request.Context(), NOTICE_CLASSID, ACTION, "FAIL_ERROR")
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	security.Audit(request.Request.Context(), NOTICE_CLASSID, ACTION, "SUCCESS")
	payload := AckAllResponsePublic{Updated: updated}
	rest.WriteHeaderAndEntitySafely(response, http.StatusOK, payload)
}

func (r *NoticeRest) Timeout(request *restful.Request, response *restful.Response) {
	ACTION := moc.Notice_TIMEOUT.String()
	if !security.HasPermissionForAction(request.Request.Context(), NOTICE_CLASSID, ACTION) {
		security.Audit(request.Request.Context(), NOTICE_CLASSID, ACTION, "FAIL")
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
		return
	}
	timeoutReq := &TimeoutRequestPublic{}
	errs := rest.SanitizeValidateReadEntity(request, SCHEMA_NOTICE_TIMEOUT, timeoutReq)
	if errs != nil {
		security.Audit(request.Request.Context(), NOTICE_CLASSID, ACTION, "FAIL_ERROR")
		response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		return
	}
	request.ReadEntity(timeoutReq)
	timeout, err := time.Parse(time.RFC3339, timeoutReq.OlderThan)
	if err != nil {
		security.Audit(request.Request.Context(), NOTICE_CLASSID, ACTION, "FAIL_ERROR")
		response.WriteError(http.StatusBadRequest, err)
		return
	}
	ptimeout, err := ptypes.TimestampProto(timeout)
	if err != nil {
		security.Audit(request.Request.Context(), NOTICE_CLASSID, ACTION, "FAIL_ERROR")
		response.WriteError(http.StatusBadRequest, err)
		return
	}
	updated, err := r.notifydb.Timeout(ptimeout)
	if err != nil {
		security.Audit(request.Request.Context(), NOTICE_CLASSID, ACTION, "FAIL_ERROR")
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	security.Audit(request.Request.Context(), NOTICE_CLASSID, ACTION, "SUCCESS")
	payload := TimeoutResponsePublic{Updated: updated}
	rest.WriteHeaderAndEntitySafely(response, http.StatusOK, payload)
}
