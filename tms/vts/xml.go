package vts

type AISTrack struct {
	MMSI        string  `xml:"MMSI"`
	Timestamp   string  `xml:"Timestamp"`
	TypeOfShip  uint8   `xml:"TypeOfShip"`
	NameOfShip  string  `xml:"NameOfShip"`
	Callsign    string  `xml:"Callsign"`
	NavStatus   uint8   `xml:"NavStatus"`
	Speed       float64 `xml:"Speed"`
	Course      float64 `xml:"Course"`
	Heading     float64 `xml:"Heading"`
	Long        float64 `xml:"Long"`
	Lat         float64 `xml:"Lat"`
	Draught     float64 `xml:"Draught"`
	ETA         uint    `xml:"ETA"`
	Size        uint    `xml:"Size"`
	IMONumber   string  `xml:"IMONumber"`
	Destination string  `xml:"Destination"`
}

type AISTracks struct {
	AISTrack []AISTrack `xml:"AisTrack"`
}

type TrackerTrack struct {
	TrackerID       uint    `xml:"TrackerId"`
	Timestamp       string  `xml:"Timestamp"`
	TrackID         uint    `xml:"TrackId"`
	TrackName       string  `xml:"TrackName"`
	TrackMMSI       string  `xml:"TrackMMSI"`
	Quality         uint    `xml:"Quality"`
	Speed           float64 `xml:"Speed"`
	Course          float64 `xml:"Course"`
	Heading         float64 `xml:"Heading"`
	Long            float64 `xml:"Long"`
	Lat             float64 `xml:"Lat"`
	TrackState      uint8   `xml:"TrackState"`
	AcquisitionType uint8   `xml:"AdquisitionType"`
	TrackType       uint8   `xml:"TrackType"`
	Range           float64 `xml:"Range"`
	Bearing         float64 `xml:"Bearing"`
}

type TrackerTracks struct {
	TrackerTrack []TrackerTrack `xml:"TrackerTrack"`
}

type TracksDataInfo struct {
	VTSCenter      string        `xml:"VTSCenter"`
	FileUpdateTime string        `xml:"FileUpdateTime"`
	AISTracks      AISTracks     `xml:"AisTracks"`
	TrackerTracks  TrackerTracks `xml:"TrackerTracks"`
}
