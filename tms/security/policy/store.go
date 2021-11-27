// Package policy provides a structure to manage security policy configuration.
package policy

import (
	"context"
	"errors"
	"fmt"
	"sync"

	securityContext "prisma/tms/security/context"
	"prisma/tms/security/message"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const MongoDbX509Mechanism = "MONGODB-X509"

type Store interface {
	Get() *Policy
	Set(*Policy) error
}

func GetStore(context context.Context) Store {
	store := Store(nil)
	switch securityContext.PolicyStoreIdFromContext(context) {
	case "mock":
		store = mockStoreInstance()
	default:
		store = mongoStoreInstance(context)
	}
	return store
}

var (
	onceMongo sync.Once
	onceMock  sync.Once
	instance  Store
)

const (
	DATABASE_NAME       = "aaa"
	DATABASE_COLLECTION = "config"
	POLICY_KEY          = "prisma.tms.security.policy"
)

type mongoStore struct {
	ctx context.Context
}

type ConfigItem struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"` // Type's bson.M
}

func defaulPolicy() *Policy {
	return &Policy{
		Description: "Montes morbi quam ut lobortis rutrum curae porttitor nascetur mus cubilia vestibulum in semper nulla id sed ornare adipiscing eu nec etiam facilisis auctor nullam.",
		Session: &SessionPolicy{
			Single:          "false",
			DurationIdle:    "20m",
			IdleConsequence: message.RoleId_Administrator.String() + "," + message.RoleId_UserManager.String(),
			DurationMaximum: "16h",
			DurationRenewal: "30m",
		},
		Password: &PasswordPolicy{
			LengthMinimum:                        "2",
			LengthMaximum:                        "128",
			Pattern:                              "[a-zA-Z0-9_@.]",
			DurationMaximum:                      "2161h",
			DurationMaximumConsequence:           "RestrictedUser",
			AuthenticateInitialConsequence:       "",
			ReuseMaximum:                         "3",
			AuthenticateFailedCountMaximum:       "5",
			AuthenticateFailedMaximumConsequence: "LOCK",
			ProhibitUserId:                       true,
		},
		User: &UserPolicy{
			InactiveDurationConsequenceLock:       "2161h",
			InactiveDurationConsequenceDeactivate: "4320h",
		},
	}
}

func getSession(ctx context.Context) (*mgo.Session, error) {
	//dialInfoValue := ctx.Value("mongodb")
	dialInfoValue := ctx.Value("mongodb")
	dialInfo, ok := dialInfoValue.(*mgo.DialInfo)
	if !ok {
		return nil, errors.New("Cast to *mgo.DialInfo failed. ctx: " + fmt.Sprint(ctx))
	}
	sess, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		return nil, err
	}
	if dialInfo.Mechanism == MongoDbX509Mechanism {
		credValue := ctx.Value("mongodb-cred")
		cred, ok := credValue.(*mgo.Credential)
		if !ok {
			return nil, errors.New("Cast to *mgo.Credential failed. ctx" + fmt.Sprint(ctx))
		}
		if sess.Login(cred) != nil {
			return nil, err
		}
	}
	return sess, err

}

func (m mongoStore) Get() *Policy {
	s, err := getSession(m.ctx)
	if err != nil {
		return defaulPolicy()
	}
	defer s.Close()
	ci := new(ConfigItem)
	c := s.DB(DATABASE_NAME).C(DATABASE_COLLECTION)
	err = c.Find(bson.M{"key": POLICY_KEY}).One(ci)
	if err != nil {
		return defaulPolicy()
	}

	p := new(Policy)
	bb, err := bson.Marshal(ci.Value)
	if err != nil {
		return defaulPolicy()
	}
	err = bson.Unmarshal(bb, p)
	if err != nil {
		return defaulPolicy()
	}
	return p
}

func (m mongoStore) Set(p *Policy) error {
	s, err := getSession(m.ctx)
	if err != nil {
		return err
	}
	defer s.Close()

	c := s.DB(DATABASE_NAME).C(DATABASE_COLLECTION)
	ci := &ConfigItem{
		Key:   POLICY_KEY,
		Value: *p,
	}
	_, err = c.Upsert(bson.M{"key": POLICY_KEY}, ci)
	return err
}

func mongoStoreInstance(ctxt context.Context) Store {
	onceMongo.Do(func() {
		instance = &mongoStore{
			ctx: ctxt,
		}
	})
	return instance
}

// Mock

func mockStoreInstance() Store {
	onceMock.Do(func() {
		instance = &mockStore{}
		instance.Set(&Policy{
			Session: &SessionPolicy{
				Single:          "false",
				DurationIdle:    "20m",
				IdleConsequence: "StandardUser",
				DurationMaximum: "16h",
				DurationRenewal: "30m",
			},
			Password: &PasswordPolicy{
				LengthMinimum:                        "2",
				LengthMaximum:                        "128",
				Pattern:                              "[a-zA-Z0-9_@.]",
				DurationMaximum:                      "2161h",
				DurationMaximumConsequence:           "RestrictedUser",
				AuthenticateInitialConsequence:       "",
				ReuseMaximum:                         "3",
				AuthenticateFailedCountMaximum:       "5",
				AuthenticateFailedMaximumConsequence: "LOCK",
				ProhibitUserId:                       true,
			},
			User: &UserPolicy{
				InactiveDurationConsequenceLock:       "2161h",
				InactiveDurationConsequenceDeactivate: "4320h",
			},
		})
	})
	return instance
}

type mockStore struct {
	policy *Policy
}

func (store *mockStore) Get() *Policy {
	return store.policy
}

func (store *mockStore) Set(policy *Policy) error {
	store.policy = policy
	return nil
}
