package mongo

import (
	"prisma/gogroup"
	"prisma/tms/client_api"
	"prisma/tms/db"
	"prisma/tms/moc"
	"unsafe"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// NoteObjectType ...
const NoteObjectType = "prisma.tms.moc.IncidentLogEntry"

// CollectionNote mongo collection
const CollectionNote = "notes"

type NoteDb struct {
	group  gogroup.GoGroup
	miscDb db.MiscDB
}

func NewMongoNoteDb(misc db.MiscDB) db.NoteDB {
	client, ok := misc.(*MongoMiscClient)
	if !ok {
		return nil
	}
	return &MongoMiscClient{
		dbconn: client.dbconn,
		ctxt:   client.ctxt,
	}
}

func (noteDb *MongoMiscClient) CreateNote(note *moc.IncidentLogEntry) (*client_api.UpsertResponse, error) {
	res, err := noteDb.Upsert(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: NoteObjectType,
			Obj: &db.GoObject{
				ID:   note.Id,
				Data: note,
			},
		},
		Ctxt: noteDb.ctxt,
		Time: &db.TimeKeeper{},
	})

	return res, err
}

func (noteDb *MongoMiscClient) FindOneNote(noteId string, isAssigned string, withDeleted bool) (*moc.IncidentLogEntry, error) {
	var note *moc.IncidentLogEntry
	var err error

	if isAssigned == "true" {
		// find a logEntry from Incident collection
		req := db.GoRequest{
			ObjectType: IncidentObjectType,
		}
		sd, ti, err := noteDb.resolveTable(&req)
		if err != nil {
			return nil, err
		}
		raw := []bson.Raw{}
		c := Coder{TypeData: sd}

		var query interface{}
		if withDeleted == true {
			query = []bson.M{
				{
					"$match": bson.M{
						"me.log.id": noteId,
					},
				},
				{
					"$addFields": bson.M{
						"me": bson.M{
							"log": bson.M{
								"$filter": bson.M{
									"input": "$me.log",
									"as":    "log",
									"cond": bson.M{
										"$and": []interface{}{
											bson.M{"$eq": []interface{}{"$$log.id", noteId}},
										},
									},
								},
							},
						},
					},
				},
			}
		} else {
			query = []bson.M{
				{
					"$match": bson.M{
						"me.log.id":      noteId,
						"me.log.deleted": false,
					},
				},
				{
					"$addFields": bson.M{
						"me": bson.M{
							"log": bson.M{
								"$filter": bson.M{
									"input": "$me.log",
									"as":    "log",
									"cond": bson.M{
										"$and": []interface{}{
											bson.M{"$eq": []interface{}{"$$log.id", noteId}},
											bson.M{"$eq": []interface{}{"$$log.deleted", false}},
										},
									},
								},
							},
						},
					},
				},
			}
		}
		err = noteDb.dbconn.DB().C(ti.Name).Pipe(query).All(&raw)

		if err != mgo.ErrNotFound {
			for _, data := range raw {
				var obj DBMiscObject
				c.DecodeTo(data, unsafe.Pointer(&obj))
				incident, ok := obj.Obj.(*moc.Incident)
				if !ok {
					return nil, db.ErrorNotFound
				}
				if len(incident.Log) > 0 {
					note = incident.Log[0]
				}
			}
		} else {
			return nil, db.ErrorNotFound
		}
	} else if isAssigned == "false" {
		// get a note from Notes table
		req := db.GoRequest{
			ObjectType: NoteObjectType,
		}
		sd, ti, err := noteDb.resolveTable(&req)
		if err != nil {
			return nil, err
		}
		raw := bson.Raw{}
		c := Coder{TypeData: sd}

		var query map[string]interface{}
		if withDeleted == true {
			query = bson.M{
				"_id": bson.ObjectIdHex(noteId),
			}
		} else {
			query = bson.M{
				"_id":        bson.ObjectIdHex(noteId),
				"me.deleted": false,
			}
		}

		err = noteDb.dbconn.DB().C(ti.Name).Find(query).One(&raw)
		if err != mgo.ErrNotFound {
			var obj DBMiscObject
			c.DecodeTo(raw, unsafe.Pointer(&obj))
			res, ok := obj.Obj.(*moc.IncidentLogEntry)
			if !ok {
				return nil, db.ErrorNotFound
			}
			note = res
			note.Id = noteId
		} else {
			return nil, db.ErrorNotFound
		}
	}

	if err == nil && note == nil {
		return nil, db.ErrorNotFound
	}

	return note, err
}

func (noteDb *MongoMiscClient) FindAllNotes() ([]db.GoGetResponse, error) {
	res, err := noteDb.Get(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: NoteObjectType,
		},
		Ctxt: noteDb.ctxt,
		Time: &db.TimeKeeper{},
	})

	return res, err
}

func (noteDb *MongoMiscClient) UpdateNote(noteId string, note *moc.IncidentLogEntry) (*client_api.UpsertResponse, error) {
	res, err := noteDb.Upsert(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: NoteObjectType,
			Obj: &db.GoObject{
				ID:   noteId,
				Data: note,
			},
		},
		Ctxt: noteDb.ctxt,
	})

	return res, err
}

func (noteDb *MongoMiscClient) RestoreNote(noteId string) error {
	query := bson.M{
		"_id": bson.ObjectIdHex(noteId),
	}

	update := bson.M{
		"$set": bson.M{
			"me.deleted": false,
		},
	}

	err := noteDb.dbconn.DB().C(CollectionNote).Update(query, update)
	return err
}

func (noteDb *MongoMiscClient) DeleteNote(noteId string) error {
	query := bson.M{
		"_id": bson.ObjectIdHex(noteId),
	}

	update := bson.M{
		"$set": bson.M{
			"me.deleted": true,
		},
	}

	err := noteDb.dbconn.DB().C(CollectionNote).Update(query, update)
	return err
}
