package mongo

import (
	"context"
	"fmt"
	"prisma/tms/client_api"
	"prisma/tms/db"
	"prisma/tms/moc"
	"unsafe"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// IncidentObjectType ...
const IncidentObjectType = "prisma.tms.moc.Incident"

// CollectionIncident mongo collection
const CollectionIncident = "incidents"

// NewMongoIncidentMiscData ...
func NewMongoIncidentMiscData(misc db.MiscDB) db.IncidentDB {
	client, ok := misc.(*MongoMiscClient)
	if !ok {
		return nil
	}
	return &MongoMiscClient{
		dbconn: client.dbconn,
		ctxt:   client.ctxt,
	}
}

func (i *MongoMiscClient) FindAllIncidents() ([]db.GoGetResponse, error) {
	res, err := i.Get(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: IncidentObjectType,
		},
		Ctxt: i.ctxt,
		Time: &db.TimeKeeper{},
	})

	return res, err
}

func (i *MongoMiscClient) FindIncidentByLogEntry(logEntryId string, withDeleted bool) (*moc.Incident, error) {
	var resIncident *moc.Incident

	incidentData, err := i.Get(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: IncidentObjectType,
		},
		Ctxt: i.ctxt,
		Time: &db.TimeKeeper{},
	})
	if err == nil {
		incidents := make([]*moc.Incident, 0)
		for _, incidentDatum := range incidentData {
			if mocIncident, ok := incidentDatum.Contents.Data.(*moc.Incident); ok {
				mocIncident.Id = incidentDatum.Contents.ID
				incidents = append(incidents, mocIncident)
			}
		}

		for _, incident := range incidents {
			isFind := false

			for _, logEntry := range incident.Log {
				if withDeleted == false && logEntry.Deleted == true {
					continue
				}

				if logEntry.Id == logEntryId {
					resIncident = incident
					isFind = true
					break
				}
			}

			if isFind == true {
				break
			}
		}
	}

	return resIncident, err
}

func (i *MongoMiscClient) UpdateIncident(incidentId string, incident *moc.Incident) (*client_api.UpsertResponse, error) {
	res, err := i.Upsert(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: IncidentObjectType,
			Obj: &db.GoObject{
				ID:   incidentId,
				Data: incident,
			},
		},
		Ctxt: i.ctxt,
	})

	return res, err
}

func (i *MongoMiscClient) RestoreIncidentLogEntry(ctxt context.Context, incidentId string, noteId string) error {
	session, err := getSession(ctxt)
	if err != nil {
		return err
	}
	defer session.Close()

	query := bson.M{
		"_id":       bson.ObjectIdHex(incidentId),
		"me.log.id": noteId,
	}
	update := bson.M{
		"$set": bson.M{
			"me.log.$.deleted": false,
		},
	}

	err = session.DB(DATABASE).C(CollectionIncident).Update(query, update)
	return err
}

func (i *MongoMiscClient) DeleteIncidentLogEntry(ctxt context.Context, incidentId string, noteId string) error {
	session, err := getSession(ctxt)
	if err != nil {
		return err
	}
	defer session.Close()

	query := bson.M{
		"_id":       bson.ObjectIdHex(incidentId),
		"me.log.id": noteId,
	}
	update := bson.M{
		"$set": bson.M{
			"me.log.$.deleted": true,
		},
	}

	err = session.DB(DATABASE).C(CollectionIncident).Update(query, update)
	return err
}

// GetIncidentWithTrackID returns all open incidents that have given track assigned to them.
func (i *MongoMiscClient) GetIncidentWithTrackID(trackID string) ([]*moc.Incident, error) {
	req := db.GoRequest{
		ObjectType: IncidentObjectType,
	}
	sd, ti, err := i.resolveTable(&req)
	if err != nil {
		return nil, err
	}
	query := bson.M{
		"me.state":           "Open",
		"me.log.entity.id":   trackID,
		"me.log.entity.type": "registry",
	}
	raw := []bson.Raw{}
	c := Coder{TypeData: sd}

	incidents := make([]*moc.Incident, 0)

	err = i.dbconn.DB().C(ti.Name).Find(query).All(&raw)
	if err != mgo.ErrNotFound {
		for _, data := range raw {
			var obj DBMiscObject
			c.DecodeTo(data, unsafe.Pointer(&obj))
			incident, ok := obj.Obj.(*moc.Incident)
			if !ok {
				return nil, fmt.Errorf("Could not recover Incident object")
			}
			incidents = append(incidents, incident)
		}
	}
	return incidents, nil
}

// GetIncidentWithMarkerID returns all open incidents that have given marker assigned to them.
func (i *MongoMiscClient) GetIncidentWithMarkerID(markerID string) ([]*moc.Incident, error) {
	req := db.GoRequest{
		ObjectType: IncidentObjectType,
	}
	sd, ti, err := i.resolveTable(&req)
	if err != nil {
		return nil, err
	}
	query := bson.M{
		"me.state": "Open",
		"me.log": bson.M{
			"$elemMatch": bson.M{
				"entity.id":   markerID,
				"entity.type": "marker",
				"deleted":     false,
			},
		},
	}
	raw := []bson.Raw{}
	c := Coder{TypeData: sd}

	incidents := make([]*moc.Incident, 0)

	err = i.dbconn.DB().C(ti.Name).Find(query).All(&raw)
	if err != mgo.ErrNotFound {
		for _, data := range raw {
			var obj DBMiscObject
			c.DecodeTo(data, unsafe.Pointer(&obj))
			incident, ok := obj.Obj.(*moc.Incident)
			if !ok {
				return nil, fmt.Errorf("Could not recover Incident object")
			}
			incident.Id = obj.Id.Hex()
			incidents = append(incidents, incident)
		}
	}
	return incidents, nil
}
