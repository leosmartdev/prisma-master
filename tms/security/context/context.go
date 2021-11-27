// Package context provides extra function to reach values from a context.
package context

import (
	"golang.org/x/net/context"
)

type key int

const RequestIdKey key = 0
const SessionIdKey key = 1
const SessionStoreIdKey key = 2
const SessionKey key = 3
const AuditStoreIdKey key = 4
const PolicyStoreIdKey key = 5

func SessionIdFromContext(context context.Context) string {
	if nil == context.Value(SessionIdKey) {
		return ""
	}
	return context.Value(SessionIdKey).(string)
}

func RequestIdFromContext(context context.Context) string {
	if nil == context.Value(RequestIdKey) {
		return ""
	}
	return context.Value(RequestIdKey).(string)
}

func SessionStoreIdFromContext(context context.Context) string {
	if nil == context.Value(SessionStoreIdKey) {
		return ""
	}
	return context.Value(SessionStoreIdKey).(string)
}

func AuditStoreIdFromContext(context context.Context) string {
	if nil == context.Value(AuditStoreIdKey) {
		return ""
	}
	return context.Value(AuditStoreIdKey).(string)
}

func PolicyStoreIdFromContext(context context.Context) string {
	if nil == context.Value(PolicyStoreIdKey) {
		return ""
	}
	return context.Value(PolicyStoreIdKey).(string)
}

type SecureObject interface {
	GetActions() []string
}

type SecureObjectRegistry interface {
	Register(applicationId string, secureObject SecureObject)
}