package mongo

import (
	"context"

	"prisma/tms/db"
	"prisma/tms/moc"
	"prisma/tms/rest"
	"prisma/tms/security/policy"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const (
	// CollectionConfig mongo collection
	CollectionConfig = "config"
)

// Brand name
type Brand struct {
	Name        string `json:"name,omitempty"`
	Version     string `json:"version,omitempty"`
	Git         string `json:"git,omitempty"`
	ReleaseDate string `json:"releaseDate,omitempty"`
}

// Configuration struct
type Configuration struct {
	MongoId bson.ObjectId                     `json:"-" bson:"_id,omitempty"`
	Id      string                            `json:"id,omitempty"`
	Lon     float32                           `json:"lon,omitempty"`
	Lat     float32                           `json:"lat,omitempty"`
	Zoom    float32                           `json:"zoom,omitempty"`
	Brand   Brand                             `json:"brand,omitempty"`
	Meta    map[string]map[string]interface{} `json:"meta,omitempty"`
	Site    *moc.Site                         `json:"site,omitempty"`
	Service *rest.Service                     `json:"service,omitempty"`
	Client  *rest.Client                      `json:"client,omitempty"`
	Policy  *policy.Policy                    `json:"policy,omitempty"`
	//TODO: this is to enable spider tracks config for 1.7
	// Should be removed in 1.8
	Spider bool `json:"showSpidertracksLayerHide,omitempty"`
}

// ConfigDb single object config
type ConfigDb struct{}

func (d *ConfigDb) Read(ctx context.Context) (*Configuration, error) {
	session, err := getSession(ctx)
	if err != nil {
		return nil, err
	}
	defer session.Close()
	var config Configuration
	err = session.DB(DATABASE).C(CollectionConfig).Find(nil).Limit(1).One(&config)
	config.Id = config.MongoId.Hex()
	if mgo.ErrNotFound == err {
		err = db.ErrorNotFound
	}
	return &config, err
}

func (d *ConfigDb) Update(ctx context.Context, config *Configuration) error {
	if config.Id == "" || !bson.IsObjectIdHex(config.Id) {
		return db.ErrorNotFound
	}
	id := bson.ObjectIdHex(config.Id)
	session, err := getSession(ctx)
	if err != nil {
		return err
	}
	defer session.Close()
	err = session.DB(DATABASE).C(CollectionConfig).UpdateId(id, config)
	config.Id = id.Hex()
	if err == mgo.ErrNotFound {
		err = db.ErrorNotFound
	}
	if mgo.IsDup(err) {
		err = db.ErrorDuplicate
	}
	return err
}

func (d *ConfigDb) Create(ctx context.Context, config *Configuration) error {
	session, err := getSession(ctx)
	if err != nil {
		return err
	}
	defer session.Close()
	mongoId := bson.NewObjectId()
	_, err = session.DB(DATABASE).C(CollectionConfig).UpsertId(mongoId, &config)
	config.Id = mongoId.Hex()
	if mgo.IsDup(err) {
		err = db.ErrorDuplicate
	}
	return err
}
