package db

import (
	. "prisma/tms"
	. "prisma/tms/client_api"
	"prisma/tms/log"
	. "prisma/tms/tmsg"

	"encoding/json"
	"time"

	"prisma/gogroup"

	"github.com/globalsign/mgo/bson"
	"github.com/golang/protobuf/jsonpb"
	pb "github.com/golang/protobuf/proto"
)

var MiscDataNotExpiredTime = time.Date(2200, 1, 1, 0, 0, 0, 0, time.UTC)

type MiscDB interface {
	Get(GoMiscRequest) ([]GoGetResponse, error)
	GetStream(request GoMiscRequest, replayFilters bson.M, pipeline []bson.M) (<-chan GoGetResponse, error)
	GetPersistentStream(request GoMiscRequest, replayFilters bson.M, pipeline []bson.M) <-chan GoGetResponse
	Upsert(GoMiscRequest) (*UpsertResponse, error)
	Expire(req GoMiscRequest) (*ExpireResponse, error)
	Delete(req GoMiscRequest) error
}

type GoGetResponse struct {
	Status   Status    `json:"status"`
	Table    string    `json:"table,omitempty"`
	Contents *GoObject `json:"contents"`
}

type GoObject struct {
	ID             string           `json:"id"`
	CreationTime   time.Time        `json:"creationTime,omitempty"`
	ExpirationTime time.Time        `json:"expirationTime,omitempty"`
	RegistryId     string           `json:"registryId,omitempty"`
	Data           interface{}      `json:"-"`
	JsonData       *json.RawMessage `json:"data"`
	//OwnerID        string           `json:"ownerID"`
}

type GoObjectNoJson GoObject

func (obj *GoObject) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		return nil
	}

	j := new(GoObjectNoJson)
	err := json.Unmarshal(b, j)
	if err != nil {
		return err
	}
	*obj = GoObject(*j)
	if obj.JsonData != nil && len(*obj.JsonData) > 0 {
		// We don't know in advance what type to unmarshal to, so we'll do
		// something generic, keep the JsonData field and hope somebody later
		// down the road can figure it out!
		err = json.Unmarshal([]byte(*obj.JsonData), &obj.Data)
	}
	log.Debug("GoObject.UnmarshalJSON: %v", log.Spew(obj))
	return err
}

func (obj *GoObject) MarshalJSON() ([]byte, error) {
	var jsonBytes []byte
	var err error
	if msg, ok := obj.Data.(pb.Message); ok {
		any, err := PackFrom(msg)
		if err != nil {
			return nil, err
		}
		jsonData, err := (&jsonpb.Marshaler{}).MarshalToString(any)
		if err != nil {
			return nil, err
		}
		jsonBytes = []byte(jsonData)
	} else {
		jsonBytes, err = json.Marshal(obj.Data)
		if err != nil {
			return nil, err
		}
	}
	obj.JsonData = new(json.RawMessage)
	*obj.JsonData = json.RawMessage(jsonBytes)
	log.Debug("GoObject.MarshalJSON: %v", log.Spew(obj))
	return json.Marshal(*obj)
}

func FromObject(obj *Object) *GoObject {
	if obj == nil {
		return nil
	}
	var msg pb.Message
	if obj.Data != nil {
		var err error
		msg, err = Unpack(obj.Data)
		if err != nil {
			log.Warn("Error unpacking request object: %v", err, obj)
			panic(err)
		}
	}
	return &GoObject{
		ID:             obj.Id,
		CreationTime:   FromTimestamp(obj.CreationTime),
		ExpirationTime: FromTimestamp(obj.ExpirationTime),
		RegistryId:     obj.RegistryId,
		Data:           msg,
	}
}

type GoMiscRequest struct {
	Req  *GoRequest
	Ctxt gogroup.GoGroup
	Time *TimeKeeper
}

func NewMiscRequest(req *GoRequest, ctxt gogroup.GoGroup) (*GoMiscRequest, error) {
	log.Debug("Request got %v", req)
	greq := GoMiscRequest{
		Req:  req,
		Ctxt: ctxt,
		Time: &TimeKeeper{},
	}
	PopulateMiscRequest(&greq)
	return &greq, nil
}

func PopulateMiscRequest(greq *GoMiscRequest) {
	req := greq.Req
	if greq.Time == nil {
		greq.Time = &TimeKeeper{}
	}

	tk := greq.Time
	if !req.ReplayTime.IsZero() {
		tk.Replay = true
		tk.ReplayTime = req.ReplayTime
		tk.StartTime = time.Now()
		tk.Speed = req.ReplaySpeed
		if tk.Speed == 0.0 {
			tk.Speed = 1.0
		}
	}
}

type GoRequest struct {
	ObjectType  string    `json:"objectType"`
	ReplayTime  time.Time `json:"replayTime,omitempty"`
	ReplaySpeed float64   `json:"replaySpeed,omitempty"`
	Obj         *GoObject `json:"obj,omitempty"`
	ExpireAll   bool      `json:"expireAll,omitempty"`
	IgnoreTime  bool      `json:"ignoreTime,omitempty"`
}

func FromRequest(req *Request) *GoRequest {
	if req == nil {
		return nil
	}

	return &GoRequest{
		ObjectType:  req.ObjectType,
		ReplayTime:  FromTimestamp(req.ReplayTime),
		ReplaySpeed: req.ReplaySpeed,
		Obj:         FromObject(req.Obj),
		ExpireAll:   req.ExpireAll,
		IgnoreTime:  req.IgnoreTime,
	}
}
