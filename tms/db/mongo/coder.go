package mongo

import (
	"github.com/globalsign/mgo/bson"
	"reflect"
	"unsafe"
)

/**
 * Coder is used mostly to pass through to a structData encode/decode, but it
 * also has the option of catching additional information during decoding.
 * Right now, it just holds the last encountered bson ObjectId. Could be
 * expanded in the future to hold misc. fields that weren't mapped into a
 * struct field and things like that.
 */
type Coder struct {
	TypeData *StructData
	LastID   bson.ObjectId
}

func (c *Coder) DecodeTo(raw bson.Raw, ptr unsafe.Pointer) uintptr {
	c.LastID = bson.ObjectId("")
	return c.TypeData.decodeTo(c, raw, ptr)
}

func (c *Coder) Encode(ptr unsafe.Pointer) bson.Raw {
	c.LastID = bson.ObjectId("")
	return c.TypeData.encode(c, ptr)
}

/******
 *  Convenience functions
 */
var (
	//StructData for a bunch of known types:
	typeData = make(map[reflect.Type]*StructData)
)

func Encode(obj interface{}) bson.Raw {
	val := reflect.ValueOf(obj)
	if val.Type().Kind() == reflect.Ptr {
		val = val.Elem()
	}
	ty := val.Type()
	if _, ok := typeData[ty]; !ok {
		typeData[ty] = NewStructData(ty, NoMap)
	}
	var raw bson.Raw
	if val.CanAddr() {
		raw = typeData[ty].EncodeIface(val.Addr().Interface())
	}
	return raw
}

func EncodeToMap(obj interface{}) bson.D {
	val := reflect.ValueOf(obj)
	if val.Type().Kind() == reflect.Ptr {
		val = val.Elem()
	}
	ty := val.Type()
	if _, ok := typeData[ty]; !ok {
		typeData[ty] = NewStructData(ty, NoMap)
	}
	return typeData[ty].EncodeIfaceToMap(val.Addr().Interface())
}

func Decode(dst interface{}, src bson.Raw) {
	val := reflect.ValueOf(dst).Elem()
	ty := val.Type()
	if _, ok := typeData[ty]; !ok {
		typeData[ty] = NewStructData(ty, NoMap)
	}
	if ty.Kind() != reflect.Ptr {
		panic("Decode only accepts pointers!")
	}
	typeData[ty].DecodeTo(src, unsafe.Pointer(val.Pointer()))
}
