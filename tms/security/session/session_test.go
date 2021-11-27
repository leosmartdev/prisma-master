package session

import (
	"prisma/tms/security/policy"
	"prisma/tms/test/context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStore(t *testing.T) {
	store := GetStore(context.Test())
	assert.NotNil(t, store, "nil store")
	assert.IsType(t, mockStore{}, store, "wrong store")
}

func TestGetStoreDefault(t *testing.T) {
	store := GetStore(context.Background())
	assert.NotNil(t, store, "nil store")
	assert.IsType(t, &mongoStore{}, store, "wrong store")
}

func TestMockStore_Create(t *testing.T) {
	store := GetStore(context.Test())
	session, err := store.Create("testOwner", []string{"testRole"})
	assert.NoError(t, err, "create err")
	assert.NotNil(t, session, "nil session")
	assert.NotEmpty(t, session.Id(), "empty SessionId")
	assert.Equal(t, "testOwner", session.GetOwner())
	assert.Len(t, session.Id(), 36, "invalid SessionId")
}

func TestMockStore_Get(t *testing.T) {
	store := GetStore(context.Test())
	createdSession, _ := store.Create("testOwner", []string{"testRole"})
	session, err := store.Get(createdSession.Id())
	assert.NoError(t, err, "get err")
	assert.NotNil(t, session, "nil session")
}

func TestMockStore_GetRoles(t *testing.T) {
	store := GetStore(context.Test())
	createdSession, _ := store.Create("testOwner", []string{"testRole"})
	session, err := store.Get(createdSession.Id())
	assert.NoError(t, err, "get err")
	assert.NotNil(t, session, "nil session")
	assert.Contains(t, session.GetRoles(), "testRole", "missing role")
}

func TestMockStore_GetExpired(t *testing.T) {
	store := GetStore(context.Test())
	if mockStore, ok := store.(mockStore); ok {
		mockStore.timeout = 0
		createdSession, _ := mockStore.Create("testOwner", []string{"testRole"})
		session, err := mockStore.Get(createdSession.Id())
		assert.EqualError(t, err, "expired", "no err")
		assert.Nil(t, session, "nil session")
	} else {
		assert.FailNow(t, "wrong store")
	}
}

func TestMockStore_GetInvalid(t *testing.T) {
	store := GetStore(context.Test())
	session, err := store.Get("qwe-invalid-id")
	assert.EqualError(t, err, "invalidId", "no err")
	assert.Nil(t, session, "nil session")
}

func TestMockStore_GetNotFound(t *testing.T) {
	store := GetStore(context.Test())
	session, err := store.Get("b12f15a2-d210-46c3-b834-c7140322f6a5")
	assert.EqualError(t, err, "notFound", "no err")
	assert.Nil(t, session, "nil session")
}

func TestMockStore_Delete(t *testing.T) {
	store := GetStore(context.Test())
	session, _ := store.Create("testOwner", []string{"testRole"})
	err := store.Delete(session.Id())
	assert.NoError(t, err, "delete err")
}

func TestMockStore_DeleteInvalid(t *testing.T) {
	store := GetStore(context.Test())
	err := store.Delete("qwe-invalid-id")
	assert.EqualError(t, err, "invalidId", "no err")
}

func TestMockStore_DeleteNotFound(t *testing.T) {
	store := GetStore(context.Test())
	err := store.Delete("b12f15a2-d210-46c3-b834-c7140322f6a5")
	assert.EqualError(t, err, "notFound", "no err")
}

func TestMongoStore_Create(t *testing.T) {
	store := GetStore(context.Test())
	policyStore := policy.GetStore(context.Test())
	policyStore.Set(&policy.Policy{
		Session: &policy.SessionPolicy{
			Single: "false",
		},
	})
	session, err := store.Create("testOwner", []string{"testRole"})
	if err == nil || "no reachable servers" != err.Error() {
		assert.NoError(t, err, "create err")
		assert.NotNil(t, session, "nil session")
		assert.NotEmpty(t, session.Id(), "empty SessionId")
		assert.Equal(t, "testOwner", session.GetOwner())
		assert.Len(t, session.Id(), 36, "invalid SessionId")
	}
}

func TestMongoStore_Delete(t *testing.T) {
	store := GetStore(context.Test())
	session, err := store.Create("testOwner", []string{"testRole"})
	if err == nil || "no reachable servers" != err.Error() {
		err = store.Delete(session.Id())
		assert.NoError(t, err, "delete err", session.Id())
	}
}

func TestMongoStore_DeleteInvalid(t *testing.T) {
	store := GetStore(context.Test())
	err := store.Delete("qwe-invalid-id")
	if err == nil || "no reachable servers" != err.Error() {
		assert.EqualError(t, err, "invalidId", "no err")
	}
}

func TestMongoStore_DeleteNotFound(t *testing.T) {
	store := GetStore(context.Test())
	err := store.Delete("b12f15a2-d210-46c3-b834-c7140322f6a5")
	if err == nil || "no reachable servers" != err.Error() {
		assert.EqualError(t, err, "notFound", "no err")
	}
}

func TestMongoStore_Get(t *testing.T) {
	store := GetStore(context.Test())
	createdSession, err := store.Create("testOwner", []string{"testRole"})
	if err == nil || "no reachable servers" != err.Error() {
		session, err := store.Get(createdSession.Id())
		assert.NoError(t, err, "get err")
		assert.NotNil(t, session, "nil session")
	}
}

func TestMongoStore_GetInvalid(t *testing.T) {
	store := GetStore(context.Test())
	session, err := store.Get("qwe-invalid-id")
	if err == nil || "no reachable servers" != err.Error() {
		assert.EqualError(t, err, "invalidId", "no err")
		assert.Nil(t, session, "nil session")
	}
}

func TestMongoStore_GetNoCache(t *testing.T) {
	store := GetStore(context.Test())
	createdSession, err := store.Create("testOwner", []string{"testRole"})
	if err == nil || "no reachable servers" != err.Error() {
		// remove from LRU cache
		for i := 0; i < 128; i++ {
			tmpSession, _ := store.Create("testOwner", []string{"testRole"})
			store.Delete(tmpSession.Id())
		}
		session, err := store.Get(createdSession.Id())
		assert.NoError(t, err, "get err")
		assert.NotNil(t, session, "nil session")
	}
}

func TestMongoStore_GetRoles(t *testing.T) {
	store := GetStore(context.Test())
	createdSession, err := store.Create("testOwner", []string{"testRole"})
	if err == nil || "no reachable servers" != err.Error() {
		session, err := store.Get(createdSession.Id())
		assert.NoError(t, err, "get err")
		assert.NotNil(t, session, "nil session")
		assert.Contains(t, session.GetRoles(), "testRole", "missing role")
	}
}

//func TestMongoStore_GetReduced(t *testing.T) {
//	policy.GetStore(context.Background()).Set(&policy.Policy{
//		Session: &policy.SessionPolicy{
//			DurationIdle:    "1s",
//			IdleConsequence: "StandardUser",
//		},
//	})
//	store := GetStore(context.Background())
//	if mStore, ok := store.(*mongoStore); ok {
//		createdSession, err := mStore.Create("testOwner", []string{"testRole"})
//		if err == nil || "no reachable servers" != err.Error() {
//			// trigger
//			storeInstance.tickerCleanup = 1 * time.Second
//			ticker := startTicker(storeInstance)
//			time.Sleep(2 * time.Second)
//			ticker.Stop()
//			session, err := mStore.Get(createdSession.Id())
//			assert.NoError(t, err, "get err")
//			assert.NotNil(t, session, "nil session")
//			assert.Contains(t, session.GetRoles(), "testRole", "missing role")
//			assert.Equal(t, message.Session_idled, session.GetState(), "wrong state")
//		}
//	} else {
//		assert.FailNow(t, "wrong store")
//	}
//}

//func TestMongoStore_CleanupExpired(t *testing.T) {
//	policy.GetStore(context.Test()).Set(&policy.Policy{
//		Session: &policy.SessionPolicy{
//			DurationMaximum:    "1s",
//		},
//	})
//	store := GetStore(context.Test())
//	createdSession, err := store.Create("testOwner", []string{"testRole"})
//	if err == nil || "no reachable servers" != err.Error() {
//		// trigger cleanup
//		storeInstance.tickerCleanup = 1 * time.Second
//		ticker := startTicker(storeInstance)
//		time.Sleep(2 * time.Second)
//		ticker.Stop()
//		_, err = store.Get(createdSession.Id())
//		assert.EqualError(t, err, "notFound", "no err")
//	}
//}

//func TestMongoStore_CleanupAbsolute(t *testing.T) {
//	policy.GetStore(context.Test()).Set(&policy.Policy{
//		Session: &policy.SessionPolicy{
//			DurationMaximum:    "1s",
//		},
//	})
//	store := GetStore(context.Test())
//	createdSession, err := store.Create("testOwner", []string{"testRole"})
//	if err == nil || "no reachable servers" != err.Error() {
//		// trigger cleanup
//		storeInstance.tickerCleanup = 1 * time.Second
//		ticker := startTicker(storeInstance)
//		time.Sleep(2 * time.Second)
//		ticker.Stop()
//		_, err := store.Get(createdSession.Id())
//		assert.EqualError(t, err, "notFound", "no err")
//	}
//}

// intermediate failures
//func TestMongoStore_CleanupRenewal(t *testing.T) {
//	store := GetStore(context.Background())
//	createdSession, err := store.Create("testOwnerRenewal", []string{"testRole"})
//	if err == nil || "no reachable servers" != err.Error() {
//		session, err := store.Get(createdSession.Id())
//		// trigger cleanup
//		//timeoutRenewal = 0
//		storeInstance.tickerCleanup = 1 * time.Second
//		ticker := startTicker(storeInstance)
//		time.Sleep(2 * time.Second)
//		ticker.Stop()
//		session, err = store.Get(createdSession.Id())
//		assert.NoError(t, err, "get err")
//		assert.NotNil(t, session, "nil session")
//	}
//}
