package feature

import (
	"math"
	"prisma/tms"
	"prisma/tms/devices"
	"prisma/tms/geojson"
	"prisma/tms/util/ais"
	"prisma/tms/util/coordsys"
)

func FromTrack(t *tms.Track, full bool, useTrackID bool) *F {
	if t == nil || len(t.Targets) == 0 {
		return nil
	}
	f := trackFeature(t, full, useTrackID)
	if f == nil {
		return nil
	}
	f.Properties["databaseId"] = t.DatabaseId

	if t.Metadata != nil && len(t.Metadata) > 0 {
		md := t.Metadata[0]
		mdFeature(f, md, full)
	}
	return f
}

func trackFeature(track *tms.Track, full bool, useTrackId bool) *F {
	target := track.Targets[0]
	var geom geojson.Object

	if target.Positions != nil {
		geom = target.Points()
	} else if target.Position != nil {
		geom = target.Point()
	}

	f := &F{}
	f.Geometry = geom

	pr := map[string]interface{}{
		"trackId":    track.Id,
		"registryId": track.RegistryId,
	}
	f.Properties = pr

	tyStr, ok := devices.DeviceType_name[int32(target.Type)]
	if ok {
		pr["type"] = "track:" + tyStr
	} else {
		pr["type"] = "track:unknown"
	}

	if !full {
		ptarget := make(map[string]interface{})
		if target.Course != nil {
			ptarget["course"] = target.Course.Value
		}
		if target.Heading != nil {
			ptarget["heading"] = target.Heading.Value
		}
		if target.Speed != nil {
			ptarget["speed"] = target.Speed.Value
		}
		if target.RateOfTurn != nil {
			ptarget["rateOfTurn"] = target.RateOfTurn.Value
		}
		if target.Time != nil {
			ptarget["time"] = tms.FromTimestamp(target.Time)
		}
		if target.Nmea != nil && target.Nmea.Vdm != nil && target.Nmea.Vdm.M1371 != nil {
			ptarget["mmsi"] = ais.FormatMMSI(int(target.Nmea.Vdm.M1371.Mmsi))
			if target.Nmea.Vdm.M1371.Pos != nil {
				ptarget["navigationalStatus"] = target.Nmea.Vdm.M1371.Pos.NavigationalStatus
			}
		}
		if target.Nmea != nil && target.Nmea.Ttm != nil {
			ptarget["number"] = target.Nmea.Ttm.Number
		}
		pr["target"] = ptarget
	} else {
		pr["target"] = target
	}

	if useTrackId {
		if len(track.RegistryId) > 0 {
			f.ID = track.RegistryId
		} else {
			f.ID = track.Id
		}
	} else {
		f.ID = target.Id.StringID()
	}

	// Everything in the database is stored in WGS-84
	pr["crs"] = coordsys.WGS84.EPSG

	return sanitizeFeature(f)
}

func mdFeature(f *F, md *tms.TrackMetadata, full bool) {
	if md == nil {
		f.Properties["metadata"] = make(map[string]interface{})
		return
	}
	if !full {
		props := make(map[string]interface{})
		props["name"] = md.Name
		if md.Nmea != nil && md.Nmea.Vdm != nil && md.Nmea.Vdm.M1371 != nil &&
			md.Nmea.Vdm.M1371.StaticVoyage != nil {
			props["shipAndCargoType"] = md.Nmea.Vdm.M1371.StaticVoyage.ShipAndCargoType
		}
		f.Properties["metadata"] = props
	} else {
		f.Properties["metadata"] = md
	}
}

func badNum(n float64) bool {
	return math.IsInf(n, 1) ||
		math.IsInf(n, -1) ||
		math.IsNaN(n)
}

func sanitizeFeature(f *F) *F {
	if f == nil {
		return nil
	}
	for k, v := range f.Properties {
		switch x := v.(type) {
		case float64:
			if badNum(float64(x)) {
				delete(f.Properties, k)
			}
		case float32:
			if badNum(float64(x)) {
				delete(f.Properties, k)
			}
		}
	}
	return f
}
