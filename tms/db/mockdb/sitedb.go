// Package mockdb provides mocks for db abstractions.
package mockdb

import (
	"context"
	"strings"

	"prisma/tms/db"
	"prisma/tms/moc"

	"github.com/hashicorp/go-memdb"
	"github.com/pborman/uuid"
)

const tableSite = "site"

var (
	schemaSite = &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"site": {
				Name: "site",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "Id"},
					},
					"name": {
						Name:    "name",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "Name"},
					},
				},
			},
		},
	}
)

type SiteDb struct {
	memDb *memdb.MemDB
}

func NewSiteDb(_ context.Context) db.SiteDB {
	memDb, err := memdb.NewMemDB(schemaSite)
	if err != nil {
		panic(err)
	}
	return &SiteDb{
		memDb: memDb,
	}
}

func (d *SiteDb) Create(ctxt context.Context, site *moc.Site) (*moc.Site, error) {
	txn := d.memDb.Txn(true)
	site.Id = strings.Replace(uuid.New(), "-", "", -1)[:24] // like mongo _id
	err := txn.Insert(tableSite, site)
	if err == nil {
		txn.Commit()
	} else {
		txn.Abort()
	}
	return site, err
}

func (d *SiteDb) FindAll(ctxt context.Context) ([]*moc.Site, error) {
	var sites []*moc.Site
	var err error
	txn := d.memDb.Txn(false)
	defer txn.Abort()
	results, err := txn.Get(tableSite, "id")
	if err == nil {
		for i := results.Next(); i != nil; i = results.Next() {
			raw, ok := i.(*moc.Site)
			if ok {
				sites = append(sites, raw)
			} else {
				err = db.ErrorCritical
			}
		}
	}
	return sites, err
}

func (d *SiteDb) Update(ctx context.Context, site *moc.Site) (*moc.Site, error) {
	panic("not implemented")
}
func (d *SiteDb) UpdateConnectionStatusBySiteId(ctx context.Context, site *moc.Site) (*moc.Site, error) {
	panic("not implemented")
}
func (d *SiteDb) Delete(ctx context.Context, siteId string) error {
	panic("not implemented")
}
func (d *SiteDb) FindByMap(ctx context.Context, searchMap map[string]string, sortFields db.SortFields) ([]*moc.Site, error) {
	panic("not implemented")
}
func (d *SiteDb) FindBySiteId(ctx context.Context, site *moc.Site) error {
	panic("not implemented")
}
func (d *SiteDb) FindById(ctx context.Context, siteId string) (*moc.Site, error) {
	panic("not implemented")
}
