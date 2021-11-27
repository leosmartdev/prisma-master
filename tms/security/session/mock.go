package session

import (
	"github.com/pborman/uuid"
	"math"
	"sync"
	"time"
)

type mockSession struct {
	InternalSession
	id string
	ownerId string
	roles []string
	lastAccess time.Time
}

func (s *mockSession) Id() string {
	return s.id
}

func (s *mockSession) GetRoles() []string {
	return s.roles
}

func (s *mockSession) GetOwner() string {
	return s.ownerId
}

type mockStore struct {
	Store
	sessionMap map[string]*mockSession
	timeout float64
}

var (
	once     sync.Once
	instance Store
)

func mockStoreInstance() Store {
	once.Do(func() {
		instance = mockStore{
			sessionMap: make(map[string]*mockSession),
			timeout: math.MaxFloat64,
		}
	})
	return instance
}

func (store mockStore) Create(owner string, roles []string) (InternalSession, error) {
	session := new(mockSession)
	session.id = uuid.New()
	session.roles = roles
	session.lastAccess = time.Now()
	session.ownerId = owner
	store.sessionMap[session.id] = session
	return session, nil
}

func (store mockStore) Get(id string) (InternalSession, error) {
	// check invalid id
	uid := uuid.Parse(id)
	if nil == uid {
		return nil, ErrorInvalidId
	}
	session := store.sessionMap[id]
	if nil == session {
		return nil, ErrorNotFound
	}
	// check expired
	elapsed := time.Since(session.lastAccess)
	if elapsed.Seconds() > store.timeout {
		return nil, ErrorExpired
	}
	// update last access
	session.lastAccess = time.Now()
	return session, nil
}

func (store mockStore) Delete(id string) error {
	// check invalid id
	uid := uuid.Parse(id)
	if nil == uid {
		return ErrorInvalidId
	}
	session := store.sessionMap[id]
	if nil == session {
		return ErrorNotFound
	}
	delete(store.sessionMap, id)
	return nil
}