package moc

import (
	"time"

	"prisma/tms"
)

// HasIncidentLogEntity checks log has an entity with id
func HasIncidentLogEntity(logs []*IncidentLogEntry, id string) (result bool) {
	var lastTime time.Time
	for _, incLog := range logs {
		if incLog.Type != "TRACK" {
			continue
		}
		if tms.FromTimestamp(incLog.Timestamp).After(lastTime) {
			lastTime = tms.FromTimestamp(incLog.Timestamp)
			if incLog.Entity.Id == id {
				result = !incLog.Deleted
			}
		}
	}
	return
}

