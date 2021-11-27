package mongo

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
	"unsafe"

	"prisma/tms/log"

	"github.com/globalsign/mgo/bson"
)

/**
 * Using the cached reflection information in StrucData, encode an object into
 * bson. In general, panic() if something goes wrong.
 */

const (
	DEBUG = false
)

// Try to convert 'obj' to the correct type by getting pointers to it or
// resolving a pointer/interface, then encode it. panic()s if anything goes
// wrong.
func (sd *StructData) EncodeIface(obj interface{}) bson.Raw {
	val := reflect.ValueOf(obj).Elem()
	for {
		ty := val.Type()
		switch ty.Kind() {
		case reflect.Ptr:
			if val.IsNil() {
				panic("Cannot encode nil pointer!")
			}
			if ty.Elem() != sd.goType {
				panic("Underlying type does not match the type in StructData")
			}
			return sd.Encode(unsafe.Pointer(val.Pointer()))
		case reflect.Interface:
			if val.IsNil() {
				panic("Cannot encode nil interface!")
			}
			val = val.Elem()
		default:
			val = val.Addr()
		}
	}
}

// Encode an object pointed to by ptr. Assume the type of this object is
// StructData.goType
func (sd *StructData) Encode(ptr unsafe.Pointer) bson.Raw {
	return sd.encode(nil, ptr)
}

// Encode an object to bson pointed to by ptr. Assume the type of this object
// is StructData.goType. If a Coder is specified, store some additional
// metadata in the coder.
func (sd *StructData) encode(c *Coder, ptr unsafe.Pointer) bson.Raw {
	if DEBUG {
		fmt.Printf("sd.Encode: %v\n", ptr)
	}
	b := bytes.NewBuffer(make([]byte, 0, sd.lastSize))
	sd.encodePtr(c, ptr, b)
	if DEBUG {
		fmt.Printf("Encoded: %v\n", log.Spew(b.Bytes()))
	}
	bytes := b.Bytes()
	ret := bson.Raw{
		Kind: 0x03,
		Data: bytes,
	}

	return ret
}

// Encode this object to bson, then decode it to a bson.D map. This is a pretty
// ugly, wasteful thing to do, but it's an easy way to do things like construct
// mongo queries and modify them before use.
func (sd *StructData) EncodeIfaceToMap(obj interface{}) bson.D {
	val := reflect.ValueOf(obj).Elem()
	for {
		ty := val.Type()
		switch ty.Kind() {
		case reflect.Ptr:
			if val.IsNil() {
				panic("Cannot encode nil pointer!")
			}
			if ty.Elem() != sd.goType {
				panic("Underlying type does not match the type in StructData")
			}
			return sd.EncodeToMap(unsafe.Pointer(val.Pointer()))
		case reflect.Interface:
			if val.IsNil() {
				panic("Cannot encode nil interface!")
			}
			val = val.Elem()
		default:
			val = val.Addr()
		}
	}
}

// Encode this object to bson, then decode it to a bson.D map. Assumes ptr is
// pointing to something of type StructData.goType. This is a pretty ugly,
// wasteful thing to do, but it's an easy way to do things like construct mongo
// queries and modify them before use.
func (sd *StructData) EncodeToMap(ptr unsafe.Pointer) bson.D {
	raw := sd.Encode(ptr)
	var ret bson.D
	err := bson.Unmarshal(raw.Data, &ret)
	if err != nil {
		panic(fmt.Sprintf("Error unmarshaling encoded object: %v", err))
	}
	return ret
}

// Get a slice of data pointing to 'ptr' and of size 'sz'
func toBytes(ptr unsafe.Pointer, sz uintptr) []byte {
	return (*[1 << 30]byte)(ptr)[:sz:sz]
}

// Encode an object of type StructData.goType pointed to by 'ptr' into 'b'
func (sd *StructData) encodePtr(c *Coder, ptr unsafe.Pointer, b *bytes.Buffer) {
	if DEBUG {
		fmt.Printf("sd.encodePtr: %v %v\n", ptr, sd.goType.Size())
	}
	sd.encodeBytes(c, toBytes(ptr, sd.goType.Size()), b)
}

// Encode an object contained in 'data' of type StructData.goType into buffer 'b'
func (sd *StructData) encodeBytes(c *Coder, data []byte, b *bytes.Buffer) {
	if sd.flatten {
		panic("It's already too late to flatten!")
	}

	if DEBUG {
		fmt.Printf("sd.encodeBytes: %v\n", &data[0])
	}

	// *** Header
	// Is just a length, but we don't know the length yet. Write zeros to
	// reserve space for it, then fill them in later.
	startlen := b.Len()
	b.Write([]byte{0, 0, 0, 0})

	// *** Body
	sd.writeBody(c, data, b)

	// *** Footer
	b.WriteByte(0)
	mylen := b.Len() - startlen
	if DEBUG {
		fmt.Printf("sd.encodeBytes footer: %v\n", mylen)
	}
	// Now that we know the byte length of the body, replace the zeros we wrote
	// above with the actual length.
	bytes := b.Bytes()[startlen:]
	binary.LittleEndian.PutUint32(bytes[0:4], uint32(mylen))
	// Cache the last length here so that we can do a better job pre-allocating
	// space ahead of time, next time. This is a pretty minor optimization.
	sd.lastSize = uint32(mylen)
}

// Iterate through all of a struct's fields and write out each one into buffer
// 'b'
func (sd *StructData) writeBody(c *Coder, data []byte, b *bytes.Buffer) {
	// *** Body
	for _, sf := range sd.orderedFields {
		name := sf.bsonName
		if c != nil && sf.typeByte == 0x07 && name == "_id" {
			// Coder wants to see _id fields
			ptr := unsafe.Pointer(&data[0])
			c.LastID = *(*bson.ObjectId)(ptr)
		}

		// Do the actual field encoding
		sfdata := data[sf.fieldOffset:]
		sd.encodeField(sfdata, b, *sf, false)
	}
}

// Write a field header (the bson type and field name) into buffer 'b'
func writeFieldHeaderLL(b *bytes.Buffer, ty byte, name string) {
	if DEBUG {
		fmt.Printf("writeFieldHeader: %v %v\n", ty, name)
	}
	// Type and name (HEADER)
	b.WriteByte(ty)
	b.Write([]byte(name))
	b.WriteByte(0)
}

func (sd *StructData) writeFieldHeader(b *bytes.Buffer, sf StructField) {
	writeFieldHeaderLL(b, sf.typeByte, sf.bsonName)
}

// Encode a field into a bson stream buffer. WARNING: This function assumes we
// are running on a LITTLE ENDIAN MACHINE.
func (sd *StructData) encodeField(
	data []byte, b *bytes.Buffer, sf StructField, force bool) {

	// Is this a custom type?
	if sf.write != nil {
		// Custom writer
		sf.write(data, b, sf)
		return
	}
	// Get a pointer to the data beginning
	ptr := unsafe.Pointer(&data[0])
	if DEBUG {
		fmt.Printf("encodeField: %v %s\n", ptr, log.Spew(sf))
	}

	// What type of data does this field contain? Most of these cases are
	// pretty straightforward and similar to each other, so only a few are
	// documented
	switch sf.field.Kind() {
	case reflect.Ptr:
		vptr := (*unsafe.Pointer)(ptr)
		if *vptr != nil {
			// If the type is a pointer and it's not nil, dereference it and
			// then write it out
			effsf := sf
			effsf.field = sf.field.Elem()
			if DEBUG {
				fmt.Printf("vptr: %p\n", *vptr)
			}
			sd.encodeField(
				toBytes(
					*vptr,
					effsf.field.Size()),
				b, effsf, true)
		} else if force {
			// It we _need_ to write this field, but it's null, then print a null!
			writeFieldHeaderLL(b, 0x0A, sf.bsonName)
		}
	case reflect.Interface:
		// TODO: This reflection stuff is probably pretty slow. I'm using it
		// because I can't figure out the memory layout for interface fields.
		// It's not documented, so it's probably not safe to to byte
		// manipulation. We'll stick with this for now.
		valptr := reflect.NewAt(sf.field, ptr)
		val := valptr.Elem()
		if DEBUG {
			fmt.Printf("encodeField Interface: %+v (%v)\n", val, val.Type())
		}
		// *** Sanity checks
		if !val.IsValid() {
			panic("For some reason this field is not valid. That's strange!")
		}
		if val.IsNil() {
			return // Do nothing for nil fields
		}
		if val.Type().Kind() != reflect.Interface {
			panic("This should be an interface")
		}
		if sf.oneofInvolved && sf.oneof == nil {
			// This is a oneof possible struct, not the actual field
			return
		}
		// *** End sanity checks

		val = val.Elem()
		if val.IsValid() && !val.IsNil() {
			if DEBUG {
				fmt.Printf("encodeField lookup: %+v (%v)\n", val, val.Type())
			}
			if sf.oneofInvolved {
				// Do we actually need to use a different name?
				var ok bool
				realsf, ok := sf.oneof[val.Type()]
				if !ok {
					panic(fmt.Sprintf("Could not find real field!\n%v\n%v\n%v", log.Spew(val), log.Spew(sf), log.Spew(sd)))
				}
				sf = *realsf
			}
			if val.Type().Kind() != reflect.Ptr {
				val = val.Addr()
			}
			if DEBUG {
				fmt.Printf("encodeField val: (%v) %v\n", val.Type(), val)
			}
			realbytes := toBytes(
				unsafe.Pointer(val.Pointer()),
				val.Type().Elem().Size())
			if sf.structData.flatten {
				sf.structData.writeBody(nil, realbytes, b)
			} else {
				sd.writeFieldHeader(b, sf)
				sf.structData.encodeBytes(nil, realbytes, b)
			}
		}

	case reflect.Struct:
		// Write out a struct. Simple: write the obj header then all the fields
		if sf.structData == nil {
			panic(fmt.Sprintf("Need struct data to encode struct! (%+v)", sf))
		}
		sd.writeFieldHeader(b, sf)
		sf.structData.encodeBytes(nil, data, b)

	case reflect.Map:
		panic(fmt.Sprintf("Map Type not yet supported (%+v)", sf))

	case reflect.Array:
		valptr := reflect.NewAt(sf.field, ptr)
		val := valptr.Elem()
		if DEBUG {
			fmt.Printf("Array: %+v %v\n", val.Type(), val)
		}
		panic(fmt.Sprintf("Array Type not yet supported (%+v)", sf))

	case reflect.Slice:
		hdr := (*reflect.SliceHeader)(ptr)
		if hdr.Len > 0 {
			sd.writeFieldHeader(b, sf)
			elemTy := sf.field.Elem()

			if sf.structData == nil {
				panic("Need a struct data to encode containers!")
			}
			sliceBytes := toBytes(unsafe.Pointer(hdr.Data),
				elemTy.Size()*uintptr(hdr.Len))
			sf.structData.encodeArray(elemTy, sliceBytes, b)
		}

	case reflect.String:
		strptr := (*string)(ptr)
		if *strptr != "" || force {
			// If the string value is not "" or we are forcing it to be written...
			sd.writeFieldHeader(b, sf) // Write the 'string' header

			strbytes := []byte(*strptr)
			l := int32(len(strbytes)) + 1
			b.Write(toBytes(unsafe.Pointer(&l), 4)) // Write the str len
			b.Write(strbytes)                       // Then the data
			b.WriteByte(0)                          // Then a 0x00 for c-compat
		}

	case reflect.Bool:
		sd.writeFieldHeader(b, sf)
		vptr := (*bool)(ptr)
		if *vptr {
			b.WriteByte(0x01)
		} else {
			b.WriteByte(0x00)
		}

	case reflect.Int:
		vptr := (*int)(ptr)
		if *vptr != 0 || force {
			// Output 32-bit int
			sd.writeFieldHeader(b, sf)
			b.Write(data[0:4])
		}
	case reflect.Int8:
		vptr := (*int8)(ptr)
		if *vptr != 0 || force {
			// Output 32-bit int
			sd.writeFieldHeader(b, sf)
			b.Write(data[0:1])
			b.Write([]byte{0, 0, 0})
		}
	case reflect.Int16:
		vptr := (*int16)(ptr)
		if *vptr != 0 || force {
			// Output 32-bit int
			sd.writeFieldHeader(b, sf)
			b.Write(data[0:2])
			b.Write([]byte{0, 0})
		}
	case reflect.Int32:
		vptr := (*int32)(ptr)
		if *vptr != 0 || force {
			if s, ok := sf.enumR[*vptr]; ok {
				// It's an enum!
				sf.typeByte = 0x02
				sd.writeFieldHeader(b, sf)

				strbytes := []byte(s)
				l := int32(len(strbytes)) + 1
				b.Write(toBytes(unsafe.Pointer(&l), 4))
				b.Write(strbytes)
				b.WriteByte(0)
			} else {
				// Output 32-bit int
				sd.writeFieldHeader(b, sf)
				b.Write(data[0:4])
			}
		}

	case reflect.Int64:
		vptr := (*int64)(ptr)
		if *vptr != 0 || force {
			// Output 32-bit int
			sd.writeFieldHeader(b, sf)
			b.Write(data[0:8])
		}

	case reflect.Uint:
		vptr := (*uint)(ptr)
		if *vptr != 0 || force {
			// Output 64-bit int since this is 32 bit signed
			sd.writeFieldHeader(b, sf)
			b.Write(data[0:4])
			b.Write([]byte{0, 0, 0, 0})
		}
	case reflect.Uint8:
		vptr := (*uint8)(ptr)
		if *vptr != 0 || force {
			// Output 32-bit int
			sd.writeFieldHeader(b, sf)
			b.Write(data[0:1])
			b.Write([]byte{0, 0, 0})
		}
	case reflect.Uint16:
		vptr := (*uint8)(ptr)
		if *vptr != 0 || force {
			// Output 32-bit int
			sd.writeFieldHeader(b, sf)
			b.Write(data[0:2])
			b.Write([]byte{0, 0})
		}

	case reflect.Uint32:
		vptr := (*uint32)(ptr)
		if *vptr != 0 || force {
			// Output 64-bit int since this is 32 bit signed
			sd.writeFieldHeader(b, sf)
			b.Write(data[0:4])
			b.Write([]byte{0, 0, 0, 0})
		}
	case reflect.Uint64:
		vptr := (*uint64)(ptr)
		if *vptr != 0 || force {
			// Output 64-bit int
			sd.writeFieldHeader(b, sf)
			b.Write(data[0:8])
		}
	case reflect.Uintptr:
		vptr := (*uint64)(ptr)
		if *vptr != 0 || force {
			// Output 64-bit int since this is 32 bit signed
			sd.writeFieldHeader(b, sf)
			b.Write(data[0:8])
		}

	case reflect.Float32:
		vptr := (*float32)(ptr)
		if *vptr != 0.0 || force {
			dbl := float64(*vptr) // upcast
			sd.writeFieldHeader(b, sf)
			b.Write(toBytes(unsafe.Pointer(&dbl), 8))
		}
	case reflect.Float64:
		vptr := (*float64)(ptr)
		if *vptr != 0.0 || force {
			sd.writeFieldHeader(b, sf)
			b.Write(data[0:8])
		}

	default:
		panic("Unknown kind!")
	}
}

func (sd *StructData) encodeArray(
	fieldty reflect.Type, data []byte, b *bytes.Buffer) {

	if sd.containerType == nil {
		panic("Trying to encode container type without containterType info!")
	}

	if DEBUG {
		fmt.Printf("encodeArray: %v, data:\n%v\n", fieldty, log.Spew(data))
	}

	// *** Header
	startlen := b.Len()
	b.Write([]byte{0, 0, 0, 0})

	// *** Body
	elemsz := fieldty.Size()
	sf := *sd.containerType
	i := 0
	for len(data) > 0 {
		sf.bsonName = fmt.Sprintf("%v", i)
		sd.encodeField(data, b, sf, true)
		data = data[elemsz:]
		i++
	}

	// *** Footer
	b.WriteByte(0)
	mylen := b.Len() - startlen
	if DEBUG {
		fmt.Printf("sd.encodeArray footer: %v\n", mylen)
	}
	bytes := b.Bytes()[startlen:]
	binary.LittleEndian.PutUint32(bytes[0:4], uint32(mylen))
	sd.lastSize = uint32(mylen)
}
