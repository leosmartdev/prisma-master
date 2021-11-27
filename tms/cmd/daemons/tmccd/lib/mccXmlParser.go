package lib

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"time"

	tsi "prisma/tms"
	"prisma/tms/sar"

	"github.com/globalsign/mgo/bson"
)

//This is a hard coded global variable for extracting xml data from incoming mcc messages
var XMLExp = regexp.MustCompile(`<\?xml version\="1.0" \?>(.|\n)*?</topMessage>`)

type TopMessage struct {
	XMLName        xml.Name       `xml:"topMessage"`
	EnvelopeHeader EnvelopeHeader `xml:"header"`
	Message        Message        `xml:"message"`
}

type EnvelopeHeader struct {
	Dest   string `xml:"dest,attr"`
	Orig   string `xml:"orig,attr"`
	Number int32  `xml:"number,attr"`
	Date   string `xml:"date,attr"`
}

type Message struct {
	UnlocatedAlertMessage UnlocatedAlertMessage `xml:"unlocatedAlertMessage,omitempty"`
	IncidentAlertMessage  IncidentAlertMessage  `xml:"incidentAlertMessage,omitempty"`
	ResolvedAlertMessage  ResolvedAlertMessage  `xml:"resolvedAlertMessage,omitempty"`
	FreeformMessage       FreeformMessage       `xml:"freeformMessage,omitempty"`
}

type Header struct {
	SiteId int32  `xml:"siteId"`
	Beacon string `xml:"beacon"`
}

type Composite struct {
	Location Locate `xml:"location"`
	Duration string `xml:"duration"`
}

type Elemental struct {
	Side        string  `xml:"side"`
	Location    Locate  `xml:"location"`
	Satellite   string  `xml:"satellite"`
	OrbitNumber string  `xml:"orbitNumber"`
	Tca         string  `xml:"tca"`
	DopplerA    Doppler `xml:"dopplerA"`
	DopplerB    Doppler `xml:"dopplerB"`
}

type MeoElemental struct {
	Satellite string `xml:"satellite"`
	Tca       string `xml:"tca"`
	Doa       Doa    `xml:"doa"`
}

type Doa struct {
	Location Locate `xml:"location"`
}

type Doppler struct {
	Location Locate `xml:"location"`
}

type Locate struct {
	Latitude  float64 `xml:"latitude,attr"`
	Longitude float64 `xml:"longitude,attr"`
}

type BeaconStruct struct {
	BcnId30           string `xml:"bcnId30"`
	BcnId15           string `xml:"bcnId15"`
	BeaconMessageType int32  `xml:"beaconMessageType"`
	MessageValid      int32  `xml:"messageValid"`
	FrameSync         string `xml:"frameSync"`
}

type Solution struct {
	SolutionId     int32        `xml:"solutionId,attr"`
	Type           int32        `xml:"type"`
	BeaconStruct   BeaconStruct `xml:"beacon"`
	UnusedPackets  int32        `xml:"unusedPackets"`
	LutId          int32        `xml:"lutId"`
	Links          Links        `xml:"links"`
	GeneratedTime  string       `xml:"generatedTime"`
	Bursts         int32        `xml:"bursts"`
	Packets        int32        `xml:"packets"`
	FirstBurstTime string       `xml:"firstBurstTime"`
	LastBurstTime  string       `xml:"lastBurstTime"`
	Lat            float32      `xml:"lat"`
	Long           float32      `xml:"long"`
	Alt            float32      `xml:"alt"`
	Frequency      float32      `xml:"frequency"`
	FrequencyDrift float32      `xml:"frequencyDrift"`
	Noise          int32        `xml:"noise"`
	Cn0            Cn0          `xml:"cn0"`
	QualityFactor  int32        `xml:"qualityFactor"`
	Iterations     int32        `xml:"iterations"`
	ErrorEllipse   ErrorEllipse `xml:"errorEllipse"`
}

type Links struct {
	Total int32  `xml:"total,attr"`
	Link  []Link `xml:"link"`
}

type Link struct {
	LutId     int32     `xml:"lutId,attr"`
	AntennaId int32     `xml:"AntennaId,attr"`
	Satellite Satellite `xml:"satellite"`
}

type Satellite struct {
	SatId      int32 `xml:"satId,attr"`
	SatNoradId int32 `xml:"satNoradId,attr"`
}

type Cn0 struct {
	AvgCN0 float32 `xml:"avgCN0"`
	MinCN0 float32 `xml:"minCN0"`
	MaxCN0 float32 `xml:"maxCN0"`
}

type SolutionMessage struct {
	Solution Solution `xml:"solution"`
}

type UnlocatedAlertMessage struct {
	Header      Header `xml:"header"`
	Tca         string `xml:"tca"`
	Satellite   string `xml:"satellite"`
	OrbitNumber string `xml:"orbitNumber"`
}

type IncidentAlertMessage struct {
	Header       Header       `xml:"header"`
	Elemental    []Elemental  `xml:"elemental"`
	MeoElemental MeoElemental `xml:"meoElemental"`
}

type ResolvedAlertMessage struct {
	Header       Header       `xml:"header"`
	Composite    Composite    `xml:"composite"`
	Elemental    []Elemental  `xml:"elemental"`
	MeoElemental MeoElemental `xml:"meoElemental"`
}

type ErrorEllipse struct {
	MajorAxis float32 `xml:"majorAxis"`
	MinorAxis float32 `xml:"minorAxis"`
	Heading   float32 `xml:"heading"`
	Radius    float32 `xml:"radius"`
	Area      float32 `xml:"area"`
}

type FreeformMessage struct {
	Subject string `xml:"subject"`
	Body    string `xml:"body"`
}

func MccxmlParser(msg []byte, protocol string) (*sar.SarsatMessage, error) {
	topMessage := TopMessage{}
	err := xml.Unmarshal(msg, &topMessage)
	if err != nil {
		return nil, err
	}

	sarsatMessage := &sar.SarsatMessage{}
	sarsatMessage.MessageNumber = topMessage.EnvelopeHeader.Number
	sarsatMessage.MessageType = sar.SarsatMessage_MCM_XML
	if topMessage.EnvelopeHeader.Date != "" {
		MessageTime, err := time.Parse("2006-01-02T15:04:05Z", topMessage.EnvelopeHeader.Date)
		if err != nil {
			return nil, err
		}
		sarsatMessage.MessageTime = tsi.ToTimestamp(MessageTime)
	}

	sarsatMessage.MessageBody = fmt.Sprintf("%+v", topMessage)
	sarsatMessage.LocalName = topMessage.EnvelopeHeader.Dest
	sarsatMessage.Protocol = protocol
	sarsatMessage.Received = true
	sarsatMessage.RemoteName = topMessage.EnvelopeHeader.Orig
	sarsatMessage.RemoteType = sar.SarsatMessage_MCC

	// Populate IncidentAlert
	if len(topMessage.Message.IncidentAlertMessage.Elemental) != 0 || topMessage.Message.IncidentAlertMessage.Header != (Header{}) {
		sarsatAlert := &sar.SarsatAlert{}
		sarsatAlert.Id = bson.NewObjectId().Hex()
		sarsatAlert.AlertType = sar.SarsatAlert_IncidentAlert
		sarsatAlert.IncidentAlertMessage = &sar.IncidentAlert{}
		sarsatAlert.SiteNumber = topMessage.Message.IncidentAlertMessage.Header.SiteId
		if len(topMessage.Message.IncidentAlertMessage.Header.Beacon) != 0 {
			beacon, err := sar.DecodeHexID(topMessage.Message.IncidentAlertMessage.Header.Beacon)
			if err != nil {
				return nil, err
			}
			sarsatAlert.Beacon = beacon
		}
		for i := 0; i < len(topMessage.Message.IncidentAlertMessage.Elemental); i++ {
			notifyTime, err := time.Parse("2006-01-02T15:04:05.000Z", topMessage.Message.IncidentAlertMessage.Elemental[i].Tca)
			if err != nil {
				return nil, err
			}
			sarsatAlert.IncidentAlertMessage.Elemental = append(sarsatAlert.IncidentAlertMessage.Elemental, &sar.Element{})
			sarsatAlert.IncidentAlertMessage.Elemental[i].NotificationTime = tsi.ToTimestamp(notifyTime)

			if topMessage.Message.IncidentAlertMessage.Elemental[i].DopplerA != (Doppler{}) {
				location := &sar.Point{}
				location.Latitude = topMessage.Message.IncidentAlertMessage.Elemental[i].DopplerA.Location.Latitude
				location.Longitude = topMessage.Message.IncidentAlertMessage.Elemental[i].DopplerA.Location.Longitude
				sarsatAlert.IncidentAlertMessage.Elemental[i].Doppler = append(sarsatAlert.IncidentAlertMessage.Elemental[i].Doppler, &sar.Doppler{})
				sarsatAlert.IncidentAlertMessage.Elemental[i].Doppler[len(sarsatAlert.IncidentAlertMessage.Elemental[i].Doppler)-1].DopplerPosition = location
				sarsatAlert.IncidentAlertMessage.Elemental[i].Doppler[len(sarsatAlert.IncidentAlertMessage.Elemental[i].Doppler)-1].Type = 0
				//TODO: Add probability here if there is one
			}

			if topMessage.Message.IncidentAlertMessage.Elemental[i].DopplerB != (Doppler{}) {
				location := &sar.Point{}
				location.Latitude = topMessage.Message.IncidentAlertMessage.Elemental[i].DopplerB.Location.Latitude
				location.Longitude = topMessage.Message.IncidentAlertMessage.Elemental[i].DopplerB.Location.Longitude
				sarsatAlert.IncidentAlertMessage.Elemental[i].Doppler = append(sarsatAlert.IncidentAlertMessage.Elemental[i].Doppler, &sar.Doppler{})
				sarsatAlert.IncidentAlertMessage.Elemental[i].Doppler[len(sarsatAlert.IncidentAlertMessage.Elemental[i].Doppler)-1].DopplerPosition = location
				sarsatAlert.IncidentAlertMessage.Elemental[i].Doppler[len(sarsatAlert.IncidentAlertMessage.Elemental[i].Doppler)-1].Type = 1
				//TODO: Add probability here if there is one
			}

		}
		if topMessage.Message.IncidentAlertMessage.MeoElemental != (MeoElemental{}) {
			sarsatAlert.IncidentAlertMessage.MeoElemental = &sar.Meoelemental{}
			if topMessage.Message.IncidentAlertMessage.MeoElemental.Doa != (Doa{}) {
				location := &sar.Point{}
				sarsatAlert.IncidentAlertMessage.MeoElemental.Doa = &sar.DOA{}
				location.Latitude = topMessage.Message.IncidentAlertMessage.MeoElemental.Doa.Location.Latitude
				location.Longitude = topMessage.Message.IncidentAlertMessage.MeoElemental.Doa.Location.Longitude
				sarsatAlert.IncidentAlertMessage.MeoElemental.Doa.DoaPosition = location
				if len(topMessage.Message.IncidentAlertMessage.MeoElemental.Tca) != 0 {
					locationTime, err := time.Parse("2006-01-02T15:04:05.000Z", topMessage.Message.IncidentAlertMessage.MeoElemental.Tca)
					if err != nil {
						return nil, err
					}
					sarsatAlert.IncidentAlertMessage.MeoElemental.NotificationTime = tsi.ToTimestamp(locationTime)
				}
			}
			if topMessage.Message.IncidentAlertMessage.MeoElemental.Satellite != "" {
				sarsatAlert.IncidentAlertMessage.MeoElemental.Satellite = topMessage.Message.IncidentAlertMessage.MeoElemental.Satellite
			}

		}
		sarsatAlert.ProcessedTime = tsi.Now()
		sarsatMessage.SarsatAlert = sarsatAlert
		return sarsatMessage, nil
	} // end of IncidentAlert population

	//populate ResolvedAlert
	if len(topMessage.Message.ResolvedAlertMessage.Elemental) != 0 || topMessage.Message.ResolvedAlertMessage.Composite != (Composite{}) || topMessage.Message.ResolvedAlertMessage.Header != (Header{}) || topMessage.Message.ResolvedAlertMessage.MeoElemental != (MeoElemental{}) {
		sarsatAlert := &sar.SarsatAlert{}
		sarsatAlert.Id = bson.NewObjectId().Hex()
		sarsatAlert.AlertType = sar.SarsatAlert_ResolvedAlert
		sarsatAlert.SiteNumber = topMessage.Message.ResolvedAlertMessage.Header.SiteId
		sarsatAlert.ResolvedAlertMessage = &sar.ResolvedAlert{}

		if len(topMessage.Message.ResolvedAlertMessage.Header.Beacon) != 0 {
			beacon, err := sar.DecodeHexID(topMessage.Message.ResolvedAlertMessage.Header.Beacon)
			if err != nil {
				return nil, err
			}
			sarsatAlert.Beacon = beacon
		}
		for i := 0; i < len(topMessage.Message.ResolvedAlertMessage.Elemental); i++ {
			notifyTime, err := time.Parse("2006-01-02T15:04:05.000Z", topMessage.Message.ResolvedAlertMessage.Elemental[i].Tca)
			if err != nil {
				return nil, err
			}
			sarsatAlert.ResolvedAlertMessage.Elemental = append(sarsatAlert.ResolvedAlertMessage.Elemental, &sar.Element{})
			sarsatAlert.ResolvedAlertMessage.Elemental[i].NotificationTime = tsi.ToTimestamp(notifyTime)
			sarsatAlert.ResolvedAlertMessage.Elemental[i].Satellite = topMessage.Message.ResolvedAlertMessage.Elemental[i].Satellite
			sarsatAlert.ResolvedAlertMessage.Elemental[i].OrbitNumber = topMessage.Message.ResolvedAlertMessage.Elemental[i].OrbitNumber

			if topMessage.Message.ResolvedAlertMessage.Elemental[i].Side == "A" || topMessage.Message.ResolvedAlertMessage.Elemental[i].Side == "B" {
				location := &sar.Point{}
				sarsatAlert.ResolvedAlertMessage.Elemental[i].Doppler = append(sarsatAlert.ResolvedAlertMessage.Elemental[i].Doppler, &sar.Doppler{})
				sarsatAlert.ResolvedAlertMessage.Elemental[i].Doppler[len(sarsatAlert.ResolvedAlertMessage.Elemental[i].Doppler)-1].DopplerPosition = &sar.Point{}
				location.Latitude = topMessage.Message.ResolvedAlertMessage.Elemental[i].Location.Latitude
				location.Longitude = topMessage.Message.ResolvedAlertMessage.Elemental[i].Location.Longitude
				sarsatAlert.ResolvedAlertMessage.Elemental[i].Doppler[len(sarsatAlert.ResolvedAlertMessage.Elemental[i].Doppler)-1].DopplerPosition = location
				if topMessage.Message.ResolvedAlertMessage.Elemental[i].Side == "A" {
					sarsatAlert.ResolvedAlertMessage.Elemental[i].Doppler[len(sarsatAlert.ResolvedAlertMessage.Elemental[i].Doppler)-1].Type = 0
				} else if topMessage.Message.ResolvedAlertMessage.Elemental[i].Side == "B" {
					sarsatAlert.ResolvedAlertMessage.Elemental[i].Doppler[len(sarsatAlert.ResolvedAlertMessage.Elemental[i].Doppler)-1].Type = 1
				}
			}
			if topMessage.Message.ResolvedAlertMessage.Elemental[i].Tca != "" {
				locationTime, err := time.Parse("2006-01-02T15:04:05.000Z", topMessage.Message.ResolvedAlertMessage.Elemental[i].Tca)
				if err != nil {
					return nil, err
				}
				sarsatAlert.ResolvedAlertMessage.Elemental[i].NotificationTime = tsi.ToTimestamp(locationTime)
			}

		}

		if topMessage.Message.ResolvedAlertMessage.Composite != (Composite{}) {
			location := &sar.Point{}
			location.Latitude = topMessage.Message.ResolvedAlertMessage.Composite.Location.Latitude
			location.Longitude = topMessage.Message.ResolvedAlertMessage.Composite.Location.Longitude
			sarsatAlert.ResolvedAlertMessage.CompositeLocation = location
			sarsatAlert.ResolvedAlertMessage.CompositeDuration = topMessage.Message.ResolvedAlertMessage.Composite.Duration
		}
		if topMessage.Message.ResolvedAlertMessage.MeoElemental != (MeoElemental{}) {
			sarsatAlert.ResolvedAlertMessage.MeoElemental = &sar.Meoelemental{}
			if topMessage.Message.ResolvedAlertMessage.MeoElemental.Doa != (Doa{}) {
				location := &sar.Point{}
				location.Latitude = topMessage.Message.ResolvedAlertMessage.MeoElemental.Doa.Location.Latitude
				location.Longitude = topMessage.Message.ResolvedAlertMessage.MeoElemental.Doa.Location.Longitude
				sarsatAlert.ResolvedAlertMessage.MeoElemental.Doa.DoaPosition = location
				if len(topMessage.Message.ResolvedAlertMessage.MeoElemental.Tca) != 0 {
					locationTime, err := time.Parse("2006-01-02T15:04:05.000Z", topMessage.Message.ResolvedAlertMessage.MeoElemental.Tca)
					if err != nil {
						return nil, err
					}
					sarsatAlert.ResolvedAlertMessage.MeoElemental.NotificationTime = tsi.ToTimestamp(locationTime)
				}
			}
			if topMessage.Message.ResolvedAlertMessage.MeoElemental.Satellite != "" {
				sarsatAlert.ResolvedAlertMessage.MeoElemental.Satellite = topMessage.Message.ResolvedAlertMessage.MeoElemental.Satellite
			}

		}
		sarsatAlert.ProcessedTime = tsi.Now()
		sarsatMessage.SarsatAlert = sarsatAlert
		return sarsatMessage, nil
	} // endo of ResolvedAlert

	//populate UnlocatedAlert
	if topMessage.Message.UnlocatedAlertMessage.Header != (Header{}) || len(topMessage.Message.UnlocatedAlertMessage.OrbitNumber) != 0 || len(topMessage.Message.UnlocatedAlertMessage.Satellite) != 0 || len(topMessage.Message.UnlocatedAlertMessage.Tca) != 0 {
		sarsatAlert := &sar.SarsatAlert{}
		sarsatAlert.Id = bson.NewObjectId().Hex()
		sarsatAlert.AlertType = sar.SarsatAlert_UnlocatedAlert
		sarsatAlert.UnlocatedAlertMessage = &sar.UnlocatedAlert{}
		if topMessage.Message.UnlocatedAlertMessage.Header != (Header{}) {
			sarsatAlert.SiteNumber = topMessage.Message.UnlocatedAlertMessage.Header.SiteId
			if len(topMessage.Message.UnlocatedAlertMessage.Header.Beacon) != 0 {
				beacon, err := sar.DecodeHexID(topMessage.Message.UnlocatedAlertMessage.Header.Beacon)
				if err != nil {
					return nil, err
				}
				sarsatAlert.Beacon = beacon
			}
		}
		if topMessage.Message.UnlocatedAlertMessage.Tca != "" {
			notifyTime, err := time.Parse("2006-01-02T15:04:05.000Z", topMessage.Message.UnlocatedAlertMessage.Tca)
			if err != nil {
				return nil, err
			}
			sarsatAlert.UnlocatedAlertMessage.NotificationTime = tsi.ToTimestamp(notifyTime)

		}
		if topMessage.Message.UnlocatedAlertMessage.OrbitNumber != "" {
			sarsatAlert.UnlocatedAlertMessage.Satellite = topMessage.Message.UnlocatedAlertMessage.Satellite
		}
		if topMessage.Message.UnlocatedAlertMessage.Satellite != "" {
			sarsatAlert.UnlocatedAlertMessage.OrbitNumber = topMessage.Message.UnlocatedAlertMessage.OrbitNumber
		}

		sarsatAlert.ProcessedTime = tsi.Now()
		sarsatMessage.SarsatAlert = sarsatAlert
		return sarsatMessage, nil

	} // end of UnlocatedAlert

	if topMessage.Message.FreeformMessage != (FreeformMessage{}) {
		sarsatMessage.FreeFormMessage = &sar.FreeForm{}
		if topMessage.Message.FreeformMessage.Subject != "" {
			sarsatMessage.FreeFormMessage.Subject = topMessage.Message.FreeformMessage.Subject
		}
		if topMessage.Message.FreeformMessage.Body != "" {
			sarsatMessage.FreeFormMessage.Body = topMessage.Message.FreeformMessage.Body
		}
		return sarsatMessage, nil
	}

	return sarsatMessage, nil
}
