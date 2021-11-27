package mongo

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"unsafe"

	"prisma/tms"
	"prisma/tms/devices"
	"prisma/tms/nmea"

	"github.com/globalsign/mgo/bson"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/wrappers"
)

var (
	SampleNMEA = "!SAVDM,1,1,1,B,35?pE`5002o=>WPJH<s993LV0000,0*6A\r\n"
	VDM        = "VDM"

	jpb = jsonpb.Marshaler{
		EnumsAsInts:  false,
		EmitDefaults: false,
		Indent:       "  ",
	}
)

func generateTargets(N int) []tms.Target {
	ret := make([]tms.Target, N)
	sample := "\u0031\u0032\u0033\u0034\u0035\u0036"
	state := make([]byte, len(sample))
	for i := 0; i < len(sample); i++ {
		state[i] = sample[i]
	}
	fmt.Printf("State is : %v", state)
	fmt.Println("    " + sample)

	for i := 0; i < N; i++ {
		ret[i] = tms.Target{
			Id: &tms.TargetID{
				Producer: &tms.SensorID{
					Site: uint32(i % 15),
					Eid:  uint32(1000 + (i % 10)),
				},
				SerialNumber: &tms.TargetID_TimeSerial{
					TimeSerial: &tms.TimeSerialNumber{
						Seconds: int64(rand.Uint64()),
						Counter: int32(rand.Uint32() % 500),
					},
				},
			},
			Type: devices.DeviceType_TV32,

			Position: &tms.Point{
				Latitude:  (rand.Float64() * 180.0) - 90.0,
				Longitude: (rand.Float64() * 360.0) - 180.0,
			},
			Course: &wrappers.DoubleValue{
				Value: rand.Float64() * 360.0,
			},
			Heading: &wrappers.DoubleValue{
				Value: rand.Float64() * 360.0,
			},
			Speed: &wrappers.DoubleValue{
				Value: rand.Float64() * 50.0,
			},
			RateOfTurn: &wrappers.DoubleValue{
				Value: rand.Float64() * 50.0,
			},

			Nmea: &nmea.Nmea{
				Vdm: &nmea.Vdm{
					M1371: &nmea.M1371{
						MessageId: 3,
						Mmsi:      352196000,
						Pos: &nmea.M1371_Position{
							PositionAccuracy:   true,
							Latitude:           27659500,
							TrueHeading:        110,
							RaimFlag:           false,
							NavigationalStatus: 5,
							SpeedOverGround:    2,
							Longitude:          -73763600,
							TimeStamp:          19,
							CourseOverGround:   2340,
							CommState:          state,
						},
					},
				},
				OriginalString: SampleNMEA,
				Format:         VDM,
			},
		}
	}

	return ret
}

func BenchmarkOldEncoder(b *testing.B) {
	targets := generateTargets(16)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tgt := &targets[i%16]
		bm := EncodeOld(tgt)
		_, _ = bson.Marshal(bm)
	}
}

func BenchmarkEncoder(b *testing.B) {
	targets := generateTargets(16)
	sd := NewStructData(reflect.TypeOf(tms.Target{}), NoMap)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tgt := &targets[i%16]
		_ = sd.Encode(unsafe.Pointer(tgt))
	}
}

func TestEncoder(t *testing.T) {
	targets := generateTargets(5)
	sd := NewStructData(reflect.TypeOf(tms.Target{}), NoMap)

	for _, tgt := range targets {
		raw := sd.Encode(unsafe.Pointer(&tgt))
		msg, err := bson.Marshal(raw)
		if err != nil {
			t.Fatalf("Error marshaling: %v", err)
		}
		var m bson.M
		err = bson.Unmarshal(msg, &m)
		if err != nil {
			t.Fatalf("Error unmarshaling: %v", err)
		}
		s, err := jpb.MarshalToString(&tgt)
		if err == nil {
			t.Logf("Orig obj: %v", s)
		} else {
			t.Fatalf("Orig obj: %v", err)
		}
		b, err := json.MarshalIndent(m, "", "  ")
		if err == nil {
			t.Logf("Unmarshaled map: %v", string(b))
		}
	}
}

func TestTargetDecoder(t *testing.T) {
	targets := generateTargets(5)
	sd := NewStructData(reflect.TypeOf(tms.Target{}), NoMap)

	for _, tgt := range targets {
		if !proto.Equal(&tgt, &tgt) {
			t.Error("Equality check is broken!")
		}

		msg := sd.Encode(unsafe.Pointer(&tgt))
		var tgt2 tms.Target
		sd.DecodeTo(msg, unsafe.Pointer(&tgt2))

		s, err := jpb.MarshalToString(&tgt)
		if err == nil {
			t.Logf("Orig obj: %v", s)
		} else {
			t.Fatalf("Orig obj: %v", err)
		}

		t.Logf("Dec obj: %+v", tgt2)
		b, err := jpb.MarshalToString(&tgt2)
		if err == nil {
			t.Logf("Enc/dec obj: %v", string(b))
		} else {
			t.Fatalf("Enc/dec obj: %v", err)
		}
		if !proto.Equal(&tgt, &tgt2) {
			t.Error("Encoded/decoded messages are not equal")
		}
	}
}

// CONV-663
//func TestTargetDecoderRawDataNil(t *testing.T) {
//	targets := generateTargets(5)
//	sd := NewStructData(reflect.TypeOf(Target{}), NoMap)
//
//	for _, tgt := range targets {
//		if !proto.Equal(&tgt, &tgt) {
//			t.Error("Equality check is broken!")
//		}
//
//		msg := sd.Encode(unsafe.Pointer(&tgt))
//		t.Logf("msg.Kind: %v", msg.Kind)
//		msg.Data = nil;
//		var tgt2 Target
//		sd.DecodeTo(msg, unsafe.Pointer(&tgt2))
//
//		s, err := jpb.MarshalToString(&tgt)
//		if err == nil {
//			t.Logf("Orig obj: %v", s)
//		} else {
//			t.Fatalf("Orig obj: %v", err)
//		}
//
//		t.Logf("Dec obj: %+v", tgt2)
//		b, err := jpb.MarshalToString(&tgt2)
//		if err == nil {
//			t.Logf("Enc/dec obj: %v", string(b))
//		} else {
//			t.Fatalf("Enc/dec obj: %v", err)
//		}
//		if !proto.Equal(&tgt, &tgt2) {
//			t.Error("Encoded/decoded messages are not equal")
//		}
//	}
//}

func BenchmarkOldDecoder(b *testing.B) {
	targets := generateTargets(16)
	encoded := make([][]byte, 16)
	for i, tgt := range targets {
		bm := EncodeOld(&tgt)
		raw, _ := bson.Marshal(bm)
		encoded[i] = raw
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		enc := encoded[i%16]
		var raw bson.M
		bson.Unmarshal(enc, &raw)
		var tgt tms.Target
		DecodeOld(&tgt, raw)
	}
}

func BenchmarkDecoder(b *testing.B) {
	targets := generateTargets(16)
	encoded := make([]bson.Raw, 16)
	sd := NewStructData(reflect.TypeOf(tms.Target{}), NoMap)
	for i, tgt := range targets {
		raw := sd.Encode(unsafe.Pointer(&tgt))
		encoded[i] = raw
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var tgt tms.Target
		raw := encoded[i%16]
		sd.DecodeTo(raw, unsafe.Pointer(&tgt))
	}
}

func BenchmarkDecoderReuseMem(b *testing.B) {
	targets := generateTargets(16)
	encoded := make([]bson.Raw, 16)
	sd := NewStructData(reflect.TypeOf(tms.Target{}), NoMap)
	for i, tgt := range targets {
		raw := sd.Encode(unsafe.Pointer(&tgt))
		encoded[i] = raw
	}
	b.ResetTimer()

	var tgt tms.Target
	for i := 0; i < b.N; i++ {
		raw := encoded[i%16]
		sd.DecodeTo(raw, unsafe.Pointer(&tgt))
	}
}
