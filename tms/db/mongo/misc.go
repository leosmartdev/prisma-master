package mongo

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"
	"unsafe"

	"prisma/gogroup"
	. "prisma/tms"
	. "prisma/tms/client_api"
	. "prisma/tms/db"
	"prisma/tms/geo"
	"prisma/tms/log"
	"prisma/tms/moc"

	"github.com/globalsign/mgo/bson"
	"github.com/golang/protobuf/jsonpb"
	pb "github.com/golang/protobuf/proto"
)

var (
	CouldNotFindTable      = errors.New("Could not locate table")
	CouldNotIdentify       = errors.New("Could not figure out identifier")
	CouldNotFound          = errors.New("Not found")
	TimeRetrieveConnection = 1 * time.Second
	timeReplayDefault      = 15 * time.Minute
	miscStreamChanSize     = 512
)

func NewMongoMiscData(ctxt gogroup.GoGroup, dbconn *MongoClient) MiscDB {
	c := &MongoMiscClient{
		dbconn: dbconn,
		ctxt:   ctxt,
	}
	return c
}

func CreateId() string {
	return bson.NewObjectId().Hex()
}

func GetId(hexId string) bson.ObjectId {
	return bson.ObjectIdHex(hexId)
}

type MongoMiscClient struct {
	dbconn *MongoClient
	ctxt   gogroup.GoGroup
}

func (c *MongoMiscClient) get(req GoMiscRequest, live bool) (<-chan GoGetResponse, error) {
	// Fix the request, if necessary
	PopulateMiscRequest(&req)

	// Get table info
	sd, ti, err := c.resolveTable(req.Req)
	if err != nil {
		return nil, err
	}

	dbsel := bson.D{}
	if req.Req.Obj != nil {

		if req.Req.Obj.Data != nil {
			obj := UnmarshalObject(ti.Type, req.Req.Obj)
			BsonStructFlattenInPlace(&dbsel, "me.",
				EncodeToMap(obj).Map())
		}

		if req.Req.Obj.ID != "" {
			if !bson.IsObjectIdHex(req.Req.Obj.ID) {
				return nil, errors.New("InvalidId")
			}
			dbsel = append(dbsel, bson.DocElem{
				Name:  "_id",
				Value: bson.ObjectIdHex(req.Req.Obj.ID),
			})
		}

		if req.Req.Obj.RegistryId != "" {
			dbsel = append(dbsel, bson.DocElem{
				Name:  "registry_id",
				Value: req.Req.Obj.RegistryId,
			})
		}
	}

	qnow := req.Time.Now()
	var dbselInit bson.D
	if !req.Req.IgnoreTime {
		dbselInit = make(bson.D, len(dbsel)+2)
		copy(dbselInit[2:], dbsel)
		dbselInit[0] = bson.DocElem{
			Name:  "etime",
			Value: bson.M{"$gte": qnow},
		}
		dbselInit[1] = bson.DocElem{
			Name:  "ctime",
			Value: bson.M{"$lte": qnow},
		}
	} else {
		dbselInit = dbsel
	}

	if live && !req.Req.ReplayTime.IsZero() {
		// If we're replaying, we need to stream out future-created objects, so
		// ditch the "ctime before now" filter
		dbselInit = dbselInit[0 : len(dbselInit)-1]
	}
	ch := make(chan GoGetResponse, 128)
	servicer := QueryServicer{
		Name:   ti.Name,
		DBConn: c.dbconn,
		Coll:   ti.Name,
		Queries: []SimpleQuery{
			SimpleQuery{
				Filter: dbselInit,
			},
			SimpleQuery{
				Filter: dbsel,
			},
		},
		Ctxt: req.Ctxt,
		ObjectCallback: func(raw bson.Raw) bson.ObjectId {
			c := Coder{TypeData: sd}
			var obj DBMiscObject
			c.DecodeTo(raw, unsafe.Pointer(&obj))
			now := time.Now()
			status := Status_Current
			if now.After(obj.ExpirationTime) {
				status = Status_Timeout
			}
			goresp := GoGetResponse{
				Status: status,
				Table:  ti.Name,
				Contents: &GoObject{
					ID:             obj.Id.Hex(),
					CreationTime:   obj.CreationTime,
					ExpirationTime: obj.ExpirationTime,
					RegistryId:     obj.RegistryId,
					Data:           obj.Obj,
				},
			}
			select {
			case ch <- goresp:
				// Done. Good
			case <-req.Ctxt.Done():
				// We're closing up. Just throw this away
				return c.LastID
			}

			return c.LastID
		},
		StatusCallback: func(s Status) {
			switch s {
			case Status_Closing:
				close(ch)
			case Status_InitialLoadDone:
				if live {
					select {
					case ch <- GoGetResponse{
						Status: Status_InitialLoadDone,
					}:
					case <-req.Ctxt.Done():
						// We're closing up. Just throw this away
					}
				}
			}
		},
	}
	if err := servicer.Service(); err != nil {
		return nil, err
	}

	return ch, nil
}

func (c *MongoMiscClient) Get(req GoMiscRequest) ([]GoGetResponse, error) {
	ch, err := c.get(req, false)
	if err != nil {
		return nil, err
	}

	ret := make([]GoGetResponse, 0, 64)

	for resp := range ch {
		ret = append(ret, resp)
	}

	return ret, nil
}

func (c *MongoMiscClient) sendRawGetResponse(ctx context.Context, informer interface{}, tableName string,
	sd *StructData, ch chan<- GoGetResponse) {
	switch data := informer.(type) {
	case bson.Raw:
		coder := Coder{TypeData: sd}
		var obj DBMiscObject
		coder.DecodeTo(data, unsafe.Pointer(&obj))
		now := time.Now()
		status := Status_Current
		if now.After(obj.ExpirationTime) {
			status = Status_Timeout
		}
		resp := GoGetResponse{
			Status: status,
			Table:  tableName,
			Contents: &GoObject{
				ID:             obj.Id.Hex(),
				CreationTime:   obj.CreationTime,
				ExpirationTime: obj.ExpirationTime,
				RegistryId:     obj.RegistryId,
				Data:           obj.Obj,
			},
		}
		select {
		case ch <- resp:
		case <-ctx.Done():
			return
		}
	case Status:
		log.Info("table %v - %v", tableName, Status_name[int32(data)])
	default:
		log.Info("table %v - data is not supported: %v", tableName, data)
	}
}

// watch runs a goroutine to stream for a collection and returns a channel for updates
func (c *MongoMiscClient) watch(req GoMiscRequest, permanent bool, replayFilters bson.M, pipeline []bson.M) chan GoGetResponse {
	sd, ti, err := c.resolveTable(req.Req)
	if err != nil {
		log.Error(err.Error())
		return nil
	}

	ch := make(chan GoGetResponse, miscStreamChanSize)
	s := NewStream(c.ctxt.Child("streamer/misc/"+ti.Name), c.dbconn, ti.Name)
	c.ctxt.Go(func() {
		defer close(ch)
		s.Watch(func(ctx context.Context, informer interface{}) {
			c.sendRawGetResponse(ctx, informer, ti.Name, sd, ch)
		}, permanent,
			NewReplay(c.dbconn, timeReplayDefault, ti.Name, "utime", replayFilters), pipeline)
	})
	return ch
}

func (c *MongoMiscClient) GetStream(request GoMiscRequest, replayFilters bson.M, pipeline []bson.M) (<-chan GoGetResponse, error) {
	ch := c.watch(request, false, replayFilters, pipeline)
	return ch, nil
}

func (c *MongoMiscClient) GetPersistentStream(request GoMiscRequest, replayFilters bson.M, pipeline []bson.M) <-chan GoGetResponse {
	ch := c.watch(request, true, replayFilters, pipeline)
	ch <- GoGetResponse{
		Status: Status_InitialLoadDone,
	}
	return ch
}

func (c *MongoMiscClient) Upsert(req GoMiscRequest) (*UpsertResponse, error) {
	if req.Req.Obj == nil || req.Req.Obj.Data == nil {
		log.Warn("Could not complete Upsert() request: obj is nil")
		return nil, BadArguments
	}

	// If we are sending OutgoingMessage, right now don't need to store that in the database
	if req.Req.ObjectType == "prisma.tms.OutgoingMessage" {
		// TODO: This needs fixing. Place outside of here.
		//return c.processOutgoingMessage(req)
		resp := &UpsertResponse{
			Status: Status_Queued,
			Id:     fmt.Sprint("FIXME"),
		}
		return resp, nil
	}

	sd, ti, err := c.resolveTable(req.Req)
	if err != nil {
		return nil, err
	}
	var id bson.ObjectId
	if req.Req.Obj.ID != "" {
		id = bson.ObjectIdHex(req.Req.Obj.ID)
	} else {
		id = bson.NewObjectId()
	}

	unmarshalled := UnmarshalObject(ti.Type, req.Req.Obj)
	dbobj := DBMiscObject{
		Id:         id,
		Obj:        unmarshalled,
		UpdateTime: time.Now(),
	}

	// Make sure the request object is valid
	valid := true
	switch ti.Inst.(type) {
	case moc.Zone:
		log.Debug("Validating the Zone")
		valid = validate(unmarshalled.(*moc.Zone))
	}

	if !valid {
		log.Warn("Could not complete Upsert() request: Invalid object")
		return nil, BadArguments
	}

	if req.Req.Obj.RegistryId != "" {
		dbobj.RegistryId = req.Req.Obj.RegistryId
	}
	dbsel := bson.M{
		"_id": id,
	}

	if req.Req.Obj != nil {
		obj := req.Req.Obj
		if !obj.CreationTime.IsZero() {
			dbobj.CreationTime = obj.CreationTime
		} else {
			dbobj.CreationTime = time.Now()
		}
		if !obj.ExpirationTime.IsZero() {
			dbobj.ExpirationTime = obj.ExpirationTime
		} else {
			// By default, expire way out there
			dbobj.ExpirationTime = time.Date(2200, 1, 1, 0, 0, 0, 0, time.UTC)
		}
	}

	db := c.dbconn.DB()
	defer c.dbconn.Release(db)
	edbobj := sd.Encode(unsafe.Pointer(&dbobj))
	var mdbobj bson.M
	bson.Unmarshal(edbobj.Data, &mdbobj)
	_, err = db.C(ti.Name).Upsert(dbsel, edbobj)
	if err != nil {
		return nil, err
	}

	resp := UpsertResponse{
		Status: Status_Current,
		Table:  ti.Name,
		Id:     id.Hex(),
	}
	return &resp, nil
}

func (c *MongoMiscClient) Delete(req GoMiscRequest) error {
	if req.Req.ObjectType == "prisma.tms.moc.Fleet" {
		response := c.DeleteFleet(&DeleteFleetRequest{
			Id: req.Req.Obj.ID,
		})
		if response.Found {
			return nil
		}
	}
	return CouldNotFound
}

func (c *MongoMiscClient) Expire(req GoMiscRequest) (*ExpireResponse, error) {
	err := c.validateLoginInfo(req)
	if err != nil {
		return nil, err
	}

	_, ti, err := c.resolveTable(req.Req)
	if err != nil {
		return nil, err
	}
	if !req.Req.ReplayTime.IsZero() {
		return nil, errors.New("ReplayTime doesn't work with Expire()")
	}

	dbsel := bson.D{
		bson.DocElem{
			Name:  "etime",
			Value: bson.M{"$gte": time.Now()},
		},
	}
	if !req.Req.ExpireAll {
		if req.Req.Obj == nil {
			return nil, errors.New("Must specify 'obj' if not expiring all")
		}

		if req.Req.Obj.Data != nil {
			obj := UnmarshalObject(ti.Type, req.Req.Obj)
			BsonStructFlattenInPlace(&dbsel, "me.",
				EncodeToMap(obj).Map())
		}

		if req.Req.Obj.ID != "" {
			dbsel = append(dbsel, bson.DocElem{
				Name:  "_id",
				Value: bson.ObjectIdHex(req.Req.Obj.ID),
			})
		}
	} else {
		// If ExpireAll, we don't need anything else
	}

	// Get a DB connection
	db := c.dbconn.DB()
	defer c.dbconn.Release(db)

	// Get a list of IDs to be expired
	idIter := db.C(ti.Name).Find(dbsel).Select(bson.M{"_id": 1}).Iter()
	dbObj := bson.M{}
	ids := make([]bson.ObjectId, 0, 32)
	for idIter.Next(dbObj) {
		id, ok := dbObj["_id"]
		if !ok {
			log.Warn("Could not get _id from misc object: %v", dbObj, idIter)
		} else {
			oid, ok := id.(bson.ObjectId)
			if !ok {
				log.Warn("_id was not an ObjectId!")
			} else {
				ids = append(ids, oid)
			}
		}
		dbObj = bson.M{}
	}

	// Expire them and insert the update record into the _live table
	newExp := time.Now()
	setObj := bson.M{
		"$set": bson.M{
			"etime": newExp,
		},
	}
	resp := ExpireResponse{
		ExpirationTime: ToTimestamp(newExp),
		NumObjs:        0,
	}
	for _, id := range ids {
		err := db.C(ti.Name).UpdateId(id, setObj)
		if err != nil {
			log.Error("Error expiring %v: %v", id, err)
		} else {
			resp.NumObjs += 1
		}
	}

	return &resp, nil
}

func (c *MongoMiscClient) resolveTable(req *GoRequest) (*StructData, *TableInfo, error) {
	if req.ObjectType != "" {
		ti := MongoTables.TableFromTypeName(req.ObjectType)
		if ti.Name == "" {
			log.Warn("Could not find table for type: %v", req.ObjectType, req)
			return nil, nil, errors.New(fmt.Sprintf("Could not locate for type: %v", req.ObjectType))
		}
		return c.getStructData(ti), &ti, nil
	}

	if req.Obj != nil && req.Obj.Data != nil {
		ti := MongoTables.TableFromType(req.Obj.Data)
		if ti.Name == "" {
			log.Warn("Could not find table for object: %v", req.Obj.Data, req)
			return nil, nil, errors.New(fmt.Sprintf("Could not find table for object: %v", req.Obj.Data))
		}
		return c.getStructData(ti), &ti, nil
	}

	log.Warn("No table specifier found in request: %v", req)
	return nil, nil, CouldNotFindTable
}

func (c *MongoMiscClient) getStructData(ti TableInfo) *StructData {
	sd := NewStructData(
		reflect.TypeOf(DBMiscObject{}),
		func(fname string) reflect.Type {
			if fname == "Obj" {
				return ti.Type
			}
			panic(fmt.Sprintf("Unknown inferface<->type mapping for %v!",
				fname))
		})
	return sd
}

func (c *MongoMiscClient) validateLoginInfo(req GoMiscRequest) error {
	// TODO change this to use security service
	return nil
}

// Convert sub-map fields to "field1.subfield" accessors. Used to make MongoDB
// queries which don't do exactly field matches
func BsonStructFlatten(m bson.D) bson.D {
	ret := make(bson.D, 0, len(m))
	BsonStructFlattenInPlace(&ret, "", m.Map())
	return ret
}

func BsonStructFlattenInPlace(ret *bson.D, prefix string, m bson.M) {
	for k, v := range m {
		switch x := v.(type) {
		case bson.M:
			BsonStructFlattenInPlace(ret, prefix+k+".", x)
		case bson.D:
			BsonStructFlattenInPlaceD(ret, prefix+k+".", x)
		default:
			*ret = append(*ret, bson.DocElem{
				Name:  prefix + k,
				Value: v,
			})
		}
	}
}

func BsonStructFlattenInPlaceD(ret *bson.D, prefix string, m bson.D) {
	for _, d := range m {
		k := d.Name
		v := d.Value
		switch x := v.(type) {
		case bson.M:
			BsonStructFlattenInPlace(ret, prefix+k+".", x)
		case bson.D:
			BsonStructFlattenInPlaceD(ret, prefix+k+".", x)
		default:
			*ret = append(*ret, bson.DocElem{
				Name:  prefix + k,
				Value: v,
			})
		}
	}
}

func UnmarshalObject(tgt reflect.Type, obj *GoObject) interface{} {
	ty := reflect.TypeOf(obj.Data)
	if ty.Kind() == reflect.Ptr {
		ty = ty.Elem()
	}
	if ty == tgt {
		return obj.Data
	}

	if obj.JsonData != nil {
		jsonBytes := *obj.JsonData
		dst := reflect.New(tgt).Interface()
		msg, ok := dst.(pb.Message)
		if ok {
			err := jsonpb.UnmarshalString(string(jsonBytes), msg)
			if err != nil {
				panic(err)
			}
			return msg
		}

		panic(fmt.Sprintf("I was expecting a protobuf message type, but not getting it: %v, %v", tgt, dst))
	}

	panic("Request object type didn't match expected and I couldn't " +
		"figure out how to unmarshal object correctly!")
}

func validate(zone *moc.Zone) bool {
	// For now, a valid zone is any zone with proper lat-lon values.
	lineStrings := zone.Poly.Lines
	valid := true
	for _, lineString := range lineStrings {
		for _, point := range lineString.Points {
			latitude := point.Latitude
			longitude := point.Longitude
			valid = valid &&
				latitude >= geo.Latitude_Min && latitude <= geo.Latitude_Max &&
				longitude >= geo.Longitude_Min && longitude <= geo.Longitude_Max
		}
	}

	if valid {
		// Only one of entry or exit alerts should be set on the zone
		valid = !(zone.CreateAlertOnEnter == true && zone.CreateAlertOnExit == true)
	}

	return valid
}

func (c *MongoMiscClient) DeleteFleet(req *DeleteFleetRequest) *DeleteFleetResponse {
	db := c.dbconn.DB()
	defer c.dbconn.Release(db)
	id := bson.ObjectIdHex(req.Id)
	info, err := db.C("fleet").RemoveAll(bson.M{"_id": id})
	if err != nil {
		log.Error("Unable to remove fleet: %v", err)
		return nil
	}
	response := &DeleteFleetResponse{Found: info.Removed > 0}
	if !response.Found {
		return response
	}

	// Unassign any vessels in the registry
	info, err = db.C("registry").UpdateAll(
		bson.M{"fleet_id": id},
		bson.M{"$unset": bson.M{"fleet_id": ""}},
	)
	if err != nil {
		log.Error("Unable to update registry: %v", err)
		return nil
	}
	response.Orphaned = uint32(info.Updated)
	return response
}
