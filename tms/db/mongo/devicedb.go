package mongo

import (
	"context"
	"reflect"
	"unsafe"

	"prisma/tms/db"
	"prisma/tms/log"
	"prisma/tms/moc"

	"prisma/gogroup"

	"prisma/tms"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const (
	collectionDevice = "devices"
)

var (
	DBDeviceSD = NewStructData(
		reflect.TypeOf(DBDevice{}),
		NoMap)
)

type DeviceDb struct {
	ctxt   context.Context // execustion context
	dbconn *MongoClient    // Connection to DB
	reg    db.RegistryDB   // Connection to registry
}

func NewMongoDeviceDb(ctxt context.Context, dbconn *MongoClient) db.DeviceDB {
	return &DeviceDb{
		dbconn: dbconn,
		//reg:    NewMongoRegistry(ctxt, dbconn),
		ctxt: ctxt,
	}
}

//Mongodb device client used by tdabased to insert devices coming from the backend
func NewMongoDeviceClient(ctxt gogroup.GoGroup, dbconn *MongoClient) db.DeviceDB {
	client := &DeviceDb{
		dbconn: dbconn,
		reg:    NewMongoRegistry(ctxt, dbconn),
		ctxt:   ctxt,
	}

	return client
}

func (d *DeviceDb) FindAll(ctx context.Context, sortFields db.SortFields) ([]*moc.Device, error) {
	session, err := getSession(ctx)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	pipe := []bson.M{
		{"$match": bson.M{}},
		{"$sort": createMongoSort(sortFields)},
	}

	items := make([]*moc.Device, 0) // make(, 0), else the python will have nil the body from rsp
	err = session.DB(DATABASE).C(collectionDevice).Pipe(pipe).All(&items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (d *DeviceDb) FindOne(ctx context.Context, deviceID string) (*moc.Device, error) {
	session, err := getSession(ctx)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	coder := Coder{TypeData: DBDeviceSD}
	dev := new(DBDevice)
	raw := &bson.Raw{}

	err = session.DB(DATABASE).C(collectionDevice).FindId(bson.ObjectIdHex(deviceID)).One(raw)
	if mgo.ErrNotFound != err {
		coder.DecodeTo(*raw, unsafe.Pointer(dev))
	} else {
		return nil, db.ErrorNotFound
	}

	return dev.ToDevice(), err
}

// FindByDevice uses DeviceId and Type
func (d *DeviceDb) FindByDevice(device *moc.Device) (*moc.Device, error) {
	conn := d.dbconn.DB()
	defer d.dbconn.Release(conn)

	coder := Coder{TypeData: DBDeviceSD}
	dev := new(DBDevice)
	raw := &bson.Raw{}
	query := bson.M{
		"deviceid": device.DeviceId,
		"type":     device.Type,
	}
	err := conn.C(collectionDevice).Find(query).One(raw)
	if mgo.ErrNotFound != err {
		coder.DecodeTo(*raw, unsafe.Pointer(dev))
	}

	return dev.ToDevice(), err
}

func (d *DeviceDb) FindNet(netID string) (*moc.Device, error) {
	conn := d.dbconn.DB()
	defer d.dbconn.Release(conn)

	coder := Coder{TypeData: DBDeviceSD}
	dev := new(DBDevice)
	raw := &bson.Raw{}
	// {"networks": {$elemMatch:{"providerid": "iridium", "subscriberid": "6301989f2903407d877f2210fb80bb54"}}}
	query := bson.M{
		"networks": bson.M{
			"$elemMatch": bson.M{
				"subscriberid": netID,
			},
		},
	}
	err := conn.C(collectionDevice).Find(query).One(raw)
	if mgo.ErrNotFound != err {
		if err != nil {
			return dev.ToDevice(), err
		}
		coder.DecodeTo(*raw, unsafe.Pointer(dev))
	}

	return dev.ToDevice(), err
}

func (d *DeviceDb) UpsertDeviceConfig(DeviceId, Type string, conf *moc.DeviceConfiguration) error {
	log.Debug("update device %v %v with %v", DeviceId, Type, conf)
	conn := d.dbconn.DB()
	defer d.dbconn.Release(conn)

	coder := Coder{TypeData: DBDeviceSD}
	dev := DBDevice{}
	raw := &bson.Raw{}
	query := bson.M{
		"deviceid": DeviceId,
		"type":     Type,
	}
	err := conn.C(collectionDevice).Find(query).One(raw)
	if mgo.ErrNotFound != err {
		coder.DecodeTo(*raw, unsafe.Pointer(&dev))
	}

	dev.Configuration = &DBDeviceConfiguration{
		Id:            dev.ID.Hex(),
		FileId:        conf.FileId,
		Configuration: conf.Configuration,
		Original:      conf.Original,
		LastUpdate:    conf.LastUpdate,
	}
	// map entity relationship
	entities := make([]*DBEntityRelationship, 0)
	for _, entity := range conf.Entities {
		if entity != nil {
			dbEntity := &DBEntityRelationship{
				Id:           entity.Id,
				Type:         entity.Type,
				UpdateTime:   tms.FromTimestamp(entity.UpdateTime),
				Relationship: moc.EntityRelationship_Relationship_name[int32(entity.Relationship)],
			}
			entities = append(entities, dbEntity)
		}
	}
	dev.Configuration.Entities = entities

	enc := Encode(&dev)
	_, err = conn.C(collectionDevice).UpsertId(dev.ID, enc)
	if mgo.IsDup(err) {
		err = db.ErrorDuplicate
	}
	return err
}

func (d *DeviceDb) RemoveVesselInfoForDevices(devices []string) error {
	conn := d.dbconn.DB()
	defer d.dbconn.Release(conn)

	selector := bson.M{
		"id": bson.M{
			"$in": devices,
		},
	}
	updator := bson.M{
		"$set": bson.M{
			"vessel_info": nil,
		},
	}
	return conn.C(collectionDevice).Update(selector, updator)
}

func (d *DeviceDb) UpsertVesselInfo(device *moc.Device, vesselInfo *moc.VesselInfo) error {
	conn := d.dbconn.DB()
	defer d.dbconn.Release(conn)

	coder := Coder{TypeData: DBDeviceSD}
	dev := DBDevice{}
	raw := &bson.Raw{}
	query := bson.M{
		"deviceid": device.DeviceId,
		"type":     device.Type,
	}
	err := conn.C(collectionDevice).Find(query).One(raw)
	if mgo.ErrNotFound != err {
		coder.DecodeTo(*raw, unsafe.Pointer(&dev))
	}

	dev.VesselInfo = &DBVesselInfo{
		ID:   vesselInfo.Id,
		Type: vesselInfo.Type,
	}

	enc := Encode(&dev)
	_, err = conn.C(collectionDevice).UpsertId(dev.ID, enc)
	if mgo.IsDup(err) {
		err = db.ErrorDuplicate
	}
	return err
}

func (d *DeviceDb) Update(ctx context.Context, device *moc.Device) error {
	session, err := getSession(ctx)
	if err != nil {
		return err
	}
	defer session.Close()
	mongoId := bson.ObjectIdHex(device.Id)
	networks := make([]*DBNetwork, 0)
	for _, network := range device.Networks {
		if network != nil {
			dbNetwork := &DBNetwork{
				SubscriberId: network.SubscriberId,
				Type:         network.Type,
				ProviderId:   network.ProviderId,
				RegistryId:   network.RegistryId,
			}
			networks = append(networks, dbNetwork)
		}
	}
	dev := DBDevice{
		ID:            mongoId,
		MongoID:       device.Id,
		RegistryId:    device.RegistryId,
		DeviceId:      device.DeviceId,
		Type:          device.Type,
		Networks:      networks,
		Configuration: d.getConfiguration(device),
		VesselInfo:    d.getVesselInfo(device),
	}
	enc := Encode(&dev)
	_, err = session.DB(DATABASE).C(collectionDevice).UpsertId(mongoId, enc)
	if mgo.IsDup(err) {
		err = db.ErrorDuplicate
	}
	return err
}

func (d *DeviceDb) Delete(ctx context.Context, deviceId string) error {
	session, err := getSession(ctx)
	if err != nil {
		return err
	}
	defer session.Close()

	mongoId := bson.ObjectIdHex(deviceId)
	err = session.DB(DATABASE).C(collectionDevice).
		Remove(bson.D{{Name: "_id", Value: mongoId}})
	if mgo.ErrNotFound == err {
		err = db.ErrorNotFound
	}
	return err
}
