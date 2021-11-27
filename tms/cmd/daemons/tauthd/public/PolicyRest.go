package public

import (
	"net/http"
	"prisma/tms/rest"
	"prisma/tms/security"
	"prisma/tms/security/policy"

	restful "github.com/orolia/go-restful"
	"github.com/go-openapi/spec"
)

var (
	// schemas
	SCHEMA_POLICY_UPDATE = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Required:   []string{"password", "session", "user"},
			Properties: map[string]spec.Schema{},
		},
	}
)

type PolicyRest struct {
	store      policy.Store
	restUpdate func()
}

func NewPolicyRest(restUpdate func()) *PolicyRest {
	return &PolicyRest{
		restUpdate: restUpdate,
	}
}

// Read the policy, by the admin.
func (policyRest *PolicyRest) Read(request *restful.Request, response *restful.Response) {
	CLASSID := security.POLICY_CLASS_ID
	ACTION := policy.Policy_READ.String()
	ctxt := request.Request.Context()
	if security.HasPermissionForAction(ctxt, CLASSID, ACTION) {
		policies := policy.GetStore(ctxt).Get()
		security.Audit(ctxt, CLASSID, ACTION, "SUCCESS")
		rest.WriteEntitySafely(response, policies)
	} else {
		security.Audit(ctxt, CLASSID, ACTION, "FAIL")
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}

// Update the policy, by the admin.
func (policyRest *PolicyRest) Update(request *restful.Request, response *restful.Response) {
	ctxt := request.Request.Context()
	CLASSID := security.POLICY_CLASS_ID
	ACTION := policy.Policy_UPDATE.String()
	if !security.HasPermissionForAction(ctxt, CLASSID, ACTION) {
		security.Audit(ctxt, CLASSID, ACTION, "FAIL")
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
		return
	}

	policyRequest := new(policy.Policy)
	errs := rest.SanitizeValidateReadEntity(request, SCHEMA_POLICY_UPDATE, policyRequest)
	if errs != nil {
		security.Audit(request.Request.Context(), CLASSID, ACTION, "FAIL_ERROR")
		response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		return
	}
	// save policy
	err := policy.GetStore(ctxt).Set(policyRequest)
	if err != nil {
		policyRequest.Description = err.Error()
	}
	security.Audit(ctxt, CLASSID, ACTION, "SUCCESS")
	rest.WriteEntitySafely(response, policyRequest)
	// update webservices, swagger
	policyRest.restUpdate()
}
