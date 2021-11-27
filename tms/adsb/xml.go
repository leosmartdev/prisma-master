package adsb

import fmt "fmt"

type ModeMessageS struct {
	Datetime    *string  `xml:"DATETIME"`
	Modes       *string  `xml:"MODES"`
	Callsign    *string  `xml:"CALLSIGN"`
	Altitude    *int32   `xml:"ALTITUDE"`
	GroundSpeed *int32   `xml:"GROUNDSPEED"`
	Track       *int32   `xml:"TRACK"`
	VRate       *int32   `xml:"VRATE"`
	AirSpeed    *int32   `xml:"AIRSPEED"`
	Latitude    *float32 `xml:"LATITUDE"`
	Longitude   *float32 `xml:"LONGITUDE"`
}

func (m ModeMessageS) String() string {
	return fmt.Sprintf(`
Datetime    = %v
Modes       = %v
Callsign    = %v
Altitude    = %v
GroundSpeed = %v
Track       = %v
VRate       = %v
Latitude    = %v
Longitude   = %v
	`,
		stringstr(m.Datetime),
		stringstr(m.Modes),
		stringstr(m.Callsign),
		int32str(m.Altitude),
		int32str(m.GroundSpeed),
		int32str(m.Track),
		int32str(m.VRate),
		float32str(m.Latitude),
		float32str(m.Longitude))
}

func int32str(v *int32) string {
	if v == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%v", *v)
}

func stringstr(v *string) string {
	if v == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%v", *v)
}

func float32str(v *float32) string {
	if v == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%v", *v)
}

func boolstr(v *bool) string {
	if v == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%v", *v)
}
