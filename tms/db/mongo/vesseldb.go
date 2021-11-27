package mongo

import (
	"context"

	"prisma/tms/db"
	"prisma/tms/log"
	"prisma/tms/moc"
	"prisma/tms/rest"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const (
	collectionVessel = "vessels"
)

type VesselDb struct {
	ctxt context.Context
}

func NewMongoVesselDb(ctxt context.Context) db.VesselDB {
	return &VesselDb{
		ctxt: ctxt,
	}
}

func (d *VesselDb) Create(ctxt context.Context, vessel *moc.Vessel) (*moc.Vessel, error) {
	session, err := getSession(ctxt)
	if err != nil {
		return nil, err
	}
	defer session.Close()
	addChildIdsToVessel(vessel)
	mongoId := bson.NewObjectId()
	vessel.Id = mongoId.Hex()
	_, err = session.DB(DATABASE).C(collectionVessel).UpsertId(mongoId, vessel)
	if mgo.IsDup(err) {
		err = db.ErrorDuplicate
	}
	return vessel, err
}

func (d *VesselDb) Update(ctxt context.Context, vessel *moc.Vessel) (*moc.Vessel, error) {
	session, err := getSession(ctxt)
	if err != nil {
		return nil, err
	}
	defer session.Close()
	addChildIdsToVessel(vessel)
	_, err = session.DB(DATABASE).C(collectionVessel).UpsertId(bson.ObjectIdHex(vessel.Id), vessel)
	if mgo.IsDup(err) {
		err = db.ErrorDuplicate
	}
	return vessel, err
}

func (d *VesselDb) Delete(ctxt context.Context, vesselId string) error {
	session, err := getSession(ctxt)
	if err != nil {
		return err
	}
	defer session.Close()
	collection := session.DB(DATABASE).C(collectionVessel)
	err = collection.Remove(bson.D{{Name: "_id", Value: bson.ObjectIdHex(vesselId)}})
	if mgo.ErrNotFound == err {
		err = db.ErrorNotFound
	}
	return err
}

func (d *VesselDb) FindByMapByPagination(ctxt context.Context, searchMap map[string]string, pagination *rest.PaginationQuery) ([]*moc.Vessel, error) {
	pipe := createPipe(searchMap, pagination)

	session, err := getSession(ctxt)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	var vessels []*moc.Vessel
	err = session.DB(DATABASE).C(collectionVessel).Pipe(pipe).All(&vessels)
	if mgo.ErrNotFound == err {
		err = db.ErrorNotFound
	}
	return vessels, err
}

func (d *VesselDb) FindByMap(ctxt context.Context, searchMap map[string]string, sortFields db.SortFields) ([]*moc.Vessel, error) {
	query := createMongoQueryFromMap(searchMap)
	session, err := getSession(ctxt)
	if err != nil {
		return nil, err
	}
	defer session.Close()
	pipe := []bson.M{
		{"$match": query},
		{"$sort": createMongoSort(sortFields)},
	}
	var vessels []*moc.Vessel
	err = session.DB(DATABASE).C(collectionVessel).Pipe(pipe).All(&vessels)
	if mgo.ErrNotFound == err {
		err = db.ErrorNotFound
	}
	return vessels, err
}

func (d *VesselDb) FindOne(ctxt context.Context, vesselId string) (*moc.Vessel, error) {
	session, err := getSession(ctxt)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	var vessel *moc.Vessel
	err = session.DB(DATABASE).C(collectionVessel).Find(bson.M{"_id": bson.ObjectIdHex(vesselId)}).One(&vessel)
	if mgo.ErrNotFound == err {
		err = db.ErrorNotFound
	}
	return vessel, err
}

// FindByDevice will use Device.Id if present, else Device.DeviceId & Device.Type
func (d *VesselDb) FindByDevice(ctxt context.Context, device *moc.Device) (*moc.Vessel, error) {
	session, err := getSession(ctxt)
	if err != nil {
		return nil, err
	}
	defer session.Close()
	var vessel *moc.Vessel
	if device.Id != "" {
		err = session.DB(DATABASE).C(collectionVessel).Find(bson.M{"devices.id": device.Id}).One(&vessel)
	} else if device.DeviceId != "" {
		query := bson.M{
			"devices": bson.M{
				"$elemMatch": bson.M{
					"type":     device.Type,
					"deviceid": device.DeviceId,
				},
			},
		}
		err = session.DB(DATABASE).C(collectionVessel).Find(query).One(&vessel)
	}
	//{"devices.networks.subscriberid": "6301989f2903407d877f2210fb80bb54"}

	if mgo.ErrNotFound == err {
		err = db.ErrorNotFound
	}
	log.Debug("found vessel %v with device %v", vessel, device)
	return vessel, err
}

func (d *VesselDb) FindAll(ctxt context.Context, sortFields db.SortFields) ([]*moc.Vessel, error) {
	session, err := getSession(ctxt)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	pipe := []bson.M{
		{"$match": bson.M{}},
		{"$sort": createMongoSort(sortFields)},
	}

	var vessels []*moc.Vessel
	err = session.DB(DATABASE).C(collectionVessel).Pipe(pipe).All(&vessels)
	if err != nil {
		return nil, err
	}

	return vessels, nil
}

func addChildIdsToVessel(vessel *moc.Vessel) {
	if vessel == nil {
		return
	}
	// add device.id if not present
	for _, device := range vessel.Devices {
		if device.Id == "" {
			device.Id = bson.NewObjectId().Hex()
		}
	}
	// add person.id if not present
	for _, person := range vessel.Crew {
		addChildIdsToPerson(person)
	}
	addChildIdsToPerson(vessel.Person)
	addChildIdsToOrganization(vessel.Organization)
}

func addChildIdsToPerson(person *moc.Person) {
	if person == nil {
		return
	}
	// add person.id if not present
	if person.Id == "" {
		person.Id = bson.NewObjectId().Hex()
	}
	// add device.id if not present
	for _, device := range person.Devices {
		if device.Id == "" {
			device.Id = bson.NewObjectId().Hex()
		}
	}
}

func addChildIdsToOrganization(organization *moc.Organization) {
	if organization == nil {
		return
	}
	// add organization.id if not present
	if organization.Id == "" {
		organization.Id = bson.NewObjectId().Hex()
	}
	// add device.id if not present
	for _, device := range organization.Devices {
		if device.Id == "" {
			device.Id = bson.NewObjectId().Hex()
		}
	}
}
