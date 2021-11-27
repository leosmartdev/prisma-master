package security

import (
	"fmt"
	securityContext "prisma/tms/security/context"
	"prisma/tms/security/database/audit"
	"prisma/tms/security/session"
	"reflect"
	"time"

	"golang.org/x/net/context"
)

const (
	SUCCESS = "SUCCESS"
	// target object not changed
	SUCCESS_NOTMODIFIED = "SUCCESS_NOTMODIFIED"
	// user does not have permission
	FAIL_PERMISSION = "FAIL_PERMISSION"
	// failure in input validation, business rule, or security policy
	FAIL_VALIDATION = "FAIL_VALIDATION"
	// target object not found
	FAIL_NOTFOUND = "FAIL_NOTFOUND"
	// critical error
	FAIL_ERROR = "FAIL_ERROR"
)

func Audit(context context.Context, classId string, action string, outcome string, payload ...interface{}) {
	objectId := ""
	userId := ""
	// process payload
	for _, load := range payload {
		payloadType := reflect.TypeOf(load)
		if "database.UserId" == payloadType.String() {
			userId = fmt.Sprint(load)
		}
	}
	AuditUserObject(context, classId, objectId, userId, action, outcome, payload...)
}
func AuditUserObject(context context.Context, classId string, objectId string, userId string, action string, outcome string, payload ...interface{}) {
	// FIXME add policy check for filtering
	// TODO add fallback to securely store when primary is offline
	if "READ" == action {
		return
	}
	requestId := securityContext.RequestIdFromContext(context)
	sessionId := securityContext.SessionIdFromContext(context)
	auditTime := time.Now()
	internalSession, err := session.GetStore(context).Get(sessionId)
	if nil == err && nil != internalSession && "" == userId {
		userId = internalSession.GetOwner()
	}
	var payloadRecord string
	if len(payload) == 1 {
		payloadRecord = fmt.Sprint(payload[0])
	}
	if len(payload) > 1 {
		payloadRecord = fmt.Sprint(payload...)
	}
	record := audit.Record{
		Created:   auditTime,
		ClassId:   classId,
		ObjectId:  objectId,
		UserId:    userId,
		Action:    action,
		Outcome:   outcome,
		Payload:   payloadRecord,
		SessionId: sessionId,
		RequestId: requestId,
	}
	auditor := audit.NewAuditor(context)
	auditor.Record(context, record)

}
