package mongo

import (
	. "prisma/tms"
	. "prisma/tms/db"
	"prisma/tms/moc"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/timestamp"
)

var (
	// Mongo-specific table info. Just the default tables, plus mongo-specific
	// Track and registry tables.
	MongoTables = NewTables(append(
		DefaultTables.Info,
		&TableInfo{
			Name:           "referenceSequence",
			Inst:           ReferenceSequence{},
			NoCappedMirror: true,
			Indexes:        []Index{},
		},

		&TableInfo{
			Name: "tracks",
			Inst: DBTrack{},
			Indexes: []Index{
				Index{
					Name: "time",
					Fields: []Field{
						ResolveField(DBTrack{}, "Time"),
					},
				},
				Index{
					Name: "update_time",
					Fields: []Field{
						ResolveField(DBTrack{}, "UpdateTime"),
					},
				},
				Index{
					Name: "track_id_time",
					Fields: []Field{
						ResolveField(DBTrack{}, "TrackID"),
						ResolveField(DBTrack{}, "Time"),
						ResolveField(DBTrack{}, "UpdateTime"),
					},
					Unique: true,
				},
			},
		},

		&TableInfo{
			Name: "activity",
			Inst: DBActivity{},
			Indexes: []Index{
				Index{
					Name: "time",
					Fields: []Field{
						ResolveField(DBActivity{}, "Time"),
					},
				},
				Index{
					Name: "request_id",
					Fields: []Field{
						ResolveField(DBActivity{}, "RequestId"),
					},
				},
				Index{
					Name: "activity_id_time",
					Fields: []Field{
						ResolveField(DBActivity{}, "ActivityId"),
						ResolveField(DBActivity{}, "Time"),
					},
					Unique: true,
				},
			},
		},

		&TableInfo{
			Name: "devices",
			Inst: DBDevice{},
		},

		&TableInfo{
			Name: "registry",
			Inst: DBRegistryEntry{},
			Indexes: []Index{
				Index{
					Name:   "registry_id",
					Unique: true,
					Fields: []Field{
						ResolveField(DBRegistryEntry{}, "Entry.RegistryId"),
					},
				},
				Index{
					Name: "target_id",
					Fields: []Field{
						ResolveField(DBRegistryEntry{}, "Entry.Target.Id"),
					},
				},
				Index{
					Name:      "location",
					GeoSphere: true,
					Fields: []Field{
						ResolveField(DBRegistryEntry{}, "Entry.Target.Position"),
					},
				},
			},
		}))
)

type ReferenceSequence struct {
	Id         bson.ObjectId `bson:"_id"`
	ObjectType string        `bson:"objectType"`
	Sequence   int           `bson:"sequence"`
}

// In order to insert something, we need access to a bson ObjectID, at a
// minimum.
type DBInsertable interface {
	GetId() bson.ObjectId
	SetId(bson.ObjectId)
}

// DBTrack specifies the Mongo schema for a track
type DBTrack struct {
	ID         bson.ObjectId  `bson:"_id"`         // A mongo id. Not really used anywhere, but required for every row.
	TrackID    string         `bson:"track_id"`    // The track's ID in protobuf format
	RegistryId string         `bson:"registry_id"` // The tracks's registry entry ID, a mongo objectid
	Time       time.Time      `bson:"time"`        // The of the metadata and target, the latest time
	UpdateTime time.Time      `bson:"update_time"` // Usually the time found in the target but could be the last the time target was extended
	Metadata   *TrackMetadata `bson:"md"`
	Target     *Target        `bson:"tgt"`
}

// DBActivity object specifies the Mongo schema for a activity
type DBActivity struct {
	ID         bson.ObjectId    `bson:"_id"`         // A mongo id. Not really used anywhere, but required for every row.
	ActivityId string           `bson:"activity_id"` // The activity's ID in protobuf format
	RegistryId string           `bson:"registry_id"` // The activity's registry entry ID, a mongo objectid
	RequestId  string           `bson:"request_id"`  // the id of the system request that references the activity
	Time       int64            `bson:"time"`        // The time of the body, the latest time
	Body       *MessageActivity `bson:"me"`
}

//DBDevice object spccifies the Mongo schema for the  device
type DBDevice struct {
	ID            bson.ObjectId          `bson:"_id"`           // A mongo id. not really used anywhere, but required for every row.
	MongoID       string                 `bson:"id"`            // used by front-end, string form of ID
	DeviceId      string                 `bson:"deviceid"`      // the device's ID in protobuf format
	RegistryId    string                 `bson:"registryid"`    // the device's registry entry ID, a mongo objectid
	Type          string                 `bson:"type"`          // This should not be a string but a DeviceType
	Networks      []*DBNetwork           `bson:"networks"`      // An array of networks that the device is using
	Configuration *DBDeviceConfiguration `bson:"configuration"` // Device configuration
	VesselInfo    *DBVesselInfo          `bson:"vessel_info"`   // Information on the vessel that own the current device
	Time          time.Time              `bson:"time"`          // The time of the device registry
}

type DBVesselInfo struct {
	ID   string `bson:"id"`
	Type string `bson:"type"`
}

type DBDeviceConfiguration struct {
	Id            string                  `bson:"id"`
	Entities      []*DBEntityRelationship `bson:"entities"`
	FileId        string                  `bson:"fileid"`
	Configuration *any.Any                `bson:"configuration"`
	Original      *any.Any                `bson:"original"`
	LastUpdate    *timestamp.Timestamp    `bson:"lastupdate"`
}

type DBEntityRelationship struct {
	Id           string    `bson:"id"`
	Type         string    `bson:"type"`
	UpdateTime   time.Time `bson:"update_time"`
	Relationship string    `bson:"relationship"`
}

type DBNetwork struct {
	SubscriberId string `bson:"subscriberid"`
	Type         string `bson:"type"`
	ProviderId   string `bson:"providerid"`
	RegistryId   string `bson:"registryid"`
}

// GetId returns bson object id from DBTrack
func (obj *DBTrack) GetId() bson.ObjectId {
	return obj.ID
}

// SetId sets bson object Id ad DBtrack ID
func (obj *DBTrack) SetId(id bson.ObjectId) {
	obj.ID = id
}

// ToDevice Convert DBDevice to protobuf Device object
func (d *DBDevice) ToDevice() *moc.Device {
	ret := &moc.Device{
		Id:         d.ID.Hex(),
		DeviceId:   d.DeviceId,
		RegistryId: d.RegistryId,
		Type:       d.Type,
	}
	if d.Configuration != nil {
		ret.Configuration = &moc.DeviceConfiguration{
			Id:            d.ID.Hex(),
			FileId:        d.Configuration.FileId,
			Configuration: d.Configuration.Configuration,
			Original:      d.Configuration.Original,
			LastUpdate:    d.Configuration.LastUpdate,
		}
		if d.Configuration.Entities != nil {
			ret.Configuration.Entities = make([]*moc.EntityRelationship, 0)
			for _, entity := range d.Configuration.Entities {
				mocEntity := &moc.EntityRelationship{
					Id:           entity.Id,
					Type:         entity.Type,
					Relationship: moc.EntityRelationship_Relationship(moc.EntityRelationship_Relationship_value[entity.Relationship]),
					UpdateTime:   ToTimestamp(entity.UpdateTime),
				}
				ret.Configuration.Entities = append(ret.Configuration.Entities, mocEntity)
			}
		}
	}
	if d.Networks != nil {
		ret.Networks = make([]*moc.Device_Network, 0)
		for _, network := range d.Networks {
			mocNetwork := &moc.Device_Network{
				SubscriberId: network.SubscriberId,
				Type:         network.Type,
				ProviderId:   network.ProviderId,
				RegistryId:   network.RegistryId,
			}
			ret.Networks = append(ret.Networks, mocNetwork)
		}
	}
	return ret
}

// ToTrack Convert DBTrack to a protobuf Track object
func (t *DBTrack) ToTrack() *Track {
	ret := &Track{
		Id: t.TrackID,
	}
	if t.Target != nil {
		ret.Targets = []*Target{t.Target}
	}
	if t.Metadata != nil {
		ret.Metadata = []*TrackMetadata{t.Metadata}
	}
	ret.DatabaseId = t.ID.Hex()
	ret.RegistryId = t.RegistryId
	return ret
}

// The mongo schema for "_live" tables which includes the change row id
type DBChangeWithID struct {
	// The change row id
	Id bson.ObjectId `bson:"_id"`

	// The _id of the updated document
	ObjId bson.ObjectId `bson:"objid"`

	// Action is one of: New, Update, Delete
	Action string `bson:"action"`
}

// Mongo schema for misc_data objects
type DBMiscObject struct {
	Id             bson.ObjectId `bson:"_id"`
	CreationTime   time.Time     `bson:"ctime"`
	ExpirationTime time.Time     `bson:"etime"`
	UpdateTime     time.Time     `bson:"utime"`
	RegistryId     string        `bson:"registry_id"`

	Obj interface{} `bson:"me"`
}

// Mongo schema for registry entries
// If redirect then use reference to Entry.id (me.id)
type DBRegistryEntry struct {
	ID      bson.ObjectId  `bson:"_id"`
	FleetID bson.ObjectId  `bson:"fleet_id"`
	Entry   *RegistryEntry `bson:"me"`
}
