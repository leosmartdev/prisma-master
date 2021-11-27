package main

import (
	"reflect"
	"testing"
	"time"

	"prisma/gogroup"
	"prisma/tms"
	"prisma/tms/devices"
	"prisma/tms/spidertracks"
	"prisma/tms/test/context"
	"prisma/tms/tmsg"
	"prisma/tms/util/ident"
	"prisma/tms/util/units"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/stretchr/testify/assert"
)

func init() {
	tmsg.GClient = tmsg.NewTsiClientStub()
}

func Test_dataToSpider(t *testing.T) {
	spdr, _ := spidertracks.Parse(data)
	type args struct {
		rawSpiderList spidertracks.Spider
		indx          int
	}
	tests := []struct {
		name    string
		args    args
		want    SpiderSimple
		wantErr bool
	}{
		{"empty spider", args{spidertracks.Spider{}, 0}, SpiderSimple{}, true},
		{"index out spider", args{spidertracks.Spider{}, 1}, SpiderSimple{}, true},
		{"valid spider", args{spdr, 0}, SpiderSimple{
			IMEI:      "300034012609560",
			Latitude:  -36.85816,
			Longitude: 174.76067,
			Altitude:  102,
			Speed:     0,
			Heading:   46,
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dataToSpider(tt.args.rawSpiderList, tt.args.indx)
			if (err != nil) != tt.wantErr {
				t.Errorf("dataToSpider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.want.Time = got.Time // time workaround
			if !assert.Equal(t, got, tt.want) {
				t.Errorf("dataToSpider() = %v, want %v", got, tt.want)
			}
		})
	}
}

// integration test instead of unit test
//func TestReturnSimplifiedSpiderList(t *testing.T) {}

func TestPopulateTrack(t *testing.T) {
	var ignoreTrack tms.Track
	testTime, _ := ptypes.TimestampProto(time.Now())
	type args struct {
		spider SpiderSimple
	}
	tests := []struct {
		name    string
		args    args
		want    *tms.Track
		wantErr bool
	}{
		{"empty track", args{SpiderSimple{}}, &ignoreTrack, false},
		{"valid track", args{SpiderSimple{
			IMEI:      "300034012609560",
			Time:      time.Now(),
			Latitude:  1,
			Longitude: 2,
			Altitude:  3,
			Speed:     4,
			Heading:   5,
		}}, &tms.Track{
			Id:         "cceb3250de37c9532062d3493aafbad4",
			RegistryId: "86cfab73f91f4ccac807208e31e9f023",
			Producer:   nil,
			Targets: []*tms.Target{{
				Id:         generateTargetID(),
				Type:       devices.DeviceType_Spidertracks,
				Time:       testTime,
				IngestTime: testTime,
				Imei:       &wrappers.StringValue{Value: "300034012609560"},
				Position: &tms.Point{
					Latitude:  1,
					Longitude: 2,
					Altitude:  3,
				},
				Speed:   &wrappers.DoubleValue{Value: units.FromMetersSecondToKnots(4)},
				Heading: &wrappers.DoubleValue{Value: 5},
			}},
			Metadata: []*tms.TrackMetadata{{
				Name:       "300034012609560",
				Time:       testTime,
				IngestTime: testTime,
				Type:       devices.DeviceType_Spidertracks,
			}},
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PopulateTrack(tt.args.spider)
			if (err != nil) != tt.wantErr {
				t.Errorf("PopulateTrack() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want == &ignoreTrack {
				return
			}
			// ignore
			tt.want.Producer = got.Producer
			tt.want.Targets[0].Id = got.Targets[0].Id
			tt.want.Targets[0].Time = got.Targets[0].Time
			tt.want.Targets[0].IngestTime = got.Targets[0].IngestTime
			tt.want.Metadata[0].Time = got.Metadata[0].Time
			tt.want.Metadata[0].IngestTime = got.Metadata[0].IngestTime
			if !assert.Equal(t, got, tt.want) {
				t.Errorf("PopulateTrack() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSendTrackToTGWAD(t *testing.T) {
	type args struct {
		ctxt  gogroup.GoGroup
		track *tms.Track
	}
	ctx := gogroup.New(context.Test(), "test")
	testTime, _ := ptypes.TimestampProto(time.Now())
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"empty track", args{ctx, &tms.Track{}}, true},
		{"nil track", args{ctx, nil}, true},
		{"valid metadata track", args{ctx, &tms.Track{
			Targets: nil,
			Metadata: []*tms.TrackMetadata{{
				Time:       testTime,
				IngestTime: testTime,
				Type:       devices.DeviceType_Spidertracks,
			}},
		}}, false},
		{"valid target track", args{ctx, &tms.Track{
			Targets: []*tms.Target{{
				Id:         generateTargetID(),
				Type:       devices.DeviceType_Spidertracks,
				Time:       testTime,
				IngestTime: testTime,
				Imei:       &wrappers.StringValue{Value: "testimei"},
			}},
		}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SendTrackToTGWAD(tt.args.ctxt, tt.args.track); (err != nil) != tt.wantErr {
				t.Errorf("SendTrackToTGWAD() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_generateTargetID(t *testing.T) {
	sn := ident.TimeSerialNumber()
	tests := []struct {
		name string
		want *tms.TargetID
	}{
		{"valid", &tms.TargetID{
			SerialNumber: &tms.TargetID_TimeSerial{TimeSerial: &sn},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateTargetID(); reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("generateTargetID() = %v, want %v", got, tt.want)
			}
		})
	}
}

const data = `
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<data xmlns="https://www.aff.gov/affSchema" version="2.23" sysID="spidertracks" rptTime="2018-05-18T16:00:20.391Z">
    <posList listType="Async">
        <acPos esn="300034012609560" UnitID="300034012609560" source="GPS" fix="3D" HDOP="8" dateTime="2018-05-18T10:01:20Z" dataCtrDateTime="2018-05-18T10:01:38Z" dataCtr="spidertracks">
            <Lat>-36.85816</Lat>
            <Long>174.76067</Long>
            <altitude units="meters">102</altitude>
            <speed units="meters/sec">0</speed>
            <heading units="Track-True">46</heading>
            <telemetry name="trackid" source="spider" type="xsd:integer" value="2401"/>
            <telemetry name="registration" source="spidertracks" type="xsd:string" value="HBEAT"/>
        </acPos>
        <acPos esn="300034012609560" UnitID="300034012609560" source="GPS" fix="3D" HDOP="8" dateTime="2018-05-18T10:02:20Z" dataCtrDateTime="2018-05-18T10:02:30Z" dataCtr="spidertracks">
            <Lat>-36.85814</Lat>
            <Long>174.76066</Long>
            <altitude units="meters">100</altitude>
            <speed units="meters/sec">0</speed>
            <heading units="Track-True">64</heading>
            <telemetry name="trackid" source="spider" type="xsd:integer" value="2401"/>
            <telemetry name="registration" source="spidertracks" type="xsd:string" value="HBEAT"/>
        </acPos>
    </posList>
</data>
`
