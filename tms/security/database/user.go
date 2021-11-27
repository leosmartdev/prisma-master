// Package database provides functions to manage auth database.
package database

import (
	"errors"
	"fmt"
	"strings"

	//	tsiDb "prisma/tms/db"
	"prisma/tms/rest"
	"prisma/tms/security/message"
	"prisma/tms/security/session"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"golang.org/x/net/context"
)

const DATABASE_NAME = "aaa"

type UserId string

type User struct {
	Id           bson.ObjectId               `bson:"_id,omitempty" json:"-"`
	UserId       UserId                      `bson:"userId" json:"userId"`
	PasswordHash string                      `bson:"passwordHash" json:"-"`
	Salt         string                      `bson:"salt" json:"-"`
	State        message.User_State          `bson:"state" json:"state"`
	Roles        []string                    `bson:"roles" json:"roles"`
	Profile      *message.UserProfile        `bson:"profile" json:"profile"`
	PasswordLog  []*message.PasswordLogEntry `bson:"passwordLog" json:"-"`
	Attempts     int                         `bson:"attempts" json:"-"`
}

var (
	ErrorDuplicate = errors.New("duplicate")
	// schema and index
	//	UserTableInfo = &tsiDb.TableInfo{
	//		Name:           "users",
	//		Inst:           User{},
	//		NoCappedMirror: true,
	//		Indexes: []tsiDb.Index{
	//			{
	//				Name: "userIdUnique",
	//				Fields: []tsiDb.Field{
	//					{"userId"},
	//				},
	//				Unique: true,
	//				Sparse: true,
	//			},
	//		},
	//	}
)

func Add(ctxt context.Context, user *User) (interface{}, error) {
	session, err := getSession(ctxt)
	if err != nil {
		return nil, err
	}
	defer session.Close()
	userCollection := session.DB(DATABASE_NAME).C("users")
	// guaranteed unique UserId via index userIdUnique
	err = userCollection.Insert(user)
	if err != nil {
		if mgo.IsDup(err) {
			err = ErrorDuplicate
		}
		return nil, err
	}
	return user.UserId, nil
}

func Update(ctxt context.Context, user *User) (*User, error) {
	session, err := getSession(ctxt)
	if err != nil {
		return nil, err
	}
	defer session.Close()
	userCollection := session.DB(DATABASE_NAME).C("users")
	return user, userCollection.Update(bson.M{"_id": user.Id}, user)
}

func FindOneByUserId(ctxt context.Context, userId UserId) (*User, error) {
	if userId == "" {
		return nil, errors.New("invalid input")
	}
	session, err := getSession(ctxt)
	if err != nil {
		return nil, err
	}
	defer session.Close()
	userCollection := session.DB("aaa").C("users")
	var newUser User
	err = userCollection.Find(bson.M{"userId": userId}).One(&newUser)
	if &newUser == nil {
		return nil, errors.New("not found")
	}
	if newUser.UserId == "" {
		return nil, errors.New("not found")
	}
	return &newUser, err
}

func FindAllNotDisabled(ctxt context.Context) ([]User, error) {
	session, err := getSession(ctxt)
	if err != nil {
		return []User{}, err
	}
	defer session.Close()
	userCollection := session.DB(DATABASE_NAME).C("users")
	var users []User
	query := bson.M{}
	query["state"] = bson.M{"$lt": 1000}
	return users, userCollection.Find(query).All(&users)
}

func FindByMapByPagination(ctxt context.Context, searchMap map[string]string, pagination *rest.PaginationQuery) ([]User, error) {
	query := createQueryFromMap(searchMap)
	if _, ok := query["state"]; !ok {
		query["state"] = bson.M{"$lt": 1000} // user was not deleted
	}
	pipe := []bson.M{
		bson.M{"$match": query},
		bson.M{"$sort": bson.M{pagination.Sort: 1}},
		bson.M{"$limit": pagination.Limit},
	}
	if pagination.AfterId != "" {
		query["userId"] = bson.M{"$gt": pagination.AfterId}
	}
	if pagination.BeforeId != "" {
		query["userId"] = bson.M{"$lt": pagination.BeforeId}
		pipe = []bson.M{
			bson.M{"$match": query},
			bson.M{"$sort": bson.M{pagination.Sort: -1}},
			bson.M{"$limit": pagination.Limit},
			bson.M{"$sort": bson.M{pagination.Sort: 1}},
		}
	}
	users := make([]User, 0)
	session, err := getSession(ctxt)
	if err != nil {
		return users, err
	}
	defer session.Close()
	collection := session.DB(DATABASE_NAME).C("users")
	return users, collection.Pipe(pipe).All(&users)
}

func FindByMap(ctxt context.Context, searchMap map[string]string) ([]User, error) {
	session, err := getSession(ctxt)
	if err != nil {
		return []User{}, err
	}
	defer session.Close()
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
	query["state"] = bson.M{"$lt": 1000}
	userCollection := session.DB(DATABASE_NAME).C("users")
	users := make([]User, 0)
	return users, userCollection.Find(query).All(&users)
}

func createQueryFromMap(searchMap map[string]string) bson.M {
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
	return query
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
