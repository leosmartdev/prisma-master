package public

import (
	"context"
	"net/http"
	"prisma/tms"
	"prisma/tms/rest"
	sec "prisma/tms/security"
	secDb "prisma/tms/security/database"
	"prisma/tms/security/message"
	"time"

	"github.com/go-openapi/spec"
	restful "github.com/orolia/go-restful"
	"github.com/pborman/uuid"
)

type ProfileRest struct{}

func SchemaProfileUpdate(ctx context.Context) spec.Schema {
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

func (profileRest *ProfileRest) Put(request *restful.Request, response *restful.Response) {
	const CLASSID = "Profile"
	ACTION := message.User_UPDATE.String()
	ctxt := request.Request.Context()
	allowed := sec.HasPermissionForClass(ctxt, CLASSID)
	if !allowed {
		sec.Audit(ctxt, CLASSID, "ANY", "FAIL")
		response.WriteErrorString(http.StatusForbidden, "")
	} else {
		userId := request.PathParameter("user-id")
		user, err := secDb.FindOneByUserId(ctxt, secDb.UserId(userId))
		if err != nil {
			response.WriteError(http.StatusBadRequest, err)
		} else {
			userRequest := new(UserRequest)
			// check policy complexity
			errs := rest.SanitizeValidateReadEntity(request, SchemaProfileUpdate(ctxt), userRequest)
			if errs != nil {
				sec.Audit(request.Request.Context(), CLASSID, "UPDATE_PASSWORD", "FAIL_ERROR")
				response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
				return
			} else {
				// update password
				allowedPassword := sec.HasPermissionForActionOnObject(ctxt, CLASSID, "UPDATE_PASSWORD", userId)
				// allowed and new password
				// FIXME: Put these checks back in
				//allowedPassword = allowedPassword && userRequest.Password != ""
				allowedPassword = userRequest.Password != ""
				if allowedPassword {
					// Password
					saltBytes := uuid.Parse(user.Salt)
					// policy check
					usedUserId := sec.EnforceProhibitUserId(ctxt, userId, userRequest.Password)
					if usedUserId {
						sec.Audit(ctxt, USER_CLASSID, ACTION, "FAIL_USER_ID", userId)
						errs = make([]rest.ErrorValidation, 1)
						errs[0] = rest.ErrorValidation{
							Property: "password",
							Rule:     "ProhibitUserId",
							Message:  "Password cannot be the username"}
						rest.WriteValidationErrsSafely(response, errs)
						return
					}
					passwordHash := computeHmac256(userRequest.Password, saltBytes)
					userRequest.Password = ""
					enforced := sec.EnforcePasswordReuseMaximum(ctxt, passwordHash, user.PasswordLog)
					if enforced {
						sec.Audit(ctxt, CLASSID, "UPDATE_PASSWORD", "FAIL_POLICY", user.UserId)
						errs = make([]rest.ErrorValidation, 1)
						errs[0] = rest.ErrorValidation{
							Property: "password",
							Rule:     "Policy",
							Message:  "Password policy - reuse"}
						rest.WriteValidationErrsSafely(response, errs)
						return
					}
					user.PasswordHash = passwordHash
					// update password log
					user.PasswordLog = append(user.PasswordLog, &message.PasswordLogEntry{
						PasswordHash: user.PasswordHash,
						Timestamp:    tms.ToTimestamp(time.Now()),
					})
					// activate user after password change
					if message.User_initialized == user.State {
						user.State = message.User_activated
						sec.Audit(ctxt, "User", "ACTIVATED", "SUCCESS", user.UserId)
					}
				}
				// update profile
				allowedProfile := sec.HasPermissionForActionOnObject(ctxt, CLASSID, "UPDATE", userId)
				// FIXME: Put these checks back in
				//if allowedProfile {
				if userRequest.Profile != nil {
					user.Profile.LastName = userRequest.Profile.LastName
					user.Profile.FirstName = userRequest.Profile.FirstName
				}
				// update user
				// FIXME: Put these checks back in
				//if allowedPassword || allowedProfile {
				if true {
					user, err = secDb.Update(ctxt, user)
					if err != nil {
						sec.Audit(ctxt, CLASSID, "UPDATE", "FAIL_ERROR", user.UserId)
						response.WriteError(http.StatusBadRequest, err)
					} else {
						if allowedPassword {
							sec.Audit(ctxt, CLASSID, "UPDATE_PASSWORD", "SUCCESS", user.UserId)
						}
						if allowedProfile {
							sec.Audit(ctxt, CLASSID, "UPDATE", "SUCCESS", user.UserId)
						}
						response.WriteHeaderAndEntity(http.StatusOK, user)
					}
				} else {
					sec.Audit(ctxt, CLASSID, "ANY", "FAIL")
					response.WriteErrorString(http.StatusForbidden, "")
				}
			}
		}
	}
}
