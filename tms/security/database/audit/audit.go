// Package audit provides functions to manage audit log(Actions of users).
package audit

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"prisma/tms/log"
	securityContext "prisma/tms/security/context"
	"prisma/tms/security/session"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type Auditor interface {
	Record(context context.Context, record Record) error
	GetRecords(context context.Context, searchQuery string) ([]Record, error)
	GetRecordsBySessionId(context context.Context, sessionId string) ([]Record, error)
	GetRecordsByUserId(context context.Context, userId string) ([]Record, error)
	GetRecordsByMap(context context.Context, searchMap map[string]string, searchQuery string) ([]Record, error)
	GetRecordsByMapByTimeQuery(context context.Context, searchMap map[string]string, timeQuery TimeQuery, searchQuery string) ([]Record, error)
}

func NewAuditor(context context.Context) Auditor {
	auditor := Auditor(nil)
	switch securityContext.AuditStoreIdFromContext(context) {
	case "mock":
		auditor = mockAuditorInstance()
	default:
		auditor = new(mongoAuditor)
	}
	return auditor
}

const auditCollection = "records"

var (
	ErrorNotFound = errors.New("notFound")
)

type Record struct {
	MongoId     bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Created     time.Time     `bson:"created" json:"created"`
	ClassId     string        `bson:"classId" json:"classId"`
	ObjectId    string        `bson:"objectId,omitempty" json:"objectId"`
	UserId      string        `bson:"userId,omitempty" json:"userId"`
	Action      string        `bson:"action" json:"action"`
	Outcome     string        `bson:"outcome" json:"outcome"`
	Payload     string        `bson:"payload,omitempty" json:"payload,omitempty"`
	SessionId   string        `bson:"sessionId" json:"sessionId"`
	RequestId   string        `bson:"requestId" json:"requestId"`
	Message     string        `bson:"message,omitempty" json:"message,omitempty"`
	Hash        []byte        `bson:"hash,omitempty" json:"hash,omitempty"`
	Signature   []byte        `bson:"signature,omitempty" json:"signature,omitempty"`
	Description string        `bson:"-" json:"description"`
}

// Records matching RecordId are excluded from results
type TimeQuery struct {
	Limit          int
	AfterRecordId  string
	BeforeRecordId string
}

// Use RecordIds to deliminate TimeFrame
type TimeCursor struct {
	FirstRecordId string
	LastRecordId  string
}

type mongoAuditor struct {
	Auditor
}

const (
	DATABASE_NAME = "aaa"
)

func (auditor *mongoAuditor) Record(ctxt context.Context, record Record) error {
	session, err := getSession(ctxt)
	if err != nil {
		return err
	}
	defer session.Close()
	collection := session.DB(DATABASE_NAME).C(auditCollection)
	return collection.Insert(record)
}

func (auditor *mongoAuditor) GetRecords(ctxt context.Context, searchMap string) ([]Record, error) {
	var query bson.M
	if searchMap != "" {
		query = createSearchQuery(searchMap)
	} else {
		query = bson.M(nil)
	}

	records, err := getRecordsFromQuery(ctxt, query)
	return records, err
}

func (auditor *mongoAuditor) GetRecordsBySessionId(ctxt context.Context, sessionId string) ([]Record, error) {
	query := bson.M{"sessionId": sessionId}
	records, err := getRecordsFromQuery(ctxt, query)
	return records, err
}

func (auditor *mongoAuditor) GetRecordsByUserId(ctxt context.Context, userId string) ([]Record, error) {
	query := bson.M{"userId": userId}
	records, err := getRecordsFromQuery(ctxt, query)
	return records, err
}

func (auditor *mongoAuditor) GetRecordsByMapByTimeQuery(ctxt context.Context, searchMap map[string]string, timeQuery TimeQuery, searchQuery string) ([]Record, error) {
	query := createQueryFromMap(ctxt, searchMap, searchQuery)
	// get before/after time
	if timeQuery.AfterRecordId != "" {
		// TODO cache for lookup time
		afterRecords, err := getRecordsWithLimitFromQuery(ctxt, bson.M{"_id": bson.ObjectIdHex(timeQuery.AfterRecordId)}, 1)
		if err != nil {
			return nil, err
		}
		if len(afterRecords) == 0 {
			return nil, ErrorNotFound
		}
		after := afterRecords[0].Created
		query["created"] = bson.M{"$gte": after}
		query["_id"] = bson.M{"$ne": bson.ObjectIdHex(timeQuery.AfterRecordId)}
	}
	if timeQuery.BeforeRecordId != "" {
		// TODO cache for lookup time
		beforeRecords, err := getRecordsWithLimitFromQuery(ctxt, bson.M{"_id": bson.ObjectIdHex(timeQuery.BeforeRecordId)}, 1)
		if err != nil {
			return nil, err
		}
		if len(beforeRecords) == 0 {
			return nil, ErrorNotFound
		}
		before := beforeRecords[0].Created
		query["created"] = bson.M{"$lte": before}
		query["_id"] = bson.M{"$ne": bson.ObjectIdHex(timeQuery.BeforeRecordId)}
	}
	return getRecordsWithLimitFromQuery(ctxt, query, timeQuery.Limit)
}

func (auditor *mongoAuditor) GetRecordsByMap(ctxt context.Context, searchMap map[string]string, searchQuery string) ([]Record, error) {
	query := createQueryFromMap(ctxt, searchMap, searchQuery)
	records, err := getRecordsFromQuery(ctxt, query)
	return records, err
}

func getRecordsFromQuery(ctxt context.Context, query bson.M) ([]Record, error) {
	records := make([]Record, 0)
	session, err := getSession(ctxt)
	if err != nil {
		return records, err
	}
	defer session.Close()
	collection := session.DB(DATABASE_NAME).C(auditCollection)
	return records, collection.Find(query).Sort("-created").All(&records)
}

func getRecordsWithLimitFromQuery(ctxt context.Context, query bson.M, limit int) ([]Record, error) {
	records := make([]Record, 0)
	session, err := getSession(ctxt)
	if err != nil {
		return records, err
	}
	defer session.Close()
	collection := session.DB(DATABASE_NAME).C(auditCollection)
	return records, collection.Find(query).Sort("-created").Limit(limit).All(&records)
}

func createQueryFromMap(_ context.Context, searchMap map[string]string, searchQuery string) bson.M {
	query := bson.M{}
	for k, v := range searchMap {
		values := strings.Split(v, ",")
		if len(values) > 1 {
			orQuery := make([]bson.M, 0)
			for _, value := range values {
				orQuery = append(orQuery, bson.M{k: value})
			}
			query["$or"] = orQuery
		} else {
			query[k] = v
		}
	}

	if searchQuery != "" {
		andQuery := make([]bson.M, 0)
		andQuery = append(andQuery, query, createSearchQuery(searchQuery))
		query = bson.M{
			"$and": andQuery,
		}
	}

	return query
}

func createSearchQuery(searchQuery string) bson.M {
	query := bson.M{}

	// match action with the search query
	// i.e. Acknowledge --> ACK
	if isContained(searchQuery, "ack") && isContained(searchQuery, "all") {
		searchQuery = "ACK_ALL"
	} else if isContained(searchQuery, "ack") {
		searchQuery = "ACK"
	} else if isContained(searchQuery, "edit") {
		searchQuery = "UPDATE"
	} else if isContained(searchQuery, "unassign") {
		if isContained(searchQuery, "note") || isContained(searchQuery, "incident") {
			searchQuery = "DELETE_NOTE|DELETE_NOTE_FILE|DELETE_NOTE_ENTITY|DETACH_NOTE"
		}
		if isContained(searchQuery, "vessel") || isContained(searchQuery, "fleet") {
			searchQuery = "REMOVE_VESSEL"
		}

		if !isContained(searchQuery, "DELETE_NOTE") && searchQuery != "REMOVE_VESSEL" {
			searchQuery = "UNASSIGN|DELETE_NOTE|DELETE_NOTE_FILE|DELETE_NOTE_ENTITY|DETACH_NOTE|REMOVE_VESSEL"
		}
	} else if isContained(searchQuery, "assign") {
		if isContained(searchQuery, "note") || isContained(searchQuery, "incident") {
			searchQuery = "ADD_NOTE|ADD_NOTE_FILE|ADD_NOTE_ENTITY"
		}
		if isContained(searchQuery, "vessel") || isContained(searchQuery, "fleet") {
			searchQuery = "ADD_VESSEL"
		}

		if !isContained(searchQuery, "ADD_NOTE") && searchQuery != "ADD_VESSEL" {
			searchQuery = "ASSIGN|ADD_NOTE|ADD_NOTE_FILE|ADD_NOTE_ENTITY|ADD_VESSEL"
		}
	}

	log.Debug("search query: %s", searchQuery)

	queryRegEx := bson.RegEx{
		Pattern: searchQuery,
		Options: "i",
	}

	query["$or"] = []interface{}{
		bson.M{
			"classId": queryRegEx,
		},
		bson.M{
			"userId": queryRegEx,
		},
		bson.M{
			"action": queryRegEx,
		},
	}

	return query
}

func isContained(str string, substr string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
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
	if dialInfo.Mechanism == session.MongoDbX509Mechanism {
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
