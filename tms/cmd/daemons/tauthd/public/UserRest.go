package public

import (
	"net/http"
	"prisma/tms"
	"prisma/tms/db/mongo"
	"prisma/tms/rest"
	"prisma/tms/security"
	sec "prisma/tms/security"
	secDb "prisma/tms/security/database"
	"prisma/tms/security/message"
	"time"

	"prisma/tms/db"

	restful "github.com/orolia/go-restful"

	"github.com/go-openapi/spec"
	"github.com/pborman/uuid"
	"golang.org/x/net/context"
)

const USER_CLASSID = "User"

type UserRequest struct {
	Id       string               `json:"id,omitempty"`
	UserId   string               `json:"userId,omitempty"`
	Password string               `json:"password"`
	Profile  *message.UserProfile `json:"profile"`
	Roles    []string
}

type UserRest struct {
	userDB db.UserDB
}

// body schema
func SchemaUserCreate(ctxt context.Context) spec.Schema {
	return spec.Schema{
		SchemaProps: spec.SchemaProps{
			Required: []string{"userId", "password"},
			Properties: map[string]spec.Schema{
				"password": {
					SchemaProps: sec.PasswordSchemaProps(ctxt),
				},
				"userId": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{2}[0],
						MaxLength: &[]int64{128}[0],
						Pattern:   "^[a-zA-Z0-9_@.]*$",
					},
				},
			},
		},
	}
}

func SchemaUserUpdate(ctx context.Context) spec.Schema {
	return spec.Schema{
		SchemaProps: spec.SchemaProps{
			Properties: map[string]spec.Schema{
				"password": {
					SchemaProps: sec.PasswordSchemaProps(ctx),
				},
			},
		},
	}
}

// parameter schema
var (
	// enums for schema
	userStateEnum            = make([]interface{}, len(message.User_State_name))
	parameterUserStateUpdate = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "user-state",
			In:       "path",
			Required: true,
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Enum: userStateEnum,
				},
			},
		},
	}
	parameterUserId = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "user-id",
			In:       "path",
			Required: true,
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					MinLength: &[]int64{1}[0],
					MaxLength: &[]int64{128}[0],
					Pattern:   "^[a-zA-Z0-9_@.]*$",
				},
			},
		},
	}
)

func NewUserRest(ctxt context.Context) *UserRest {
	// enums for schema
	index := 0
	for _, state := range message.User_State_name {
		userStateEnum[index] = state
		index++
	}
	user := new(secDb.User)
	// UserId
	user.UserId = "admin"
	// Password
	saltBytes := uuid.NewRandom()
	inputedHash := computeHmac256("admin", saltBytes)
	user.PasswordHash = inputedHash
	// Log password
	user.PasswordLog = make([]*message.PasswordLogEntry, 0, 1)
	user.PasswordLog = append(user.PasswordLog, &message.PasswordLogEntry{
		PasswordHash: user.PasswordHash,
		Timestamp:    tms.ToTimestamp(time.Now()),
	})
	// Salt
	user.Salt = saltBytes.String()
	// State
	user.State = message.User_initialized
	// Profile
	user.Profile = &message.UserProfile{}
	// Roles
	user.Roles = []string{message.RoleId_name[int32(message.RoleId_Administrator)]}
	_, err := secDb.Add(ctxt, user)
	if err == nil {
		sec.Audit(ctxt, "User", "CREATE", "SUCCESS", user.UserId, user.Roles)
	}

	r := &UserRest{
		userDB: mongo.NewMongoUserDb(),
	}
	return r
}

func (userRest *UserRest) Get(request *restful.Request, response *restful.Response) {
	ACTION := message.User_READ.String()
	ctxt := request.Request.Context()
	allowed := sec.HasPermissionForAction(ctxt, USER_CLASSID, ACTION)
	if !allowed {
		sec.Audit(ctxt, USER_CLASSID, ACTION, "FAIL")
		response.WriteErrorString(http.StatusForbidden, "")
	} else {
		var err error
		var users []secDb.User
		pagination, ok := rest.SanitizePagination(request)
		if ok {
			// search map from request query parameters
			searchMap := make(map[string]string)
			if request.QueryParameter("state") != "" {
				searchMap["state"] = request.QueryParameter("state")
			}
			pagination.Sort = "userId"
			users, err = secDb.FindByMapByPagination(ctxt, searchMap, pagination)
			if len(users) > 0 {
				// update pagination. before + after
				pagination.AfterId = string(users[len(users)-1].UserId)
				pagination.BeforeId = string(users[0].UserId)
				rest.AddPaginationHeaderSafely(request, response, pagination)
			}
		} else {
			users, err = secDb.FindAllNotDisabled(ctxt)
		}
		if err != nil {
			sec.Audit(ctxt, USER_CLASSID, ACTION, "SUCCESS_ERROR")
			response.WriteError(http.StatusInternalServerError, err)
		} else {
			sec.Audit(ctxt, USER_CLASSID, ACTION, "SUCCESS")
			response.WriteEntity(users)
		}
	}
}

func (r *UserRest) GetOne(req *restful.Request, rsp *restful.Response) {
	ACTION := message.User_READ.String()
	ctx := req.Request.Context()

	if !rest.Authorized(ctx, rsp, USER_CLASSID, ACTION) {
		return
	}

	userID, errs := rest.SanitizeValidatePathParameter(req, parameterUserId)
	if !rest.Valid(errs, ctx, rsp, USER_CLASSID, ACTION) {
		return
	}

	v, err := r.userDB.FindOne(ctx, userID)
	if !db.ErrorFree(err, ctx, rsp, USER_CLASSID, ACTION) {
		return
	}

	security.Audit(ctx, USER_CLASSID, ACTION, security.SUCCESS)
	rest.WriteHeaderAndEntitySafely(rsp, http.StatusOK, v)
}

func (userRest *UserRest) UpdateState(request *restful.Request, response *restful.Response) {
	ctxt := request.Request.Context()
	userStateString, errs := rest.SanitizeValidatePathParameter(request, parameterUserStateUpdate)
	userId, errs2 := rest.SanitizeValidatePathParameter(request, parameterUserId)
	errs = append(errs, errs2...)
	if errs != nil {
		security.Audit(ctxt, USER_CLASSID, "ANY", "FAIL_ERROR")
		rest.WriteValidationErrsSafely(response, errs)
	} else {
		user, err := secDb.FindOneByUserId(ctxt, secDb.UserId(userId))
		if err != nil {
			response.WriteError(http.StatusBadRequest, err)
		} else {
			userState := message.User_State(message.User_State_value[userStateString])
			user.State = userState
			user.Attempts = 0
			// determine action
			var action string
			switch {
			case message.User_initialized == userState:
				action = message.User_INITIALIZE.String()
			case message.User_locked == userState:
				action = message.User_LOCK.String()
			case message.User_activated == userState:
				action = message.User_ACTIVATE.String()
			case message.User_deactivated == userState:
				action = message.User_DEACTIVATE.String()
			}
			// update user
			user, err = secDb.Update(ctxt, user)
			if err != nil {
				AuditUser(ctxt, user, action, "FAIL_ERROR", user.UserId)
				response.WriteError(http.StatusInternalServerError, err)
			} else {
				AuditUser(ctxt, user, action, "SUCCESS")
				rest.WriteEntitySafely(response, user)
			}
		}
	}
}

func (userRest *UserRest) Put(request *restful.Request, response *restful.Response) {
	ACTION := message.User_UPDATE.String()
	ctx := request.Request.Context()
	allowed := sec.HasPermissionForAction(ctx, USER_CLASSID, ACTION)
	if !allowed {
		sec.Audit(ctx, USER_CLASSID, ACTION, "FAIL")
		response.WriteErrorString(http.StatusForbidden, "")
	} else {
		userId := request.PathParameter("user-id")
		user, err := secDb.FindOneByUserId(ctx, secDb.UserId(userId))
		if err != nil {
			response.WriteError(http.StatusBadRequest, err)
		} else {
			userRequest := new(UserRequest)
			err = request.ReadEntity(&userRequest)
			if err != nil {
				response.WriteError(http.StatusBadRequest, err)
			} else {
				// update password
				if userRequest.Password != "" {
					// check policy complexity
					errs := rest.SanitizeValidate(userRequest, SchemaUserUpdate(ctx))
					if errs != nil {
						sec.Audit(ctx, USER_CLASSID, ACTION, "FAIL", errs)
						response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
						return
					}
					// Password
					saltBytes := uuid.NewRandom()
					inputedHash := computeHmac256(userRequest.Password, saltBytes)
					userRequest.Password = ""
					user.PasswordHash = inputedHash
					// Salt
					user.Salt = saltBytes.String()
				}
				// update profile
				if userRequest.Profile != nil {
					user.Profile.LastName = userRequest.Profile.LastName
					user.Profile.FirstName = userRequest.Profile.FirstName
				}
				// update roles
				if userRequest.Roles != nil {
					// TODO check if roles have changed, if changed then check permission
					allowedRole := sec.HasPermissionForAction(ctx, USER_CLASSID, "UPDATE_ROLE")
					if allowedRole {
						// reset roles
						user.Roles = nil
						for _, roleId := range userRequest.Roles {
							// check if valid roleId
							if _, ok := message.RoleId_value[roleId]; ok {
								// check if roleId is present
								if !stringInSlice(roleId, user.Roles) {
									// add role
									user.Roles = append(user.Roles, roleId)
								}
							}
						}
						sec.Audit(ctx, USER_CLASSID, "UPDATE_ROLE", "SUCCESS", user.Roles)
					}
				}
				// update user
				user, err = secDb.Update(ctx, user)
				if err != nil {
					sec.Audit(ctx, USER_CLASSID, ACTION, "FAIL_ERROR", user.UserId)
					response.WriteError(http.StatusBadRequest, err)
				} else {
					AuditUser(ctx, user, ACTION, "SUCCESS", user.UserId)
					response.WriteHeaderAndEntity(http.StatusOK, user)
				}
			}
		}
	}
}

func (userRest *UserRest) Delete(request *restful.Request, response *restful.Response) {
	ACTION := message.User_DEACTIVATE.String()
	ctxt := request.Request.Context()
	allowed := sec.HasPermissionForAction(ctxt, USER_CLASSID, ACTION)
	if !allowed {
		sec.Audit(ctxt, USER_CLASSID, ACTION, "FAIL")
		response.WriteErrorString(http.StatusForbidden, "")
	} else {
		userId := request.PathParameter("user-id")
		user, err := secDb.FindOneByUserId(ctxt, secDb.UserId(userId))
		if err != nil {
			sec.Audit(ctxt, USER_CLASSID, ACTION, "FAIL_NOTFOUND")
			response.WriteError(http.StatusBadRequest, err)
		} else {
			user.UserId = secDb.UserId(string(user.UserId) + "__" + uuid.New())
			// set disabled aka inactive
			user.State = message.User_deactivated
			// update user
			user, err = secDb.Update(ctxt, user)
			if err != nil {
				sec.Audit(ctxt, USER_CLASSID, ACTION, "FAIL_ERROR", userId)
				response.WriteError(http.StatusBadRequest, err)
			} else {
				// special action RENAME
				AuditUser(ctxt, &secDb.User{UserId: secDb.UserId(userId)}, "RENAME", "SUCCESS", user.UserId)
				AuditUser(ctxt, user, ACTION, "SUCCESS", userId)
				response.WriteHeaderAndEntity(http.StatusOK, user)
			}
		}
	}
}

func (userRest *UserRest) Post(request *restful.Request, response *restful.Response) {
	ACTION := message.User_CREATE.String()
	ctxt := request.Request.Context()
	allowed := sec.HasPermissionForAction(ctxt, USER_CLASSID, ACTION)
	if !allowed {
		sec.Audit(ctxt, USER_CLASSID, ACTION, "FAIL")
		response.WriteErrorString(http.StatusForbidden, "")
	} else {
		userRequest := new(UserRequest)
		// check policy complexity
		errs := rest.SanitizeValidateReadEntity(request, SchemaUserCreate(ctxt), userRequest)
		if errs != nil {
			sec.Audit(ctxt, USER_CLASSID, ACTION, "FAIL", errs)
			response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
			return
		}
		user := new(secDb.User)
		// UserId
		user.UserId = secDb.UserId(userRequest.UserId)
		// Password
		saltBytes := uuid.NewRandom()
		inputedHash := computeHmac256(userRequest.Password, saltBytes)
		userRequest.Password = ""
		user.PasswordHash = inputedHash
		// Salt
		user.Salt = saltBytes.String()
		// State
		user.State = message.User_initialized
		// Roles
		user.Roles = make([]string, 0, len(userRequest.Roles))
		for _, roleId := range userRequest.Roles {
			if _, ok := message.RoleId_value[roleId]; ok {
				user.Roles = append(user.Roles, roleId)
			}
		}
		// Log password
		user.PasswordLog = make([]*message.PasswordLogEntry, 0, 1)
		user.PasswordLog = append(user.PasswordLog, &message.PasswordLogEntry{
			PasswordHash: user.PasswordHash,
			Timestamp:    tms.ToTimestamp(time.Now()),
		})
		// Profile
		user.Profile = &message.UserProfile{}
		if userRequest.Profile != nil {
			user.Profile.LastName = userRequest.Profile.LastName
			user.Profile.FirstName = userRequest.Profile.FirstName
		}
		// store
		newUserId, err := secDb.Add(ctxt, user)
		if err != nil {
			if err == secDb.ErrorDuplicate {
				sec.Audit(ctxt, USER_CLASSID, ACTION, "FAIL_DUPLICATE", err)
				errs = make([]rest.ErrorValidation, 1)
				errs[0] = rest.ErrorValidation{
					Property: "userId",
					Rule:     "Duplicate",
					Message:  "User ID is duplicate"}
				rest.WriteValidationErrsSafely(response, errs)
			} else {
				sec.Audit(ctxt, USER_CLASSID, ACTION, "FAIL_ERROR", err)
				response.WriteError(http.StatusInternalServerError, err)
			}
		} else {
			addedUser, err := secDb.FindOneByUserId(ctxt, user.UserId)
			if err != nil {
				sec.Audit(ctxt, USER_CLASSID, ACTION, "SUCCESS_ERROR", user.UserId, newUserId)
				response.WriteError(http.StatusInternalServerError, err)
			} else {
				AuditUser(ctxt, addedUser, ACTION, "SUCCESS", user.UserId, user.Roles, addedUser.Id)
				protoUser := mapUser(addedUser)
				rest.WriteHeaderAndProtoSafely(response, http.StatusCreated, protoUser)
			}
		}
	}
}

func AuditUser(context context.Context, user *secDb.User, action string, outcome string, payload ...interface{}) {
	security.AuditUserObject(context, USER_CLASSID, string(user.UserId), "", action, outcome, payload...)
}
