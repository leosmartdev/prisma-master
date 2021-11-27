package tms

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	geo "prisma/tms/geojson"

	"github.com/globalsign/mgo/bson"

	"github.com/golang/protobuf/jsonpb"
)

// When storing targets in a map indexed by their ID, the TargetID
// struct synthesized by protobuf are inappropriate -- they
// contain pointers which may not compare properly. The ID struct should be
// used instead. Utility functions to convert from TargetID are
// provided
type ID struct {
	a, b uint64
	s    string
	ty   uint8
}

// Checks if an array of Target values are valid. valid returns true, invalid returns false
func ValidTargets(targets []*Target) bool {
	// assume valid
	var isValid bool = true
	for i := range targets {
		isValid = ValidPositionPoint(targets[i].GetPosition())
		if !(isValid) {
			break
		}
	}
	return isValid
}

// Checks if Position is valid
func ValidPositionPoint(point *Point) bool {
	// assume valid
	var isValid bool = true
	if nil != point {
		isValid = !((point.Latitude == 0) && (point.Longitude == 0))
	}
	return isValid
}

func FromTargetID(id *TargetID) ID {
	if id == nil {
		return ID{}
	}
	a := (uint64(id.Producer.Site) << 32) | (uint64(id.Producer.Eid))
	var ty uint8 = 1
	var b uint64
	switch x := id.SerialNumber.(type) {
	case *TargetID_TimeSerial:
		b = (uint64(x.TimeSerial.Seconds) << 32) | (uint64(x.TimeSerial.Counter))
	}
	return ID{ty: ty, a: a, b: b}
}

func (id *TargetID) ID() ID {
	return FromTargetID(id)
}

//LookupDevID is a helper function that Returns devices specific comm ID in a target
func (t *Target) LookupDevID() string {
	if t.GetMmsi() != "" {
		return t.Mmsi
	}
	if t.GetImei() != nil {
		return t.Imei.Value
	}
	if t.GetNodeid() != nil {
		return t.Nodeid.Value
	}
	return ""
}

func (t *Target) Point() *geo.Point {
	pos := t.GetPosition()
	if pos == nil {
		return nil
	}
	point := geo.NewPointFromLatLng(pos.Latitude, pos.Longitude)
	point.Coordinates.Z = pos.Altitude
	return point
}

func (t *Target) Points() *geo.MultiPoint {
	pos := t.GetPositions()
	if pos == nil {
		return nil
	}
	coords := make([]geo.Position, 0, len(pos))
	for _, p := range pos {
		coords = append(coords, geo.Position{X: p.Longitude, Y: p.Latitude, Z: p.Altitude})
	}
	return &geo.MultiPoint{Coordinates: coords}
}

func (t *Track) Point() *geo.Point {
	if len(t.Targets) == 0 {
		return nil
	}
	return t.Targets[0].Point()
}

func (t *Track) Geometry() geo.Object {
	if len(t.Targets) == 0 {
		return nil
	}
	tgt := t.Targets[0]
	if tgt.GetPositions() != nil {
		return tgt.Points()
	}
	return tgt.Point()
}

func (id *TargetID) StringID() string {
	if id == nil {
		return ""
	}

	ret := ""

	switch x := id.SerialNumber.(type) {
	case *TargetID_TimeSerial:
		g := x.TimeSerial
		ret = ret + fmt.Sprintf("sec:%v:ctr:%v", g.Seconds, g.Counter)
	}

	if id.Producer != nil {
		ret = ret + fmt.Sprintf(":site:%v:eid:%v", id.Producer.Site, id.Producer.Eid)
	}

	return ret
}

func atou(s string) uint32 {
	num, err := strconv.ParseUint(s, 0, 32)
	if err != nil {
		panic(err)
	}
	return uint32(num)
}

func FromTargetStringID(s string) *TargetID {
	m := make(map[string]string)
	arr := strings.Split(s, ":")
	for i := 1; i < len(arr); i = i + 2 {
		m[arr[i-1]] = arr[i]
	}

	ret := new(TargetID)

	site, hasSite := m["site"]
	eid, hasEid := m["eid"]

	if hasSite || hasEid {
		ret.Producer = new(SensorID)
		if hasSite {
			ret.Producer.Site = atou(site)
		}
		if hasEid {
			ret.Producer.Eid = atou(eid)
		}
	}

	if secStr, ok := m["sec"]; ok {
		// It's got a generic identifier!
		sec := atou(secStr)
		ctr := atou(m["ctr"])
		ret.SerialNumber = &TargetID_TimeSerial{
			TimeSerial: &TimeSerialNumber{
				Seconds: int64(sec),
				Counter: int32(ctr),
			},
		}
	}
	return ret
}

func (t *Track) MarshalJSON() ([]byte, error) {
	m := jsonpb.Marshaler{
		Indent: "  ",
	}
	s, err := m.MarshalToString(t)
	if err != nil {
		return nil, err
	}
	return []byte(s), err
}
func (t *Track) UnmarshalJSON(b []byte) error {
	return jsonpb.UnmarshalString(string(b), t)
}

func (t *Target) MarshalJSON() ([]byte, error) {
	m := jsonpb.Marshaler{
		Indent: "  ",
	}
	s, err := m.MarshalToString(t)
	if err != nil {
		return nil, err
	}
	return []byte(s), err
}
func (t *Target) UnmarshalJSON(b []byte) error {
	return jsonpb.UnmarshalString(string(b), t)
}

func (t *TrackMetadata) MarshalJSON() ([]byte, error) {
	m := jsonpb.Marshaler{
		Indent: "  ",
	}
	s, err := m.MarshalToString(t)
	if err != nil {
		return nil, err
	}
	return []byte(s), err
}
func (t *TrackMetadata) UnmarshalJSON(b []byte) error {
	return jsonpb.UnmarshalString(string(b), t)
}

func (t *TargetID) MarshalJSON() ([]byte, error) {
	m := jsonpb.Marshaler{
		Indent: "  ",
	}
	s, err := m.MarshalToString(t)
	if err != nil {
		return nil, err
	}
	return []byte(s), err
}
func (t *TargetID) UnmarshalJSON(b []byte) error {
	return jsonpb.UnmarshalString(string(b), t)
}

func (t *Track) Time() time.Time {
	var tm time.Time
	for _, tgt := range t.Targets {
		t := FromTimestamp(tgt.Time)
		if tm.Before(t) {
			tm = t
		}
	}
	for _, md := range t.Metadata {
		t := FromTimestamp(md.Time)
		if tm.Before(t) {
			tm = t
		}
	}
	return tm
}

func (t *Track) Rect() (minX, minY, minZ, maxX, maxY, maxZ float64) {
	if len(t.Targets) == 0 {
		return 0, 0, 0, 0, 0, 0
	}
	tgt := t.Targets[0]
	var bbox geo.BBox
	if tgt.Points() != nil {
		bbox = tgt.Points().CalculatedBBox()
	} else {
		bbox = tgt.Point().CalculatedBBox()
	}
	return bbox.Min.X, bbox.Min.Y, 0, bbox.Max.X, bbox.Max.Y, 0
}

func (t *Track) HasPosition() bool {
	tgt := t.Targets[0]
	return tgt.Points() != nil || tgt.Point() != nil
}

func (t *Track) LookupID() string {
	if t.RegistryId != "" {
		return t.RegistryId
	}
	return t.Id
}

// TrackExtension is used to extend expiration date for a given track
type TrackExtension struct {
	Track   *Track
	Updated time.Time
	Next    time.Time
	Expires time.Time
	Count   int32
}

// Db populates TrackExtensionDB from TrackExtension
func (t TrackExtension) Db() (*TrackExtensionDb, error) {
	return &TrackExtensionDb{
		Track:   t.Track,
		Updated: t.Updated,
		Next:    t.Next,
		Expires: t.Expires,
		Count:   t.Count,
	}, nil
}

// TrackExtensionDb is the DB structure for TrackExtension
type TrackExtensionDb struct {
	Track   *Track
	Updated time.Time
	Next    time.Time
	Expires time.Time
	Count   int32
}

// Proto populate TrackExtension from TrackExtensionDb
func (t TrackExtensionDb) Proto() (*TrackExtension, error) {
	ex := &TrackExtension{
		Track:   &Track{},
		Updated: t.Updated,
		Next:    t.Next,
		Expires: t.Expires,
		Count:   t.Count,
	}

	jtrack, err := bson.MarshalJSON(t.Track)
	if err != nil {
		return nil, err
	}
	bson.UnmarshalJSON(jtrack, ex.Track)
	err = jsonpb.UnmarshalString(string(jtrack), ex.Track)
	return ex, err
}
