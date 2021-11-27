package mongo

/******
 *  Mongo has a bunch of special data types (e.g. GeoJSON, timestamps) which we
 *  want to use. Also, we have some oddball types (e.g. DoubleValue) which
 *  convert to normal types (e.g. double). This file contains custom
 *  encoders/decoders for those data types.
 */
import (
	"prisma/tms"

	"C"
	"bytes"
	"fmt"
	"reflect"
	"time"
	"unsafe"

	"github.com/golang/protobuf/ptypes/wrappers"

	"github.com/globalsign/mgo/bson"
)

var (
	CustomWriters = map[reflect.Type]CustomWriter{
		// *** Mongo types
		reflect.TypeOf(bson.ObjectId("")): ObjectIdWriter,
		reflect.TypeOf(time.Time{}):       TimeWriter,

		// *** Wrapper types
		reflect.TypeOf(&wrappers.DoubleValue{}): DoubleValueWriter,
		reflect.TypeOf(&wrappers.FloatValue{}):  FloatValueWriter,
		reflect.TypeOf(&wrappers.Int64Value{}):  Int64ValueWriter,
		reflect.TypeOf(&wrappers.UInt64Value{}): UInt64ValueWriter,
		reflect.TypeOf(&wrappers.Int32Value{}):  Int32ValueWriter,
		reflect.TypeOf(&wrappers.UInt32Value{}): UInt32ValueWriter,
		reflect.TypeOf(&wrappers.BoolValue{}):   BoolValueWriter,
		reflect.TypeOf(&wrappers.StringValue{}): StringValueWriter,
		reflect.TypeOf(&wrappers.BytesValue{}):  BytesValueWriter,

		// *** Geography types
		reflect.TypeOf(&tms.Point{}): PointWriter,
	}

	CustomReaders = map[reflect.Type]CustomReader{
		// *** Mongo types
		reflect.TypeOf(bson.ObjectId("")): ObjectIdReader,
		reflect.TypeOf(time.Time{}):       TimeReader,

		// *** Wrapper types
		reflect.TypeOf(&wrappers.DoubleValue{}): DoubleValueReader,
		reflect.TypeOf(&wrappers.FloatValue{}):  FloatValueReader,
		reflect.TypeOf(&wrappers.Int64Value{}):  Int64ValueReader,
		reflect.TypeOf(&wrappers.UInt64Value{}): UInt64ValueReader,
		reflect.TypeOf(&wrappers.Int32Value{}):  Int32ValueReader,
		reflect.TypeOf(&wrappers.UInt32Value{}): UInt32ValueReader,
		reflect.TypeOf(&wrappers.BoolValue{}):   BoolValueReader,
		reflect.TypeOf(&wrappers.StringValue{}): StringValueReader,
		reflect.TypeOf(&wrappers.BytesValue{}):  BytesValueReader,

		// *** Geography types
		reflect.TypeOf(&tms.Point{}): PointReader,
	}
)

func TimeWriter(data []byte, b *bytes.Buffer, sf StructField) {
	ptr := unsafe.Pointer(&data[0])
	time := *(*time.Time)(ptr)
	writeFieldHeaderLL(b, 0x09, sf.bsonName)
	var unixMS int64 = time.UnixNano() / 1000000
	b.Write(toBytes(unsafe.Pointer(&unixMS), 8))
}

func TimeReader(raw bson.Raw, ptrToField unsafe.Pointer, sf StructField) uintptr {
	ptr := unsafe.Pointer(&raw.Data[0])
	ip := (*int64)(ptr)

	tp := (*time.Time)(ptrToField)
	if raw.Kind != 0x09 {
		panic("Expected time type for decoding to time.Time")
	}
	if len(raw.Data) < 8 {
		panic("Not enough data left for time!")
	}
	*tp = time.Unix((*ip / 1000), ((*ip)%1000)*1000000)
	return 8
}

func ObjectIdWriter(data []byte, b *bytes.Buffer, sf StructField) {
	if len(data) < 8 {
		panic("Not enough data left to constitute an ObjectID")
	}
	ptr := unsafe.Pointer(&data[0])
	oid := (*bson.ObjectId)(ptr)
	if oid.Valid() {
		writeFieldHeaderLL(b, 0x07, sf.bsonName)
		bytes := []byte(*oid)
		if len(bytes) != 12 {
			panic("invalid oid!")
		}
		b.Write(bytes)
	}
}

func ObjectIdReader(raw bson.Raw, ptrToField unsafe.Pointer, sf StructField) uintptr {
	oid := (*bson.ObjectId)(ptrToField)
	if raw.Kind != 0x07 {
		panic("Expected 0x07(ObjectId) to fill in an ObjectID!")
	}
	if len(raw.Data) < 12 {
		panic("Not enough data left to constitute an ObjectID")
	}
	*oid = bson.ObjectId(raw.Data[0:12])
	return 12
}

func DoubleValueWriter(data []byte, b *bytes.Buffer, sf StructField) {
	ptr := unsafe.Pointer(&data[0])
	dv := (**wrappers.DoubleValue)(ptr)
	if dv != nil && (*dv) != nil {
		dv := *dv
		writeFieldHeaderLL(b, 0x01, sf.bsonName)
		if DEBUG {
			fmt.Printf("Value: %v\n", dv.Value)
		}
		b.Write(toBytes(unsafe.Pointer(&dv.Value), 8))
	}
}

func DoubleValueReader(raw bson.Raw, ptrToField unsafe.Pointer, sf StructField) uintptr {
	ptr := unsafe.Pointer(&raw.Data[0])
	vp := (**wrappers.DoubleValue)(ptrToField)
	if *vp == nil {
		*vp = new(wrappers.DoubleValue)
	}
	v := *vp
	if raw.Kind != 0x01 {
		panic(fmt.Sprintf("Expected a double to fill DoubleValue, not 0x%x!\n\t%v\n\t%v", raw.Kind, sf, raw.Data))
	}
	v.Value = *(*float64)(ptr)
	return 8
}

func FloatValueWriter(data []byte, b *bytes.Buffer, sf StructField) {
	ptr := unsafe.Pointer(&data[0])
	vp := (**wrappers.FloatValue)(ptr)
	if vp != nil && (*vp) != nil {
		v := *vp
		writeFieldHeaderLL(b, 0x01, sf.bsonName)
		if DEBUG {
			fmt.Printf("Value: %v\n", v.Value)
		}
		dbl := float64(v.Value)
		b.Write(toBytes(unsafe.Pointer(&dbl), 8))
	}
}
func FloatValueReader(raw bson.Raw, ptrToField unsafe.Pointer, sf StructField) uintptr {
	ptr := unsafe.Pointer(&raw.Data[0])
	vp := (**wrappers.FloatValue)(ptrToField)
	if *vp == nil {
		*vp = new(wrappers.FloatValue)
	}
	v := *vp
	if raw.Kind != 0x01 {
		panic(fmt.Sprintf("Expected a double to fill FloatValue, not 0x%x!\n\t%v\n\t%v", raw.Kind, sf, raw.Data))
	}
	v.Value = float32(*(*float64)(ptr))
	return 4
}

func Int64ValueWriter(data []byte, b *bytes.Buffer, sf StructField) {
	ptr := unsafe.Pointer(&data[0])
	vp := (**wrappers.Int64Value)(ptr)
	if vp != nil && (*vp) != nil {
		v := *vp
		writeFieldHeaderLL(b, 0x12, sf.bsonName)
		if DEBUG {
			fmt.Printf("Value: %v\n", v.Value)
		}
		b.Write(toBytes(unsafe.Pointer(&v.Value), 8))
	}
}
func Int64ValueReader(raw bson.Raw, ptrToField unsafe.Pointer, sf StructField) uintptr {
	ptr := unsafe.Pointer(&raw.Data[0])
	vp := (**wrappers.Int64Value)(ptrToField)
	if *vp == nil {
		*vp = new(wrappers.Int64Value)
	}
	v := *vp
	if raw.Kind != 0x12 {
		panic(fmt.Sprintf("Expected an int64 to fill Int64Value, not 0x%x!\n\t%v\n\t%v", raw.Kind, sf, raw.Data))
	}
	v.Value = *(*int64)(ptr)
	return 8
}

func UInt64ValueWriter(data []byte, b *bytes.Buffer, sf StructField) {
	ptr := unsafe.Pointer(&data[0])
	vp := (**wrappers.UInt64Value)(ptr)
	if vp != nil && (*vp) != nil {
		v := *vp
		writeFieldHeaderLL(b, 0x12, sf.bsonName)
		if DEBUG {
			fmt.Printf("Value: %v\n", v.Value)
		}
		b.Write(toBytes(unsafe.Pointer(&v.Value), 8))
	}
}
func UInt64ValueReader(raw bson.Raw, ptrToField unsafe.Pointer, sf StructField) uintptr {
	ptr := unsafe.Pointer(&raw.Data[0])
	vp := (**wrappers.UInt64Value)(ptrToField)
	if *vp == nil {
		*vp = new(wrappers.UInt64Value)
	}
	v := *vp
	if raw.Kind != 0x01 {
		panic(fmt.Sprintf("Expected an int64 to fill UInt64Value, not 0x%x!\n\t%v\n\t%v", raw.Kind, sf, raw.Data))
	}
	v.Value = uint64(*(*int64)(ptr))
	return 8
}

func Int32ValueWriter(data []byte, b *bytes.Buffer, sf StructField) {
	ptr := unsafe.Pointer(&data[0])
	vp := (**wrappers.Int32Value)(ptr)
	if vp != nil && (*vp) != nil {
		v := *vp
		writeFieldHeaderLL(b, 0x10, sf.bsonName)
		if DEBUG {
			fmt.Printf("Value: %v\n", v.Value)
		}
		b.Write(toBytes(unsafe.Pointer(&v.Value), 4))
	}
}
func Int32ValueReader(raw bson.Raw, ptrToField unsafe.Pointer, sf StructField) uintptr {
	ptr := unsafe.Pointer(&raw.Data[0])
	vp := (**wrappers.Int32Value)(ptrToField)
	if *vp == nil {
		*vp = new(wrappers.Int32Value)
	}
	v := *vp
	if raw.Kind != 0x10 {
		panic(fmt.Sprintf("Expected an int32 to fill Int32Value, not 0x%x!\n\t%v\n\t%v", raw.Kind, sf, raw.Data))
	}
	v.Value = *(*int32)(ptr)
	return 4
}

func UInt32ValueWriter(data []byte, b *bytes.Buffer, sf StructField) {
	ptr := unsafe.Pointer(&data[0])
	vp := (**wrappers.FloatValue)(ptr)
	if vp != nil && (*vp) != nil {
		v := *vp
		writeFieldHeaderLL(b, 0x12, sf.bsonName)
		if DEBUG {
			fmt.Printf("Value: %v\n", v.Value)
		}
		i64 := int64(v.Value)
		b.Write(toBytes(unsafe.Pointer(&i64), 8))
	}
}
func UInt32ValueReader(raw bson.Raw, ptrToField unsafe.Pointer, sf StructField) uintptr {
	ptr := unsafe.Pointer(&raw.Data[0])
	vp := (**wrappers.UInt32Value)(ptrToField)
	if *vp == nil {
		*vp = new(wrappers.UInt32Value)
	}
	v := *vp
	if raw.Kind != 0x12 {
		panic(fmt.Sprintf("Expected an int64 to fill UInt32Value, not 0x%x!\n\t%v\n\t%v", raw.Kind, sf, raw.Data))
	}
	v.Value = uint32(*(*int64)(ptr))
	return 8
}

func BoolValueWriter(data []byte, b *bytes.Buffer, sf StructField) {
	ptr := unsafe.Pointer(&data[0])
	vp := (**wrappers.BoolValue)(ptr)
	if vp != nil && (*vp) != nil {
		v := *vp
		writeFieldHeaderLL(b, 0x08, sf.bsonName)
		if DEBUG {
			fmt.Printf("Value: %v\n", v.Value)
		}
		b.Write(toBytes(unsafe.Pointer(&v.Value), 1))
	}
}

func BoolValueReader(raw bson.Raw, ptrToField unsafe.Pointer, sf StructField) uintptr {
	ptr := unsafe.Pointer(&raw.Data[0])
	vp := (**wrappers.BoolValue)(ptrToField)
	if *vp == nil {
		*vp = new(wrappers.BoolValue)
	}
	v := *vp
	if raw.Kind != 0x08 {
		panic(fmt.Sprintf("Expected a bool to fill BoolValue, not 0x%x!\n\t%v\n\t%v", raw.Kind, sf, raw.Data))
	}
	v.Value = *(*bool)(ptr)
	return 1
}

func StringValueWriter(data []byte, b *bytes.Buffer, sf StructField) {
	ptr := unsafe.Pointer(&data[0])
	vp := (**wrappers.StringValue)(ptr)
	if vp != nil && (*vp) != nil {
		v := *vp
		writeFieldHeaderLL(b, 0x02, sf.bsonName)

		strbytes := []byte(v.Value)
		l := int32(len(strbytes)) + 1
		b.Write(toBytes(unsafe.Pointer(&l), 4))
		b.Write(strbytes)
		b.WriteByte(0)

		if DEBUG {
			fmt.Printf("Value: %v\n", v.Value)
		}
	}
}

func StringValueReader(raw bson.Raw, ptrToField unsafe.Pointer, sf StructField) uintptr {
	ptr := unsafe.Pointer(&raw.Data[0])
	vp := (**wrappers.StringValue)(ptrToField)
	if *vp == nil {
		*vp = new(wrappers.StringValue)
	}
	v := *vp
	if raw.Kind != 0x02 {
		panic(fmt.Sprintf("Expected a double to fill StringValue, not 0x%x!\n\t%v\n\t%v", raw.Kind, sf, raw.Data))
	}
	sz := (*int32)(ptr)
	str := string(raw.Data[4 : 3+*sz])
	v.Value = str
	return 4 + uintptr(*sz)
}

func BytesValueWriter(data []byte, b *bytes.Buffer, sf StructField) {
	ptr := unsafe.Pointer(&data[0])
	vp := (**wrappers.BytesValue)(ptr)
	if vp != nil && (*vp) != nil {
		v := *vp
		hdr := (*reflect.SliceHeader)(ptr)
		if hdr.Len > 0 {
			if sf.structData == nil {
				panic("Need a struct data to encode containers!")
			}
			sf.structData.writeFieldHeader(b, sf)
			elemTy := sf.field.Elem()

			sliceBytes := toBytes(unsafe.Pointer(&v.Value),
				elemTy.Size()*uintptr(hdr.Len))
			sf.structData.encodeArray(elemTy, sliceBytes, b)
		}
	}
}

func BytesValueReader(raw bson.Raw, ptrToField unsafe.Pointer, sf StructField) uintptr {
	ptr := unsafe.Pointer(&raw.Data[0])
	vp := (**wrappers.BytesValue)(ptrToField)
	if *vp == nil {
		*vp = new(wrappers.BytesValue)
	}
	v := *vp
	if raw.Kind != 0x08 {
		panic(fmt.Sprintf("Expected a []byte to fill BytesValue, not 0x%x!\n\t%v\n\t%v", raw.Kind, sf, raw.Data))
	}
	v.Value = *(*[]byte)(ptr)
	return sf.structData.decodeArray(raw, ptrToField)
}

// Write a Point coordinate to the byte buffer. This implementation is super
// ugly -- we're pre-computing the entire byte stream and dropping in the
// lat/long bytes. Yeah, we could make this prettier, but this will be the
// fastest implementation.
func PointWriter(data []byte, b *bytes.Buffer, sf StructField) {
	pptr := (*unsafe.Pointer)(unsafe.Pointer(&data[0]))
	if pptr != nil && *pptr != nil {
		ptr := *pptr
		longitudeOffset := unsafe.Offsetof(tms.Point{}.Longitude)
		latitudeOffset := unsafe.Offsetof(tms.Point{}.Latitude)

		pointbytes := toBytes(ptr, unsafe.Sizeof(tms.Point{}))
		writeFieldHeaderLL(b, 0x03, sf.bsonName)

		header := []byte{
			// *** struct length
			byte(pointOuterLength),
			byte(pointOuterLength >> 8),
			byte(pointOuterLength >> 16),
			byte(pointOuterLength >> 24),

			// *** type: "Point"
			0x02,               // string type
			't', 'y', 'p', 'e', // field name
			0x00,       // field name c_str sentinel
			6, 0, 0, 0, // string length (6) incl. \0
			'P', 'o', 'i', 'n', 't',
			0x00, // The \0

			// *** coordinates: [lat, long]
			0x04, // array type
			'c', 'o', 'o', 'r', 'd', 'i', 'n', 'a', 't', 'e', 's',
			0x00, // field name c_str sentinel

			// *** inner struct for coordinates array
			byte(pointInnerLength),
			byte(pointInnerLength >> 8),
			byte(pointInnerLength >> 16),
			byte(pointInnerLength >> 24),
		}
		b.Write(header)
		b.Write([]byte{
			// *** "0": longitude
			0x01,            // double type
			byte('0'), 0x00, //field name "0"
		})
		b.Write(pointbytes[longitudeOffset : longitudeOffset+8])
		b.Write([]byte{
			// *** "1": latitude
			0x01,            // double type
			byte('1'), 0x00, //field name "1"
		})
		b.Write(pointbytes[latitudeOffset : latitudeOffset+8])

		b.Write([]byte{
			// *** inner struct end sentinel
			0x00,

			// *** outer struct end sentinel
			0x00,
		})
	}
}

const (
	pointOuterLength = int32(45 + 8 + 8) // length of whole struct
	pointInnerLength = int32(11 + 8 + 8) // length of inner "coordinates" struct
)

var (
	typePointBytes = []byte{
		0x02,               // string type
		't', 'y', 'p', 'e', // field name
		0x00,       // field name c_str sentinel
		6, 0, 0, 0, // string length (6) incl. \0
		'P', 'o', 'i', 'n', 't',
		0x00, // The \0
	}

	coordHeader = []byte{
		// *** coordinates: [lat, long]
		0x04, // array type
		'c', 'o', 'o', 'r', 'd', 'i', 'n', 'a', 't', 'e', 's',
		0x00, // field name c_str sentinel

		// *** inner struct for coordinates array
		byte(pointInnerLength),
		byte(pointInnerLength >> 8),
		byte(pointInnerLength >> 16),
		byte(pointInnerLength >> 24),
	}
)

func PointReader(raw bson.Raw, ptrToField unsafe.Pointer, sf StructField) uintptr {
	// *** Sanity checks
	if raw.Kind != 0x03 {
		panic(fmt.Sprintf("Wrong bson type (0x%x)!", raw.Kind))
	}

	sz := uintptr(*(*int32)(unsafe.Pointer(&raw.Data[0])))
	if sz > uintptr(len(raw.Data)) {
		panic("Message corrupt: initial size extends beyond buffer")
	}

	if raw.Data[sz-1] != 0 {
		panic("Message corrupt: doesn't end with '0'")
	}
	// *** End sanity checks

	vp := (**tms.Point)(ptrToField)
	if *vp == nil {
		*vp = new(tms.Point)
	}
	v := *vp

	// Ditch outside raw.Data -- keep only our body
	raw.Data = raw.Data[4 : sz-1]
	gotType := false
	gotCoord := false
	for len(raw.Data) > 0 {
		if len(raw.Data) < 2 {
			panic("Message corrupt: " +
				"less than 2 bytes left before done decoding!")
		}

		if bytes.HasPrefix(raw.Data, typePointBytes) {
			// This is the type: "Point" field
			raw.Data = raw.Data[len(typePointBytes):]
			// We don't have to do anything with it, just verify it's here
			gotType = true
		} else if bytes.HasPrefix(raw.Data, coordHeader) {
			lngPtr := (*float64)(unsafe.Pointer(&raw.Data[len(coordHeader)+3]))
			v.Longitude = *lngPtr
			latPtr := (*float64)(unsafe.Pointer(&raw.Data[len(coordHeader)+14]))
			v.Latitude = *latPtr
			gotCoord = true
			raw.Data = raw.Data[len(coordHeader)+23:]
		} else {
			name := C.GoString((*C.char)(unsafe.Pointer(&raw.Data[1])))
			panic(fmt.Sprintf(
				"Found unexpected field '%v' in Point struct!", name))
		}
	}

	if !gotType {
		panic("Couldn't find type field in point!")
	}

	if !gotCoord {
		panic("Couldn't find type field in point!")
	}

	return sz
}
