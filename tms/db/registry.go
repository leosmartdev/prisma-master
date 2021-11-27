package db

import (
	"prisma/tms"
	. "prisma/tms/client_api"
	. "prisma/tms/devices"
	"prisma/tms/util/ais"
	"strconv"
	"strings"

	"prisma/gogroup"
)

type RegistryDB interface {
	// ----- For use by Internal API only, do not use on the client side

	// Insert/update the given RegistryEntry in the database.
	Upsert(*tms.RegistryEntry) error

	// Returns the RegistryEntry from the database matching the given track id. This method is guaranteed to return a
	// non-nil RegistryEntry. If none is found, creates a blank RegistryEntry with the given track id
	// (does NOT store it in the database) and returns it.
	// FIXME: It previously didn't store it in the database, now it appears to do so
	GetOrCreate(trackId string) (*tms.RegistryEntry, error)

	Get(registryID string) (*tms.RegistryEntry, error)

	// ------ Following methods can be used anywhere in the code

	// See all registry updates
	GetStream(RegistryRequest) (<-chan *GoRegistryEntry, error)

	// Search through the registry
	Search(RegistrySearchRequest) ([]*RegistrySearchResult, error)
	SearchV1(RegistrySearchV1) ([]*tms.RegistryEntry, error)

	// Get all the registered vessels last seen within 'maxd' meters of lat,long
	GetNear(lat float64, long float64, maxd float64) ([]*tms.RegistryEntry, error)

	// Add new vessel registration entry
	Assign(AssignRequest) (*AssignResponse, error)

	AddToIncident(regitryId string, incidentId string) (bool, error)
	RemoveFromIncident(registryId string, incidentId string) (bool, error)

	GetSit185Messages(startDateTime int, endDateTime int) ([]*tms.RegistryEntry, error)
}

// FIXME: include in newer search or refactor into a separate one
type RegistrySearchV1 struct {
	Ctxt      gogroup.GoGroup
	Query     string `json:"query"`
	Page      Page
	InFleet   bool
	FleetId   string
	SearchMap map[string]string
}

type RegistrySearchRequest struct {
	Query string
	Limit int
}

type RegistrySearchResult struct {
	RegistryID string            `json:"registryId"`
	Label      string            `json:"label"`
	LabelType  string            `json:"labelType"`
	Matches    []*tms.MatchField `json:"matches"`
}

type RegistryRequest struct {
	Ctxt gogroup.GoGroup

	// Just get this one registry entry
	RegistryId string

	// Get misc data associated with also?
	GetAssociated bool

	// Filtering criteria:
	Type DeviceType
}

type GoRegistryEntry struct {
	tms.RegistryEntry

	Associated []interface{}
}

func RegistryTargetSearchFields(t *tms.Track) []*tms.SearchField {
	fields := []*tms.SearchField{}

	add := func(name string, value string) {
		value = strings.TrimSpace(value)
		fields = append(fields, &tms.SearchField{Name: name, Value: value})
	}

	if len(t.Targets) > 0 {
		target := t.Targets[0]
		if target.Nmea != nil {
			if target.Nmea.Vdm != nil {
				if target.Nmea.Vdm.M1371 != nil {
					m := target.Nmea.Vdm.M1371
					add("mmsi", ais.FormatMMSI(int(m.Mmsi)))
				}
			}
		}
		if target.Imei != nil {
			add("imei", target.Imei.Value)
		}
		if target.Nodeid != nil {
			add("ingenuNodeId", target.Nodeid.Value)
		}
		if target.Sarmsg != nil {
			if target.Sarmsg.SarsatAlert != nil {
				if target.Sarmsg.SarsatAlert.Beacon != nil {
					add("sarsatBeaconId", target.Sarmsg.SarsatAlert.Beacon.HexId)
				}
			}
		}
	}

	return fields
}

func RegistryMetadataSearchFields(track *tms.Track) []*tms.SearchField {
	fields := []*tms.SearchField{}
	add := func(name string, value string) {
		value = strings.TrimSpace(value)
		fields = append(fields, &tms.SearchField{Name: name, Value: value})
	}

	if len(track.Metadata) == 0 {
		return fields
	}
	md := track.Metadata[0]
	if md.Name != "" {
		add("name", md.Name)
	}
	if md.Nmea != nil {
		if md.Nmea.Vdm != nil {
			if md.Nmea.Vdm.M1371 != nil {
				m := md.Nmea.Vdm.M1371
				add("mmsi", ais.FormatMMSI(int(m.Mmsi)))
				if m.StaticVoyage != nil {
					sv := m.StaticVoyage
					if sv.ImoNumber > 0 {
						add("imo", strconv.Itoa(int(sv.ImoNumber)))
					}
					if sv.CallSign != "" {
						add("callSign", sv.CallSign)
					}
					if sv.Destination != "" {
						add("destination", sv.Destination)
					}
				}
			}
		}
	}
	return fields
}

func RegistryKeywords(fieldSets ...[]*tms.SearchField) []string {
	keywords := []string{}
	seen := make(map[string]bool)
	for _, fields := range fieldSets {
		for _, field := range fields {
			if _, ok := seen[field.Name]; !ok {
				seen[field.Name] = true
				keywords = append(keywords, field.Value)
			}
		}
	}
	return keywords
}

func SetRegistryLabel(r *tms.RegistryEntry) {
	if r.Assignment != nil && r.Assignment.Label != "" {
		r.LabelType = "name"
		r.Label = r.Assignment.Label
	} else if len(r.MetadataFields) > 0 {
		r.LabelType = r.MetadataFields[0].Name
		r.Label = r.MetadataFields[0].Value
	} else if len(r.TargetFields) > 0 {
		r.LabelType = r.TargetFields[0].Name
		r.Label = r.TargetFields[0].Value
	} else {
		r.LabelType = ""
		r.Label = ""
	}
}

// For a given search string, which fields does this match to?
func RegistryMatches(search string, r *tms.RegistryEntry) []*tms.MatchField {
	results := []*tms.MatchField{}
	fieldSets := [][]*tms.SearchField{r.MetadataFields, r.TargetFields}
	seen := make(map[string]bool)
	for _, fields := range fieldSets {
		for _, field := range fields {
			if _, ok := seen[field.Name]; !ok {
				seen[field.Name] = true
				results = append(results, &tms.MatchField{
					Name:      field.Name,
					Value:     field.Value,
					Highlight: strings.HasPrefix(field.Value, search),
				})
			}
		}
	}
	return results
}
