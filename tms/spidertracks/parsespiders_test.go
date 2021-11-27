package spidertracks

import (
	"encoding/xml"
	"testing"

	"github.com/json-iterator/go/assert"
)

func TestSpider(t *testing.T) {
	TelemetryOne := Telemetry{
		"trackid",
		"spider",
		"xsd:integer",
		"2401",
	}
	TelemetryTwo := Telemetry{
		"registration",
		"spidertracks",
		"xsd:string",
		"HBEAT",
	}
	TelemetryList := []Telemetry{TelemetryOne, TelemetryTwo}
	DummySpiderOne := AcPos{
		"300034012609560",
		"300034012609560",
		"GPS",
		"3D",
		"8",
		"2018-05-18T10:01:20Z",
		"2018-05-18T10:01:38Z",
		"spidertracks",
		-36.85816,
		174.76067,
		102,
		0,
		46,
		TelemetryList,
	}
	DummySpiderTwo := AcPos{
		"300034012609560",
		"300034012609560",
		"GPS",
		"3D",
		"8",
		"2018-05-18T10:02:20Z",
		"2018-05-18T10:02:30Z",
		"spidertracks",
		-36.85814,
		174.76066,
		100,
		0,
		64,
		TelemetryList,
	}
	acPosList := []AcPos{DummySpiderOne, DummySpiderTwo}
	DummySpider := Spider{
		xml.Name{Space: "https://www.aff.gov/affSchema", Local: "data"},
		acPosList,
	}

	tt := []struct {
		name    string
		spiders Spider
	}{
		{"Standard Spider Test", DummySpider},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			spdr, err := Parse(data)
			t.Logf("%+v", spdr)
			assert.NoError(t, err)
			assert.Equal(t, tc.spiders, spdr)
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
