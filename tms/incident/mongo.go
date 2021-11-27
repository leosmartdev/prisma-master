// Package incident provides functions to work with incident store.
package incident

import (
	"context"
	"errors"
	"fmt"

	"prisma/tms/db/mongo"
	"prisma/tms/log"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type mongoIdCreator struct {
	IdCreator
	ctxt context.Context
}

func (creator *mongoIdCreator) Next(prefixer IdPrefixer) string {
	prefix := prefixer.Prefix()
	change := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"sequence": 1}},
		ReturnNew: true,
	}
	session, err := getSession(creator.ctxt)
	defer session.Close()
	if err != nil {
		log.Error("Error getting mongo session %+v", err)
		return prefix
	}
	collection := session.DB("trident").C("referenceSequence")
	reference := new(mongo.ReferenceSequence)
	_, err = collection.Find(bson.M{"objectType": prefix}).Apply(change, reference)
	if err != nil {
		reference.Id = bson.NewObjectId()
		reference.ObjectType = prefix
		reference.Sequence = 1000
		err = collection.Insert(reference)
		if err != nil {
			log.Error("Error creating referenceSequence %v", err)
		}
	}
	return prefix + fmt.Sprint(reference.Sequence)
}

func mongoIdCreatorInstance(ctxt context.Context) IdCreator {
	return &mongoIdCreator{
		ctxt: ctxt,
	}
}

func getSession(ctxt context.Context) (*mgo.Session, error) {
	dialInfoValue := ctxt.Value("mongodb")
	var session *mgo.Session
	dialInfo, ok := dialInfoValue.(*mgo.DialInfo)
	if !ok {
		return nil, errors.New("Cast to *mgo.DialInfo failed. ctx" + fmt.Sprint(ctxt))
	}
	session, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		return nil, err
	}
	if dialInfo.Mechanism == mongo.MongoDbX509Mechanism {
		credValue := ctxt.Value("mongodb-cred")
		cred, ok := credValue.(*mgo.Credential)
		if !ok {
			return nil, errors.New("Cast to *mgo.Credential failed. ctx" + fmt.Sprint(ctxt))
		}
		if session.Login(cred) != nil {
			return nil, err
		}
	}
	return session, err
}
