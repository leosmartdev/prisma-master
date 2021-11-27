// Code generated by parse_nmea; DO NOT EDIT
package nmea

import "fmt"
import "strings"
import "strconv"

const (
	// PrefixVTG prefix
	PrefixVTG = "VTG"
)

// VTG represents fix data.
type CoreVTG struct {
	CourseOverGround float64

	CourseOverGroundValidity bool

	DegreesTrue string

	DegreesTrueValidity bool

	CourseOverGroundMagnetic float64

	CourseOverGroundMagneticValidity bool

	DegreesMagnetic string

	DegreesMagneticValidity bool

	SpeedOverGroundKnots float64

	SpeedOverGroundKnotsValidity bool

	Knots string

	KnotsValidity bool

	SpeedOverGroundKph float64

	SpeedOverGroundKphValidity bool

	Kilometers string

	KilometersValidity bool

	ModeIndicator string

	ModeIndicatorValidity bool
}

type VTG struct {
	BaseSentence
	CoreVTG
}

func NewVTG(sentence BaseSentence) *VTG {
	s := new(VTG)
	s.BaseSentence = sentence

	s.CourseOverGroundValidity = false

	s.DegreesTrueValidity = false

	s.CourseOverGroundMagneticValidity = false

	s.DegreesMagneticValidity = false

	s.SpeedOverGroundKnotsValidity = false

	s.KnotsValidity = false

	s.SpeedOverGroundKphValidity = false

	s.KilometersValidity = false

	s.ModeIndicatorValidity = false

	return s
}

func (s *VTG) parse() error {
	var err error

	if s.Format != PrefixVTG {
		err = fmt.Errorf("%s is not a %s", s.Format, PrefixVTG)
		return err
	}

	if len(s.Fields) == 0 {
		return nil
	} else {
		if s.Fields[0] != "" {
			i, err := strconv.ParseFloat(s.Fields[0], 64)
			if err != nil {
				return fmt.Errorf("VTG decode variation error: %s", s.Fields[0])
			} else {
				s.CoreVTG.CourseOverGround = float64(i)
				s.CoreVTG.CourseOverGroundValidity = true
			}

		}
	}

	if len(s.Fields) == 1 {
		return nil
	} else {
		if s.Fields[1] != "" {
			s.DegreesTrue = s.Fields[1]
			s.DegreesTrueValidity = true
		}
	}

	if len(s.Fields) == 2 {
		return nil
	} else {
		if s.Fields[2] != "" {
			i, err := strconv.ParseFloat(s.Fields[2], 64)
			if err != nil {
				return fmt.Errorf("VTG decode variation error: %s", s.Fields[2])
			} else {
				s.CoreVTG.CourseOverGroundMagnetic = float64(i)
				s.CoreVTG.CourseOverGroundMagneticValidity = true
			}

		}
	}

	if len(s.Fields) == 3 {
		return nil
	} else {
		if s.Fields[3] != "" {
			s.DegreesMagnetic = s.Fields[3]
			s.DegreesMagneticValidity = true
		}
	}

	if len(s.Fields) == 4 {
		return nil
	} else {
		if s.Fields[4] != "" {
			i, err := strconv.ParseFloat(s.Fields[4], 64)
			if err != nil {
				return fmt.Errorf("VTG decode variation error: %s", s.Fields[4])
			} else {
				s.CoreVTG.SpeedOverGroundKnots = float64(i)
				s.CoreVTG.SpeedOverGroundKnotsValidity = true
			}

		}
	}

	if len(s.Fields) == 5 {
		return nil
	} else {
		if s.Fields[5] != "" {
			s.Knots = s.Fields[5]
			s.KnotsValidity = true
		}
	}

	if len(s.Fields) == 6 {
		return nil
	} else {
		if s.Fields[6] != "" {
			i, err := strconv.ParseFloat(s.Fields[6], 64)
			if err != nil {
				return fmt.Errorf("VTG decode variation error: %s", s.Fields[6])
			} else {
				s.CoreVTG.SpeedOverGroundKph = float64(i)
				s.CoreVTG.SpeedOverGroundKphValidity = true
			}

		}
	}

	if len(s.Fields) == 7 {
		return nil
	} else {
		if s.Fields[7] != "" {
			s.Kilometers = s.Fields[7]
			s.KilometersValidity = true
		}
	}

	if len(s.Fields) == 8 {
		return nil
	} else {
		if s.Fields[8] != "" {
			s.ModeIndicator = s.Fields[8]
			s.ModeIndicatorValidity = true
		}
	}

	return nil
}

func (s *VTG) Encode() (string, error) {
	var Raw string

	if s.Format != PrefixVTG {
		err := fmt.Errorf("Sentence format %s is not a VTG sentence", s.Format)
		return "", err
	}

	Raw = s.SOS + s.Talker + s.Format

	if s.CourseOverGroundValidity == true {

		if len(Raw) > len(strings.TrimSuffix(Raw, ",")) {

			Raw = Raw + strconv.FormatFloat(s.CoreVTG.CourseOverGround, 'f', -1, 64)

		} else {
			Raw = Raw + "," + strconv.FormatFloat(s.CoreVTG.CourseOverGround, 'f', -1, 64)
		}

	} else if len(Raw) > len(strings.TrimSuffix(Raw, ",,")) {
		Raw = Raw + ","
	} else {
		Raw = Raw + ",,"
	}

	if s.DegreesTrueValidity == true {

		if len(Raw) > len(strings.TrimSuffix(Raw, ",")) {

			Raw = Raw + s.CoreVTG.DegreesTrue

		} else {
			Raw = Raw + "," + s.CoreVTG.DegreesTrue
		}

	} else if len(Raw) > len(strings.TrimSuffix(Raw, ",,")) {
		Raw = Raw + ","
	} else {
		Raw = Raw + ",,"
	}

	if s.CourseOverGroundMagneticValidity == true {

		if len(Raw) > len(strings.TrimSuffix(Raw, ",")) {

			Raw = Raw + strconv.FormatFloat(s.CoreVTG.CourseOverGroundMagnetic, 'f', -1, 64)

		} else {
			Raw = Raw + "," + strconv.FormatFloat(s.CoreVTG.CourseOverGroundMagnetic, 'f', -1, 64)
		}

	} else if len(Raw) > len(strings.TrimSuffix(Raw, ",,")) {
		Raw = Raw + ","
	} else {
		Raw = Raw + ",,"
	}

	if s.DegreesMagneticValidity == true {

		if len(Raw) > len(strings.TrimSuffix(Raw, ",")) {

			Raw = Raw + s.CoreVTG.DegreesMagnetic

		} else {
			Raw = Raw + "," + s.CoreVTG.DegreesMagnetic
		}

	} else if len(Raw) > len(strings.TrimSuffix(Raw, ",,")) {
		Raw = Raw + ","
	} else {
		Raw = Raw + ",,"
	}

	if s.SpeedOverGroundKnotsValidity == true {

		if len(Raw) > len(strings.TrimSuffix(Raw, ",")) {

			Raw = Raw + strconv.FormatFloat(s.CoreVTG.SpeedOverGroundKnots, 'f', -1, 64)

		} else {
			Raw = Raw + "," + strconv.FormatFloat(s.CoreVTG.SpeedOverGroundKnots, 'f', -1, 64)
		}

	} else if len(Raw) > len(strings.TrimSuffix(Raw, ",,")) {
		Raw = Raw + ","
	} else {
		Raw = Raw + ",,"
	}

	if s.KnotsValidity == true {

		if len(Raw) > len(strings.TrimSuffix(Raw, ",")) {

			Raw = Raw + s.CoreVTG.Knots

		} else {
			Raw = Raw + "," + s.CoreVTG.Knots
		}

	} else if len(Raw) > len(strings.TrimSuffix(Raw, ",,")) {
		Raw = Raw + ","
	} else {
		Raw = Raw + ",,"
	}

	if s.SpeedOverGroundKphValidity == true {

		if len(Raw) > len(strings.TrimSuffix(Raw, ",")) {

			Raw = Raw + strconv.FormatFloat(s.CoreVTG.SpeedOverGroundKph, 'f', -1, 64)

		} else {
			Raw = Raw + "," + strconv.FormatFloat(s.CoreVTG.SpeedOverGroundKph, 'f', -1, 64)
		}

	} else if len(Raw) > len(strings.TrimSuffix(Raw, ",,")) {
		Raw = Raw + ","
	} else {
		Raw = Raw + ",,"
	}

	if s.KilometersValidity == true {

		if len(Raw) > len(strings.TrimSuffix(Raw, ",")) {

			Raw = Raw + s.CoreVTG.Kilometers

		} else {
			Raw = Raw + "," + s.CoreVTG.Kilometers
		}

	} else if len(Raw) > len(strings.TrimSuffix(Raw, ",,")) {
		Raw = Raw + ","
	} else {
		Raw = Raw + ",,"
	}

	if s.ModeIndicatorValidity == true {

		if len(Raw) > len(strings.TrimSuffix(Raw, ",")) {

			Raw = Raw + s.CoreVTG.ModeIndicator

		} else {
			Raw = Raw + "," + s.CoreVTG.ModeIndicator
		}
	}

	check := Checksum(Raw)

	Raw = Raw + check

	return Raw, nil

}