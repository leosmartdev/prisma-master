package mongo

import (
	"context"
	"errors"

	"prisma/tms/db"
	"prisma/tms/security/database"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const (
	collectionUsers = "users"
	dbName          = "aaa"
)

type UserDb struct{}

func NewMongoUserDb() db.UserDB {
	return &UserDb{}
}

func (d *UserDb) FindOne(ctx context.Context, userID string) (*database.User, error) {
	if userID == "" {
		return nil, errors.New("expected not empty user id, got empty")
	}

	session, err := getSession(ctx)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	newUser := new(database.User)
	err = session.DB(dbName).C(collectionUsers).
		Find(bson.M{"userId": userID}).
		One(newUser)

	if mgo.ErrNotFound == err {
		err = db.ErrorNotFound
	}
	return newUser, err
}
