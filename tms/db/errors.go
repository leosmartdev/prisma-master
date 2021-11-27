package db

import (
	"context"
	"errors"
	"net/http"
	"prisma/tms/log"
	"prisma/tms/security"

	"prisma/tms/rest"

	restful "github.com/orolia/go-restful"
)

var (
	ErrorNotFound  = errors.New("notFound")
	ErrorLocked    = errors.New("locked")
	ErrorCritical  = errors.New("critical")
	ErrorDuplicate = errors.New("duplicate")
	ErrorBadID = errors.New("badID")
)

func ErrorFree(err error, ctx context.Context, response *restful.Response, classId string, action string) bool {
	errorFree := (err == nil)

	if !errorFree {
		var errs []rest.ErrorValidation
		switch err {
		case ErrorNotFound:
			errs = append(errs, rest.ErrorValidation{
				Property: classId,
				Rule:     "NotFound",
				Message:  "Not found"})
			security.Audit(ctx, classId, action, security.FAIL_NOTFOUND)
			// TODO create json error, have ErrorNotFound include id/search criteria
			response.WriteHeaderAndEntity(http.StatusNotFound, errs)
		case ErrorDuplicate:
			errs = append(errs, rest.ErrorValidation{
				Property: classId,
				Rule:     "Constraint",
				Message:  "Duplicate"})
			security.Audit(ctx, classId, action, security.FAIL_VALIDATION)
			// TODO create json error, have ErrorDuplicate include field causing constraint failure
			rest.WriteValidationErrsSafely(response, errs)
		default:
			log.Error("unexpected error: %v", err)
			security.Audit(ctx, classId, action, security.FAIL_ERROR)
			response.WriteError(http.StatusInternalServerError, err)
		}
	}

	return errorFree
}
