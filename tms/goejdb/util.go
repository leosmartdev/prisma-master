package goejdb

// #cgo LDFLAGS: -lejdb
// #include <ejdb/ejdb.h>
import "C"

import (
	"encoding/json"
	"unsafe"

	"github.com/globalsign/mgo/bson"
)

func bson_oid_from_string(oid *string) *C.bson_oid_t {
	if oid == nil {
		return nil
	}

	c_oid := C.CString(*oid)
	defer C.free(unsafe.Pointer(c_oid))

	ret := new(C.bson_oid_t)
	C.bson_oid_from_string(ret, c_oid)
	return ret
}

func bson_oid_to_string(oid *C.bson_oid_t) string {
	var c_str [25]C.char
	char_ptr := (*C.char)(unsafe.Pointer(&c_str))
	C.bson_oid_to_string(oid, char_ptr)
	return C.GoString(char_ptr)
}

func bson_to_byte_slice(bson *C.bson) []byte {
	size := int(C.bson_size(bson))
	data := C.bson_data(bson)
	ptr_data := (*[maxslice]byte)(unsafe.Pointer(data))
	ret := make([]byte, size)
	copy(ret, (*ptr_data)[:size])
	return ret
}

func bson_from_byte_slice(bsdata []byte) *C.bson {
	c_bson := new(C.bson)

	buff := C.malloc(C.size_t(len(bsdata)))
	ptr_buff := (*[maxslice]byte)(unsafe.Pointer(buff))
	for i := 0; i < len(bsdata); i++ {
		(*ptr_buff)[i] = bsdata[i]
	}

	C.bson_init_finished_data(c_bson, (*C.char)(buff))
	return c_bson
}

func bson_from_json(j string) *C.bson {
	var m map[string]interface{}
	json.Unmarshal(([]byte)(j), &m)
	bytes, _ := bson.Marshal(&m)
	return bson_from_byte_slice(bytes)
}
