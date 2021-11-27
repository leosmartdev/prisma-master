package public

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"

	"prisma/tms/log"
	"prisma/tms/public"
	"prisma/tms/rest"
	sec "prisma/tms/security"
	securityContext "prisma/tms/security/context"
	secDb "prisma/tms/security/database"
	"prisma/tms/security/message"
	"prisma/tms/security/session"
	"prisma/tms/ws"

	"prisma/tms/envelope"

	"github.com/go-openapi/spec"
	restful "github.com/orolia/go-restful"
	"github.com/pborman/uuid"
	"golang.org/x/net/context"
)

type SessionRest struct {
	publisher *ws.Publisher
}

func NewSessionRest(ctxt context.Context, publisher *ws.Publisher) *SessionRest {
	session.SetPublisher(ctxt, &SessionPublisher{
		ctxt:      ctxt,
		publisher: publisher,
	})
	return &SessionRest{
		publisher: publisher,
	}
}

// SessionPublisher is used for workaround of cycle import
type SessionPublisher struct {
	ctxt      context.Context
	publisher *ws.Publisher
}

func (p *SessionPublisher) Publish(action message.Session_Action, s session.InternalSession) {
	const TOPIC = "Session"
	envelope := envelope.Envelope{
		Type:   TOPIC + "/" + action.String(),
		Source: s.Id(),
		Contents: &envelope.Envelope_Session{Session: &message.Session{
			State: s.GetState(),
			User: &message.User{
				UserId: s.GetOwner(),
			},
			Permissions: sec.PermissionsFromSession(p.ctxt, s),
		}},
	}
	p.publisher.Publish(TOPIC, envelope)
}

func SchemaSessionCreate(ctx context.Context) spec.Schema {
	return spec.Schema{
		SchemaProps: spec.SchemaProps{
			Required: []string{"userName", "token"},
			Properties: map[string]spec.Schema{
				"userName": {
					SchemaProps: spec.SchemaProps{
						Type:      []string{"string"},
						MinLength: &[]int64{1}[0],
						MaxLength: &[]int64{255}[0],
					},
				},
				"token": {
					SchemaProps: sec.PasswordSchemaProps(ctx),
				},
			},
		},
	}
}

func (sessionRest *SessionRest) Post(request *restful.Request, response *restful.Response) {
	ctxt := request.Request.Context()
	loginRequest := new(public.LoginRequest)
	shsc := SchemaSessionCreate(ctxt)
	errs := rest.SanitizeValidateReadEntity(request, shsc, loginRequest)
	if errs != nil {
		log.Debug("%+v", errs)
		sec.Audit(ctxt, "User", "AUTHENTICATE", "FAIL_ERROR")
		response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		return
	}

	password := loginRequest.Token
	loginRequest.Token = ""
	username := loginRequest.UserName
	user, err := secDb.FindOneByUserId(ctxt, secDb.UserId(username))
	if err == nil {
		decodedSalt := uuid.Parse(user.Salt)
		inputedHash := computeHmac256(password, []byte(decodedSalt))
		password = ""
		// validate login info
		if isValidUser(*user, username, inputedHash) {
			// policy check
			roles := user.Roles
			enforced, enforcedRoles := sec.EnforcePasswordDuration(ctxt, user.PasswordLog)
			if enforced {
				roles = enforcedRoles
			}
			enforced, enforcedRoles = sec.EnforceAuthenticateInitial(ctxt, user.State)
			if enforced {
				roles = enforcedRoles
			} else if message.User_initialized == user.State {
				// activate user, after first login (after policy checks)
				user.State = message.User_activated
				_, err = secDb.Update(ctxt, user)
				if err == nil {
					sec.Audit(ctxt, "User", message.User_ACTIVATE.String(), "SUCCESS", user.UserId)
				}
			}
			// check if old session (re-authenticate)
			sessionCookie, err := request.Request.Cookie("id")
			oldSessionId := ""
			if err == nil {
				// TODO add check to see if owner of session
				oldSessionId = sessionCookie.Value
				_ = session.GetStore(ctxt).Delete(oldSessionId)
				sec.Audit(request.Request.Context(), "Session", message.Session_TERMINATE.String(), "SUCCESS", oldSessionId, request.Request.RemoteAddr)
				// publish
				const TOPIC = "Session"
				envelope := envelope.Envelope{
					Type:   TOPIC + "/" + message.Session_TERMINATE.String(),
					Source: oldSessionId,
					Contents: &envelope.Envelope_Session{Session: &message.Session{
						State: message.Session_terminated,
					}},
				}
				if nil != sessionRest.publisher {
					sessionRest.publisher.Publish(TOPIC, envelope)
				}
			}
			// reset attempts after successful
			user.Attempts = 0
			_, err = secDb.Update(ctxt, user)
			// create session
			storedSession, err := session.GetStore(ctxt).Create(username, roles)
			if nil == err {
				// create session cookie
				//fmt.Println(request.Request.TLS.TLSUnique)
				sessionCookie := &http.Cookie{
					Name:     "id",
					Value:    storedSession.Id(),
					Path:     "/",
					Secure:   true,
					HttpOnly: true,
				}
				responseWriter := response.ResponseWriter
				http.SetCookie(responseWriter, sessionCookie)
				sessionPublic := new(message.Session)
				sessionPublic.User = mapUser(user)
				sessionPublic.Permissions = sec.PermissionsFromRolesString(roles)
				sessionPublic.State = storedSession.GetState()
				// add sessionId for audit log
				request.Request = request.Request.WithContext(context.WithValue(ctxt, securityContext.SessionIdKey, storedSession.Id()))
				sec.Audit(ctxt, "User", "AUTHENTICATE", "SUCCESS", user.UserId, storedSession.Id())
				sec.Audit(ctxt, "Session", "CREATE", "SUCCESS", user.UserId, storedSession.Id())
				rest.WriteProtoSafely(response, sessionPublic)
			} else {
				sec.Audit(ctxt, "User", "AUTHENTICATE", "FAIL_ERROR", loginRequest, err)
			}
		} else {
			errs = make([]rest.ErrorValidation, 1)
			if user.State == message.User_deactivated {
				sec.Audit(ctxt, "User", "AUTHENTICATE", "FAIL_DEACTIVATED", loginRequest, request.Request.RemoteAddr)
				errs[0] = rest.ErrorValidation{
					Property: "userName",
					Rule:     "Policy",
					Message:  "account is deactivated"}
			} else if user.State == message.User_locked {
				sec.Audit(ctxt, "User", "AUTHENTICATE", "FAIL_LOCKED", loginRequest, request.Request.RemoteAddr)
				errs[0] = rest.ErrorValidation{
					Property: "userName",
					Rule:     "Policy",
					Message:  "account is locked"}
			} else if string(user.UserId) != username {
				sec.Audit(ctxt, "User", "AUTHENTICATE", "FAIL_MISMATCH", loginRequest, request.Request.RemoteAddr)
				errs[0] = rest.ErrorValidation{
					Property: "userName",
					Rule:     "Policy",
					Message:  "username mismatch"}
			} else {
				// update attempts
				user.Attempts = user.Attempts + 1
				enforced, enforcedState := sec.EnforceAuthenticateFailedCountMaximum(ctxt, user.Attempts)
				if enforced {
					user.State = enforcedState
				}
				_, err = secDb.Update(ctxt, user)
				if user.State == message.User_locked {
					sec.Audit(ctxt, "User", "AUTHENTICATE", "FAIL_LOCKED", loginRequest, request.Request.RemoteAddr)
					errs[0] = rest.ErrorValidation{
						Property: "userName",
						Rule:     "Policy",
						Message:  "account is locked"}
				} else {
					sec.Audit(ctxt, "User", "AUTHENTICATE", "FAIL_PASSWORD", loginRequest, request.Request.RemoteAddr, user.Attempts)
					errs[0] = rest.ErrorValidation{
						Property: "token",
						Rule:     "Policy",
						Message:  "password is incorrect"}
				}
			}
			rest.WriteValidationErrsSafely(response, errs)
		}
	} else {
		sec.Audit(ctxt, "User", "AUTHENTICATE", "FAIL_NOTFOUND", loginRequest, request.Request.RemoteAddr)
		errs = make([]rest.ErrorValidation, 1)
		errs[0] = rest.ErrorValidation{
			Property: "userName",
			Rule:     "Policy",
			Message:  "username not found"}
		rest.WriteValidationErrsSafely(response, errs)
	}
}

func (sessionRest *SessionRest) Get(request *restful.Request, response *restful.Response) {
	sessionCookie, err := request.Request.Cookie("id")
	if err != nil {
		response.WriteError(http.StatusBadRequest, err)
		return
	}
	ctxt := request.Request.Context()
	sessionStored, err := session.GetStore(ctxt).Get(sessionCookie.Value)
	if err != nil {
		response.WriteError(http.StatusBadRequest, err)
		return
	}
	sessionPublic := new(message.Session)
	username := sessionStored.GetOwner()
	user, err := secDb.FindOneByUserId(ctxt, secDb.UserId(username))
	if err == nil {
		sessionPublic.User = mapUser(user)
	}
	sessionPublic.Permissions = sec.PermissionsFromSession(ctxt, sessionStored)
	sessionPublic.State = sessionStored.GetState()
	rest.WriteProtoSafely(response, sessionPublic)
}

func (sessionRest *SessionRest) Delete(request *restful.Request, response *restful.Response) {
	sessionCookie, err := request.Request.Cookie("id")
	sessionId := ""
	if err == nil {
		// TODO add check to see if owner of session
		sessionId = sessionCookie.Value
		_ = session.GetStore(request.Request.Context()).Delete(sessionId)
		sec.Audit(request.Request.Context(), "Session", message.Session_TERMINATE.String(), "SUCCESS", sessionId, request.Request.RemoteAddr)
	}
	responseWriter := response.ResponseWriter
	// delete cookie
	sessionCookie = &http.Cookie{
		Name:   "id",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
		//		Secure: true,
		HttpOnly: true,
	}

	http.SetCookie(responseWriter, sessionCookie)
	response.WriteHeaderAndEntity(http.StatusAccepted, "{}")
	// publish
	const TOPIC = "Session"
	envelope := envelope.Envelope{
		Type:   TOPIC + "/" + message.Session_TERMINATE.String(),
		Source: sessionId,
		Contents: &envelope.Envelope_Session{Session: &message.Session{
			State: message.Session_terminated,
		}},
	}
	if nil != sessionRest.publisher {
		sessionRest.publisher.Publish(TOPIC, envelope)
	}
}

// computeHmac256 calculates the HMAC for the given password and salt, used for comparison.
func computeHmac256(password string, salt []byte) string {
	key := []byte(salt)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(password))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func mapUser(user *secDb.User) *message.User {
	return &message.User{
		UserId:  string(user.UserId),
		Profile: user.Profile,
		Roles:   user.Roles,
		State:   user.State,
	}
}

func isValidUser(user secDb.User, username, inputedHash string) bool {
	return string(user.UserId) == username &&
		hmac.Equal([]byte(user.PasswordHash), []byte(inputedHash)) &&
		user.State != message.User_deactivated &&
		user.State != message.User_locked
}
