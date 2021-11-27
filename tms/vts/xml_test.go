package vts

import (
	"encoding/xml"
	"fmt"
	"testing"
)

func TestAISXML(t *testing.T) {
	doc := `
<?xml version="1.0" standalone="no"?>
<TracksDataInfo xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="VTMISTracksData.xsd">
	<VTSCenter>Ras Laffan</VTSCenter>
	<FileUpdateTime>2018-09-11T22:49:15.458Z</FileUpdateTime>
	<AisTracks>
		<AisTrack>
			<MMSI>576254000</MMSI>
			<Timestamp>2018-09-11T22:44:50.000Z</Timestamp>
			<TypeOfShip>52</TypeOfShip>
			<NameOfShip>KIRKCONNELL TIDE    </NameOfShip>
			<Callsign>YJVV6  </Callsign>
			<NavStatus>1</NavStatus>
			<Speed>0.05144</Speed>
			<Course>241.7</Course>
			<Heading>344.0</Heading>
			<Long>50.18861</Long>
			<Lat>26.62047</Lat>
			<Draught>4.8</Draught>
			<ETA>613696</ETA>
			<Size>23265924</Size>
			<IMONumber>9582180</IMONumber>
			<Destination>RASTANURA FRAIGHTER </Destination>
		</AisTrack>
	</AisTracks>
</TracksDataInfo>
`

	want := TracksDataInfo{
		VTSCenter:      "Ras Laffan",
		FileUpdateTime: "2018-09-11T22:49:15.458Z",
		AISTracks: AISTracks{
			AISTrack: []AISTrack{
				{
					MMSI:        "576254000",
					Timestamp:   "2018-09-11T22:44:50.000Z",
					TypeOfShip:  52,
					NameOfShip:  "KIRKCONNELL TIDE    ",
					Callsign:    "YJVV6  ",
					NavStatus:   1,
					Speed:       0.05144,
					Course:      241.7,
					Heading:     344.0,
					Long:        50.18861,
					Lat:         26.62047,
					Draught:     4.8,
					ETA:         613696,
					Size:        23265924,
					IMONumber:   "9582180",
					Destination: "RASTANURA FRAIGHTER ",
				},
			},
		},
		TrackerTracks: TrackerTracks{},
	}

	have := TracksDataInfo{}
	if err := xml.Unmarshal([]byte(doc), &have); err != nil {
		t.Fatalf("xml decoding failed: %v", err)
	}
	strhave := fmt.Sprintf("%v", have)
	strwant := fmt.Sprintf("%v", want)
	if strhave != strwant {
		t.Fatalf("\n have: %+v \n want: %+v", strhave, strwant)
	}

}

func TestRadarXML(t *testing.T) {
	doc := `
<?xml version="1.0" standalone="no"?>
<TracksDataInfo xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="VTMISTracksData.xsd">
	<VTSCenter>Ras Laffan</VTSCenter>
	<FileUpdateTime>2018-09-11T22:49:15.458Z</FileUpdateTime>
	<AisTracks/>
	<TrackerTracks>
		<TrackerTrack>
			<TrackerId>1</TrackerId>
			<Timestamp>2018-09-11T22:44:50.000Z</Timestamp>
			<TrackId>13</TrackId>
			<TrackName>Contact Bravo</TrackName>
			<TrackMMSI>576254000</TrackMMSI>
			<Quality>1</Quality>
			<Speed>12.4</Speed>
			<Course>42.1</Course>
			<Heading>42.5</Heading>
			<Long>50.18861</Long>
			<Lat>26.62047</Lat>
			<TrackState>1</TrackState>
			<AdquisitionType>1</AdquisitionType>
			<TrackType>1</TrackType>
			<Range>4.3</Range>
			<Bearing>91</Bearing>
		</TrackerTrack>
	</TrackerTracks>
</TracksDataInfo>
`

	want := TracksDataInfo{
		VTSCenter:      "Ras Laffan",
		FileUpdateTime: "2018-09-11T22:49:15.458Z",
		TrackerTracks: TrackerTracks{
			TrackerTrack: []TrackerTrack{
				{
					TrackerID:       1,
					Timestamp:       "2018-09-11T22:44:50.000Z",
					TrackID:         13,
					TrackName:       "Contact Bravo",
					TrackMMSI:       "576254000",
					Quality:         1,
					Speed:           12.4,
					Course:          42.1,
					Heading:         42.5,
					Long:            50.18861,
					Lat:             26.62047,
					TrackState:      1,
					AcquisitionType: 1,
					TrackType:       1,
					Range:           4.3,
					Bearing:         91,
				},
			},
		},
		AISTracks: AISTracks{},
	}
	have := TracksDataInfo{}
	if err := xml.Unmarshal([]byte(doc), &have); err != nil {
		t.Fatalf("xml decoding failed: %v", err)
	}
	strhave := fmt.Sprintf("%v", have)
	strwant := fmt.Sprintf("%v", want)
	if strhave != strwant {
		t.Fatalf("\n have: %+v \n want: %+v", strhave, strwant)
	}
}
