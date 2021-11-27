package public

import (
	"net/http"

	"prisma/tms/db"
	"prisma/tms/log"
	"prisma/tms/rest"
	"prisma/tms/security"
	"prisma/tms/moc"

	restful "github.com/orolia/go-restful" 
	"io"
)

// authorized checks if the user in the context has permission to perform the action on the class id.
// Writes to audit log and response.
// Returns false if not authorized, and caller should not continue processing the transaction.
func authorized(request *restful.Request, response *restful.Response, classId string, action string) bool {
	ctxt := request.Request.Context()
	authorized := security.HasPermissionForAction(ctxt, classId, action)
	if !authorized {
		security.Audit(ctxt, classId, action, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
	return authorized
}

// valid checks if validation errors are present.
// Writes to audit log and response.
// Returns false if validation errors, and caller should not continue processing the transaction.
func valid(errs []rest.ErrorValidation, request *restful.Request, response *restful.Response, classId string, action string) bool {
	ctxt := request.Request.Context()
	valid := errs == nil
	if !valid {
		security.Audit(ctxt, classId, action, security.FAIL_VALIDATION)
		rest.WriteValidationErrsSafely(response, errs)
	}
	return valid
}

func errorFree(err error, request *restful.Request, response *restful.Response, classId string, action string) bool {
	ctxt := request.Request.Context()
	errorFree := err == nil
	if !errorFree {
		var errs []rest.ErrorValidation
		switch err {
		case db.ErrorNotFound:
			errs = append(errs, rest.ErrorValidation{
				Property: classId,
				Rule:     "NotFound",
				Message:  "Not found"})
			security.Audit(ctxt, classId, action, security.FAIL_NOTFOUND)
			// TODO create json error, have ErrorNotFound include id/search criteria
			response.WriteHeaderAndEntity(http.StatusNotFound, errs)
		case db.ErrorDuplicate, ErrDuplicateEntityLogIncident:
			errs = append(errs, rest.ErrorValidation{
				Property: classId,
				Rule:     "Constraint",
				Message:  "Duplicate"})
			security.Audit(ctxt, classId, action, security.FAIL_VALIDATION)
			// TODO create json error, have ErrorDuplicate include field causing constraint failure
			rest.WriteValidationErrsSafely(response, errs)
		case db.ErrorBadID:
			errs = append(errs, rest.ErrorValidation{
				Property: classId,
				Rule:     "bad mongo hex ID",
				Message:  "Is not Hex ObjectId"})
			security.Audit(ctxt, classId, action, security.FAIL_VALIDATION)
			response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		case io.EOF:
			errs = append(errs, rest.ErrorValidation{
				Property: classId,
				Rule:     "EOF",
				Message:  "Unexpected end of data"})
			security.Audit(ctxt, classId, action, security.FAIL_VALIDATION)
			response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		default:
			log.Error(err.Error(), err)
			security.Audit(ctxt, classId, action, security.FAIL_ERROR)
			response.WriteHeaderAndEntity(http.StatusInternalServerError, err.Error())
		}
	}
	return errorFree
}

// duplicatingDevice is used to check duplicating devices
// return id of device or empty string
func duplicatingDevice(vessel *moc.Vessel) string {
	duplicate := make(map[string]bool)
	for i := range vessel.Devices {
		if _, ok := duplicate[vessel.Devices[i].Id]; ok {
			return vessel.Devices[i].Id
		} else {
			duplicate[vessel.Devices[i].Id] = true
		}
	}
	return ""
}
