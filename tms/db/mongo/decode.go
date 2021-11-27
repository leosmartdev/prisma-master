package mongo

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	"prisma/tms/log"

	"github.com/globalsign/mgo/bson"
)

func (sd *StructData) DecodeTo(raw bson.Raw, ptr unsafe.Pointer) uintptr {
	return sd.decodeTo(nil, raw, ptr)
}

func (sd *StructData) decodeTo(c *Coder, raw bson.Raw, ptr unsafe.Pointer) uintptr {
	data := raw.Data
	// **** Sanity checks
	if ptr == nil {
		panic("Cannot decode to NIL pointer!")
	}

	// check data CONV-663
	if raw.Data == nil {
		panic(fmt.Sprintf("mongo.decode Message corrupt: nil data raw.Data=nil raw.Kind=%v raw=%v\n",
			raw.Kind, raw))
	}

	sz := uintptr(*(*int32)(unsafe.Pointer(&data[0])))
	if sz > uintptr(len(data)) {
		panic("Message corrupt: initial size extends beyond buffer")
	}

	if data[sz-1] != 0 {
		panic("Message corrupt: doesn't end with '0'")
	}
	//  *** End sanity checks

	// Ditch outside data -- keep only our body
	data = data[4 : sz-1]

	if DEBUG {
		fmt.Printf("DecodeTo '%v' sz: %v bytes: %v\n",
			sd.goType, sz, log.Spew(data))
	}

	for len(data) > 0 {
		if len(data) < 2 {
			panic("Message corrupt: less than 2 bytes left before done decoding!")
		}

		kind := data[0]
		//name := C.GoString((*C.char)(unsafe.Pointer(&data[1])))
		var name string
		cString(&name, data[1:])
		sf, ok := sd.fields[name]
		beginByte := uintptr(2 + len(name))
		if DEBUG {
			fmt.Printf("next field: %v '%v' %v\n%s",
				kind, name, ok, log.Spew(data))
		}
		var decodedBytes uintptr
		var remove uintptr
		// this condition is made to skip fields that have prefix name xxx_
		if !strings.HasPrefix(name, "xxx_") {
			if !ok {
				// Hmmm... Issue a warning here? A panic is too much.
				decodedBytes = sd.skipField(
					bson.Raw{
						Kind: kind,
						Data: data[beginByte:],
					}, name)
			} else {
				decodedBytes =
					sd.decodeField(
						bson.Raw{
							Kind: kind,
							Data: data[beginByte:],
						},
						unsafe.Pointer(uintptr(ptr)+sf.fieldOffset),
						*sf)
			}

			if c != nil && kind == 0x07 && name == "_id" {
				// Coder wants to know the "_id"
				c.LastID = bson.ObjectId(data[beginByte : beginByte+12])
			}

			remove = beginByte + decodedBytes
			if DEBUG {
				fmt.Printf("decodeField used %v bytes, loping off %v\n",
					decodedBytes, remove)
			}
		}

		if len(data) < int(remove) {
			data = data[len(data):]
		} else {
			data = data[remove:]
		}
	}
	return sz
}

func cString(dst *string, data []byte) {
	// Return a string pointing to a c-string at the start of 'data'
	slen := 0
	for data[slen] != 0 && slen < len(data) {
		slen++
	}
	hdr := (*reflect.StringHeader)(unsafe.Pointer(dst))
	hdr.Data = uintptr(unsafe.Pointer(&data[0]))
	hdr.Len = slen
}

func (sd *StructData) skipField(raw bson.Raw, name string) uintptr {
	if DEBUG {
		fmt.Printf("skipField: 0x%x %v\n%v\n%v\n",
			raw.Kind, name, log.Spew(raw.Data), log.Spew(sd))
	}

	valptr := unsafe.Pointer(&raw.Data[0])
	// Do different stuff depending on the bson type
	switch raw.Kind {

	// BSON Double
	case 0x01:
		return 8

	// BSON string
	case 0x02:
		sz := (*int32)(valptr)
		return 4 + uintptr(*sz)

	// BSON embedded document
	case 0x03:
		sz := (*int32)(valptr)
		return uintptr(*sz)

	// Mongo ObjectID
	case 0x07:
		return 12

	// UTC timestamp
	case 0x09:
		return 8
	// BSON null
	case 0x0A:
		return 0

	// BSON 32-bit int
	case 0x10:
		return 4

	// BSON 64-bit int
	case 0x12:
		return 8

	// Unknown type
	default:
		panic(fmt.Sprintf(
			"Encountered unsupported BSON type (0x%x) in field %v, %+v",
			raw.Kind, name, sd))
	}
}

func (sd *StructData) decodeFlattened(
	raw bson.Raw, ptr unsafe.Pointer, sf StructField) uintptr {

	// This struct is transparent. We need to get a handle to its subfield
	// then pass this call through to it.
	if len(sd.fields) != 1 {
		// This can be fixed later on, but there probably isn't a use case
		// for flattened structs with more than one field, esp. since field
		// names could then become ambiguous
		panic("Can only decode 'flattened' struct with one field")
	}
	for _, sf := range sd.fields {
		// This body should only execute once due to "if" above

		// get a pointer to the field
		fieldptr := unsafe.Pointer(uintptr(ptr) + sf.fieldOffset)
		ret := sd.decodeField(raw, fieldptr, *sf)
		return ret
	}
	panic("Could not find field in flattened data structure!")
}

// Decode bson data to a field.
// Ideally, we'd break this out into a number of functions. Go's code
// optimizer, however, has been reported to not inline switch cases or panic
// functions. So making a bunch of small functions would have some overhead we
// want to avoid. SOOO, long function it is! I apologize.
func (sd *StructData) decodeField(
	raw bson.Raw, ptrToField unsafe.Pointer, sf StructField) uintptr {

	data := raw.Data
	kind := raw.Kind
	if DEBUG {
		fmt.Printf("decodeField: 0x%x %v\n%v\n",
			kind, sf.bsonName, log.Spew(data))
	}
	if sf.read != nil {
		return sf.read(raw, ptrToField, sf)
	}

	// Do we need to allocate this pointer?
	if sf.field.Kind() == reflect.Ptr {
		// It's important that we use **struct{} instead of *uintptr. When we
		// set *fptr and it's a **struct{}, go knows it's a pointer we're
		// setting and tells the GC about this pointer. With *uintptr, it
		// doesn't tell the GC.
		fptr := (**struct{})(ptrToField)
		if *fptr == nil {
			// And it's nil. Gotta allocate
			newPtr := reflect.New(sf.field.Elem())
			*fptr = (*struct{})(unsafe.Pointer(newPtr.Pointer()))
		}
		ptrToField = unsafe.Pointer(*fptr)
		sf.field = sf.field.Elem()
	} else if sf.field.Kind() == reflect.Interface {
		fieldPtr := reflect.NewAt(sf.field, ptrToField)
		field := fieldPtr.Elem()
		if sf.structData == nil {
			panic("Need struct data to decode interface field!")
		}
		newVal := reflect.New(sf.structData.goType)
		field.Set(newVal)
		ptrToField = unsafe.Pointer(newVal.Pointer())
		sf.field = sf.structData.goType
	}

	// Was this type originally flattened?
	if sf.structData != nil && sf.structData.flatten {
		// If so, we need to handle it appropriately
		return sf.structData.decodeFlattened(raw, ptrToField, sf)
	}

	var valptr unsafe.Pointer
	if len(data) > 0 {
		valptr = unsafe.Pointer(&data[0])
	}

	// Do different stuff depending on the bson type
	switch kind {

	// BSON Double
	case 0x01:
		vp := (*float64)(valptr)
		// Switch based on target type
		switch sf.field.Kind() {
		case reflect.Float64:
			*(*float64)(ptrToField) = *vp
		case reflect.Float32:
			*(*float32)(ptrToField) = float32(*vp)
		default:
			panic(fmt.Sprintf(
				"Expected a float to decode to a float! Got a %v instead.",
				sf.field))
		}
		return 8

	// BSON string
	case 0x02:
		sz := (*int32)(valptr)
		str := string(data[4 : 3+*sz])
		// Switch based on target type
		switch sf.field.Kind() {
		case reflect.Int32:
			// This may be an enum
			if i, ok := sf.enum[str]; ok {
				*(*int32)(ptrToField) = i
			} else {
				panic(fmt.Sprintf("Could not resolve '%v' to enum value. "+
					"StructField.enum: %+v", str, sf.enum))
			}

		case reflect.String:
			*(*string)(ptrToField) = str
		default:
			panic(fmt.Sprintf(
				"Expected a string to decode to a string! Got a %v instead.",
				sf.field))
		}
		return 4 + uintptr(*sz)

	// BSON embedded document
	case 0x03:
		sz := (*int32)(valptr)
		if sf.structData == nil {
			panic("Need struct data to decode!")
		}

		if sf.field.Kind() == reflect.Struct {
			ret := sf.structData.DecodeTo(raw, ptrToField)
			if ret != uintptr(*sz) {
				panic("DecodeTo size doesn't match expectation")
			}
		} else {
			panic("Unsupported target type for document decoding")
		}

		return uintptr(*sz)

	// BSON array
	case 0x04:
		if sf.structData == nil {
			panic("Need struct data to array!")
		}
		return sf.structData.decodeArray(raw, ptrToField)

	// BSON bool
	case 0x08:
		vp := (*byte)(valptr)
		// Switch based on target type
		switch sf.field.Kind() {
		case reflect.Bool:
			*(*bool)(ptrToField) = (*vp != 0)
		default:
			panic(fmt.Sprintf(
				"Expected an int64 to decode to some sort of number!"+
					" Got a %v instead.",
				sf.field))
		}
		return 1

	// BSON Null
	case 0x0A:
		return 0

	// BSON 32-bit int
	case 0x10:
		vp := (*int32)(valptr)
		// Switch based on target type
		switch sf.field.Kind() {
		case reflect.Int64:
			*(*int64)(ptrToField) = int64(*vp)
		case reflect.Uint64:
			*(*uint64)(ptrToField) = uint64(*vp)
		case reflect.Int32:
			*(*int32)(ptrToField) = int32(*vp)
		case reflect.Uint32:
			*(*uint32)(ptrToField) = uint32(*vp)
		case reflect.Int16:
			*(*int16)(ptrToField) = int16(*vp)
		case reflect.Uint16:
			*(*uint16)(ptrToField) = uint16(*vp)
		case reflect.Int8:
			*(*int8)(ptrToField) = int8(*vp)
		case reflect.Uint8:
			*(*uint8)(ptrToField) = uint8(*vp)
		case reflect.Int:
			*(*int)(ptrToField) = int(*vp)
		case reflect.Uint:
			*(*uint)(ptrToField) = uint(*vp)
		default:
			panic(fmt.Sprintf(
				"Expected an int64 to decode to some sort of number!"+
					" Got a %v instead.",
				sf.field))
		}
		return 4

	// BSON 64-bit int
	case 0x12:
		vp := (*int64)(valptr)
		// Switch based on target type
		switch sf.field.Kind() {
		case reflect.Int64:
			*(*int64)(ptrToField) = *vp
		case reflect.Uint64:
			*(*uint64)(ptrToField) = uint64(*vp)
		case reflect.Int32:
			*(*int32)(ptrToField) = int32(*vp)
		case reflect.Uint32:
			*(*uint32)(ptrToField) = uint32(*vp)
		case reflect.Int16:
			*(*int16)(ptrToField) = int16(*vp)
		case reflect.Uint16:
			*(*uint16)(ptrToField) = uint16(*vp)
		case reflect.Int8:
			*(*int8)(ptrToField) = int8(*vp)
		case reflect.Uint8:
			*(*uint8)(ptrToField) = uint8(*vp)
		case reflect.Int:
			*(*int)(ptrToField) = int(*vp)
		case reflect.Uint:
			*(*uint)(ptrToField) = uint(*vp)
		default:
			panic(fmt.Sprintf(
				"Expected an int64 to decode to some sort of number!"+
					" Got a %v instead for field %v.",
				sf.field, sf.bsonName))
		}
		return 8

	// Unknown type
	default:
		panic(fmt.Sprintf(
			"Encountered HERE unsupported BSON type (0x%x) as field '%v'",
			kind, sf.bsonName))
	}
}

func (sd *StructData) decodeArray(raw bson.Raw, ptr unsafe.Pointer) uintptr {
	// TODO: This method uses a bunch of reflection stuff. This is to ensure
	// that we get this correct, rather than risk misunderstanding memory
	// allocation issues. I don't think this method is used by
	// performance-sensitive decoders, so there's no huge need for speed. If,
	// however, it ends up being a bottleneck, we'll have to re-think it.

	data := raw.Data
	// **** Sanity checks
	if ptr == nil {
		panic("Cannot decode to NIL pointer!")
	}

	if sd.containerType == nil {
		panic("Need containerType to decode array!")
	}

	sz := uintptr(*(*int32)(unsafe.Pointer(&data[0])))
	if sz > uintptr(len(data)) {
		panic("Message corrupt: initial size extends beyond buffer")
	}

	if data[sz-1] != 0 {
		panic("Message corrupt: doesn't end with '0'")
	}
	//  *** End sanity checks

	// Ditch outside data -- keep only our body
	data = data[4 : sz-1]

	if DEBUG {
		fmt.Printf("decodeAray '%v' sz: %v bytes: %v\n",
			sd.goType, sz, log.Spew(data))
	}

	sf := *sd.containerType
	arrayVal := reflect.NewAt(sd.goType, ptr).Elem()
	if sd.goType.Kind() == reflect.Slice {
		arrayVal.Set(reflect.MakeSlice(sd.goType, 0, 1))
	} else {
		panic("Unsupported container type")
	}

	i := 0
	for len(data) > 0 {
		if len(data) < 2 {
			panic("Message corrupt: less than 2 bytes left before done decoding!")
		}

		if sd.goType.Kind() == reflect.Slice && arrayVal.Cap() <= i {
			// We don't have the capacity. Increase the size of the slice

			// Keep old slice
			oldArray := reflect.New(arrayVal.Type()).Elem()
			oldArray.Set(arrayVal)

			// Allocate new, bigger array
			arrayVal.Set(reflect.MakeSlice(sd.goType, oldArray.Len(),
				oldArray.Cap()*2))
			// Copy the contents of the old array
			for j := 0; j < oldArray.Len(); j++ {
				arrayVal.Index(j).Set(oldArray.Index(j))
			}
		}

		kind := data[0]
		var name string
		cString(&name, data[1:])
		sf.bsonName = name

		// Increase length
		arrayVal.SetLen(i + 1)
		ptrToField := unsafe.Pointer(arrayVal.Index(i).UnsafeAddr())

		var decodedBytes uintptr
		decodedBytes =
			sd.decodeField(
				bson.Raw{
					Kind: kind,
					Data: data[2+len(name):],
				},
				ptrToField,
				sf)
		remove := 2 + uintptr(len(name)) + decodedBytes
		if DEBUG {
			fmt.Printf("decodeArray used %v bytes, loping off %v\n",
				decodedBytes, remove)
		}
		data = data[remove:]
		i++
	}
	return sz
}
