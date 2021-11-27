package mongo

import (
	"context"
	"fmt"
	"strings"

	"prisma/gogroup"
	"prisma/tms"
	"prisma/tms/db"
	"prisma/tms/moc"
)

const objectIncident = "prisma.tms.moc.Incident"
const typeGeoJSON = "Feature"
const typeGeometry = "Point"

type SarmapDb struct {
	group      gogroup.GoGroup
	miscDb     db.MiscDB
	trackDb    db.TrackDB
	registryDb db.RegistryDB
}

// NewMongoSarmapDb initializes a SarmapDB struct.
func NewMongoSarmapDb(group gogroup.GoGroup, mongoClient *MongoClient) db.SarmapDB {
	return &SarmapDb{
		group:      group,
		miscDb:     NewMongoMiscData(group, mongoClient),
		trackDb:    NewMongoTracks(group, mongoClient),
		registryDb: NewMongoRegistry(group, mongoClient),
	}
}

// FindAll recovers all the targets assigned to an open incident
func (d *SarmapDb) FindAll(ctx context.Context) ([]*moc.GeoJsonFeaturePoint, error) {
	geoJsons := make([]*moc.GeoJsonFeaturePoint, 0)
	// get all incidents
	incidentData, err := d.miscDb.Get(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: objectIncident,
		},
		Ctxt: d.group,
		Time: &db.TimeKeeper{},
	})
	if err != nil {
		return nil, fmt.Errorf("SARMAP %s", err)
	}
	incidents := make([]*moc.Incident, 0)
	for _, v := range incidentData {
		if mocIncident, ok := v.Contents.Data.(*moc.Incident); ok {
			// only open incidents are going to be forwarded
			if mocIncident.State != moc.Incident_Open {
				continue
			}
			mocIncident.Id = v.Contents.ID
			incidents = append(incidents, mocIncident)
		}
	}
	// get all entities from incident log entries
	for _, incident := range incidents {
		// filter non-relevant log entries
		if len(incident.Log) > 0 {
			tmp := make([]*moc.IncidentLogEntry, 0)
			for _, logEntry := range incident.Log {
				if !logEntry.Deleted && logEntry.Entity != nil {
					tmp = append(tmp, logEntry)
				}
			}
			incident.Log = tmp
		}
		// process entity log entries
		for _, le := range incident.Log {
			if le.Entity.Type == "registry" {
				registry, err := d.registryDb.Get(le.Entity.Id)
				if err != nil {
					continue
				}
				target := registry.GetTarget()
				if target != nil {
					//SARSAT data contains targets with one position or multiple positions
					// The way sarmap intergration works does not allow us to GeoJSON with other geometries the point
					// Concequentely we split multi points to multiple GeoJSONFeaturePoint
					if target.GetPositions() != nil {
						for _, ps := range target.Positions {
							geoJsons = append(geoJsons, newGeoJSONFeaturePoint(target, incident, ps))
						}
					}
					if target.GetPosition() != nil {
						geoJsons = append(geoJsons, newGeoJSONFeaturePoint(target, incident, target.Position))
					}

				}
			}
		}
	}
	return geoJsons, nil
}

func newGeoJSONFeaturePoint(tgt *tms.Target, inc *moc.Incident, ps *tms.Point) *moc.GeoJsonFeaturePoint {
	return &moc.GeoJsonFeaturePoint{
		Type: typeGeoJSON,
		Properties: map[string]string{
			"label":      inc.GetName(),
			"deviceId":   tgt.LookupDevID(),
			"deviceType": strings.ToLower(tgt.GetType().String()),
		},
		Geometry: &moc.GeoJsonGeometryPoint{
			Type:        typeGeometry,
			Coordinates: []float64{ps.Longitude, ps.Latitude},
		},
	}
}
