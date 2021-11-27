package session

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	securityContext "prisma/tms/security/context"
	"prisma/tms/security/message"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/pborman/uuid"
)

const (
	DATABASE_NAME        = "aaa"
	COLLECTION           = "sessions"
	TICKER_CLEANUP       = 1 * time.Minute
	MongoDbX509Mechanism = "MONGODB-X509"
)

type mongoSession struct {
	InternalSession `bson:"-"`
	MongoId         bson.ObjectId         `bson:"_id,omitempty" json:"-"`
	SessionId       string                `bson:"sessionId" json:"-"`
	OwnerId         string                `bson:"ownerId" json:"-"`
	LastAccess      time.Time             `bson:"lastAccess" json:"-"`
	Created         time.Time             `bson:"created" json:"-"`
	Roles           []string              `bson:"roles" json:"-"`
	State           message.Session_State `bson:"state" json:"-"`
}

func (session *mongoSession) Id() string {
	return session.SessionId
}

func (session *mongoSession) GetRoles() []string {
	return session.Roles
}

func (session *mongoSession) GetState() message.Session_State {
	return session.State
}

func (session *mongoSession) GetOwner() string {
	return session.OwnerId
}

type mongoStore struct {
	Store
	tickerCleanup time.Duration
	context       context.Context
	publisher     Publisher
	ticker        *time.Ticker
}

var (
	mongoStoreOnce  sync.Once
	tickerStartOnce sync.Once
	storeInstance   *mongoStore
)

func mongoStoreInstance(ctxt context.Context) Store {
	mongoStoreOnce.Do(func() {
		storeInstance = &mongoStore{
			tickerCleanup: TICKER_CLEANUP,
			context:       ctxt,
		}
	})
	storeInstance.context = ctxt
	return storeInstance
}

func mongoSetPublisher(_ context.Context, publisher Publisher) error {
	tickerStartOnce.Do(func() {
		storeInstance.ticker = startTicker(storeInstance)
	})
	storeInstance.publisher = publisher
	return nil
}

func startTicker(store *mongoStore) *time.Ticker {
	ticker := time.NewTicker(store.tickerCleanup)
	go func() {
		for range ticker.C {
			mSession, err := getSession(store.context)
			if err != nil {
				panic(err)
			}
			collection := mSession.DB(DATABASE_NAME).C(COLLECTION)
			allSessions := make([]*mongoSession, 0)
			err = collection.Find(nil).All(&allSessions)
			if nil == err {
				for _, checkSession := range allSessions {
					// check idle
					enforced, _ := EnforceSessionIdle(store.context, checkSession.LastAccess)
					if enforced && (checkSession.State != message.Session_idled) {
						checkSession.State = message.Session_idled
						collection.UpdateId(checkSession.MongoId, checkSession)
						store.publish(message.Session_IDLE, checkSession)
						// TODO cycle import fix: security.AuditUserObject(store.context, "Session", string(checkSession.OwnerId), "SYSTEM", message.Session_IDLE.String(), "SUCCESS", checkSession.SessionId)
					}
					// check renewal, policy check
					enforced = EnforceSessionRenewal(store.context, checkSession.Created)
					if enforced && (checkSession.State != message.Session_renewing) {
						checkSession.State = message.Session_renewing
						collection.UpdateId(checkSession.MongoId, checkSession)
					}
					// check absolute, policy check
					enforced = EnforceSessionDuration(store.context, checkSession.Created)
					if enforced && (checkSession.State != message.Session_expired) {
						checkSession.State = message.Session_expired
						collection.UpdateId(checkSession.MongoId, checkSession)
					}
				}
				//// delete expired
				collection.Remove(bson.M{"state": message.Session_expired})
			}
			mSession.Close()
		}
	}()
	return ticker
}

func (store mongoStore) Create(owner string, roles []string) (InternalSession, error) {
	mSession, err := getSession(store.context)
	if err != nil {
		return nil, err
	}
	collection := mSession.DB(DATABASE_NAME).C(COLLECTION)
	defer mSession.Close()
	sessionId := uuid.New()
	// check policy
	if EnforceSessionSingle(store.context) {
		query := bson.M{"ownerId": owner}
		result := collection.Find(query)
		targetSession := new(mongoSession)
		iterator := result.Iter()
		for iterator.Next(targetSession) {
			targetSession.State = message.Session_terminated
			store.publish(message.Session_TERMINATE, targetSession)
			// TODO add audit log entry
			targetSession = new(mongoSession)
		}
		collection.Remove(query)
	}
	// create new session
	newSession := new(mongoSession)
	newSession.SessionId = sessionId
	newSession.OwnerId = owner
	newSession.Roles = roles
	newSession.State = message.Session_activated
	newSession.Created = time.Now()
	newSession.LastAccess = newSession.Created
	return newSession, collection.Insert(newSession)
}

func (store mongoStore) Get(id string) (InternalSession, error) {
	// check invalid id
	uid := uuid.Parse(id)
	if nil == uid {
		return nil, ErrorInvalidId
	}
	session := new(mongoSession)
	// if in context then return
	if store.context.Value(securityContext.SessionKey) != nil {
		session = store.context.Value(securityContext.SessionKey).(*mongoSession)
		if id == session.Id() {
			return session, nil
		}
	}
	mSession, err := getSession(store.context)
	if err != nil {
		return nil, err
	}
	collection := mSession.DB(DATABASE_NAME).C(COLLECTION)
	defer mSession.Close()
	// check mapping for mongo _id
	cond := bson.M{"sessionId": id}
	result := collection.Find(cond)
	if count, _ := result.Count(); count == 0 {
		return nil, ErrorNotFound
	}
	err = result.One(session)
	if err != nil {
		return nil, err
	}
	// check expired
	if session.State == message.Session_expired {
		return nil, ErrorExpired
	}
	// check idle, check policy
	enforced, _ := EnforceSessionIdle(store.context, session.LastAccess)
	if enforced {
		session.State = message.Session_idled
	}
	// last access update
	session.LastAccess = time.Now()
	collection.Update(cond, session)
	return session, err
}

func (store mongoStore) Delete(id string) error {
	// check invalid id
	uid := uuid.Parse(id)
	if nil == uid {
		return ErrorInvalidId
	}
	mSession, err := getSession(store.context)
	if err != nil {
		return err
	}
	collection := mSession.DB(DATABASE_NAME).C(COLLECTION)
	defer mSession.Close()
	// check mapping for mongo _id
	cond := bson.M{"sessionId": id}
	result := collection.Find(cond)
	if count, _ := result.Count(); count == 0 {
		return ErrorNotFound
	}
	return collection.Remove(cond)
}

func (store mongoStore) publish(action message.Session_Action, session *mongoSession) {
	if nil != store.publisher {
		store.publisher.Publish(action, session)
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
