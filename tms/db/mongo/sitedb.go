package mongo

import (
	"context"

	"prisma/tms/db"
	"prisma/tms/moc"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const (
	collectionSite = "sites"
)

type SiteDb struct {
	ctx context.Context
}

func NewSiteDb(ctx context.Context) db.SiteDB {
	return &SiteDb{
		ctx: ctx,
	}
}

func (d *SiteDb) Create(ctx context.Context, site *moc.Site) (*moc.Site, error) {
	session, err := getSession(ctx)
	if err != nil {
		return nil, err
	}
	defer session.Close()
	mongoId := bson.NewObjectId()
	site.Id = mongoId.Hex()
	_, err = session.DB(DATABASE).C(collectionSite).UpsertId(mongoId, site)
	if mgo.IsDup(err) {
		err = db.ErrorDuplicate
	}
	return site, err
}

func (d *SiteDb) FindAll(ctx context.Context) ([]*moc.Site, error) {
	session, err := getSession(ctx)
	if err != nil {
		return nil, err
	}
	defer session.Close()
	var sites []*moc.Site
	err = session.DB(DATABASE).C(collectionSite).Find(nil).All(&sites)
	if mgo.ErrNotFound == err {
		err = db.ErrorNotFound
	}
	return sites, err
}

func (d *SiteDb) Update(ctx context.Context, site *moc.Site) (*moc.Site, error) {
	if !bson.IsObjectIdHex(site.Id) {
		return nil, db.ErrorNotFound
	}
	session, err := getSession(ctx)
	if err != nil {
		return nil, err
	}
	defer session.Close()
	_, err = session.DB(DATABASE).C(collectionSite).UpsertId(bson.ObjectIdHex(site.Id), site)
	if mgo.IsDup(err) {
		err = db.ErrorDuplicate
	}
	return site, err
}

func (d *SiteDb) FindBySiteId(ctx context.Context, site *moc.Site) error {
	query := map[string]uint32{
		"siteid": site.SiteId,
	}
	session, err := getSession(ctx)
	if err != nil {
		return err
	}
	defer session.Close()
	err = session.DB(DATABASE).C(collectionSite).Find(query).One(site)
	if mgo.ErrNotFound == err {
		err = db.ErrorNotFound
	}
	return err
}

func (d *SiteDb) FindById(ctx context.Context, siteId string) (*moc.Site, error) {
	if !bson.IsObjectIdHex(siteId) {
		return nil, db.ErrorBadID
	}
	session, err := getSession(ctx)
	if err != nil {
		return nil, err
	}
	defer session.Close()
	site := moc.Site{}
	err = session.DB(DATABASE).C(collectionSite).FindId(bson.ObjectIdHex(siteId)).One(&site)
	if mgo.ErrNotFound == err {
		err = db.ErrorNotFound
	}
	return &site, err
}

func (d *SiteDb) UpdateConnectionStatusBySiteId(ctx context.Context, site *moc.Site) (*moc.Site, error) {
	query := map[string]uint32{
		"siteid": site.SiteId,
	}
	session, err := getSession(ctx)
	if err != nil {
		return nil, err
	}
	defer session.Close()
	err = session.DB(DATABASE).C(collectionSite).Update(query, bson.M{"$set": bson.M{"connectionstatus": site.ConnectionStatus}})
	if mgo.IsDup(err) {
		err = db.ErrorDuplicate
	}
	return site, err
}

func (d *SiteDb) Delete(ctx context.Context, siteId string) error {
	session, err := getSession(ctx)
	if err != nil {
		return err
	}
	defer session.Close()
	collection := session.DB(DATABASE).C(collectionSite)
	err = collection.Remove(bson.D{{Name: "_id", Value: bson.ObjectIdHex(siteId)}})
	if mgo.ErrNotFound == err {
		err = db.ErrorNotFound
	}
	return err
}

func (d *SiteDb) FindByMap(ctx context.Context, searchMap map[string]string, sortFields db.SortFields) ([]*moc.Site, error) {
	query := createMongoQueryFromMap(searchMap)
	session, err := getSession(ctx)
	if err != nil {
		return nil, err
	}
	defer session.Close()
	pipe := []bson.M{
		{"$match": query},
		{"$sort": createMongoSort(sortFields)},
	}
	var sites []*moc.Site
	err = session.DB(DATABASE).C(collectionSite).Pipe(pipe).All(&sites)
	if mgo.ErrNotFound == err {
		err = db.ErrorNotFound
	}
	return sites, err
}
