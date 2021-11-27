package mongo

import (
	"context"

	"prisma/tms"
	"prisma/tms/db"
	"prisma/tms/log"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const (
	// CollectionMulticast mongo collection
	CollectionMulticast = "multicasts"
)

type MulticastDb struct {
}

func NewMulticastDb(_ context.Context) *MulticastDb {
	return &MulticastDb{}
}

func (d *MulticastDb) Create(ctx context.Context, mc *tms.Multicast) error {
	session, err := getSession(ctx)
	if err != nil {
		return err
	}
	defer session.Close()
	var mongoId bson.ObjectId
	if mc.Id == "" {
		mongoId := bson.NewObjectId()
		mc.Id = mongoId.Hex()
	} else {
		mongoId = bson.ObjectIdHex(mc.Id)
	}
	_, err = session.DB(DATABASE).C(CollectionMulticast).UpsertId(mongoId, mc)
	if mgo.IsDup(err) {
		err = db.ErrorDuplicate
	}
	return err
}

func (d *MulticastDb) Find(ctx context.Context, mcId string) (*tms.Multicast, error) {
	if !bson.IsObjectIdHex(mcId) {
		return nil, db.ErrorNotFound
	}
	session, err := getSession(ctx)
	if err != nil {
		return nil, err
	}
	defer session.Close()
	var mc *tms.Multicast
	err = session.DB(DATABASE).C(CollectionMulticast).FindId(bson.ObjectIdHex(mcId)).One(&mc)
	if mgo.ErrNotFound == err {
		err = db.ErrorNotFound
	}
	return mc, err
}

func (d *MulticastDb) FindAll(ctx context.Context) ([]*tms.Multicast, error) {
	session, err := getSession(ctx)
	if err != nil {
		return nil, err
	}
	defer session.Close()
	var mcs []*tms.Multicast
	err = session.DB(DATABASE).C(CollectionMulticast).Find(nil).All(&mcs)
	if mgo.ErrNotFound == err {
		err = db.ErrorNotFound
	}
	return mcs, err
}

func (d *MulticastDb) FindByMapByState(ctx context.Context, searchMap map[string]string, states []tms.Transmission_State) ([]*tms.Multicast, error) {
	var query bson.M
	mapQuery := createMongoQueryFromMap(searchMap)
	if len(states) > 0 {
		var stateIn []int32
		for _, state := range states {
			stateIn = append(stateIn, int32(state))
		}
		query = bson.M{"$and": []bson.M{
			{"transmissions.state": bson.M{
				"$in": stateIn,
			}},
			mapQuery,
		}}
	} else {
		query = mapQuery
	}
	session, err := getSession(ctx)
	if err != nil {
		return nil, err
	}
	defer session.Close()
	var mcs []*tms.Multicast
	err = session.DB(DATABASE).C(CollectionMulticast).Find(query).All(&mcs)
	if mgo.ErrNotFound == err {
		err = db.ErrorNotFound
	}
	return mcs, err
}

func (d *MulticastDb) Update(ctx context.Context, mc *tms.Multicast) error {
	log.Debug("updating multicast %v transmissions %v", mc.Id, mc.Transmissions)
	if !bson.IsObjectIdHex(mc.Id) {
		return db.ErrorNotFound
	}
	session, err := getSession(ctx)
	if err != nil {
		return err
	}
	defer session.Close()
	_, err = session.DB(DATABASE).C(CollectionMulticast).UpsertId(bson.ObjectIdHex(mc.Id), mc)
	if mgo.IsDup(err) {
		err = db.ErrorDuplicate
	}
	return err
}

func (d *MulticastDb) UpdateTransmission(ctx context.Context, tr *tms.Transmission) error {
	log.Debug("updating multicast %v transmission %v", tr.ParentId, tr)
	if !bson.IsObjectIdHex(tr.ParentId) {
		return db.ErrorNotFound
	}
	session, err := getSession(ctx)
	if err != nil {
		return err
	}
	defer session.Close()
	err = session.DB(DATABASE).C(CollectionMulticast).Update(
		bson.M{
			"_id":              bson.ObjectIdHex(tr.ParentId),
			"transmissions.id": tr.Id,
		},
		bson.M{
			"$set": bson.M{
				"transmissions.$": tr,
			},
		},
	)
	if mgo.IsDup(err) {
		err = db.ErrorDuplicate
	}
	return err
}

func (d *MulticastDb) Delete(ctx context.Context, multicastId string) error {
	session, err := getSession(ctx)
	if err != nil {
		return err
	}
	defer session.Close()
	collection := session.DB(DATABASE).C(CollectionMulticast)
	err = collection.Remove(bson.D{{Name: "_id", Value: bson.ObjectIdHex(multicastId)}})
	if mgo.ErrNotFound == err {
		err = db.ErrorNotFound
	}
	return err
}
