package database

import (
	"prisma/tms/security"
	"prisma/tms/test/context"
	"testing"
)

func TestAudit(t *testing.T) {
	userId := UserId("testuser")
	security.Audit(context.Test(), "classid", "action", "outcome", userId)
}

func TestAuditNoPayload(t *testing.T) {
	security.Audit(context.Test(), "classid", "action", "outcome")
}

func TestAuditPayload(t *testing.T) {
	userId := UserId("testuser")
	security.Audit(context.Test(), "classid", "action", "outcome", userId, "more")
}

//func TestFindByMap(t *testing.T) {
//	searchMap := make(map[string]string)
//	searchMap["roles"] = "Administrator"
//	users, err := FindByMap(context.Test(), searchMap)
//	if err == nil || "no reachable servers" != err.Error() {
//		assert.NotNil(t, users, "nil array")
//		assert.NotEmpty(t, users, "empty array")
//	}
//}