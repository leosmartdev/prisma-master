package db

import (
	"prisma/tms"
	"prisma/tms/nmea"
	"prisma/tms/sar"
	"reflect"
	"testing"

	"github.com/golang/protobuf/ptypes/wrappers"
)

var searchMetadataFieldTests = []struct {
	name   string
	track  *tms.Track
	fields []*tms.SearchField
}{
	{
		"ais",
		&tms.Track{
			Metadata: []*tms.TrackMetadata{
				&tms.TrackMetadata{
					Name: "BIG GUN",
					Nmea: &nmea.Nmea{
						Vdm: &nmea.Vdm{
							M1371: &nmea.M1371{
								Mmsi: 123456789,
								StaticVoyage: &nmea.M1371_Static{
									ImoNumber:   98765,
									CallSign:    "ORO2017",
									Destination: "SINGAPORE",
								},
							},
						},
					},
				},
			},
		},
		[]*tms.SearchField{
			&tms.SearchField{Name: "name", Value: "BIG GUN"},
			&tms.SearchField{Name: "mmsi", Value: "123456789"},
			&tms.SearchField{Name: "imo", Value: "98765"},
			&tms.SearchField{Name: "callSign", Value: "ORO2017"},
			&tms.SearchField{Name: "destination", Value: "SINGAPORE"},
		},
	},
}

func TestRegistryMetadataSearchFields(t *testing.T) {
	for _, test := range searchMetadataFieldTests {
		t.Run(test.name, func(t *testing.T) {
			sf := RegistryMetadataSearchFields(test.track)
			if !reflect.DeepEqual(sf, test.fields) {
				t.Errorf("\n have: %+v \n want: %+v", sf, test.fields)
			}
		})
	}
}

var searchTargetFieldTests = []struct {
	name   string
	track  *tms.Track
	fields []*tms.SearchField
}{
	{
		"ais",
		&tms.Track{
			Targets: []*tms.Target{
				&tms.Target{
					Nmea: &nmea.Nmea{
						Vdm: &nmea.Vdm{
							M1371: &nmea.M1371{
								Mmsi: 123456789,
							},
						},
					},
				},
			},
		},
		[]*tms.SearchField{
			&tms.SearchField{Name: "mmsi", Value: "123456789"},
		},
	},
	{
		"imei",
		&tms.Track{
			Targets: []*tms.Target{
				&tms.Target{
					Imei: &wrappers.StringValue{Value: "6543"},
				},
			},
		},
		[]*tms.SearchField{
			&tms.SearchField{Name: "imei", Value: "6543"},
		},
	},
	{
		"ingenuNodeId",
		&tms.Track{
			Targets: []*tms.Target{
				&tms.Target{
					Nodeid: &wrappers.StringValue{Value: "5432"},
				},
			},
		},
		[]*tms.SearchField{
			&tms.SearchField{Name: "ingenuNodeId", Value: "5432"},
		},
	},
	{
		"sarsatBeaconId",
		&tms.Track{
			Targets: []*tms.Target{
				&tms.Target{
					Sarmsg: &sar.SarsatMessage{
						SarsatAlert: &sar.SarsatAlert{
							Beacon: &sar.Beacon{
								HexId: "deadbeef",
							},
						},
					},
				},
			},
		},
		[]*tms.SearchField{
			&tms.SearchField{Name: "sarsatBeaconId", Value: "deadbeef"},
		},
	},
}

func TestRegistryTargetSearchFields(t *testing.T) {
	for _, test := range searchTargetFieldTests {
		t.Run(test.name, func(t *testing.T) {
			sf := RegistryTargetSearchFields(test.track)
			if !reflect.DeepEqual(sf, test.fields) {
				t.Errorf("\n have: %+v \n want: %+v", sf, test.fields)
			}
		})
	}
}

var keywordTests = []struct {
	name     string
	fields   []*tms.SearchField
	fields2  []*tms.SearchField
	keywords []string
}{
	{
		"ais",
		[]*tms.SearchField{
			&tms.SearchField{Name: "mmsi", Value: "123456789"},
			&tms.SearchField{Name: "imo", Value: "98765"},
			&tms.SearchField{Name: "callSign", Value: "ORO2017"},
			&tms.SearchField{Name: "destination", Value: "SINGAPORE"},
			&tms.SearchField{Name: "name", Value: "BIG GUN"},
		},
		[]*tms.SearchField{},
		[]string{"123456789", "98765", "ORO2017", "SINGAPORE", "BIG GUN"},
	},
	{
		"mix",
		[]*tms.SearchField{
			&tms.SearchField{Name: "mmsi", Value: "123456789"},
			&tms.SearchField{Name: "imo", Value: "98765"},
			&tms.SearchField{Name: "callSign", Value: "ORO2017"},
			&tms.SearchField{Name: "destination", Value: "SINGAPORE"},
			&tms.SearchField{Name: "name", Value: "BIG GUN"},
		},
		[]*tms.SearchField{
			&tms.SearchField{Name: "mmsi", Value: "123456789"},
		},
		[]string{"123456789", "98765", "ORO2017", "SINGAPORE", "BIG GUN"},
	},
}

func TestRegistryKeywords(t *testing.T) {
	for _, test := range keywordTests {
		t.Run(test.name, func(t *testing.T) {
			kw := RegistryKeywords(test.fields, test.fields2)
			if !reflect.DeepEqual(kw, test.keywords) {
				t.Errorf("\n have: %+v \n want: %+v", kw, test.keywords)
			}
		})
	}
}

var matchTests = []struct {
	name     string
	search   string
	registry *tms.RegistryEntry
	expected []*tms.MatchField
}{
	{
		"name",
		"12M",
		&tms.RegistryEntry{
			MetadataFields: []*tms.SearchField{
				&tms.SearchField{Name: "name", Value: "12MONKEYS"},
				&tms.SearchField{Name: "mmsi", Value: "123456789"},
			},
		},
		[]*tms.MatchField{
			&tms.MatchField{Name: "name", Value: "12MONKEYS", Highlight: true},
			&tms.MatchField{Name: "mmsi", Value: "123456789"},
		},
	},
	{
		"mmsi",
		"123",
		&tms.RegistryEntry{
			MetadataFields: []*tms.SearchField{
				&tms.SearchField{Name: "name", Value: "12MONKEYS"},
				&tms.SearchField{Name: "mmsi", Value: "123456789"},
			},
		},
		[]*tms.MatchField{
			&tms.MatchField{Name: "name", Value: "12MONKEYS"},
			&tms.MatchField{Name: "mmsi", Value: "123456789", Highlight: true},
		},
	},
	{
		"name and mmsi",
		"12",
		&tms.RegistryEntry{
			MetadataFields: []*tms.SearchField{
				&tms.SearchField{Name: "name", Value: "12MONKEYS"},
				&tms.SearchField{Name: "mmsi", Value: "123456789"},
			},
		},
		[]*tms.MatchField{
			&tms.MatchField{Name: "name", Value: "12MONKEYS", Highlight: true},
			&tms.MatchField{Name: "mmsi", Value: "123456789", Highlight: true},
		},
	},
	{
		"none",
		"XXX",
		&tms.RegistryEntry{
			MetadataFields: []*tms.SearchField{
				&tms.SearchField{Name: "name", Value: "12MONKEYS"},
				&tms.SearchField{Name: "mmsi", Value: "123456789"},
			},
		},
		[]*tms.MatchField{
			&tms.MatchField{Name: "name", Value: "12MONKEYS"},
			&tms.MatchField{Name: "mmsi", Value: "123456789"},
		},
	},
}

func TestRegistryMatches(t *testing.T) {
	for _, test := range matchTests {
		t.Run(test.name, func(t *testing.T) {
			matches := RegistryMatches(test.search, test.registry)
			if !reflect.DeepEqual(matches, test.expected) {
				t.Errorf("\n have: %+v \n want: %+v", matches, test.expected)
			}
		})
	}
}
