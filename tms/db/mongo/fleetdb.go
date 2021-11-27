package mongo

import (
	"context"

	"prisma/tms/db"
	"prisma/tms/moc"
	"prisma/tms/rest"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const (
	collectionFleet = "fleets"
)

type FleetDb struct {
	ctxt context.Context
}

func NewFleetDb(ctxt context.Context) db.FleetDB {
	return &FleetDb{
		ctxt: ctxt,
	}
}

func (d *FleetDb) Create(ctxt context.Context, fleet *moc.Fleet) (*moc.Fleet, error) {
	session, err := getSession(ctxt)
	if err != nil {
		return nil, err
	}
	defer session.Close()
	collection := session.DB(DATABASE).C(collectionFleet)
	mongoId := bson.NewObjectId()
	fleet.Id = mongoId.Hex()
	addChildIdsToFleet(fleet)
	_, err = collection.UpsertId(mongoId, fleet)
	if mgo.IsDup(err) {
		err = db.ErrorDuplicate
	}
	return fleet, err
}

func (d *FleetDb) Update(ctxt context.Context, fleet *moc.Fleet) (*moc.Fleet, error) {
	session, err := getSession(ctxt)
	if err != nil {
		return nil, err
	}
	defer session.Close()
	collection := session.DB(DATABASE).C(collectionFleet)
	addChildIdsToFleet(fleet)
	_, err = collection.UpsertId(bson.ObjectIdHex(fleet.Id), fleet)
	if mgo.IsDup(err) {
		err = db.ErrorDuplicate
	}
	return fleet, err
}

func (d *FleetDb) RemoveVessel(ctxt context.Context, fleetId string, vesselId string) error {
	session, err := getSession(ctxt)
	if err != nil {
		return err
	}
	defer session.Close()
	collection := session.DB(DATABASE).C(collectionFleet)
	err = collection.Update(
		bson.M{"_id": bson.ObjectIdHex(fleetId)},
		bson.M{
			"$pull": bson.M{
				"vessels": bson.M{"id": bson.ObjectIdHex(vesselId)},
			},
		},
	)
	if mgo.ErrNotFound == err {
		err = db.ErrorNotFound
	}
	return err
}

func (d *FleetDb) AddVessel(ctxt context.Context, fleetId string, vessel *moc.Vessel) error {
	session, err := getSession(ctxt)
	if err != nil {
		return err
	}
	defer session.Close()
	collection := session.DB(DATABASE).C(collectionFleet)
	addChildIdsToVessel(vessel)
	err = collection.Update(
		bson.M{"_id": bson.ObjectIdHex(fleetId)},
		bson.M{
			"$push": bson.M{
				"vessels": vessel,
			},
		},
	)
	if mgo.ErrNotFound == err {
		err = db.ErrorNotFound
	}
	return err
}

func (d *FleetDb) UpdateVessel(ctxt context.Context, fleetId string, vessel *moc.Vessel) error {
	session, err := getSession(ctxt)
	if err != nil {
		return err
	}
	defer session.Close()
	collection := session.DB(DATABASE).C(collectionFleet)
	addChildIdsToVessel(vessel)
	err = collection.Update(
		bson.M{
			"_id":        bson.ObjectIdHex(fleetId),
			"vessels.id": vessel.Id,
		},
		bson.M{
			"$set": bson.M{
				"vessels.$": vessel,
			},
		},
	)
	if mgo.ErrNotFound == err {
		err = db.ErrorNotFound
	}
	return err
}

func (d *FleetDb) Delete(ctxt context.Context, fleetId string) error {
	session, err := getSession(ctxt)
	if err != nil {
		return err
	}
	defer session.Close()
	collection := session.DB(DATABASE).C(collectionFleet)
	err = collection.Remove(bson.D{{Name: "_id", Value: bson.ObjectIdHex(fleetId)}})
	if mgo.ErrNotFound == err {
		err = db.ErrorNotFound
	}
	return err
}

func (d *FleetDb) FindOne(ctxt context.Context, fleetId string) (*moc.Fleet, error) {
	fleet := new(moc.Fleet)
	session, err := getSession(ctxt)
	if err != nil {
		return fleet, err
	}
	defer session.Close()

	err = session.DB(DATABASE).C(collectionFleet).
		FindId(bson.ObjectIdHex(fleetId)).
		One(fleet)
	if mgo.ErrNotFound == err {
		err = db.ErrorNotFound
	}
	return fleet, err
}

func (d *FleetDb) FindAll(ctxt context.Context) ([]*moc.Fleet, error) {
	session, err := getSession(ctxt)
	if err != nil {
		return nil, err
	}
	defer session.Close()
	collection := session.DB(DATABASE).C(collectionFleet)
	var fleets []*moc.Fleet
	return fleets, collection.Find(nil).All(&fleets)
}

func (d *FleetDb) FindByMapByPagination(ctxt context.Context, searchMap map[string]string, pagination *rest.PaginationQuery) ([]*moc.Fleet, error) {
	pipe := createPipe(searchMap, pagination)
	session, err := getSession(ctxt)
	if err != nil {
		return nil, err
	}
	defer session.Close()
	var fleets []*moc.Fleet
	err = session.DB(DATABASE).C(collectionFleet).Pipe(pipe).All(&fleets)
	if mgo.ErrNotFound == err {
		err = db.ErrorNotFound
	}
	return fleets, err
}

func (d *FleetDb) FindByMap(ctxt context.Context, searchMap map[string]string) ([]*moc.Fleet, error) {
	panic("not implemented")
	var fleets []*moc.Fleet
	var err error
	return fleets, err
}

func addChildIdsToFleet(fleet *moc.Fleet) {
	addChildIdsToPerson(fleet.Person)
	addChildIdsToOrganization(fleet.Organization)

}
