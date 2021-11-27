package public

import (
	"net/http"

	"prisma/gogroup"
	"prisma/tms"
	"prisma/tms/db"
	"prisma/tms/db/mongo"
	"prisma/tms/envelope"
	"prisma/tms/incident"
	"prisma/tms/log"
	"prisma/tms/moc"
	"prisma/tms/rest"
	"prisma/tms/security"
	"prisma/tms/ws"

	"github.com/go-openapi/spec"
	restful "github.com/orolia/go-restful"
	"golang.org/x/net/context"
)

const (
	NOTE_CLASSID = "IncidentLogEntry"
	OBJECT_NOTE  = "prisma.tms.moc." + NOTE_CLASSID
)

var (
	// incidentRest
	incidentRest *IncidentRest

	// LogEntry Types for the notes
	NoteTypes = [...]string{
		"NOTE",
		"NOTE_FILE",
	}

	// parameters with schema
	PARAMETER_NOTE_ID = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "note-id",
			In:       "path",
			Required: true,
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					MinLength: &[]int64{24}[0],
					MaxLength: &[]int64{24}[0],
					Pattern:   "[0-9a-fA-F]{24}",
				},
			},
		},
		SimpleSchema: spec.SimpleSchema{
			Type:   "string",
			Format: "hexadecimal",
		},
	}

	PARAMETER_IS_ASSIGNED = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "is-assigned",
			In:       "path",
			Required: true,
		},
		SimpleSchema: spec.SimpleSchema{
			Type:    "boolean",
			Default: "false",
		},
	}

	// schema
	SCHEMA_NOTE = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Required: []string{"type"},
			Properties: map[string]spec.Schema{
				"type": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{2}[0],
						MaxLength: &[]int64{255}[0],
					},
				},
				"id": {
					SwaggerSchemaProps: spec.SwaggerSchemaProps{
						ReadOnly: true,
					},
				},
				"timestamp": {
					SwaggerSchemaProps: spec.SwaggerSchemaProps{
						ReadOnly: true,
					},
				},
				"note": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{2}[0],
						MaxLength: &[]int64{2000}[0],
					},
				},
				"attachment": {
					SwaggerSchemaProps: spec.SwaggerSchemaProps{
						ReadOnly: true,
					},
				},
			},
		},
	}
)

type NoteRest struct {
	client     *mongo.MongoClient
	noteDb     db.NoteDB
	incidentDb db.IncidentDB
	group      gogroup.GoGroup
	publisher  *ws.Publisher
}

func NewNoteRest(group gogroup.GoGroup, client *mongo.MongoClient, idPrefixer incident.IdPrefixer, publisher *ws.Publisher) *NoteRest {
	incidentRest = NewIncidentRest(group, client, idPrefixer, publisher)
	miscDb := mongo.NewMongoMiscData(group, client)

	return &NoteRest{
		client:     client,
		noteDb:     mongo.NewMongoNoteDb(miscDb),
		incidentDb: mongo.NewMongoIncidentMiscData(mongo.NewMongoMiscData(group, client)),
		group:      group,
		publisher:  publisher,
	}
}

func (noteRest *NoteRest) Create(request *restful.Request, response *restful.Response) {
	ACTION := moc.IncidentLogEntry_CREATE.String()
	ctxt := request.Request.Context()
	if security.HasPermissionForAction(ctxt, NOTE_CLASSID, ACTION) {
		noteRequest := new(moc.IncidentLogEntry)

		err := rest.SanitizeValidateReadEntity(request, SCHEMA_NOTE, noteRequest)
		if err == nil {
			// set id and timestamp
			noteRequest.Id = mongo.CreateId()
			noteRequest.Timestamp = tms.Now()

			// if attachment then get metadata
			if nil != noteRequest.Attachment {
				ACTION = moc.IncidentLogEntry_ADD_FILE.String()
				err := populateFileMetadata(noteRest.client, noteRequest.Attachment)
				if err != nil {
					security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_VALIDATION)
					response.WriteError(http.StatusBadRequest, err)
					return
				}
			}

			upsertResponse, err := noteRest.noteDb.CreateNote(noteRequest)

			if err == nil {
				noteRequest.Id = upsertResponse.Id

				AuditIncidentLogEntry(ctxt, noteRequest, ACTION, security.SUCCESS)
				rest.WriteHeaderAndEntitySafely(response, http.StatusCreated, noteRequest)
				noteRest.Publish(ACTION, noteRequest)
			} else {
				log.Error("unexpected error: %v", err)
				security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_ERROR)
				response.WriteError(http.StatusInternalServerError, err)
			}
		} else {
			security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_VALIDATION)
			response.WriteHeaderAndEntity(http.StatusBadRequest, err)
		}
	} else {
		security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}

func (noteRest *NoteRest) ReadAll(request *restful.Request, response *restful.Response) {
	ACTION := moc.IncidentLogEntry_READ.String()
	ctxt := request.Request.Context()
	if security.HasPermissionForAction(ctxt, NOTE_CLASSID, ACTION) {
		resNotes := make([]*moc.IncidentLogEntry, 0)

		// Read assigned notes
		incidentData, err := noteRest.incidentDb.FindAllIncidents()

		if err == nil {
			incidents := make([]*moc.Incident, 0)
			for _, incidentDatum := range incidentData {
				if mocIncident, ok := incidentDatum.Contents.Data.(*moc.Incident); ok {
					mocIncident.Id = incidentDatum.Contents.ID
					incidents = append(incidents, mocIncident)
				}
			}

			// filter response
			for _, incident := range incidents {
				filterResponseLogEntry(incident)

				for _, logEntry := range incident.Log {
					logEntry.Assigned = true
					isNote := false

					for _, noteType := range NoteTypes {
						if logEntry.Type == noteType {
							isNote = true
							break
						}
					}

					if isNote == false {
						continue
					}

					resNotes = append(resNotes, logEntry)
				}
			}

			// Read unassigned notes
			noteData, err := noteRest.noteDb.FindAllNotes()

			if err == nil {
				notes := make([]*moc.IncidentLogEntry, 0)
				for _, noteDatum := range noteData {
					if mocNote, ok := noteDatum.Contents.Data.(*moc.IncidentLogEntry); ok {
						mocNote.Id = noteDatum.Contents.ID
						notes = append(notes, mocNote)
					}
				}

				notes = filterNotes(notes)

				for _, note := range notes {
					note.Assigned = false

					resNotes = append(resNotes, note)
				}

				security.Audit(ctxt, NOTE_CLASSID, ACTION, security.SUCCESS)
				rest.WriteEntitySafely(response, resNotes)
			} else {
				log.Error("unexpected error: %v", err)
				security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_ERROR)
				response.WriteError(http.StatusInternalServerError, err)
			}
		} else {
			log.Error("unexpected error: %v", err)
			security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_ERROR)
			response.WriteError(http.StatusInternalServerError, err)
		}
	} else {
		security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}

func (noteRest *NoteRest) ReadOne(request *restful.Request, response *restful.Response) {
	ACTION := moc.IncidentLogEntry_READ.String()
	ctxt := request.Request.Context()
	if security.HasPermissionForAction(ctxt, NOTE_CLASSID, ACTION) {
		noteId, errs := rest.SanitizeValidatePathParameter(request, PARAMETER_NOTE_ID)
		isAssigned, err := rest.SanitizeValidatePathParameter(request, PARAMETER_IS_ASSIGNED)
		errs = append(errs, err...)
		if errs == nil {
			note, err := noteRest.noteDb.FindOneNote(noteId, isAssigned, false)
			if err == nil {
				var incident *moc.Incident

				if isAssigned == "true" {
					incident, err = noteRest.incidentDb.FindIncidentByLogEntry(noteId, false)
				} else if isAssigned == "false" {
					incident, err = nil, nil
				}

				if err == nil {
					res := new(moc.IncidentLogEntryResponse)
					res.Note = note
					res.Incident = incident

					security.Audit(ctxt, NOTE_CLASSID, ACTION, security.SUCCESS)
					rest.WriteEntitySafely(response, res)
				} else {
					if db.ErrorNotFound == err {
						security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_NOTFOUND)
						response.WriteError(http.StatusNotFound, err)
					} else {
						log.Error("unexpected error: %v", err)
						security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_ERROR)
						response.WriteError(http.StatusInternalServerError, err)
					}
				}
			} else {
				if db.ErrorNotFound == err {
					security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_NOTFOUND)
					response.WriteError(http.StatusNotFound, err)
				} else {
					log.Error("unexpected error: %v", err)
					security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_ERROR)
					response.WriteError(http.StatusInternalServerError, err)
				}
			}
		} else {
			security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_VALIDATION, noteId)
			response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		}
	} else {
		security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}

func (noteRest *NoteRest) Update(request *restful.Request, response *restful.Response) {
	ACTION := moc.IncidentLogEntry_UPDATE.String()
	ctxt := request.Request.Context()
	if security.HasPermissionForAction(ctxt, NOTE_CLASSID, ACTION) {
		noteRequest := new(moc.IncidentLogEntry)

		noteId, errs := rest.SanitizeValidatePathParameter(request, PARAMETER_NOTE_ID)
		err := rest.SanitizeValidateReadEntity(request, SCHEMA_NOTE, noteRequest)
		errs = append(errs, err...)
		if errs == nil {
			// add timestamp
			noteRequest.Timestamp = tms.Now()
			// if attachment then get metadata
			if nil != noteRequest.Attachment {
				err := populateFileMetadata(noteRest.client, noteRequest.Attachment)
				if err != nil {
					security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_VALIDATION)
					response.WriteError(http.StatusBadRequest, err)
					return
				}
			}

			_, err := noteRest.noteDb.UpdateNote(noteId, noteRequest)
			if err == nil {
				noteRequest.Id = noteId
				AuditIncidentLogEntry(ctxt, noteRequest, ACTION, security.SUCCESS)
				rest.WriteEntitySafely(response, noteRequest)
				noteRest.Publish(ACTION, noteRequest)
			} else {
				log.Error("unexpected error: %v", err)
				security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_ERROR, noteRequest.Id)
				response.WriteError(http.StatusInternalServerError, err)
			}
		} else {
			security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_VALIDATION)
			response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		}
	} else {
		security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}

func (noteRest *NoteRest) Delete(request *restful.Request, response *restful.Response) {
	ACTION := moc.IncidentLogEntry_DELETE.String()
	ctxt := request.Request.Context()
	if security.HasPermissionForAction(ctxt, NOTE_CLASSID, ACTION) {
		noteId, errs := rest.SanitizeValidatePathParameter(request, PARAMETER_NOTE_ID)
		if errs == nil {
			err := noteRest.noteDb.DeleteNote(noteId)

			if err == nil {
				note, err := noteRest.noteDb.FindOneNote(noteId, "false", true)
				if err == nil {
					security.Audit(ctxt, NOTE_CLASSID, ACTION, security.SUCCESS)
					rest.WriteEntitySafely(response, note)
				} else {
					if db.ErrorNotFound == err {
						security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_NOTFOUND)
						response.WriteError(http.StatusNotFound, err)
					} else {
						log.Error("unexpected error: %v", err)
						security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_ERROR)
						response.WriteError(http.StatusInternalServerError, err)
					}
				}
			} else {
				log.Error("unexpected error: %v", err)
				security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_ERROR, noteId)
				response.WriteError(http.StatusInternalServerError, err)
			}
		} else {
			security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_VALIDATION, noteId)
			response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		}
	} else {
		security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}

// Delete from Note collection, Add to Incident's log
func (noteRest *NoteRest) Assign(request *restful.Request, response *restful.Response) {
	ACTION := moc.IncidentLogEntry_ASSIGN.String()
	ctxt := request.Request.Context()
	if security.HasPermissionForAction(ctxt, NOTE_CLASSID, ACTION) {
		noteId, errs := rest.SanitizeValidatePathParameter(request, PARAMETER_NOTE_ID)
		incidentId, err := rest.SanitizeValidatePathParameter(request, PARAMETER_INCIDENT_ID)
		errs = append(errs, err...)
		if errs == nil {
			note, err := noteRest.noteDb.FindOneNote(noteId, "false", false)

			clientDb := noteRest.client.DB()
			defer noteRest.client.Release(clientDb)
			// if the note is unassigned, simply remove it from Note collection
			// then add it to incident's log list
			if err == nil {
				err := noteRest.noteDb.DeleteNote(noteId)
				if err == nil {
					incident, err := getIncident(incidentRest, incidentId)
					if err == nil {
						// add timestamp
						note.Timestamp = tms.Now()
						// set assigned flag as true
						note.Assigned = true
						// if attachment then get metadata
						if nil != note.Attachment {
							err := populateFileMetadata(noteRest.client, note.Attachment)
							if err != nil {
								security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_VALIDATION)
								response.WriteError(http.StatusBadRequest, err)
								return
							}
						}
						// check if a note exists in incident's log list
						var err error
						isExist := false
						for _, logEntry := range incident.Log {
							if noteId == logEntry.Id {
								isExist = true
								break
							}
						}

						if isExist == false {
							// if not exist, add it to the list
							incident.Log = append(incident.Log, note)
							_, err = noteRest.incidentDb.UpdateIncident(incidentId, incident)
						} else {
							// if exist, set the incident's logEntry's deleted flag as false
							err = noteRest.incidentDb.RestoreIncidentLogEntry(ctxt, incidentId, noteId)
						}

						if err == nil {
							incident, err = getIncident(incidentRest, incidentId)
							filterResponseLogEntry(incident)
							AuditIncident(ctxt, incident, moc.Incident_ADD_NOTE.String(), security.SUCCESS, note)
							rest.WriteEntitySafely(response, incident)
						} else {
							log.Error("unexpected error: %v", err)
							security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_ERROR, incident.Id)
							response.WriteError(http.StatusInternalServerError, err)
						}
					} else {
						if db.ErrorNotFound == err {
							security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_NOTFOUND)
							response.WriteError(http.StatusNotFound, err)
						} else {
							log.Error("unexpected error: %v", err)
							security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_ERROR)
							response.WriteError(http.StatusInternalServerError, err)
						}
					}
				} else {
					log.Error("unexpected error: %v", err)
					security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_ERROR, noteId)
					response.WriteError(http.StatusInternalServerError, err)
				}
			} else {
				if db.ErrorNotFound == err {
					// if the note is assigned, remove it from legacy incident's log list
					// then add it to target incident's log list
					note, err := noteRest.noteDb.FindOneNote(noteId, "true", false)
					if err == nil {
						incident, err := noteRest.incidentDb.FindIncidentByLogEntry(noteId, false)
						if err == nil {
							// if target incident is just current incident, pass
							if incident.Id == incidentId {
								filterResponseLogEntry(incident)
								AuditIncident(ctxt, incident, moc.Incident_ADD_NOTE.String(), security.SUCCESS, note)
								rest.WriteEntitySafely(response, incident)
							} else {
								// delete from legacy incident's log list
								err = noteRest.incidentDb.DeleteIncidentLogEntry(ctxt, incident.Id, noteId)
								if err == nil {
									incident, err := getIncident(incidentRest, incidentId)
									if err == nil {
										// add timestamp
										note.Timestamp = tms.Now()
										// set assigned flag as true
										note.Assigned = true
										// if attachment then get metadata
										if nil != note.Attachment {
											err := populateFileMetadata(noteRest.client, note.Attachment)
											if err != nil {
												security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_VALIDATION)
												response.WriteError(http.StatusBadRequest, err)
												return
											}
										}
										// check if a note exists in incident's log list
										var err error
										isExist := false
										for _, logEntry := range incident.Log {
											if noteId == logEntry.Id {
												isExist = true
												break
											}
										}

										if isExist == false {
											// if not exist, add it to the list
											incident.Log = append(incident.Log, note)
											_, err = noteRest.incidentDb.UpdateIncident(incidentId, incident)
										} else {
											// if exist, set the incident's logEntry's deleted flag as false
											noteRest.incidentDb.RestoreIncidentLogEntry(ctxt, incidentId, noteId)
										}

										if err == nil {
											incident, err = getIncident(incidentRest, incidentId)
											filterResponseLogEntry(incident)
											AuditIncident(ctxt, incident, moc.Incident_ADD_NOTE.String(), security.SUCCESS, note)
											rest.WriteEntitySafely(response, incident)
										} else {
											log.Error("unexpected error: %v", err)
											security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_ERROR, incident.Id)
											response.WriteError(http.StatusInternalServerError, err)
										}
									} else {
										if db.ErrorNotFound == err {
											security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_NOTFOUND)
											response.WriteError(http.StatusNotFound, err)
										} else {
											log.Error("unexpected error: %v", err)
											security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_ERROR)
											response.WriteError(http.StatusInternalServerError, err)
										}
									}
								} else {
									log.Error("unexpected error: %v", err)
									security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_ERROR, incident.Id)
									response.WriteError(http.StatusInternalServerError, err)
								}
							}
						} else {
							if db.ErrorNotFound == err {
								security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_NOTFOUND)
								response.WriteError(http.StatusNotFound, err)
							} else {
								log.Error("unexpected error: %v", err)
								security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_ERROR)
								response.WriteError(http.StatusInternalServerError, err)
							}
						}
					} else {
						if db.ErrorNotFound == err {
							security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_NOTFOUND)
							response.WriteError(http.StatusNotFound, err)
						} else {
							log.Error("unexpected error: %v", err)
							security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_ERROR)
							response.WriteError(http.StatusInternalServerError, err)
						}
					}
				} else {
					log.Error("unexpected error: %v", err)
					security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_ERROR)
					response.WriteError(http.StatusInternalServerError, err)
				}
			}
		} else {
			security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_VALIDATION, noteId)
			response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
		}
	} else {
		security.Audit(ctxt, NOTE_CLASSID, ACTION, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}

func (noteRest *NoteRest) Publish(action string, note *moc.IncidentLogEntry) {
	envelope := envelope.Envelope{
		Type:     NOTE_CLASSID + "/" + action,
		Contents: &envelope.Envelope_Note{Note: note},
	}
	noteRest.publisher.Publish(NOTE_CLASSID, envelope)
}

func filterNotes(notes []*moc.IncidentLogEntry) []*moc.IncidentLogEntry {
	// filter deleted note
	if len(notes) > 0 {
		tmp := make([]*moc.IncidentLogEntry, 0)
		for _, note := range notes {
			if !note.Deleted {
				tmp = append(tmp, note)
			}
		}

		return tmp
	}

	return notes
}

func AuditIncidentLogEntry(context context.Context, note *moc.IncidentLogEntry, action string, outcome string, payload ...interface{}) {
	security.AuditUserObject(context, NOTE_CLASSID, note.Id, "", action, outcome, payload...)
}
