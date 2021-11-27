package audit

import (
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	"prisma/tms/security/session"
	"prisma/tms/test/context"
	"testing"
	"time"
	"reflect"
)

var (
	testRecord = Record{
		Created:   time.Now(),
		ClassId:   "testClassId",
		ObjectId:  "objectId",
		Action:    "testAction",
		Outcome:   "SUCCESS",
		Payload:   "payload",
		SessionId: uuid.New(),
		RequestId: uuid.New(),
	}
)

func clone(i interface{}) interface{} {
	// Wrap argument to reflect.Value, dereference it and return back as interface{}
	return reflect.Indirect(reflect.ValueOf(i)).Interface()
}

func TestNewAuditor(t *testing.T) {
	auditor := NewAuditor(context.Test())
	assert.NotNil(t, auditor, "not nil")
}

func TestRecord(t *testing.T) {
	auditor := NewAuditor(context.Test())
	_ = auditor.Record(context.Test(), testRecord)
}

//func TestMongoAuditor_GetRecordsByMap(t *testing.T) {
//	userId := "testuser"
//	testRecord.UserId = userId
//	auditor := NewAuditor(context.Test())
//	auditor.Record(context.Test(), testRecord)
//	searchMap := make(map[string]string)
//	searchMap["classId"] = "testClassId"
//	searchMap["action"] = "testAction"
//	records, err := auditor.GetRecordsByMap(context.Test(), searchMap)
//	if err == nil || "no reachable servers" != err.Error() {
//		assert.NotNil(t, records, "nil array")
//		assert.NotEmpty(t, records, "empty array")
//		t.Log(len(records))
//		t.Log(records)
//	}
//}

//func TestMongoAuditor_GetRecordsByMapByTimeQuery(t *testing.T) {
//	logger := log.New(os.Stdout, "[mgo] ", log.LUTC|log.Lshortfile)
//	mgo.SetLogger(logger)
//	mgo.SetDebug(true)
//	userId := "testuser"
//	testRecord.UserId = userId
//	auditor := NewAuditor(context.Test())
//	auditor.Record(context.Test(), testRecord)
//	searchMap := make(map[string]string)
//	searchMap["classId"] = "testClassId"
//	searchMap["action"] = "testAction"
//	timeQuery := TimeQuery{
//		Limit: 10,
//		AfterRecordId: testRecord.MongoId.Hex(),
//		//BeforeRecordId: "59daa4ff81f7a4ce5fbea726",//"59dad67481f7a4ce5fbea739",
//	}
//	records, err := auditor.GetRecordsByMapByTimeQuery(context.Background(), searchMap, timeQuery)
//	if err == nil || "no reachable servers" != err.Error() {
//		assert.NotNil(t, records, "nil array")
//		assert.NotEmpty(t, records, "empty array")
//		t.Log(len(records))
//		t.Log(records)
//	}
//}

//func TestMongoAuditor_GetRecordsByMapMultipleValue(t *testing.T) {
//	auditor := NewAuditor(context.Test())
//	classId := "testClassId" + time.Now().String()
//	testRecord1 := clone(testRecord).(Record)
//	testRecord1.ClassId = classId
//	testRecord1.Action = "testAction1"
//	auditor.Record(context.Test(), testRecord1)
//	testRecord2 := clone(testRecord).(Record)
//	testRecord2.ClassId = classId
//	testRecord2.Action = "testAction2"
//	auditor.Record(context.Test(), testRecord2)
//	searchMap := make(map[string]string)
//	searchMap["classId"] = classId
//	searchMap["action"] = "testAction1,testAction2"
//	records, err := auditor.GetRecordsByMap(context.Test(), searchMap)
//	if err == nil || "no reachable servers" != err.Error() {
//		assert.NotNil(t, records, "nil array")
//		assert.NotEmpty(t, records, "empty array")
//		assert.Len(t, records, 2)
//		t.Log(records)
//	}
//}

//func TestMongoAuditor_GetRecordsByMapQuery(t *testing.T) {
//	//logger := log.New(os.Stdout, "[mgo] ", log.LUTC|log.Lshortfile)
//	//mgo.SetLogger(logger)
//	//mgo.SetDebug(true)
//	auditor := NewAuditor(context.Test())
//	searchMap := make(map[string]string)
//	searchMap["classId"] = "testClassId"
//	searchMap["action"] = "testAction,testAction1"
//	//searchMap["objectId"] = "59428c0563209e0ff9617bfe"
//	records, err := auditor.GetRecordsByMap(context.Test(), searchMap)
//	if err == nil || "no reachable servers" != err.Error() {
//		assert.NotNil(t, records, "nil array")
//		assert.NotEmpty(t, records, "empty array")
//		t.Log(records)
//	}
//}

func TestMongoAuditor_GetRecordsBySessionId(t *testing.T) {
	session, err := session.GetStore(context.Test()).Create("testuser", []string{})
	if err == nil || "no reachable servers" != err.Error() {
		sessionId := session.Id()
		testRecord.SessionId = sessionId
		auditor := NewAuditor(context.Test())
		auditor.Record(context.Test(), testRecord)
		records, err := auditor.GetRecordsBySessionId(context.Test(), sessionId)
		if err == nil || "no reachable servers" != err.Error() {
			assert.NotNil(t, records, "nil array")
			assert.NotEmpty(t, records, "empty array")
		}
	}
}

func TestMongoAuditor_GetRecordsByUserId(t *testing.T) {
	userId := "testuser"
	testRecord.UserId = userId
	auditor := NewAuditor(context.Test())
	auditor.Record(context.Test(), testRecord)
	records, err := auditor.GetRecordsByUserId(context.Test(), userId)
	if err == nil || "no reachable servers" != err.Error() {
		assert.NotNil(t, records, "nil array")
		assert.NotEmpty(t, records, "empty array")
	}
}
