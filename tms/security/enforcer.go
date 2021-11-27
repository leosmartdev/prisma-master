package security

import (
	"context"
	"prisma/tms"
	"prisma/tms/security/message"
	"prisma/tms/security/policy"
	"strconv"
	"strings"
	"time"

	"github.com/go-openapi/spec"
)

// Session Enforce in session/enforcer

// Checks password age is within limit
// returns roles
func EnforcePasswordDuration(ctxt context.Context, passwordLog []*message.PasswordLogEntry) (bool, []string) {
	durationMaximum, err := time.ParseDuration(policy.GetStore(ctxt).Get().Password.DurationMaximum)
	if err == nil && len(passwordLog) > 0 {
		lastPasswordCreated := passwordLog[len(passwordLog)-1].Timestamp
		duration := time.Since(tms.FromTimestamp(lastPasswordCreated))
		if duration > durationMaximum {
			roleString := policy.GetStore(ctxt).Get().Password.DurationMaximumConsequence
			return true, strings.Split(roleString, ",")
		}
	}
	return false, nil
}

// Checks initial login
// returns roles
func EnforceAuthenticateInitial(ctxt context.Context, state message.User_State) (bool, []string) {
	if message.User_initialized == state {
		roleString := policy.GetStore(ctxt).Get().Password.AuthenticateInitialConsequence
		if "" != roleString {
			return true, strings.Split(roleString, ",")
		}
	}
	return false, nil
}

func EnforceAuthenticateFailedCountMaximum(ctxt context.Context, attempts int) (bool, message.User_State) {
	failedCount, err := strconv.Atoi(policy.GetStore(ctxt).Get().Password.AuthenticateFailedCountMaximum)
	if err == nil {
		if attempts > failedCount {
			consequence := policy.GetStore(ctxt).Get().Password.GetAuthenticateFailedMaximumConsequence()
			if message.User_LOCK.String() == consequence {
				return true, message.User_locked
			}
		}
	}
	return false, message.User_nonstate
}

func EnforcePasswordReuseMaximum(ctxt context.Context, passwordHash string, passwordLog []*message.PasswordLogEntry) bool {
	remainingCount, err := strconv.Atoi(policy.GetStore(ctxt).Get().Password.ReuseMaximum)
	if err == nil {
		// last log entry is the latest
		index := len(passwordLog) - 1
		// until no more elements or specified count is checked
		for index > -1 && remainingCount > 0 {
			if passwordHash == passwordLog[index].PasswordHash {
				return true
			}
			remainingCount--
			index--
		}
	}
	return false
}

func EnforceInactiveDuration(ctxt context.Context, inactiveDuration time.Duration) (bool, message.User_State) {
	userState := message.User_nonstate
	durationLock, err := time.ParseDuration(policy.GetStore(ctxt).Get().User.InactiveDurationConsequenceLock)
	durationDeactivate, err := time.ParseDuration(policy.GetStore(ctxt).Get().User.InactiveDurationConsequenceDeactivate)
	if err == nil {
		if inactiveDuration > durationDeactivate {
			return true, message.User_deactivated
		}
		if inactiveDuration > durationLock {
			return true, message.User_locked
		}
	}
	return false, userState
}

func EnforceProhibitUserId(ctxt context.Context, userId string, password string) bool {
	prohibtUserId := policy.GetStore(ctxt).Get().Password.ProhibitUserId
	if prohibtUserId && userId == password {
		return true
	}
	return false
}

func PasswordSchemaProps(ctxt context.Context) spec.SchemaProps {
	minLength, _ := strconv.ParseInt(policy.GetStore(ctxt).Get().Password.LengthMinimum, 10, 0)
	maxLength, _ := strconv.ParseInt(policy.GetStore(ctxt).Get().Password.LengthMaximum, 10, 0)
	return spec.SchemaProps{
		Type:      []string{"string"},
		MinLength: &[]int64{minLength}[0],
		MaxLength: &[]int64{maxLength}[0],
		Pattern:   policy.GetStore(ctxt).Get().Password.Pattern,
	}
}

// returns roles
func ConsequenceSessionIdle(ctxt context.Context) []string {
	roleString := policy.GetStore(ctxt).Get().Session.IdleConsequence
	if "" != roleString {
		return strings.Split(roleString, ",")
	}
	return nil
}
