package mongo

import (
	"context"

	"prisma/tms"
	"prisma/tms/db"
	"prisma/tms/log"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// CollectionTransmission mongo collection
const CollectionTransmission = "transmissions"

type TransmissionDb struct {
	ctx context.Context
	db  *MongoClient
}

func NewTransmissionDb(ctx context.Context, dbconn *MongoClient) *TransmissionDb {
	return &TransmissionDb{
		ctx: ctx,
		db:  dbconn,
	}
}

func (t *TransmissionDb) Create(tr *tms.Transmission) error {
	session, err := getSession(t.ctx)
	if err != nil {
		return err
	}
	defer session.Close()
	mongoId := bson.NewObjectId()
	tr.Id = mongoId.Hex()
	collection := session.DB(DATABASE).C(CollectionTransmission)
	_, err = collection.UpsertId(mongoId, tr)
	if mgo.IsDup(err) {
		err = db.ErrorDuplicate
	}
	return err
}

func (t *TransmissionDb) Update(tr *tms.Transmission) error {
	log.Debug("updating transmission %v with state %v", tr.MessageId, tr.State)
	session, err := getSession(t.ctx)
	if err != nil {
		return err
	}
	defer session.Close()
	collection := session.DB(DATABASE).C(CollectionTransmission)
	_, err = collection.UpsertId(bson.ObjectIdHex(tr.Id), tr)
	if mgo.IsDup(err) {
		err = db.ErrorDuplicate
	}
	return err
}

func (t *TransmissionDb) FindByID(id string) (*tms.Transmission, error) {
	var tr tms.Transmission
	var err error
	session, err := getSession(t.ctx)
	if err != nil {
		return &tr, err
	}
	defer session.Close()
	err = session.DB(DATABASE).C(CollectionTransmission).
		FindId(bson.ObjectIdHex(id)).One(&tr)
	if mgo.ErrNotFound == err {
		err = db.ErrorNotFound
	}
	return &tr, err
}

// Status updates the transmissions state and status code
func (t *TransmissionDb) StatusById(id string, state tms.Transmission_State, status int32) error {
	session := t.db.DB().Session
	defer session.Close()
	tid := bson.ObjectIdHex(id)
	collection := session.DB(DATABASE).C(CollectionTransmission)
	return collection.UpdateId(tid,
		bson.M{"$set": bson.M{
			"status": status,
			"state":  state,
		}})
}

// Status updates the transmissions state and status code
func (t *TransmissionDb) Status(messageId string, state tms.Transmission_State, status int32) error {
	log.Debug("updating transmission %v with state %v", messageId, state)
	var tr tms.Transmission
	session := t.db.DB().Session
	defer session.Close()
	collection := session.DB(DATABASE).C(CollectionTransmission)
	if err := collection.Find(bson.M{"messageid": messageId}).One(&tr); err != nil {
		return err
	}
	return t.StatusById(tr.Id, state, status)
}

func (t *TransmissionDb) ClearMessageId(messageId string) error {
	log.Debug("clearing transmission %v", messageId)
	session := t.db.DB().Session
	defer session.Close()
	collection := session.DB(DATABASE).C(CollectionTransmission)
	err := collection.Update(
		bson.M{
			"messageid": messageId,
		},
		bson.M{
			"$set": bson.M{
				"messageid": "",
			},
		},
	)
	return err
}

// PacketStatus multiple Packet to one Transmission
func (t *TransmissionDb) PacketStatus(id string, packet *tms.Packet) error {
	session := t.db.DB().Session
	defer session.Close()
	collection := session.DB(DATABASE).C(CollectionTransmission)
	err := collection.Update(
		bson.M{
			"_id":               bson.ObjectIdHex(id),
			"packets.messageid": packet.MessageId,
		},
		bson.M{
			"$set": bson.M{
				"packets.$": packet,
			},
		},
	)
	return err
}

// PacketStatusSingle is supposing we have one packet / transmission for now
func (t *TransmissionDb) PacketStatusSingle(requestId, state string, status int32) (*tms.Transmission, error) {
	log.Debug("updating packet %v with state %v", requestId, state)
	session := t.db.DB().Session
	defer session.Close()
	var tr tms.Transmission
	collection := session.DB(DATABASE).C(CollectionTransmission)
	if err := collection.Find(bson.M{"messageid": requestId}).One(&tr); err != nil {
		return nil, err
	}
	// state
	if len(tr.Packets) > 0 {
		tr.Packets[0].State = tms.Transmission_State(tms.Transmission_State_value[state])
	}
	tr.State = tms.Transmission_State(tms.Transmission_State_value[state])
	err := collection.UpdateId(bson.ObjectIdHex(tr.Id),
		bson.M{"$set": bson.M{"packets.0.state": state}})
	if err != nil {
		return nil, err
	}
	// status
	if len(tr.Packets) > 0 {
		tr.Packets[0].Status = &tms.ResponseStatus{
			Code:    status,
			Message: "",
		}
	}
	tr.Status = &tms.ResponseStatus{
		Code:    status,
		Message: "",
	}
	err = collection.UpdateId(bson.ObjectIdHex(tr.Id),
		bson.M{"$set": bson.M{"packets.0.status.code": tr.Status}})
	return &tr, err
}
