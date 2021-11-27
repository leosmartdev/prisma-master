package public

import (
	restful "github.com/orolia/go-restful" 
	"net/http"
	sec "prisma/tms/security"
)

type ObjectRest struct {
}

func (objectRest *ObjectRest) Get(request *restful.Request, response *restful.Response) {
	const CLASSID = "Object"
	const ACTION = "READ"
	allowed := sec.HasPermissionForAction(request.Request.Context(), CLASSID, ACTION)
	if !allowed {
		sec.Audit(request.Request.Context(), CLASSID, ACTION, "FAIL")
		response.WriteErrorString(http.StatusForbidden, "")
	} else {
		sec.Audit(request.Request.Context(), CLASSID, ACTION, "SUCCESS")
		response.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
		response.ResponseWriter.Write([]byte(OBJECT_ACTION_JSON))
	}
}

const OBJECT_ACTION_JSON = `[{"ObjectClassId":"User","ObjectActions":[{"ObjectAction":"CREATE"},{"ObjectAction":"READ"},{"ObjectAction":"UPDATE"},{"ObjectAction":"DELETE"},{"ObjectAction":"UPDATE_ROLE"}]},{"ObjectClassId":"Role","ObjectActions":[{"ObjectAction":"READ"}]},{"ObjectClassId":"Profile","ObjectActions":[{"ObjectAction":"READ"},{"ObjectAction":"UPDATE"},{"ObjectAction":"UPDATE_PASSWORD"}]},{"ObjectClassId":"Zone","ObjectActions":[{"ObjectAction":"CREATE"},{"ObjectAction":"READ"},{"ObjectAction":"UPDATE"},{"ObjectAction":"DELETE"}]},{"ObjectClassId":"Alert","ObjectActions":[{"ObjectAction":"READ"},{"ObjectAction":"ACKNOWLEDGE"}]},{"ObjectClassId":"Vessel","ObjectActions":[{"ObjectAction":"CREATE"},{"ObjectAction":"READ"},{"ObjectAction":"UPDATE"},{"ObjectAction":"DELETE"},{"ObjectAction":"UPDATE_DEVICE"}]},{"ObjectClassId":"Device","ObjectActions":[{"ObjectAction":"READ"}]},{"ObjectClassId":"Fleet","ObjectActions":[{"ObjectAction":"CREATE"},{"ObjectAction":"READ"},{"ObjectAction":"UPDATE"},{"ObjectAction":"DELETE"},{"ObjectAction":"UPDATE_VESSEL"}]}]`
