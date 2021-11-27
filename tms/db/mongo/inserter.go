package mongo

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unsafe"

	. "prisma/tms"
	"prisma/tms/db"
	"prisma/tms/devices"
	"prisma/tms/log"
	"prisma/tms/moc"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	pb "github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
)

var (
	DBTrackSD = NewStructData(
		reflect.TypeOf(DBTrack{}),
		NoMap)
)

// For a particular TrackID, get the last track in the database occurring
// before 'time'. Used to supplement tracks.
func (d *MongoTrackClient) GetLast(id string, time time.Time) *DBTrack {
	metaTable := MongoTables.TableFromType(DBTrack{})
	if metaTable.Name == "" {
		panic("Could not find table name for DBTrack type")
	}

	mdb := d.dbconn.DB()
	defer d.dbconn.Release(mdb)

	q := mdb.C(metaTable.Name).Find(bson.M{
		dbtrack_TimeField: bson.M{"$lte": time},
		"track_id":        id,
	}).Sort("-" + dbtrack_TimeField).Limit(1)
	log.TraceMsg("Metadata look query: %v", q)
	iter := q.Iter()

	var res bson.Raw
	if iter.Next(&res) {
		c := Coder{TypeData: DBTrackSD}
		resp := new(DBTrack)
		c.DecodeTo(res, unsafe.Pointer(resp))
		return resp
	} else {
		if iter.Err() != nil {
			log.Warn("Error getting last for data supplementation: %v", iter.Err())
			return nil
		}
	}
	return nil
}

// Supplement a track with an ID, registry ID, and missing target/metadata
func (d *MongoTrackClient) SupplementTrack(t *DBTrack) (*DBTrack, error) {
	now := time.Now()
	pnow, err := ptypes.TimestampProto(now)

	if err != nil {
		return nil, err
	}
	if t.Metadata != nil && t.Target != nil {
		mdTime := FromTimestamp(t.Metadata.Time)
		tgtTime := FromTimestamp(t.Target.Time)
		if mdTime.Before(tgtTime) {
			t.Time = tgtTime
		} else {
			t.Time = mdTime
		}
	} else if t.Metadata != nil {
		t.Time = FromTimestamp(t.Metadata.Time)
		t.UpdateTime = t.Time
	} else if t.Target != nil {
		t.Time = FromTimestamp(t.Target.Time)
	} else {
		return nil, nil
	}
	if t.TrackID == "" {
		return nil, fmt.Errorf("Track doesn't have an ID!, the track is: %+v", t)
	}

	if t.Target != nil {
		if t.Target.Repeat {
			t.UpdateTime = now
			t.Target.UpdateTime = pnow
		} else {
			if t.Target.UpdateTime == nil {
				t.UpdateTime = FromTimestamp(t.Target.Time)
				t.Target.UpdateTime = t.Target.Time
			} else {
				t.UpdateTime = now
				t.Target.UpdateTime = pnow
			}
		}
	}

	if t.Metadata == nil || t.Target == nil {
		last := d.GetLast(t.TrackID, t.Time)
		if last != nil {
			if t.Metadata == nil {
				t.Metadata = last.Metadata
			}
			if t.Target == nil {
				t.Target = last.Target
			}
		}
	}
	return t, nil
}

// SupplementBeforeInsert an arbitrary object, depending on its type
func (d *MongoTrackClient) SupplementBeforeInsert(track *Track, src interface{}) (DBInsertable, error) {
	switch x := src.(type) {
	case Target:
		return d.SupplementTrack(&DBTrack{TrackID: track.Id, RegistryId: track.RegistryId, Target: &x})
	case *Target:
		return d.SupplementTrack(&DBTrack{TrackID: track.Id, RegistryId: track.RegistryId, Target: x})
	case TrackMetadata:
		return d.SupplementTrack(&DBTrack{TrackID: track.Id, RegistryId: track.RegistryId, Metadata: &x})
	case *TrackMetadata:
		return d.SupplementTrack(&DBTrack{TrackID: track.Id, RegistryId: track.RegistryId, Metadata: x})
	}
	return nil, errors.New("Object could not be supplemented to DBInsertable")
}

func (d *DeviceDb) getVesselInfo(device *moc.Device) *DBVesselInfo {
	if device.VesselInfo == nil {
		return nil
	}
	return &DBVesselInfo{
		Type: device.VesselInfo.Type,
		ID:   device.VesselInfo.Id,
	}
}

func (d *DeviceDb) getConfiguration(device *moc.Device) *DBDeviceConfiguration {
	if device.Configuration == nil {
		return nil
	}
	conf := &DBDeviceConfiguration{
		Id:            device.Id,
		FileId:        device.Configuration.FileId,
		Configuration: device.Configuration.Configuration,
		Original:      device.Configuration.Original,
		LastUpdate:    device.Configuration.LastUpdate,
	}
	if device.Configuration.Entities != nil {
		conf.Entities = make([]*DBEntityRelationship, 0)
		for _, entity := range device.Configuration.Entities {
			dbEntity := &DBEntityRelationship{
				Id:           entity.Id,
				Type:         entity.Type,
				UpdateTime:   FromTimestamp(entity.UpdateTime),
				Relationship: moc.EntityRelationship_Relationship_name[int32(entity.Relationship)],
			}
			conf.Entities = append(conf.Entities, dbEntity)
		}
	}
	return conf
}

//Insert API entry point intert all device data
func (d *DeviceDb) Insert(device *moc.Device) error {
	if device == nil {
		return errors.New("nil device")
	}
	var err error
	id := bson.NewObjectId()
	device.Id = id.Hex()
	if device.Configuration != nil {
		device.Configuration.Id = device.Id
	}
	// map network for db
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
	dev := &DBDevice{
		ID:            id,
		MongoID:       id.Hex(),
		RegistryId:    device.RegistryId,
		DeviceId:      device.DeviceId,
		Type:          device.Type,
		Networks:      networks,
		VesselInfo:    d.getVesselInfo(device),
		Configuration: d.getConfiguration(device),
	}
	table := MongoTables.TableFromType(dev)
	if table.Name == "" {
		return fmt.Errorf("Could not find table for object: %v", dev)
	}
	enc := Encode(dev)
	conn := d.dbconn.DB()
	defer d.dbconn.Release(conn)
	// mongo indexes guarantee constraints, no need for defensive queries
	err = conn.C(table.Name).Insert(enc)
	if err != nil {
		log.Error("inserting device object %v into database: %v", dev, err)
		return err
	}
	return nil
}

// Insert API entry poing insert all activity data
func (d *MongoActivityClient) Insert(activity *MessageActivity) error {
	if activity == nil {
		return errors.New("nil activity")
	}

	return d.insert(activity)
}

//Insert an arbitrary activity object into the database.
//Returns an error if the type is not correct
func (d *MongoActivityClient) insert(activity *MessageActivity) error {
	var err error

	a := &DBActivity{
		ActivityId: activity.ActivityId,
		RegistryId: activity.RegistryId,
		Body:       activity,
		Time:       FromTimestamp(activity.Time).UnixNano(),
	}

	// *** select which table we are going to insert to
	table := MongoTables.TableFromType(a)
	if table.Name == "" {
		return errors.New(fmt.Sprintf("Could not find table for object: %v", a))
	}

	a.ID = bson.NewObjectId()

	enc := Encode(a)

	mdb := d.dbconn.DB()
	defer d.dbconn.Release(mdb)

	err = mdb.C(table.Name).Insert(enc)
	if err != nil {
		log.Error("inserting activity object %v into database: %v", a, table, d, err)
		return err
	}
	return nil
}

// Insert an arbitrary object into the database. Returns an error is the type
// is not recognized.
func (d *MongoTrackClient) insert(track *Track, orig interface{}) error {
	// FIXME: For now, we just ignore GPS sentences
	target, ok := orig.(*Target)
	if ok && target.Type == devices.DeviceType_GPS {
		return nil
	}
	// *** Supplement the object
	obj, err := d.SupplementBeforeInsert(track, orig)
	if err != nil {
		log.Error("Error supplementing object: %v", err)
		return err
	}
	if obj == nil {
		panic("Got no error AND nil object!")
	}

	// *** Figure out into which table the object should be inserted
	ti := MongoTables.TableFromType(obj)
	if ti.Name == "" {
		return errors.New(fmt.Sprintf("Could not location table for obj: %v", obj))
	}

	// *** Figure out a mongo objectid for the object
	id := obj.GetId()
	if !id.Valid() {
		id = bson.NewObjectId()
		obj.SetId(id)
	}
	encoded := Encode(obj)

	mdb := d.dbconn.DB()
	defer d.dbconn.Release(mdb)

	err = mdb.C(ti.Name).Insert(encoded)
	if err != nil {
		if strings.HasPrefix(err.Error(), "E11000 duplicate key error collection") {
			log.Info("inserting object into database: %v", err, obj, ti, d)
		} else {
			log.Error("inserting object into database: %v", err, obj, ti, d)
		}
		return err
	}

	if track.RegistryId != "" {
		d.updateRegistry(track)
	}

	return nil
}

// API entry point: insert all data associated with a protobuf Track message
func (d *MongoTrackClient) Insert(track *Track) error {
	if track == nil {
		return errors.New("nil track")
	}

	var reterr error
	// insert data
	for _, meta := range track.Metadata {
		err := d.insert(track, meta)
		if err != nil {
			reterr = err
		}
	}

	// insert target, discard track
	for _, target := range track.Targets {
		err := d.insert(track, target)
		if err != nil {
			return err
		}
	}
	// insert track
	//err := d.insert(track)
	//if err != nil {
	//	reterr = err
	//}
	return reterr
}

func (d *MongoTrackClient) syncRegistry(registryId string) {
	d.condRegistry.L.Lock()
	defer d.condRegistry.L.Unlock()
	for {
		d.muRegistryInserts.Lock()
		_, ok := d.registryInserts[registryId]
		d.muRegistryInserts.Unlock()
		if ok {
			d.condRegistry.Wait()
		} else {
			d.muRegistryInserts.Lock()
			d.registryInserts[registryId] = struct{}{}
			d.muRegistryInserts.Unlock()
			return
		}
	}
}

// updates track with registryId
func (d *MongoTrackClient) updateRegistry(track *Track) {
	if track == nil {
		log.Error("No track specified")
		return
	}
	d.syncRegistry(track.RegistryId)
	defer d.condRegistry.Broadcast()
	defer func() {
		d.muRegistryInserts.Lock()
		delete(d.registryInserts, track.RegistryId)
		d.muRegistryInserts.Unlock()
	}()
	regt, err := d.reg.GetOrCreate(track.RegistryId)
	if err != nil && !mgo.IsDup(err) {
		log.Error("Error getting registry track: %v %v", track.RegistryId, err)
		return
	}

	// **** Replace target, if newer
	tgt := GetNewest(track.Targets)
	if tgt != nil &&
		(regt.Target == nil ||
			FromTimestamp(regt.Target.Time).Before(FromTimestamp(tgt.Time))) {
		regt.Target = tgt
	}

	// **** Merge metadata
	if len(track.Metadata) > 0 {
		if regt.Metadata == nil {
			md := track.Metadata[len(track.Metadata)-1]
			regt.Metadata = pb.Clone(md).(*TrackMetadata)
		}
		for i := len(track.Metadata) - 1; i >= 0; i-- {
			pb.Merge(regt.Metadata, track.Metadata[i])
		}

		regt.MetadataFields = db.RegistryMetadataSearchFields(track)
	}
	if len(track.Targets) > 0 {
		regt.TargetFields = db.RegistryTargetSearchFields(track)
	}
	regt.Keywords = db.RegistryKeywords(regt.TargetFields, regt.MetadataFields)
	db.SetRegistryLabel(regt)

	// *** Add device type
	if tgt != nil {
		regt.DeviceType = tgt.Type
	}
	err = d.reg.Upsert(regt)
	if err != nil && !mgo.IsDup(err) {
		log.Error("Error upserting registry track: %v", err, regt)
		return
	}

}

func GetNewest(tgts []*Target) *Target {
	newestIdx := -1
	var newestT time.Time
	for i, tgt := range tgts {
		t := FromTimestamp(tgt.Time)
		if newestIdx == -1 || newestT.After(t) {
			newestT = t
			newestIdx = i
		}
	}

	if newestIdx == -1 {
		return nil
	}
	return tgts[newestIdx]
}
