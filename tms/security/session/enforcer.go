package session

import (
	"context"
	"time"
	"strings"
	"prisma/tms/security/policy"
	"strconv"
)

func EnforceSessionDuration(ctxt context.Context, created time.Time) bool {
	durationMaximum, err := time.ParseDuration(policy.GetStore(ctxt).Get().Session.DurationMaximum)
	if err == nil {
		return time.Since(created) > durationMaximum
	}
	return false
}

func EnforceSessionRenewal(ctxt context.Context, created time.Time) bool {
	durationMaximum, err := time.ParseDuration(policy.GetStore(ctxt).Get().Session.DurationMaximum)
	if err == nil {
		durationRenewal, err := time.ParseDuration(policy.GetStore(ctxt).Get().Session.DurationRenewal)
		if err == nil {
			return durationMaximum > time.Since(created) && time.Since(created) > (durationMaximum - durationRenewal)
		}
	}
	return false
}

// returns roles
func EnforceSessionIdle(ctxt context.Context, lastAccess time.Time) (bool, []string) {
	durationIdle, err := time.ParseDuration(policy.GetStore(ctxt).Get().Session.DurationIdle)
	if err == nil {
		roleString := policy.GetStore(ctxt).Get().Session.IdleConsequence
		if time.Since(lastAccess) > durationIdle && "" != roleString {
			return true, strings.Split(roleString, ",")
		}
	}
	return false, nil
}

func EnforceSessionSingle(ctxt context.Context) bool {
	single, _ := strconv.ParseBool(policy.GetStore(ctxt).Get().Session.Single)
	return single
}