package mongo

/**
 * This file contains the OLD bson decoder. We should delete it soon.
 */

import (
	"prisma/tms/log"

	"fmt"
	"reflect"
	"strings"

	"github.com/globalsign/mgo/bson"
	pb "github.com/golang/protobuf/proto"
)

func DecodeOld(dst interface{}, src bson.M) error {
	return decode(reflect.ValueOf(dst), src)
}

func decode(target reflect.Value, src interface{}) error {
	srcVal := reflect.ValueOf(src)
	var srcType reflect.Type
	if srcVal.IsValid() {
		srcType = srcVal.Type()
	}

	targetType := target.Type()
	log.Trace("Trying to decode %v (%v) to %v (%v)",
		src, srcType, target, targetType)

	_, useBson := bsonEncode[targetType]
	if useBson && targetType == srcType {
		target.Set(srcVal)
	}

	switch targetType.Kind() {
	case reflect.Ptr:
		if target.IsNil() {
			target.Set(reflect.New(targetType.Elem()))
		}
		return decode(target.Elem(), src)
	case reflect.Interface:
		if target.IsNil() {
			return NilInterfaceError
		}
		return decode(target.Elem(), src)
	}

	if targetType.Kind() == reflect.Struct {
		srcMap, ok := src.(bson.M)
		if !ok {
			log.Warn("Expected bson.M, got: %v", src)
			return UnexpectedType
		}

		sprops := pb.GetProperties(targetType)
		for i := 0; i < target.NumField(); i++ {
			ft := target.Type().Field(i)
			if strings.HasPrefix(ft.Name, "XXX_") {
				continue
			}
			fieldName := properties(ft, ft.Type).OrigName
			log.Trace("Decode looking for field '%v' in map", fieldName)

			valueForField, ok := srcMap[fieldName]
			if !ok {
				continue
			}
			delete(srcMap, fieldName)

			// Handle enums, which have an underlying type of int32,
			// and may appear as strings. We do this while handling
			// the struct so we have access to the enum info.
			// The case of an enum appearing as a number is handled
			// by the recursive call to unmarshalValue.
			valueString, valueIsString := valueForField.(string)
			if enum := sprops.Prop[i].Enum; valueIsString && enum != "" {
				vmap := pb.EnumValueMap(enum)
				// Don't need to do unquoting; valid enum names
				// are from a limited character set.
				n, ok := vmap[valueString]
				if !ok {
					return fmt.Errorf("unknown value %q for enum %s", valueString, enum)
				}
				f := target.Field(i)
				if f.Kind() == reflect.Ptr { // proto2
					f.Set(reflect.New(f.Type().Elem()))
					f = f.Elem()
				}
				f.SetInt(int64(n))
				continue
			}

			log.Trace("Trying to decode simple field '%s'", fieldName)
			if err := decode(target.Field(i), valueForField); err != nil {
				return err
			}
		}
		// Check for any oneof fields.
		for fname, raw := range srcMap {
			if oop, ok := sprops.OneofTypes[fname]; ok {
				nv := reflect.New(oop.Type.Elem())
				target.Field(oop.Field).Set(nv)
				if err := decode(nv.Elem().Field(0), raw); err != nil {
					return err
				}
				delete(srcMap, fname)
			}
		}
		if len(srcMap) > 0 {
			// Pick any field to be the scapegoat.
			var f string
			for fname := range srcMap {
				f = fname
				break
			}
			return fmt.Errorf("unknown field %q in %v", f, targetType)
		}
		return nil
	}

	// Any other type:
	if target.CanSet() {
		if srcVal.Type().ConvertibleTo(target.Type()) {
			target.Set(srcVal.Convert(target.Type()))
			return nil
		}
	}
	log.Warn("Cannot decode %v to %v", srcVal, target)
	return UnimplementedType
}
