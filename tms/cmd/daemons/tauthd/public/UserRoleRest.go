package public

import (
	restful "github.com/orolia/go-restful"
	"net/http"
	"prisma/tms/security"
	"log"
	secDb "prisma/tms/security/database"
	"prisma/tms/security/message"
)

const (
	CLASSID_ROLE = "Role"
)

type UserRoleRest struct{}

// AssignedRoles (RC-09)
func (userRoleRest *UserRoleRest) GetUserRole(request *restful.Request, response *restful.Response) {
	userId := request.PathParameter("user-id")
	user, err := secDb.FindOneByUserId(request.Request.Context(), secDb.UserId(userId))
	if err != nil {
		response.WriteError(http.StatusBadRequest, err)
	}
	if user == nil {
		response.WriteErrorString(http.StatusNotFound, "")
	} else {
		if user.Roles == nil {
			user.Roles = make([]string, 0)
		}
		response.WriteEntity(user.Roles)
	}
}

// AssignedUsers (RC-11)
func (userRoleRest *UserRoleRest) GetRoleUser(request *restful.Request, response *restful.Response) {
	const ACTION = "READ"
	allowed := security.HasPermissionForAction(request.Request.Context(), CLASSID_ROLE, ACTION)
	if !allowed {
		security.Audit(request.Request.Context(), CLASSID_ROLE, ACTION, "FAIL")
		response.WriteErrorString(http.StatusForbidden, "")
	} else {
		roleId := request.PathParameter("role-id")
		if _, ok := message.RoleId_value[roleId]; ok {
			// search map from request query parameters
			searchMap := make(map[string]string)
			searchMap["roles"] = roleId
			users, err := secDb.FindByMap(request.Request.Context(), searchMap)
			if err != nil {
				response.WriteError(http.StatusInternalServerError, err)
			} else {
				response.WriteEntity(users)
			}
		} else {
			response.WriteErrorString(http.StatusBadRequest, "")
		}
	}
}

// AssignUser (RC-10)
func (userRoleRest *UserRoleRest) Put(request *restful.Request, response *restful.Response) {
	log.Println("**UserRoleRest.Put")
	roleId := request.PathParameter("role-id")
	userId := request.PathParameter("user-id")
	// check if valid roleId
	if _, ok := message.RoleId_value[roleId]; ok {
		// get user
		user, err := secDb.FindOneByUserId(request.Request.Context(), secDb.UserId(userId))
		if err != nil {
			response.WriteError(http.StatusBadRequest, err)
		} else {
			if user == nil {
				response.WriteErrorString(http.StatusNotFound, "")
			} else {
				log.Printf("user=%v", user)
				// check if roleId is present
				if stringInSlice(roleId, user.Roles) {
					response.WriteHeaderAndEntity(http.StatusNotModified, user.Roles)
				} else {
					// add role
					user.Roles = append(user.Roles, roleId)
					// update user
					user, err := secDb.Update(request.Request.Context(), user)
					if err != nil {
						response.WriteError(http.StatusInternalServerError, err)
					} else {
						response.WriteEntity(user.Roles)
					}
				}
			}
		}
	} else {
		response.WriteErrorString(http.StatusNotFound, "")
	}
}

// DeassignUser (RC-18)
func (userRoleRest *UserRoleRest) Delete(request *restful.Request, response *restful.Response) {
	log.Println("**UserRoleRest.Delete")
	roleId := request.PathParameter("role-id")
	userId := request.PathParameter("user-id")
	// check if valid roleId
	if _, ok := message.RoleId_value[roleId]; ok {
		// get user
		user, err := secDb.FindOneByUserId(request.Request.Context(), secDb.UserId(userId))
		if err != nil {
			response.WriteError(http.StatusBadRequest, err)
		}
		if user == nil {
			response.WriteErrorString(http.StatusNotFound, "")
		} else {
			// check and remove if roleId is present
			updatedRoles, found := removeStringInSlice(roleId, user.Roles)
			if found {
				// remove role
				user.Roles = updatedRoles
				// update user
				user, err := secDb.Update(request.Request.Context(), user)
				if err != nil {
					response.WriteError(http.StatusInternalServerError, err)
				} else {
					response.WriteEntity(user.Roles)
				}
			} else {
				response.WriteHeaderAndEntity(http.StatusNotModified, user.Roles)
			}
		}
	} else {
		response.WriteErrorString(http.StatusNotFound, "")
	}
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func removeStringInSlice(r string, s []string) ([]string, bool) {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...), true
		}
	}
	return s, false
}
