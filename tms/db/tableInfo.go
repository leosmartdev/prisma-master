package db

import (
	"fmt"
	"reflect"
	"strings"

	"prisma/tms"
	"prisma/tms/log"
	"prisma/tms/marker"
	"prisma/tms/moc"

	pb "github.com/golang/protobuf/proto"
)

var (
	DefaultTables = NewTables([]*TableInfo{
		{
			Name:        "zones",
			Inst:        moc.Zone{},
			Indexes:     MiscIndexes,
			FeatureType: tms.FeatureCategory_ZoneFeature,
		},
		{
			Name:    "transmissions",
			Inst:    tms.Transmission{},
			Indexes: MiscIndexes,
		},
		&TableInfo{
			Name: "geofences",
			Inst: moc.GeoFence{},
		},
		{
			Name:           "incidents",
			Inst:           moc.Incident{},
			NoCappedMirror: true,
			Indexes: []Index{
				{
					Name:   "incidentIdUnique",
					Unique: true,
					Fields: []Field{
						{"me.incidentId"},
					},
				},
			},
		},
		{
			Name: "notes",
			Inst: moc.IncidentLogEntry{},
		},
		{
			Name: "markers",
			Inst: marker.Marker{},
		},
		{
			Name: "markerImages",
			Inst: marker.MarkerImage{},
		},
		{
			Name: "icons",
			Inst: moc.Icon{},
		},
		{
			Name: "iconImages",
			Inst: moc.IconImage{},
		},
		{
			Name:    "notices",
			Inst:    moc.Notice{},
			Indexes: MiscIndexes,
		},
		{
			Name: "activity",
			Inst: tms.MessageActivity{},
			Indexes: []Index{
				{
					Name: "registry_id",
					Fields: []Field{
						{"registry_id"},
					},
				},
			},
		},
		{
			Name: "remoteSites",
			Inst: moc.RemoteSite{},
		},
		{
			Name: "sit915",
			Inst: moc.Sit915{},
		},
		{
			Name: "mapconfig",
			Inst: moc.MapConfig{},
		},
		{
			Name: "filtertracks",
			Inst: moc.FilterTracks{},
		},
	})

	MiscIndexes = []Index{
		{
			Name: "ctime",
			Fields: []Field{
				{"ctime"},
			},
		},
		{
			Name: "track_id",
			Fields: []Field{
				{"track_id"},
			},
		},
	}
)

type Tables struct {
	Info            []*TableInfo
	TypeToTable     map[reflect.Type]*TableInfo
	TypeNameToTable map[string]*TableInfo
	NameToTable     map[string]*TableInfo
}

func NewTables(tables []*TableInfo) *Tables {
	ret := &Tables{
		Info:            tables,
		TypeToTable:     make(map[reflect.Type]*TableInfo),
		TypeNameToTable: make(map[string]*TableInfo),
		NameToTable:     make(map[string]*TableInfo),
	}
	ret.Update()
	return ret
}

type TableInfo struct {
	Name               string
	Inst               interface{}
	Type               reflect.Type
	Indexes            []Index
	FeatureType        tms.FeatureCategory
	ContainsAssociated bool
	// if true then no _live table is created
	NoCappedMirror bool
}

func (ti TableInfo) TypeName() string {
	return fmt.Sprintf("%v.%v", ti.Type.PkgPath(), ti.Type.Name())
}

type Field []string
type Index struct {
	Name      string
	GeoSphere bool
	TextIndex bool
	Fields    []Field
	Unique    bool
	Sparse    bool
}

func (t *Tables) Update() {
	for _, ti := range t.Info {
		t.NameToTable[ti.Name] = ti
		ti.Type = reflect.TypeOf(ti.Inst)
		t.TypeToTable[ti.Type] = ti
		valPtr := reflect.New(ti.Type)
		if msg, ok := valPtr.Interface().(pb.Message); ok {
			t.TypeNameToTable[pb.MessageName(msg)] = ti
		}
	}
}

func (t *Tables) GetInfo(tblname string) TableInfo {
	if ti, ok := t.NameToTable[tblname]; ok {
		return *ti
	}
	return TableInfo{}
}

func (t *Tables) TableFromType(obj interface{}) TableInfo {
	ty := reflect.TypeOf(obj)
	ti, ok := t.TypeToTable[ty]
	if ok {
		return *ti
	}

	if ty.Kind() == reflect.Ptr {
		ti, ok := t.TypeToTable[ty.Elem()]
		if ok {
			return *ti
		}
	}
	return TableInfo{}
}

func (t *Tables) TableFromTypeName(typename string) TableInfo {
	log.TraceMsg("TableFromTypeName, table types: %v", t.TypeNameToTable)
	if strings.Index(typename, "type.googleapis.com/") == 0 {
		typename = typename[len("type.googleapis.com/"):]
	}
	typename = strings.Replace(typename, "/", ".", -1)
	ti, ok := t.TypeNameToTable[typename]
	if ok {
		return *ti
	}
	log.Warn("Could not find table info for typename '%v', %v", typename, t.TypeNameToTable)
	return TableInfo{}
}
