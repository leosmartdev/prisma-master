package db

import (
	"fmt"
	"reflect"
	"strings"
)

// Resolve & check a golang path to subfield (from a root/base type) to a
// JSON/BSON path. If the path is invalid, panic.
func ResolveName(obj interface{}, path string) string {
	return ResolveNameArr(reflect.TypeOf(obj), strings.Split(path, "."))
}

func ResolveField(obj interface{}, path string) Field {
	return Field(strings.Split(ResolveName(obj, path), "."))
}

func getFieldName(sf reflect.StructField) string {
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

func ResolveNameArr(root reflect.Type, pathparts []string) string {
	if len(pathparts) == 0 {
		return ""
	}

	if root.Kind() == reflect.Ptr {
		return ResolveNameArr(root.Elem(), pathparts)
	}

	if root.Kind() != reflect.Struct {
		panic("Type with field to be resolved must be a struct!")
	}

	field, ok := root.FieldByName(pathparts[0])
	if !ok {
		panic(fmt.Sprintf("Could not find field '%v' in type '%v'!",
			pathparts[0], root))
	}

	if len(pathparts) > 1 {
		return getFieldName(field) + "." + ResolveNameArr(field.Type, pathparts[1:])
	}
	return getFieldName(field)
}
