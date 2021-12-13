// Package service provides functions to determine acceptable Rest endpoints.
package service

import (
	"net/http"
	"strings"

	"prisma/tms/rest"
	sec "prisma/tms/security"
	securityContext "prisma/tms/security/context"
	"prisma/tms/security/session"

	"github.com/go-openapi/spec"
	restful "github.com/orolia/go-restful"
	"github.com/pborman/uuid"
	"golang.org/x/net/context"
)

const (
	ActionAny = "ANY"
)

var (
	COOKIE_SESSION_ID = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "id",
			In:       "cookie",
			Required: true,
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					MinLength: &[]int64{36}[0],
					MaxLength: &[]int64{36}[0],
					Pattern:   "[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}",
				},
			},
		},
		SimpleSchema: spec.SimpleSchema{
			Type:   "string",
			Format: "RFC4122",
		},
	}
	dialInfoValue   interface{}
	credentialValue interface{}
)

func SetDialInfo(dialInfo interface{}) {
	dialInfoValue = dialInfo

}
func SetCredential(credential interface{}) {
	credentialValue = credential
}

func RequestIdContextFilter(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	// TODO input sanitation
	requestId := request.Request.Header.Get("X-Request-ID")
	if requestId == "" {
		requestId = request.Request.Header.Get("X-Devtools-Request-Id")
		if requestId == "" {
			requestId = uuid.NewRandom().String()
		}
	}
	// add requestId for audit log
	request.Request = request.Request.WithContext(context.WithValue(request.Request.Context(), securityContext.RequestIdKey, requestId))
	// add MongoDB dial info
	request.Request = request.Request.WithContext(context.WithValue(request.Request.Context(), "mongodb", dialInfoValue))
	// add mongodb creds
	request.Request = request.Request.WithContext(context.WithValue(request.Request.Context(), "mongodb-cred", credentialValue))
	// add requestId to response CONV-1424
	if requestId != "" {
		response.AddHeader("X-Request-ID", requestId)
	}
	// enable headers to be read by javascript
	response.AddHeader(restful.HEADER_AccessControlExposeHeaders, strings.Join([]string{"Link", "X-Request-Id"}, ","))
	response.AddHeader(restful.HEADER_AccessControlAllowHeaders, strings.Join([]string{"Link", "X-Request-Id"}, ","))
	chain.ProcessFilter(request, response)
}

// checks cookie, session
func SessionIdContextFilter(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	routePath := request.SelectedRoutePath()
	if strings.HasSuffix(routePath, "/") {
		routePath = routePath[:len(routePath)-1]
	}
	check := !(strings.HasSuffix(routePath, "/session") ||
		strings.HasSuffix(routePath, "/api/v2/config.json") ||
		strings.HasSuffix(routePath, "/api/v2/sarmap.json") ||
		strings.HasSuffix(routePath, "/api/v2/apidocs.json"))
	if check {
		sessionID, errs := rest.SanitizeValidateCookie(request, COOKIE_SESSION_ID)
		if len(errs) > 0 {
			sec.Audit(request.Request.Context(), "Session", "READ", sec.FAIL_VALIDATION)
			// unauthorized
			response.WriteErrorString(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		} else {
			// check session
			sessionStored, err := session.GetStore(request.Request.Context()).Get(sessionID)
			if session.ErrorNotFound == err {
				sec.Audit(request.Request.Context(), "Session", "READ", sec.FAIL_NOTFOUND, sessionID)
				response.WriteErrorString(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
			} else if session.ErrorInvalidId == err {
				sec.Audit(request.Request.Context(), "Session", "READ", sec.FAIL_VALIDATION, sessionID)
				response.WriteErrorString(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
			} else if session.ErrorExpired == err {
				sec.Audit(request.Request.Context(), "Session", "READ", "FAIL_EXPIRED", sessionID)
				response.WriteErrorString(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
			} else {
				// add sessionId for audit log
				request.Request = request.Request.WithContext(context.WithValue(request.Request.Context(), securityContext.SessionIdKey, sessionID))
				// add session in context
				request.Request = request.Request.WithContext(context.WithValue(request.Request.Context(), securityContext.SessionKey, sessionStored))
				chain.ProcessFilter(request, response)
			}
		}
	} else {
		chain.ProcessFilter(request, response)
	}
}

type RouteAuthorizer struct {
	routeMatchers []rest.RouteMatcher
}

func (a *RouteAuthorizer) Add(routeMatcher rest.RouteMatcher) {
	if a.routeMatchers == nil {
		a.routeMatchers = make([]rest.RouteMatcher, 0)
	}
	a.routeMatchers = append(a.routeMatchers, routeMatcher)
}

func (a *RouteAuthorizer) Authorize(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	routePath := request.SelectedRoutePath()
	ctx := request.Request.Context()
	if strings.HasSuffix(routePath, "/") {
		routePath = routePath[:len(routePath)-1]
	}
	// exclusions.  if excluded then unprotected
	check := !(strings.HasSuffix(routePath, "/session") ||
		strings.HasSuffix(routePath, "/api/v2/config.json") ||
		strings.HasSuffix(routePath, "/api/v2/sarmap.json") ||
		strings.HasSuffix(routePath, "/api/v2/apidocs.json"))
	if check {
		allowed := false
		for _, routeMatcher := range a.routeMatchers {
			match, classId := routeMatcher.MatchRoute(routePath)
			if match {
				allowed = sec.HasPermissionForClass(ctx, classId)
				if !allowed {
					sec.Audit(ctx, classId, ActionAny, sec.FAIL_PERMISSION, routePath)
				}
				break
			}
		}
		// user
		if strings.HasSuffix(routePath, "/user") ||
			strings.HasSuffix(routePath, "/user/{user-id}") ||
			strings.HasSuffix(routePath, "/user/{user-id}/state/{user-state}") {
			allowed = sec.HasPermissionForClass(request.Request.Context(), "User")
			if !allowed {
				sec.Audit(request.Request.Context(), "User", "ANY", "FAIL", routePath)
			}
		}
		// profile
		if strings.HasSuffix(routePath, "/profile/{user-id}") {
			allowed = sec.HasPermissionForClass(request.Request.Context(), "Profile")
			if !allowed {
				sec.Audit(request.Request.Context(), "Profile", "ANY", "FAIL", routePath)
			}
		}
		// role
		if strings.HasSuffix(routePath, "/role") ||
			strings.HasSuffix(routePath, "role/{role-id}/user") {
			allowed = sec.HasPermissionForClass(request.Request.Context(), "Role")
			if !allowed {
				sec.Audit(request.Request.Context(), "Role", "ANY", "FAIL", routePath)
			}
		}
		// object
		if strings.HasSuffix(routePath, "/object") {
			allowed = sec.HasPermissionForClass(request.Request.Context(), "Object")
			if !allowed {
				sec.Audit(request.Request.Context(), "Object", "ANY", "FAIL", routePath)
			}
		}
		// fleet
		if strings.HasSuffix(routePath, "/fleet") || strings.HasSuffix(routePath, "/fleet/{fleet-id}") || strings.HasSuffix(routePath, "/fleet/meta") {
			allowed = sec.HasPermissionForClass(request.Request.Context(), "Fleet")
			if !allowed {
				sec.Audit(request.Request.Context(), "Fleet", "ANY", "FAIL", routePath)
			}
		}
		// device
		if strings.HasSuffix(routePath, "/device") || strings.HasSuffix(routePath, "/device/{id}") || strings.HasSuffix(routePath, "/device/meta") {
			allowed = sec.HasPermissionForClass(request.Request.Context(), "Device")
			if !allowed {
				sec.Audit(request.Request.Context(), "Device", "ANY", "FAIL", routePath)
			}
		}
		// rule
		if strings.HasSuffix(routePath, "/rule") ||
			strings.HasSuffix(routePath, "/rule/meta") ||
			strings.HasSuffix(routePath, "/rule/{rule-id}") ||
			strings.HasSuffix(routePath, "/rule/{rule-id}/state/{state-id}") {
			allowed = sec.HasPermissionForClass(request.Request.Context(), sec.RULE_CLASS_ID)
			if !allowed {
				sec.Audit(request.Request.Context(), sec.RULE_CLASS_ID, "ANY", "FAIL", routePath)
			}
		}
		// vessel
		if strings.HasSuffix(routePath, "/vessel") || strings.HasSuffix(routePath, "/vessel/{vessel-id}") || strings.HasSuffix(routePath, "/vessel/meta") {
			allowed = sec.HasPermissionForClass(request.Request.Context(), "Vessel")
			if !allowed {
				sec.Audit(request.Request.Context(), "Vessel", "ANY", "FAIL", routePath)
			}
		}
		// zone
		if strings.HasSuffix(routePath, "/zone") ||
			strings.HasSuffix(routePath, "/zone/geo") ||
			strings.HasSuffix(routePath, "/zone/{zone-id}") {
			allowed = sec.HasPermissionForClass(request.Request.Context(), "Zone")
			if !allowed {
				sec.Audit(request.Request.Context(), "Zone", "ANY", "FAIL", routePath)
			}
		}
		// geofence
		if strings.HasSuffix(routePath, "/geofence") || strings.HasSuffix(routePath, "/geofence/{geofence-id}") {
			allowed = sec.HasPermissionForClass(request.Request.Context(), "Geofence")
			if !allowed {
				sec.Audit(request.Request.Context(), "Geofence", "ANY", "FAIL", routePath)
			}
		}
		// search
		if strings.HasSuffix(routePath, "/search/tables/{tables}") {
			allowed = sec.HasPermissionForClass(request.Request.Context(), "Search")
			if !allowed {
				sec.Audit(request.Request.Context(), "Search", "ANY", "FAIL", routePath)
			}
		}
		// audit
		if strings.HasSuffix(routePath, "/audit") ||
			strings.HasSuffix(routePath, "/audit/session/{session-id}") ||
			strings.HasSuffix(routePath, "/audit/user/{user-id}") {
			allowed = sec.HasPermissionForClass(request.Request.Context(), "Audit")
			if !allowed {
				sec.Audit(request.Request.Context(), "Audit", "ANY", "FAIL", routePath)
			}
		}
		// notices
		if strings.HasSuffix(routePath, "/notice") ||
			strings.HasSuffix(routePath, "/notice/history") ||
			strings.HasSuffix(routePath, "/notice/new") ||
			strings.HasSuffix(routePath, "/notice/{id}/ack") ||
			strings.HasSuffix(routePath, "/notice/all/ack") ||
			strings.HasSuffix(routePath, "/notice/timeout") {
			allowed = sec.HasPermissionForClass(request.Request.Context(), "Notice")
			if !allowed {
				sec.Audit(request.Request.Context(), "Notice", "ANY", "FAIL", routePath)
			}
		}
		// device
		if strings.HasSuffix(routePath, "/device") ||
			strings.HasSuffix(routePath, "/device/vessel/{vessel-id}") ||
			strings.HasSuffix(routePath, "/device/{device-id}") {
			allowed = sec.HasPermissionForClass(request.Request.Context(), "Device")
			if !allowed {
				sec.Audit(request.Request.Context(), "Device", "ANY", "FAIL", routePath)
			}
		}
		// incident
		if strings.HasSuffix(routePath, "/incident") ||
			strings.HasSuffix(routePath, "/incident/{incident-id}") ||
			strings.HasSuffix(routePath, "/incident/{incident-id}/processing") ||
			strings.HasSuffix(routePath, "/incident/{incident-id}/log") ||
			strings.HasSuffix(routePath, "/incident/{incident-id}/log/{log-id}") ||
			strings.HasSuffix(routePath, "/incident/{incident-id}/log-detach/{log-id}") ||
			strings.HasSuffix(routePath, "/incident/{incident-id}/state/{incident-state}") ||
			strings.HasSuffix(routePath, "/incident/{incident-id}/assignee/{user-id}") ||
			strings.HasSuffix(routePath, "/audit/incident/{incident-id}") {
			allowed = sec.HasPermissionForClass(request.Request.Context(), "Incident")
			if !allowed {
				sec.Audit(request.Request.Context(), "Incident", "ANY", "FAIL", routePath)
			}
		}
		// note
		if strings.HasSuffix(routePath, "/note") ||
			strings.HasSuffix(routePath, "/note/{note-id}") ||
			strings.HasSuffix(routePath, "/note/{note-id}/{is-assigned}") ||
			strings.HasSuffix(routePath, "/note/{note-id}/assignee/{incident-id}") {
			allowed = sec.HasPermissionForClass(request.Request.Context(), "IncidentLogEntry")
			if !allowed {
				sec.Audit(request.Request.Context(), "Note", "ANY", "FAIL", routePath)
			}
		}
		// marker
		if strings.HasSuffix(routePath, "/marker") ||
			strings.HasSuffix(routePath, "/marker/{marker-id}") ||
			strings.HasSuffix(routePath, "/marker/image") {
			allowed = sec.HasPermissionForClass(request.Request.Context(), "Marker")
			if !allowed {
				sec.Audit(request.Request.Context(), "Marker", "ANY", "FAIL", routePath)
			}
		}
		// icon
		if strings.HasSuffix(routePath, "/icon") ||
			strings.HasSuffix(routePath, "/icon/{icon-id}") ||
			strings.HasSuffix(routePath, "/icon/image") {
			allowed = sec.HasPermissionForClass(request.Request.Context(), "Icon")
			if !allowed {
				sec.Audit(request.Request.Context(), "Icon", "ANY", "FAIL", routePath)
			}
		}
		// file
		if strings.HasSuffix(routePath, "/file") ||
			strings.HasSuffix(routePath, "/file/{file-id}") {
			allowed = sec.HasPermissionForClass(request.Request.Context(), "File")
			if !allowed {
				sec.Audit(request.Request.Context(), "File", "ANY", "FAIL", routePath)
			}
		}
		// view
		if strings.HasSuffix(routePath, "/view") ||
			strings.HasSuffix(routePath, "/view/stream") {
			allowed = sec.HasPermissionForClass(request.Request.Context(), "View")
			if !allowed {
				sec.Audit(request.Request.Context(), "View", "ANY", "FAIL", routePath)
			}
		}
		// multicast
		if strings.HasSuffix(routePath, "/multicast/device/{device-id}") {
			allowed = sec.HasPermissionForClass(request.Request.Context(), "Multicast")
			if !allowed {
				sec.Audit(request.Request.Context(), "Multicast", "ANY", "FAIL", routePath)
			}
		}
		// remotesite
		if strings.HasSuffix(routePath, "/remotesite") ||
			strings.HasSuffix(routePath, "/remotesite/{remotesite-id}") {
			allowed = sec.HasPermissionForClass(request.Request.Context(), "RemoteSite")
			if !allowed {
				sec.Audit(request.Request.Context(), "RemoteSite", "ANY", "FAIL", routePath)
			}
		}
		// track
		if strings.HasSuffix(routePath, "/track") ||
			strings.HasSuffix(routePath, "/track/{track-id}") ||
			strings.HasSuffix(routePath, "/track/{registry-id}") {
			allowed = sec.HasPermissionForClass(request.Request.Context(), "Track")
			if !allowed {
				sec.Audit(request.Request.Context(), "Track", "ANY", "FAIL", routePath)
			}
		}
		// track history
		if strings.HasSuffix(routePath, "/history/{registry-id}") ||
			strings.HasSuffix(routePath, "/history-database/{database-id}") {
			allowed = sec.HasPermissionForClass(request.Request.Context(), "History")
			if !allowed {
				sec.Audit(request.Request.Context(), "Track", "ANY", "FAIL", routePath)
			}
		}
		// registry
		if strings.HasSuffix(routePath, "/registry/{registry-id}") ||
			strings.HasSuffix(routePath, "/search/registry") {
			allowed = sec.HasPermissionForClass(request.Request.Context(), "Registry")
			if !allowed {
				sec.Audit(request.Request.Context(), "Registry", "ANY", "FAIL", routePath)
			}
		}
		// policy
		if strings.HasSuffix(routePath, "/policy") {
			allowed = sec.HasPermissionForClass(request.Request.Context(), "Policy")
			if !allowed {
				sec.Audit(request.Request.Context(), "Policy", "ANY", "FAIL", routePath)
			}
		}
		// config
		if strings.HasSuffix(routePath, "/config") {
			allowed = sec.HasPermissionForClass(request.Request.Context(), sec.CLASSIDConfig)
			if !allowed {
				sec.Audit(request.Request.Context(), sec.CLASSIDConfig, "ANY", "FAIL", routePath)
			}
		}
		// sit915
		if strings.HasSuffix(routePath, "/sit915") ||
			strings.HasSuffix(routePath, "/sit915/{comm-link-type}/{remotesite-id}") ||
			strings.HasSuffix(routePath, "/sit915/ack/{message-id}") ||
			strings.HasSuffix(routePath, "/sit915/retry/{message-id}") ||
			strings.HasSuffix(routePath, "/sit915/{message-id}") {
			allowed = sec.HasPermissionForClass(request.Request.Context(), sec.CLASSIDSit915)
			if !allowed {
				sec.Audit(request.Request.Context(), sec.CLASSIDSit915, "ANY", "FAIL", routePath)
			}
		}
		// mapconfig
		if strings.HasSuffix(routePath, "/mapconfig") ||
			strings.HasSuffix(routePath, "/mapconfig/set") {
			allowed = sec.HasPermissionForClass(request.Request.Context(), "MapConfig")
			if !allowed {
				sec.Audit(request.Request.Context(), "MapConfig", "ANY", "FAIL", routePath)
			}
		}
		// message
		if strings.HasSuffix(routePath, "/message") {
			allowed = sec.HasPermissionForClass(request.Request.Context(), sec.CLASSIDMessage)
			if !allowed {
				sec.Audit(request.Request.Context(), sec.CLASSIDMessage, "ANY", "FAIL", routePath)
			}
		}
		if !allowed {
			// forbidden
			chain.Target = func(request *restful.Request, response *restful.Response) {
				response.WriteErrorString(http.StatusForbidden, http.StatusText(http.StatusForbidden))
			}
		}
	}
	chain.ProcessFilter(request, response)
}
