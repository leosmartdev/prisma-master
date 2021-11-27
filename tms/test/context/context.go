// Package context provides a mock for context.
package context

import (
	"context"
	"time"

	securityContext "prisma/tms/security/context"

	"github.com/globalsign/mgo"
)

type TestingContext interface {
	context.Context
}

type testingContext struct {
	context context.Context
}

func Test() TestingContext {
	tContext := testingContext{
		context: context.Background(),
	}
	nContext := context.WithValue(tContext, securityContext.SessionStoreIdKey, "mock")
	nContext = context.WithValue(nContext, securityContext.AuditStoreIdKey, "mock")
	nContext = context.WithValue(nContext, securityContext.PolicyStoreIdKey, "mock")
	dialInfo, _ := mgo.ParseURL("mongodb://:27017")
	nContext = context.WithValue(nContext, "mongodb", dialInfo)
	return nContext
}

func (tContext testingContext) Deadline() (deadline time.Time, ok bool) {
	return tContext.context.Deadline()
}

func (tContext testingContext) Done() <-chan struct{} {
	return tContext.context.Done()
}

func (tContext testingContext) Err() error {
	return tContext.context.Err()
}

func (tContext testingContext) Value(key interface{}) interface{} {
	return tContext.context.Value(key)
}

func (tContext testingContext) String() string {
	return "Testing Context"
}

func Background() context.Context {
	return context.Background()
}
