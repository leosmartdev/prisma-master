package rule

import (
	"errors"
	"io"
	"sync"
	"time"

	"prisma/tms/db/mongo"
	"prisma/tms/log"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// Store is an interface to determine methods for storing rules
type Store interface {
	// Delete a rule from a store
	DeleteRule(id string) error
	// Insert or update a record for a rule in a store. Return an id, an error
	UpsertRule(rule Rule) error
	// Get a rule by id
	GetRule(id string) (*Rule, error)
	// Get all rules
	GetAll() []*Rule
	// Get a rule by an operand type
	GetByType(operandType OperandType) []*Rule
}

type ruleIf interface {
	GetOperandType() OperandType
	GetAll() *Rule_IfAll
	GetAny() *Rule_IfAny
}

// StorageMongoDBRule is a store for storing rules in mongodb
type StorageMongoDBRule struct {
	mu           sync.Mutex
	database     *mgo.Database
	clientDB     *mongo.MongoClient
	hashMap      map[string]Rule
	lastModified time.Time
}

// NewStorageMongoDBRule returns an instance of mongo store
func NewStorageMongoDBRule(client *mongo.MongoClient) (*StorageMongoDBRule, error) {
	st := StorageMongoDBRule{
		hashMap:  make(map[string]Rule),
		clientDB: client,
	}
	if client == nil {
		log.Warn("Using without mongodb")
		return &st, nil
	}
	st.database = st.clientDB.DB()
	return &st, nil
}

func (st *StorageMongoDBRule) checkConnection() {
	if err := st.database.Session.Ping(); err == io.EOF {
		log.Error("A connection was lost")
		log.Info("Try to reconnect")
		st.database = st.clientDB.DB()
		log.Info("connected")
	}
}

// Updating the hashmap.
// It doesn't use a mutex, cause it is a private function and should be call using mutex
func (st *StorageMongoDBRule) updatedHashMap() error {
	if st.database == nil {
		return nil
	}
	st.checkConnection()
	var ltTime struct {
		Time time.Time
	}
	if err := st.database.C("rules").FindId(0).One(&ltTime); err == nil &&
		ltTime.Time.Equal(st.lastModified) {
		return nil
	}
	query := st.database.C("rules").Find(bson.M{"_id": bson.M{"$nin": []int{0}}})
	// TODO: do smt, if it contains records a lot. What is about RAM....
	if n, err := query.Count(); err != nil || n > 99999 {
		if err != nil {
			return err
		}
		return errors.New("a lot of records")
	}
	var mongoRule Rule
	iter := query.Iter()
	for iter.Next(&mongoRule) {
		st.hashMap[mongoRule.Id] = mongoRule
	}
	st.lastModified = ltTime.Time
	return iter.Close()
}

func (st *StorageMongoDBRule) UpsertRule(rule Rule) error {
	st.mu.Lock()
	defer st.mu.Unlock()
	st.hashMap[rule.Id] = rule
	if st.database == nil {
		log.Warn("Using without mongodb")
		return nil
	}
	st.checkConnection()
	_, err := st.database.C("rules").Upsert(bson.M{"_id": rule.Id}, rule)
	if err == nil {
		_, err = st.database.C("rules").Upsert(bson.M{"_id": 0}, bson.M{"time": time.Now()})
	}
	return err
}

func (st *StorageMongoDBRule) DeleteRule(id string) error {
	st.mu.Lock()
	defer st.mu.Unlock()
	delete(st.hashMap, id)
	if st.database == nil {
		log.Warn("Using without mongodb")
		return nil
	}
	st.checkConnection()
	return st.database.C("rules").RemoveId(id)
}

func (st *StorageMongoDBRule) GetRule(id string) (*Rule, error) {
	st.mu.Lock()
	defer st.mu.Unlock()
	if retRule, ok := st.hashMap[id]; ok {
		return &retRule, nil
	}
	if st.database == nil {
		return nil, errors.New("not found")
	}
	st.checkConnection()
	var retRule Rule
	if err := st.database.C("rules").Find(bson.M{"id": id}).One(&retRule); err != nil {
		return nil, err
	}
	return &retRule, nil
}

func (st *StorageMongoDBRule) getByType(operandType OperandType, rIf ruleIf) bool {
	if rIf == nil {
		return false
	}
	if rIf.GetOperandType() == operandType {
		return true
	}
	if rIf.GetAll() != nil {
		return st.getByType(operandType, rIf.GetAll())
	} else if rIf.GetAny() != nil {
		return st.getByType(operandType, rIf.GetAny())
	}
	return false
}

func (st *StorageMongoDBRule) GetByType(operandType OperandType) []*Rule {
	st.mu.Lock()
	defer st.mu.Unlock()
	ret := make([]*Rule, 0, len(st.hashMap))
	for _, val := range st.hashMap {
		if st.getByType(operandType, val.GetAll()) {
			ret = append(ret, &val)
		}
		if st.getByType(operandType, val.GetAny()) {
			ret = append(ret, &val)
		}
	}
	return ret
}

func (st *StorageMongoDBRule) GetAll() []*Rule {
	st.mu.Lock()
	defer st.mu.Unlock()
	// The first we should be sure about our data is recent
	if err := st.updatedHashMap(); err != nil {
		log.Info("got err: %v", err)
	}
	ret := make([]*Rule, 0, len(st.hashMap))
	for _, val := range st.hashMap {
		obj := val
		ret = append(ret, &obj)
	}
	return ret
}
