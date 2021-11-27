package public

import (
	"net/http"

	sec "prisma/tms/security"

	restful "github.com/orolia/go-restful" 
)

type RoleRest struct {
}

func (roleRest *RoleRest) Get(request *restful.Request, response *restful.Response) {
	const CLASSID = "Role"
	const ACTION = "READ"
	allowed := sec.HasPermissionForAction(request.Request.Context(), CLASSID, ACTION)
	if !allowed {
		sec.Audit(request.Request.Context(), CLASSID, ACTION, "FAIL")
		response.WriteErrorString(http.StatusForbidden, "")
	} else {
		sec.Audit(request.Request.Context(), CLASSID, ACTION, "SUCCESS")
		response.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
		response.ResponseWriter.Write([]byte(ROLE_OBJECT_ACTION_JSON))
	}
}

const ROLE_OBJECT_ACTION_JSON = `[{"RoleName":"StandardUser","RoleId":"StandardUser","Permissions":[{"ObjectClassId":"Profile","ObjectActions":[{"ObjectAction":"READ"},{"ObjectAction":"UPDATE"},{"ObjectAction":"UPDATE_PASSWORD"}]},{"ObjectClassId":"Zone","ObjectActions":[{"ObjectAction":"CREATE"},{"ObjectAction":"READ"},{"ObjectAction":"UPDATE"},{"ObjectAction":"DELETE"}]},{"ObjectClassId":"Alert","ObjectActions":[{"ObjectAction":"READ"},{"ObjectAction":"ACKNOWLEDGE"}]},{"ObjectClassId":"Incident","ObjectActions":[{"ObjectAction":"READ"}]}]},{"RoleName":"UserManager","RoleId":"UserManager","Permissions":[{"ObjectClassId":"User","ObjectActions":[{"ObjectAction":"CREATE"},{"ObjectAction":"READ"},{"ObjectAction":"UPDATE"},{"ObjectAction":"DEACTIVATE"},{"ObjectAction":"UPDATE_ROLE"}]},{"ObjectClassId":"Role","ObjectActions":[{"ObjectAction":"READ"}]}]},{"RoleName":"FleetManager","RoleId":"FleetManager","Permissions":[{"ObjectClassId":"Fleet","ObjectActions":[{"ObjectAction":"CREATE"},{"ObjectAction":"READ"},{"ObjectAction":"UPDATE"},{"ObjectAction":"DELETE"},{"ObjectAction":"UPDATE_VESSEL"}]},{"ObjectClassId":"Vessel","ObjectActions":[{"ObjectAction":"CREATE"},{"ObjectAction":"READ"},{"ObjectAction":"UPDATE"},{"ObjectAction":"DELETE"},{"ObjectAction":"UPDATE_DEVICE"}]},{"ObjectClassId":"Device","ObjectActions":[{"ObjectAction":"READ"}]}]},{"RoleName":"IncidentManager","RoleId":"IncidentManager","Permissions":[{"ObjectClassId":"Incident","ObjectActions":[{"ObjectAction":"CREATE"},{"ObjectAction":"READ"},{"ObjectAction":"UPDATE"},{"ObjectAction":"DELETE"},{"ObjectAction":"OPEN"},{"ObjectAction":"CLOSE"},{"ObjectAction":"ASSIGN"},{"ObjectAction":"UNASSIGN"},{"ObjectAction":"ARCHIVE"}]},{"ObjectClassId":"File","ObjectActions":[{"ObjectAction":"READ"},{"ObjectAction":"CREATE"},{"ObjectAction":"DELETE"}]}]},{"RoleName":"Administrator","RoleId":"Administrator","Permissions":[{"ObjectClassId":"User","ObjectActions":[{"ObjectAction":"CREATE"},{"ObjectAction":"READ"},{"ObjectAction":"UPDATE"},{"ObjectAction":"DEACTIVATE"},{"ObjectAction":"UPDATE_ROLE"}]},{"ObjectClassId":"Role","ObjectActions":[{"ObjectAction":"READ"}]},{"ObjectClassId":"Vessel","ObjectActions":[{"ObjectAction":"CREATE"},{"ObjectAction":"READ"},{"ObjectAction":"UPDATE"},{"ObjectAction":"DELETE"},{"ObjectAction":"UPDATE_DEVICE"}]},{"ObjectClassId":"Device","ObjectActions":[{"ObjectAction":"READ"}]},{"ObjectClassId":"Fleet","ObjectActions":[{"ObjectAction":"CREATE"},{"ObjectAction":"READ"},{"ObjectAction":"UPDATE"},{"ObjectAction":"DELETE"},{"ObjectAction":"UPDATE_VESSEL"}]},{"ObjectClassId":"Profile","ObjectActions":[{"ObjectAction":"READ"},{"ObjectAction":"UPDATE"},{"ObjectAction":"UPDATE_PASSWORD"}]},{"ObjectClassId":"Zone","ObjectActions":[{"ObjectAction":"CREATE"},{"ObjectAction":"READ"},{"ObjectAction":"UPDATE"},{"ObjectAction":"DELETE"}]},{"ObjectClassId":"Alert","ObjectActions":[{"ObjectAction":"READ"},{"ObjectAction":"ACKNOWLEDGE"}]},{"ObjectClassId":"File","ObjectActions":[{"ObjectAction":"READ"},{"ObjectAction":"CREATE"},{"ObjectAction":"DELETE"}]}]}]`
