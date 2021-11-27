package mongo

import (
	"fmt"
	"prisma/tms/db"
	"prisma/tms/moc"
	"unsafe"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// RemoteSiteObjectType ...
const RemoteSiteObjectType = "prisma.tms.moc.RemoteSite"

// CollectionRemoteSite mongo collection
const CollectionRemoteSite = "remoteSites"

func NewMongoRemoteSiteMiscData(misc db.MiscDB) db.RemoteSiteDB {
	client, ok := misc.(*MongoMiscClient)
	if !ok {
		return nil
	}
	return &MongoMiscClient{
		dbconn: client.dbconn,
		ctxt:   client.ctxt,
	}
}

func (remoteSiteDb *MongoMiscClient) UpsertRemoteSite(remoteSite *moc.RemoteSite) error {
	_, err := remoteSiteDb.Upsert(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: RemoteSiteObjectType,
			Obj: &db.GoObject{
				ID:   remoteSite.Id,
				Data: remoteSite,
			},
		},
		Ctxt: remoteSiteDb.ctxt,
	})

	return err
}

func (remoteSiteDb *MongoMiscClient) FindAllRemoteSites() ([]*moc.RemoteSite, error) {
	var remoteSites = make([]*moc.RemoteSite, 0)

	req := db.GoRequest{
		ObjectType: RemoteSiteObjectType,
	}
	structData, tableInfo, err := remoteSiteDb.resolveTable(&req)
	if err != nil {
		return remoteSites, err
	}

	raw := []bson.Raw{}
	coder := Coder{TypeData: structData}

	query := bson.M{
		"me.deleted": false,
	}

	err = remoteSiteDb.dbconn.DB().C(tableInfo.Name).Find(query).All(&raw)
	if err != mgo.ErrNotFound {
		for _, data := range raw {
			var obj DBMiscObject
			coder.DecodeTo(data, unsafe.Pointer(&obj))
			remoteSite, ok := obj.Obj.(*moc.RemoteSite)
			if !ok {
				return remoteSites, fmt.Errorf("Could not fetch RemoteSite object")
			}

			remoteSites = append(remoteSites, remoteSite)
		}
	}

	return remoteSites, err
}

func (remoteSiteDb *MongoMiscClient) FindOneRemoteSite(id string, withDeleted bool) (*moc.RemoteSite, error) {
	var remoteSite *moc.RemoteSite

	remoteSiteData, err := remoteSiteDb.Get(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: RemoteSiteObjectType,
			Obj: &db.GoObject{
				ID: id,
			},
		},
		Ctxt: remoteSiteDb.ctxt,
		Time: &db.TimeKeeper{},
	})

	if err == nil {
		remoteSites := make([]*moc.RemoteSite, 0)

		for _, remoteSiteDatum := range remoteSiteData {
			if mocRemoteSite, ok := remoteSiteDatum.Contents.Data.(*moc.RemoteSite); ok && (mocRemoteSite.Deleted == false || withDeleted == true) {
				remoteSites = append(remoteSites, mocRemoteSite)
			}
		}

		if len(remoteSites) > 0 {
			remoteSite = remoteSites[0]
		} else {
			err = db.ErrorNotFound
		}
	}

	return remoteSite, err
}

func (remoteSiteDb *MongoMiscClient) FindOneRemoteSiteByCscode(cscode string) (*moc.RemoteSite, error) {
	var remoteSite *moc.RemoteSite

	remoteSiteData, err := remoteSiteDb.Get(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: RemoteSiteObjectType,
			Obj: &db.GoObject{
				Data: &moc.RemoteSite{
					Cscode: cscode,
				},
			},
		},
		Ctxt: remoteSiteDb.ctxt,
		Time: &db.TimeKeeper{},
	})

	if err == nil {
		remoteSites := make([]*moc.RemoteSite, 0)

		for _, remoteSiteDatum := range remoteSiteData {
			if mocRemoteSite, ok := remoteSiteDatum.Contents.Data.(*moc.RemoteSite); ok {
				remoteSites = append(remoteSites, mocRemoteSite)
			}
		}

		if len(remoteSites) > 0 {
			remoteSite = remoteSites[0]
		} else {
			err = db.ErrorNotFound
		}
	}

	return remoteSite, err
}

func (remoteSiteDb *MongoMiscClient) DeleteRemoteSite(id string) error {
	query := bson.M{
		"_id": bson.ObjectIdHex(id),
	}

	update := bson.M{
		"$set": bson.M{
			"me.deleted": true,
		},
	}

	err := remoteSiteDb.dbconn.DB().C(CollectionRemoteSite).Update(query, update)

	return err
}
