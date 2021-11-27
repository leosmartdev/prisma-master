package mongo

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"unsafe"

	"prisma/tms/log"

	"github.com/globalsign/mgo/bson"
	"github.com/golang/protobuf/proto"
)

/***
 *  This struct caches information about structs gained through reflection. We
 *  then use this information to do bson encoding and decoding. Caching
 *  increases the speed of encoding/decoding significantly as we can mostly use
 *  untyped pointers and byte offsets into structs instead of the reflection
 *  API. The reflection API is very, very slow so running it once at startup is
 *  the only acceptable way to use it.
 *
 *  Additionally, if the object is a protobuf message, use some
 *  protobuf-specific stuff to make a more rational encoding.
 */
type StructData struct {
	goType        reflect.Type            // The underlying go type
	fields        map[string]*StructField // If this is indeed a struct, information about each field
	orderedFields []*StructField
	containerType *StructField // If this is a container(map, array, slice)
	lastSize      uint32       // Last time we wrote one of these, how many bytes was it?
	flatten       bool         // Should we ignore this struct and write its contents as if part of the parent?
}

type CustomWriter func([]byte, *bytes.Buffer, StructField)
type CustomReader func(bson.Raw, unsafe.Pointer, StructField) uintptr

// Information about each field in a struct
type StructField struct {
	// The reflection data for this field
	field reflect.Type

	// Offset into the field
	fieldOffset uintptr

	// If it's a struct or pointer to a struct, fill this in
	structData *StructData

	// Encoding name
	bsonName string

	// Was this a "oneof" protobuf field?
	oneof map[reflect.Type]*StructField

	// Is this field involved in a oneof?
	oneofInvolved bool

	// The bson type
	typeByte byte

	// An enum
	enum  map[string]int32
	enumR map[int32]string

	// Special type printer
	write CustomWriter

	// Special type decoder
	read CustomReader
}

func NoMap(string) reflect.Type {
	panic("No interface mapping known!")
}

// Get type data for type 'ty'. If a field of type 'interface' is encountered,
// we need to be able to resolve it to a concrete type. 'ifMap' can be
// specified to do that mapping.
func NewStructData(ty reflect.Type, ifMap func(string) reflect.Type) *StructData {
	log.TraceMsg("Building struct data for %v", ty)
	switch ty.Kind() {
	case reflect.Ptr:
		return NewStructData(ty.Elem(), ifMap)
	case reflect.Slice,
		reflect.Array,
		reflect.Map:
		// These are significantly different
		return newStructDataContainer(ty, ifMap)
	default:
		// Only return stuff for pointers and structs!
		return nil
	case reflect.Struct:
		// We can deal directly... continue
	}

	log.TraceMsg("Building REAL struct data for %+v", ty)
	data := StructData{
		goType:   ty,
		fields:   make(map[string]*StructField),
		lastSize: 0,
		flatten:  false,
	}

	// *** Try to get protobuf-centric info
	var protodata *proto.StructProperties = nil
	var pbMsg proto.Message
	var prototy reflect.Type = nil
	if ty.Implements(reflect.TypeOf(&pbMsg).Elem()) {
		prototy = ty
	} else if reflect.PtrTo(ty).Implements(reflect.TypeOf(&pbMsg).Elem()) {
		prototy = reflect.PtrTo(ty)
	}
	if prototy != nil && prototy.Kind() != reflect.Struct {
		protodata = proto.GetProperties(prototy.Elem())
		log.TraceMsg("Protodata for %v:\n%v", ty, log.Spew(protodata))
	}

	//  *** Get information for each type field
	for i := 0; i < ty.NumField(); i++ {
		var prop *proto.Properties = nil
		if protodata != nil {
			prop = protodata.Prop[i]
		}
		f := newStructField(ty.Field(i), nil, prop, ifMap)
		if f != nil {
			if f.oneofInvolved {
				f.oneof = make(map[reflect.Type]*StructField)
			}

			data.fields[f.bsonName] = f
		}
	}

	// *** Get information for ONEOF fields
	if protodata != nil {
		log.TraceMsg("Evaluating protodata for %v, %v:\n%v", log.Spew(ty), log.Spew(prototy), log.Spew(protodata))
		for _, oneof := range protodata.OneofTypes {
			log.TraceMsg("Protodata.OneofTypes %v:\n%v", ty, log.Spew(oneof))
			if oneof == nil {
				continue
			}
			if oneof.Field >= ty.NumField() {
				panic("Oneof has invalid field number!")
			}
			sf := ty.Field(oneof.Field)
			f := newStructField(sf, &oneof.Type, oneof.Prop, ifMap)
			if f != nil {
				bsn := getBsonName(sf)
				storageField, ok := data.fields[bsn]
				if !ok {
					panic(fmt.Sprintf("Could not find storage field for oneof under '%v'. Should be: %v!", bsn, log.Spew(sf)))
				}
				storageField.oneof[oneof.Type] = f
				if oneof.Type.Kind() == reflect.Ptr {
					storageField.oneof[oneof.Type.Elem()] = f
				} else {
					storageField.oneof[reflect.PtrTo(oneof.Type)] = f
				}

				data.fields[f.bsonName] = f
			}
		}
	}
	data.sortFields()
	if DEBUG {
		fmt.Println(ty)
		fmt.Println(data.fields)
	}
	return &data
}

func newStructDataContainer(ty reflect.Type, ifMap func(string) reflect.Type) *StructData {
	data := StructData{
		goType:   ty,
		fields:   nil,
		lastSize: 0,
		flatten:  false,
		containerType: newStructField(
			reflect.StructField{
				Type:   ty.Elem(),
				Offset: 0,
			}, nil, nil, ifMap),
	}
	data.sortFields()
	return &data
}

func newStructField(sf reflect.StructField,
	realType *reflect.Type,
	prop *proto.Properties,
	ifMap func(string) reflect.Type) *StructField {
	f := StructField{
		field:         sf.Type,
		fieldOffset:   sf.Offset,
		structData:    nil,
		oneof:         nil,
		oneofInvolved: isOneof(sf),
	}
	if prop != nil {
		f.bsonName = prop.OrigName
	} else {
		f.bsonName = getBsonName(sf)
	}

	if strings.HasPrefix(f.bsonName, "XXX_") {
		// Ignore fields beginning with XXX_
		return nil
	}

	//Ensure that bsonName is a valid c_string (contains no 0x00 bytes). I
	//don't see how this situation could possibly arise, though lets check
	//anyway.
	nameBytes := []byte(f.bsonName)
	for _, b := range nameBytes {
		if b == 0x00 {
			panic("Cannot use name '%v' -- it contains a 0x00 byte")
		}
	}

	effType := sf.Type
	if realType != nil {
		effType = *realType
	}
	if !f.oneofInvolved && effType.Kind() == reflect.Interface {
		effType = ifMap(sf.Name)
	}

	fullCustom := true
	if customW, ok := CustomWriters[effType]; ok {
		f.write = customW
	} else {
		fullCustom = false
	}
	if customR, ok := CustomReaders[effType]; ok {
		f.read = customR
	} else {
		fullCustom = false
	}

	if prop != nil {
		if prop.Enum != "" {
			f.enum = proto.EnumValueMap(prop.Enum)
			f.enumR = make(map[int32]string)
			for k, v := range f.enum {
				f.enumR[v] = k
			}
		}
	}

	if !fullCustom {
		f.structData = NewStructData(effType, ifMap)
	}

	f.typeByte = typeByte(effType)
	return &f
}

func isOneof(sf reflect.StructField) bool {
	return sf.Tag.Get("protobuf_oneof") != ""
}

func getBsonName(sf reflect.StructField) string {
	pboo := sf.Tag.Get("protobuf_oneof")
	if pboo != "" {
		return pboo
	}

	pb := sf.Tag.Get("protobuf")
	if pb != "" {
		arr := strings.Split(pb, ",")
		for _, f := range arr {
			if strings.HasPrefix(f, "name=") {
				arr := strings.Split(f, "=")
				if len(arr) >= 2 {
					return string(arr[1])
				}
			}
		}
	}

	bson := sf.Tag.Get("bson")
	if bson != "" {
		arr := strings.Split(bson, ",")
		return string(arr[0])
	}

	return sf.Name
}

func typeByte(ty reflect.Type) byte {
	if ty == reflect.TypeOf(bson.ObjectId("")) {
		return 0x07
	}

	switch ty.Kind() {
	case reflect.Ptr:
		return typeByte(ty.Elem())
	case reflect.Struct,
		reflect.Map:
		return 0x03
	case reflect.Array,
		reflect.Slice:
		return 0x04
	case reflect.String:
		return 0x02
	case reflect.Bool:
		return 0x08
	case reflect.Interface:
		return 0x0A
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32:
		return 0x10
	case reflect.Int64:
		return 0x12
	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16:
		return 0x10
	case reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr:
		return 0x12
	case reflect.Float32,
		reflect.Float64:
		return 0x01
	}
	panic(fmt.Sprintf("Cannot encode this type: %v!", ty))
}

func (sd *StructData) sortFields() {
	names := make([]string, 0, len(sd.fields))
	for name, _ := range sd.fields {
		names = append(names, name)
	}

	sort.Sort(sort.StringSlice(names))

	sd.orderedFields = make([]*StructField, 0, len(sd.fields))
	for _, name := range names {
		sd.orderedFields = append(sd.orderedFields, sd.fields[name])
	}
}
