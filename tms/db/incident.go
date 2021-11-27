package db

import (
	"context"
	"prisma/tms/client_api"
	"prisma/tms/moc"
)

// IncidentDB is a composition of MiscDB and extra behavior specific to Incident
type IncidentDB interface {
	MiscDB
	GetIncidentWithTrackID(trackID string) ([]*moc.Incident, error)
	GetIncidentWithMarkerID(markerID string) ([]*moc.Incident, error)
	FindAllIncidents() ([]GoGetResponse, error)
	FindIncidentByLogEntry(logEntryId string, withDeleted bool) (*moc.Incident, error)
	UpdateIncident(incidentId string, incident *moc.Incident) (*client_api.UpsertResponse, error)
	RestoreIncidentLogEntry(ctxt context.Context, incidentId string, noteId string) error
	DeleteIncidentLogEntry(ctxt context.Context, incidentId string, noteId string) error
}
