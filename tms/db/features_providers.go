package db

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"prisma/tms"
	"prisma/tms/client_api"
	"prisma/tms/feature"
	"prisma/tms/log"

	"prisma/gogroup"
)

// GetProviders returns array of FeaturesProvider
func GetProviders(types []*tms.FeatureType) []FeaturesProvider {
	var ret []FeaturesProvider

	if len(types) == 0 {
		types = append(types, &tms.FeatureType{Type: tms.FeatureCategory_AllFeatures})
	}

	addedCategories := map[tms.FeatureCategory]struct{}{} // A set of FeatureCategories already added

	// Iterate through the feature types requested
	for _, ty := range types {
		if ty.Type == tms.FeatureCategory_SiteFeature || ty.Type == tms.FeatureCategory_AllFeatures {
			if _, alreadyAdded := addedCategories[tms.FeatureCategory_SiteFeature]; !alreadyAdded {
				ret = append(ret, &SitesFeatureProvider{})
				log.Debug("Adding provider:\n%v", ret)
				addedCategories[tms.FeatureCategory_SiteFeature] = struct{}{}
			}
		}
		if ty.Type == tms.FeatureCategory_TrackFeature || ty.Type == tms.FeatureCategory_AllFeatures {
			if _, alreadyAdded := addedCategories[tms.FeatureCategory_TrackFeature]; !alreadyAdded {
				// A tracks feature is requested and we haven't already added one
				ret = append(ret, &TracksFeatureProvider{})
				log.Debug("Adding provider:\n%v", ret)
				addedCategories[tms.FeatureCategory_TrackFeature] = struct{}{}
			}
		}

		if ty.Type != tms.FeatureCategory_TrackFeature {
			found := false
			// Not a track feature. Look up in tables for misc_data feature
			for _, ti := range DefaultTables.Info {
				if ti.FeatureType == ty.Type || (ty.Type == tms.FeatureCategory_AllFeatures &&
					ti.FeatureType != tms.FeatureCategory_UnknownFeature) {
					// Yes! A match.
					if _, alreadyAdded := addedCategories[ti.FeatureType]; !alreadyAdded {
						// And we haven't added it yet
						ret = append(ret, &MiscDataFeatureProvider{
							table: ti,
						})
						addedCategories[ti.FeatureType] = struct{}{}
						found = true
					}
				}
			}

			if !found {
				panic(fmt.Sprintf("Could not find provider for type: %v", ty.Type))
			}
		}
	}
	return ret
}

type FeaturesProvider interface {
	Init(*client_api.ViewRequest, gogroup.GoGroup, TrackDB, MiscDB, SiteDB, DeviceDB) error

	// Search for feature details _OR_ associated objects. Only ONE of the
	// three return values can be non-nil
	DetailsStream(GoDetailRequest) (<-chan GoFeatureDetail, error)

	// Get the history for a feature
	History(GoHistoryRequest) (<-chan *feature.F, error)

	Destroy()

	// Get all feature updates
	Service(chan<- FeatureUpdate) error
}

type ProviderCommon struct {
	req *client_api.ViewRequest

	ctxt gogroup.GoGroup
}

func (p *ProviderCommon) commonInit(req *client_api.ViewRequest, group gogroup.GoGroup) error {
	p.req = req
	p.ctxt = group
	return nil
}

func fieldName(sf reflect.StructField) string {
	pbtag := sf.Tag.Get("protobuf")
	if pbtag != "" {
		arr := strings.Split(pbtag, ",")
		for _, part := range arr {
			if strings.HasPrefix(part, "json=") {
				return strings.TrimPrefix(part, "json=")
			}
		}
	}

	if sf.Name != "" {
		nameRunes := []rune(sf.Name)
		nameRunes[0] = unicode.ToLower(nameRunes[0])
		return string(nameRunes)
	}

	return ""
}

func addAllToFeature(pref string, ft *feature.F, obj interface{}) {
	val := reflect.ValueOf(obj)
	switch val.Type().Kind() {
	case reflect.Ptr:
		if val.IsNil() {
			return
		}
		addAllToFeature(pref, ft, val.Elem().Interface())
		return
	default:
		// Can't handle this guy!
		return
	case reflect.Struct:
		// Actually handle this case
	}

	for i := 0; i < val.NumField(); i++ {
		sf := val.Type().Field(i)
		if strings.HasPrefix(sf.Name, "XXX_") {
			continue
		}

		name := pref + "." + fieldName(sf)
		f := val.Field(i)
		if f.Type().Kind() == reflect.Ptr {
			if f.IsNil() {
				continue
			}
			f = f.Elem()
		}
		if f.Type().Kind() == reflect.Struct {
			addAllToFeature(name, ft, f.Interface())
		} else {
			if _, ok := ft.Properties[name]; !ok {
				// Only add it if not previously added
				ft.Properties[name] = f.Interface()
			}
		}
	}
}
