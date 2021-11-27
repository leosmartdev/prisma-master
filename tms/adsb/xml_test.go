package adsb

import (
	"encoding/xml"
	"reflect"
	"testing"
)

func TestXML(t *testing.T) {

	doc1 := `
<MODESMESSAGE>
	<DATETIME>20070622141943</DATETIME>
	<MODES>400F2B</MODES>
	<CALLSIGN>BAW134</CALLSIGN>
	<ALTITUDE>120300</ALTITUDE>
	<GROUNDSPEED>451</GROUNDSPEED>
	<TRACK>234</TRACK>
	<VRATE>0</VRATE>
	<AIRSPEED></AIRSPEED>
	<LATITUDE>1.10</LATITUDE>
	<LONGITUDE>104.10</LONGITUDE>
</MODESMESSAGE>
	`
	want1 := ModeMessageS{
		Datetime:    &[]string{"20070622141943"}[0],
		Modes:       &[]string{"400F2B"}[0],
		Callsign:    &[]string{"BAW134"}[0],
		Altitude:    &[]int32{120300}[0],
		GroundSpeed: &[]int32{451}[0],
		Track:       &[]int32{234}[0],
		VRate:       &[]int32{0}[0],
		AirSpeed:    &[]int32{0}[0],
		Latitude:    &[]float32{1.10}[0],
		Longitude:   &[]float32{104.10}[0],
	}

	doc2 := `
	<MODESMESSAGE>
	<MODES>A3CD7D</MODES>
	<DATETIME>20210105180726</DATETIME>
	<ALTITUDE>30000</ALTITUDE>
	<GROUNDSPEED>420</GROUNDSPEED>
	<TRACK>240</TRACK>
	<VRATE>0</VRATE>
	<LATITUDE>40.1851485</LATITUDE>
	<LONGITUDE>-76.156976</LONGITUDE>
	</MODESMESSAGE>`

	want2 := ModeMessageS{
		Datetime:    &[]string{"20210105180726"}[0],
		Modes:       &[]string{"A3CD7D"}[0],
		Callsign:    nil,
		Altitude:    &[]int32{30000}[0],
		GroundSpeed: &[]int32{420}[0],
		Track:       &[]int32{240}[0],
		VRate:       &[]int32{0}[0],
		AirSpeed:    nil,
		Latitude:    &[]float32{40.1851485}[0],
		Longitude:   &[]float32{-76.156976}[0],
	}

	tt := []struct {
		name       string
		msg        string
		want ModeMessageS
	}{
		{"sample doc from Qatar documentation", doc1, want1},
		{"sample doc from Lab receiver", doc2, want2},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			have := ModeMessageS{}
	if err := xml.Unmarshal([]byte(tc.msg), &have); err != nil {
		t.Fatalf("%s: xml decoding failed: %v", tc.name,err)
	}
	if !reflect.DeepEqual(have, tc.want) {
		t.Fatalf("%s: \n have: %+v \n want: %+v", tc.name,have, tc.want)
	}
		})
	}

	
	
}
