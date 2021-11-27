package mongo

/******
 * This file contains the OLD bson encoder. We should delete it soon.
 */
import (
	"prisma/tms/log"

	"errors"
	"reflect"
	"strings"
	"time"

	"github.com/globalsign/mgo/bson"
	pb "github.com/golang/protobuf/proto"
)

var (
	NilInterfaceError = errors.New("Cannot decode to nil interface")
	UnexpectedType    = errors.New("Got unexpected type")
	UnimplementedType = errors.New("Type recieved for decoding not yet implemented")

	bsonEncode = map[reflect.Type]bool{
		reflect.TypeOf(time.Time{}): true,
	}
)

func EncodeOld(src interface{}) bson.M {
	s := reflect.ValueOf(src).Elem()
	return encode(s)
}

func encode(s reflect.Value) bson.M {
	dst := make(bson.M)
	if !s.IsValid() {
		return dst
	}

	switch s.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		if s.IsNil() {
			return dst
		}
	}

	if s.Kind() == reflect.Ptr {
		return encode(s.Elem())
	}

	for i := 0; i < s.NumField(); i++ {
		value := s.Field(i)
		valueField := s.Type().Field(i)
		if strings.HasPrefix(valueField.Name, "XXX_") {
			continue
		}

		encodeField(&dst, value, valueField)
	}
	return dst
}

func encodeField(dst *bson.M, value reflect.Value,
	valueField reflect.StructField) {
	// IsNil will panic on most value kinds.
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		if value.IsNil() {
			return
		}
	}

	// Detect and do not emit default defaults
	switch value.Kind() {
	case reflect.Bool:
		if !value.Bool() {
			return
		}
	case reflect.Int32, reflect.Int64:
		if value.Int() == 0 {
			return
		}
	case reflect.Uint32, reflect.Uint64:
		if value.Uint() == 0 {
			return
		}
	case reflect.Float32, reflect.Float64:
		if value.Float() == 0 {
			return
		}
	case reflect.String:
		if value.Len() == 0 {
			return
		}
	case reflect.Interface:
		// Oneof fields need special handling.
		if valueField.Tag.Get("protobuf_oneof") != "" {
			// value is an interface containing &T{real_value}.
			sv := value.Elem().Elem() // interface -> *T -> T
			value = sv.Field(0)
			valueField = sv.Type().Field(0)
		} else {
			encodeField(dst, value.Elem(), valueField)
			return
		}
	}

	_, skip := bsonEncode[value.Type()]
	if value.CanInterface() && skip {
		name := valueField.Tag.Get("bson")
		if name == "" {
			name = valueField.Name
		}
		(*dst)[name] = value.Interface()
	} else {
		prop := properties(valueField, value.Type())
		name := prop.OrigName
		(*dst)[name] = encodePbField(prop, value)
	}
}

func encodePbField(prop pb.Properties, v reflect.Value) interface{} {
	v = reflect.Indirect(v)

	switch v.Kind() {
	case reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32,
		reflect.Float64,
		reflect.Complex64,
		reflect.Complex128,
		reflect.String:
		return v.Interface()
	case reflect.Struct:
		msg, ok := v.Addr().Interface().(pb.Message)
		if !ok {
			log.Error("Cannot encode struct field which is struct but not a proto message: %v (%v)", v, v.Type())
			return nil
		}
		return EncodeOld(msg)
	case reflect.Ptr,
		reflect.Interface:
		if v.IsNil() {
			return nil
		}
		return encodePbField(prop, v.Elem())
	case reflect.Slice,
		reflect.Array:
		s := make([]interface{}, 0)
		for i := 0; i < v.Len(); i++ {
			s = append(s, encodePbField(prop, v.Index(i)))
		}
		return s
	case reflect.Map:
		m := make(map[interface{}]interface{})
		for _, k := range v.MapKeys() {
			val := v.MapIndex(k)
			m[encodePbField(prop, k)] = encodePbField(prop, val)
		}
		return m
	}

	if v.CanInterface() {
		log.Trace("Just using Interface() to encode type '%v'. This may not be right....", v.Type())
		return v.Interface()
	} else {
		log.Error("Could not encode field: %v (%v)", v, prop)
		return nil
	}
}

func properties(valueField reflect.StructField, ty reflect.Type) pb.Properties {
	var prop pb.Properties
	prop.Init(ty, valueField.Name, valueField.Tag.Get("protobuf"), &valueField)
	return prop
}
