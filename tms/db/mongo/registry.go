package mongo

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	"prisma/gogroup"
	"prisma/tms"
	"prisma/tms/client_api"
	tmsdb "prisma/tms/db"
	"prisma/tms/log"
	"prisma/tms/util/ident"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

var (
	registryEntrySD = NewStructData(reflect.TypeOf(DBRegistryEntry{}), NoMap)
	noRedirect      = bson.DocElem{
		Name: "$or",
		Value: []interface{}{
			bson.M{dbRegEntry_Redirect: ""},
			bson.M{dbRegEntry_Redirect: bson.M{"$exists": true}},
		},
	}
)

func NewMongoRegistry(ctxt gogroup.GoGroup, dbconn *MongoClient) tmsdb.RegistryDB {
	c := &MongoRegistryClient{
		dbconn: dbconn,
		ctxt:   ctxt,
	}
	miscdb := NewMongoMiscData(ctxt, dbconn)
	if miscdb != nil {
		c.associated = tmsdb.NewAssociatedDataProvider(miscdb)
	}
	return c
}

type MongoRegistryClient struct {
	dbconn     *MongoClient
	associated *tmsdb.AssociatedDataProvider
	ctxt       gogroup.GoGroup
}

func (c *MongoRegistryClient) Assign(req client_api.AssignRequest) (*client_api.AssignResponse, error) {
	var registryId string
	switch req.Type {
	case "imei":
		registryId = ident.With("imei", req.Lookup).Hash()
	case "OmnicomSolar":
		registryId = ident.With("OmnicomSolar", req.Lookup).Hash()
	default:
		return nil, fmt.Errorf("unknown type %v", req.Type)
	}

	r, err := c.GetOrCreate(registryId)
	if err != nil {
		return nil, err
	}

	// Add the registration entry
	if r.Assignment == nil {
		r.Assignment = &tms.Assignment{}
	}
	r.Assignment.Type = req.Type
	r.Assignment.Lookup = req.Lookup
	r.Assignment.InFleet = true
	if req.Label != "" {
		r.Assignment.Label = req.Label
	}
	if req.FleetId != "" {
		r.Assignment.FleetId = req.FleetId
	}

	err = c.Upsert(r)
	if err != nil {
		return nil, err
	}

	resp := &client_api.AssignResponse{
		RegistryId: registryId,
	}

	return resp, nil
}

func (c *MongoRegistryClient) Upsert(rt *tms.RegistryEntry) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	ti := MongoTables.TableFromType(DBRegistryEntry{})
	dbEntry := DBRegistryEntry{
		Entry: rt,
	}
	if rt.DatabaseId != "" {
		dbEntry.ID = bson.ObjectIdHex(rt.DatabaseId)
	} else {
		dbEntry.ID = bson.NewObjectId()
		rt.DatabaseId = dbEntry.ID.Hex()
	}
	if rt.Assignment == nil {
		rt.Assignment = &tms.Assignment{}
	}
	if rt.Assignment.FleetId != "" {
		dbEntry.FleetID = bson.ObjectIdHex(rt.Assignment.FleetId)
	}

	dbsel := bson.M{
		"_id": dbEntry.ID,
	}
	bsonrt := registryEntrySD.Encode(unsafe.Pointer(&dbEntry))

	db := c.dbconn.DB()
	defer c.dbconn.Release(db)
	log.TraceMsg("About to upsert: %v, %v", log.Spew(dbsel), log.Spew(dbEntry))
	ci, err := db.C(ti.Name).Upsert(dbsel, bsonrt)
	if err != nil {
		return err
	}
	log.TraceMsg("Upserted registry track: %v", ci)

	return nil
}

func (c *MongoRegistryClient) GetOrCreate(regId string) (*tms.RegistryEntry, error) {
	if regId == "" {
		return nil, errors.New("empty regId")
	}
	dbsel := bson.M{
		"me.registry_id": regId,
	}
	ti := MongoTables.TableFromType(DBRegistryEntry{})
	db := c.dbconn.DB()
	defer c.dbconn.Release(db)
	// Get
	iter := db.C(ti.Name).Find(dbsel).Iter()
	var raw bson.Raw
	if iter.Next(&raw) {
		var rt DBRegistryEntry
		registryEntrySD.DecodeTo(raw, unsafe.Pointer(&rt))
		// FIXME: It is awkward to insert a blank entry which gets updated
		// immedately anyway.
		return rt.Entry, nil //c.Upsert(rt.Entry)
	}
	if iter.Err() != nil {
		return nil, iter.Err()
	}
	// Create
	entry := &tms.RegistryEntry{
		RegistryId:     regId,
		Keywords:       []string{},
		TargetFields:   []*tms.SearchField{},
		MetadataFields: []*tms.SearchField{},
	}
	return entry, c.Upsert(entry)
}

func (c *MongoRegistryClient) Get(regID string) (*tms.RegistryEntry, error) {
	if regID == "" {
		return nil, errors.New("empty regId")
	}
	dbsel := bson.M{
		"me.registry_id": regID,
	}
	ti := MongoTables.TableFromType(DBRegistryEntry{})
	db := c.dbconn.DB()
	defer c.dbconn.Release(db)
	// Get
	var raw bson.Raw
	err := db.C(ti.Name).Find(dbsel).One(&raw)
	if err != nil {
		if err.Error() == "not found" {
			return nil, tmsdb.ErrorNotFound
		}
		return nil, err
	}
	var rt DBRegistryEntry
	registryEntrySD.DecodeTo(raw, unsafe.Pointer(&rt))
	return rt.Entry, nil
}

func (c *MongoRegistryClient) Query(dbsel bson.M, page tmsdb.Page) ([]*tms.RegistryEntry, error) {
	ti := MongoTables.TableFromType(DBRegistryEntry{})

	db := c.dbconn.DB()
	defer c.dbconn.Release(db)

	q := db.C(ti.Name).Find(dbsel).Sort("me.search_fields.name")
	if page.Number > 0 && page.Length > 0 {
		q.Skip((page.Number - 1) * page.Length)
	}
	if page.Length > 0 {
		q.Limit(page.Length)
	}
	iter := q.Iter()

	ret := make([]*tms.RegistryEntry, 0, 32)
	var raw bson.Raw
	for iter.Next(&raw) {
		var rt DBRegistryEntry
		registryEntrySD.DecodeTo(raw, unsafe.Pointer(&rt))
		if rt.Entry != nil {
			rt.Entry.DatabaseId = rt.ID.Hex()
		}
		ret = append(ret, rt.Entry)
	}
	if iter.Err() != nil {
		return nil, iter.Err()
	}

	return ret, nil
}

func (c *MongoRegistryClient) GetNear(lat float64, long float64, maxd float64) ([]*tms.RegistryEntry, error) {
	dbsel := bson.M{
		dbRegEntry_Position: bson.M{
			"$near": bson.M{
				"$geometry": bson.M{
					"type":        "Point",
					"coordinates": []float64{long, lat},
				},
				"$maxDistance": maxd,
			},
		},
		noRedirect.Name: noRedirect.Value,
	}

	return c.Query(dbsel, tmsdb.Page{})
}

func (c *MongoRegistryClient) GetStream(req tmsdb.RegistryRequest) (<-chan *tmsdb.GoRegistryEntry, error) {
	log.Debug("Stream request: %v", log.Spew(req))
	if req.GetAssociated && req.RegistryId == "" {
		return nil, errors.New("Cannot specify GetAssociated with a specific ID")
	}

	ti := MongoTables.TableFromType(DBRegistryEntry{})

	if req.RegistryId == "" {
		return nil, fmt.Errorf("No request id")
	}

	query := bson.D{{
		"me.registry_id", req.RegistryId,
	}}
	entries := make(chan *tmsdb.GoRegistryEntry, 64)
	servicer := QueryServicer{
		Name:   ti.Name,
		DBConn: c.dbconn,
		Coll:   ti.Name,
		Queries: []SimpleQuery{
			SimpleQuery{
				Filter: query,
			},
		},
		Ctxt:          req.Ctxt,
		NumDeliverers: DecodeThreads,
		ObjectCallback: func(raw bson.Raw) bson.ObjectId {
			if raw.Data != nil {
				c := Coder{TypeData: registryEntrySD}
				entry := new(DBRegistryEntry)
				c.DecodeTo(raw, unsafe.Pointer(entry))
				if entry.Entry != nil {
					entry.Entry.DatabaseId = entry.ID.Hex()
					ge := &tmsdb.GoRegistryEntry{
						RegistryEntry: *entry.Entry,
					}
					select {
					case entries <- ge:
					case <-req.Ctxt.Done():
						return c.LastID
					}
				}
				return c.LastID
			}
			return bson.ObjectId("")
		},
		StatusCallback: func(s client_api.Status) {
			if s == client_api.Status_Closing {
				close(entries)
			}
		},
	}
	err := servicer.Service()
	//servicer.Debug = true
	if err != nil {
		return nil, err
	}

	if !req.GetAssociated {
		log.TraceMsg("Returning entries without associated data")
		return entries, nil
	}

	asschan, err := c.associated.Get(tmsdb.AssociatedReq{
		Ctxt:       req.Ctxt,
		RegistryId: req.RegistryId,
	})
	if err != nil {
		req.Ctxt.Cancel(err)
		return nil, err
	}

	entriesWithAss := make(chan *tmsdb.GoRegistryEntry, 0)
	req.Ctxt.Go(func() {
		// OK -- So we gotta wait for entries and associated data and append
		// the associate data to the entiries. Also, don't send any entries
		// until the associate data's backlog is done. After the backlog is
		// done, send a new entry with the associated entries any time there's
		// new data and immediately after the backlog
		ass := make(map[string]*tmsdb.GoObject)
		defer close(entriesWithAss)

		var entry *tmsdb.GoRegistryEntry
		assBacklog := true
		sendEntry := func() {
			if entry != nil {
				entry.Associated = nil
				for _, obj := range ass {
					entry.Associated = append(entry.Associated, obj)
				}
				entriesWithAss <- entry
			}
		}
		var ok bool
		for {
			select {
			case entry, ok = <-entries:
				log.TraceMsg("Got entry: %v, %v, %v", assBacklog, ok, entry)
				if !ok {
					req.Ctxt.Cancel(nil)
					return
				}
				if !assBacklog {
					sendEntry()
				}
			case obj, ok := <-asschan:
				log.Debug("Got ass: %v, %v", ok, obj)
				if !ok {
					req.Ctxt.Cancel(nil)
					return
				}
				switch obj.Status {
				case client_api.Status_InitialLoadDone:
					assBacklog = false
					sendEntry()
				case client_api.Status_Timeout, client_api.Status_LeftGeoRange:
					if obj.Contents != nil {
						delete(ass, obj.Contents.ID)
					}
				case client_api.Status_Current:
					if obj.Contents != nil {
						ass[obj.Contents.ID] = obj.Contents
						if !assBacklog {
							sendEntry()
						}
					}
				}
			}
		}
	})

	return entriesWithAss, nil
}

func (c *MongoRegistryClient) Search(req tmsdb.RegistrySearchRequest) ([]*tmsdb.RegistrySearchResult, error) {
	log.Debug("Request is : %+v", req)
	dbsel := bson.M{}
	if len(req.Query) > 0 {
		pattern := "^" + req.Query
		regex := bson.RegEx{Pattern: pattern}
		dbsel["me.keywords"] = bson.M{"$regex": regex}
	}

	ti := MongoTables.TableFromType(DBRegistryEntry{})
	db := c.dbconn.DB()
	defer c.dbconn.Release(db)

	q := db.C(ti.Name).Find(dbsel).Sort("me.label")
	if req.Limit > 0 {
		q = q.Limit(req.Limit)
	}
	iter := q.Iter()
	results := make([]*tmsdb.RegistrySearchResult, 0)
	var raw bson.Raw
	for iter.Next(&raw) {
		var r DBRegistryEntry
		registryEntrySD.DecodeTo(raw, unsafe.Pointer(&r))
		results = append(results, &tmsdb.RegistrySearchResult{
			RegistryID: r.Entry.RegistryId,
			Label:      r.Entry.Label,
			LabelType:  r.Entry.LabelType,
			Matches:    tmsdb.RegistryMatches(req.Query, r.Entry),
		})
	}
	if iter.Err() != nil {
		return nil, iter.Err()
	}
	return results, nil
}

// Fleet specific stuff should be eventually moved in to the main search
// or this should be refactored
func (c *MongoRegistryClient) SearchV1(req tmsdb.RegistrySearchV1) ([]*tms.RegistryEntry, error) {
	log.Debug("Request is : %+v", req)
	dbsel := bson.M{}
	if len(req.Query) > 0 {
		pattern := strings.ToTitle(req.Query) + ".*"
		regex := bson.RegEx{Pattern: pattern}
		dbsel["me.search_fields.tags"] = bson.M{
			"$in": []bson.RegEx{regex},
		}
	}
	if len(req.SearchMap) > 0 {
		for k, v := range req.SearchMap {
			values := strings.Split(v, ",")
			if len(values) > 1 {
				orQuery := make([]bson.M, 0)
				for _, value := range values {
					orQuery = append(orQuery, bson.M{k: value})
				}
				dbsel["$or"] = orQuery
			} else {
				dbsel[k] = v
			}
		}
	}
	// FIXME: Fleet should be a type, this is a device
	if req.InFleet {
		dbsel["me.assignment.in_fleet"] = true
	}
	if len(req.FleetId) > 0 {
		dbsel["fleet_id"] = bson.ObjectIdHex(req.FleetId)
	}
	log.Debug("dbsel is: %+v", dbsel)
	return c.Query(dbsel, req.Page)
}

func (c *MongoRegistryClient) Unassign(req client_api.UnassignRequest) (*client_api.UnassignResponse, error) {
	db := c.dbconn.DB()
	defer c.dbconn.Release(db)
	info, err := db.C("registry").UpdateAll(
		bson.M{"me.registry_id": req.RegistryId},
		bson.M{"$unset": bson.M{"me.assignment": ""}},
	)
	if err != nil {
		return nil, err
	}
	ok := info.Updated > 0
	return &client_api.UnassignResponse{Ok: ok}, nil
}

func (c *MongoRegistryClient) AddToIncident(registryID string, incidentID string) (bool, error) {
	db := c.dbconn.DB()
	defer c.dbconn.Release(db)
	dbsel := bson.M{
		"me.registry_id": registryID,
	}
	dbupd := bson.M{
		"$addToSet": bson.M{
			"me.incidents": incidentID,
		},
	}
	err := db.C("registry").Update(dbsel, dbupd)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *MongoRegistryClient) RemoveFromIncident(registryID string, incidentID string) (bool, error) {
	db := c.dbconn.DB()
	defer c.dbconn.Release(db)
	dbsel := bson.M{
		"me.registry_id": registryID,
	}
	dbupd := bson.M{
		"$pull": bson.M{
			"me.incidents": incidentID,
		},
	}
	err := db.C("registry").Update(dbsel, dbupd)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *MongoRegistryClient) GetSit185Messages(startDateTime int, endDateTime int) ([]*tms.RegistryEntry, error) {
	db := c.dbconn.DB()
	defer c.dbconn.Release(db)

	query := bson.M{
		"me.target.type":                "SARSAT",
		"me.target.sarmsg.message_type": "SIT_185",
		"me.target.sarmsg.received":     true,
		"me.target.time.seconds": bson.M{
			"$gte": startDateTime,
			"$lt":  endDateTime,
		},
	}

	coder := Coder{TypeData: registryEntrySD}
	registries := make([]*tms.RegistryEntry, 0)
	raw := []bson.Raw{}

	err := db.C("registry").Find(query).All(&raw)

	if err == mgo.ErrNotFound {
		return registries, nil
	}

	for _, data := range raw {
		registry := new(DBRegistryEntry)

		coder.DecodeTo(data, unsafe.Pointer(registry))

		registries = append(registries, registry.Entry)
	}

	return registries, err
}
