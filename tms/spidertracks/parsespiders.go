// Package spidertracks provides functions and structures to parse responses
// from servers and keep info about spider tracks.
package spidertracks

import (
	"encoding/xml"
)

type Spider struct {
	XMLName xml.Name `xml:"data"`
	PosList []AcPos  `xml:"posList>acPos"`
}

type AcPos struct {
	Esn             string      `xml:"esn,attr"`
	UnitID          string      `xml:"UnitID,attr"`
	Source          string      `xml:"source,attr"`
	Fix             string      `xml:"fix,attr"`
	HDOP            string      `xml:"HDOP,attr"`
	DateTime        string      `xml:"dateTime,attr"`
	DataCtrDateTime string      `xml:"dataCtrDateTime,attr"`
	DataCtr         string      `xml:"dataCtr,attr"`
	Lat             float64     `xml:"Lat"`
	Long            float64     `xml:"Long"`
	Altitude        int         `xml:"altitude"`
	Speed           int         `xml:"speed"`
	Heading         int         `xml:"heading"`
	Telemetry       []Telemetry `xml:"telemetry"`
}

type Telemetry struct {
	Name   string `xml:"name,attr"`
	Source string `xml:"source,attr"`
	Type   string `xml:"type,attr"`
	Value  string `xml:"value,attr"`
}

func Parse(datai string) (Spider, error) {
	var spiderList Spider
	err := xml.Unmarshal([]byte(datai), &spiderList)
	return spiderList, err
}
