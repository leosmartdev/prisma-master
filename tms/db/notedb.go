package db

import (
	"prisma/tms/client_api"
	"prisma/tms/moc"
)

type NoteDB interface {
	MiscDB
	CreateNote(note *moc.IncidentLogEntry) (*client_api.UpsertResponse, error)
	FindOneNote(noteId string, isAssigned string, withDeleted bool) (*moc.IncidentLogEntry, error)
	FindAllNotes() ([]GoGetResponse, error)
	UpdateNote(noteId string, note *moc.IncidentLogEntry) (*client_api.UpsertResponse, error)
	RestoreNote(noteId string) error
	DeleteNote(noteId string) error
}
