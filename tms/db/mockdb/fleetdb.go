package mockdb

import (
	"context"
	"strings"

	"prisma/tms/db"
	"prisma/tms/moc"
	"prisma/tms/rest"

	"github.com/hashicorp/go-memdb"
	"github.com/pborman/uuid"
)

const tableFleet = "fleet"

var (
	schema = &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"fleet": {
				Name: "fleet",
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

type FleetDb struct {
	memDb *memdb.MemDB
}

func NewFleetDb(_ context.Context) db.FleetDB {
	memDb, err := memdb.NewMemDB(schema)
	if err != nil {
		panic(err)
	}
	return &FleetDb{
		memDb: memDb,
	}
}

func (d *FleetDb) Create(ctxt context.Context, fleet *moc.Fleet) (*moc.Fleet, error) {
	txn := d.memDb.Txn(true)
	fleet.Id = strings.Replace(uuid.New(), "-", "", -1)[:24] // like mongo _id
	err := txn.Insert(tableFleet, fleet)
	if err == nil {
		txn.Commit()
	} else {
		txn.Abort()
	}
	return fleet, err
}

func (d *FleetDb) Update(ctxt context.Context, fleet *moc.Fleet) (*moc.Fleet, error) {
	txn := d.memDb.Txn(true)
	err := txn.Insert(tableFleet, fleet)
	if err == nil {
		txn.Commit()
	} else {
		txn.Abort()
	}
	return fleet, err
}

func (d *FleetDb) Delete(ctxt context.Context, fleetId string) error {
	txn := d.memDb.Txn(true)
	fleet, err := d.FindOne(ctxt, fleetId)
	if err == nil {
		err = txn.Delete(tableFleet, fleet)
		if err == nil {
			txn.Commit()
		} else {
			txn.Abort()
		}
	}
	return err
}

func (d *FleetDb) FindOne(ctxt context.Context, fleetId string) (*moc.Fleet, error) {
	var fleet *moc.Fleet
	txn := d.memDb.Txn(false)
	defer txn.Abort()
	result, err := txn.Get(tableFleet, "id", fleetId)
	if err == nil {
		for i := result.Next(); i != nil; i = result.Next() {
			raw, ok := i.(*moc.Fleet)
			if ok {
				fleet = raw
			} else {
				err = db.ErrorCritical
			}
		}
		if err == nil && fleet == nil {
			err = db.ErrorNotFound
		}
	}
	return fleet, err
}

func (d *FleetDb) FindAll(ctxt context.Context) ([]*moc.Fleet, error) {
	var fleets []*moc.Fleet
	var err error
	txn := d.memDb.Txn(false)
	defer txn.Abort()
	results, err := txn.Get(tableFleet, "id")
	if err == nil {
		for i := results.Next(); i != nil; i = results.Next() {
			raw, ok := i.(*moc.Fleet)
			if ok {
				fleets = append(fleets, raw)
			} else {
				err = db.ErrorCritical
			}
		}
	}
	return fleets, err
}

func (d *FleetDb) FindByMapByPagination(ctxt context.Context, searchMap map[string]string, pagination *rest.PaginationQuery) ([]*moc.Fleet, error) {
	var fleets []*moc.Fleet
	txn := d.memDb.Txn(false)
	defer txn.Abort()
	results, err := txn.Get(tableFleet, "id")
	if err == nil {
		count := 0
		for i := results.Next(); i != nil && count < pagination.Limit; i = results.Next() {
			raw, ok := i.(*moc.Fleet)
			if ok {
				fleets = append(fleets, raw)
				count += 1
			} else {
				err = db.ErrorCritical
			}
		}
	}
	return fleets, err
}

func (d *FleetDb) FindByMap(ctxt context.Context, searchMap map[string]string) ([]*moc.Fleet, error) {
	panic("not implemented")
}

func (d *FleetDb) RemoveVessel(ctxt context.Context, fleetId string, vesselId string) error {
	panic("not implemented")
}

func (d *FleetDb) AddVessel(ctxt context.Context, fleetId string, vessel *moc.Vessel) error {
	panic("not implemented")
}

func (d *FleetDb) UpdateVessel(ctxt context.Context, fleetId string, vessel *moc.Vessel) error {
	panic("not implemented")
}
