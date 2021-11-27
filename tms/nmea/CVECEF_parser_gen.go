// Code generated by parse_nmea; DO NOT EDIT
package nmea

import "fmt"
import "strings"
import "strconv"

const (
	// PrefixCVECEF prefix
	PrefixCVECEF = "CVECEF"
)

// CVECEF represents fix data.
type CoreCVECEF struct {
	TrackID uint32

	TrackIDValidity bool

	MsttID uint32

	MsttIDValidity bool

	TrackConfidence float64

	TrackConfidenceValidity bool

	SensorNames string

	SensorNamesValidity bool

	CvX float64

	CvXValidity bool

	CvXvx float64

	CvXvxValidity bool

	CvXY float64

	CvXYValidity bool

	CvXvY float64

	CvXvYValidity bool

	CvXZ float64

	CvXZValidity bool

	CvXvz float64

	CvXvzValidity bool

	CvVxvx float64

	CvVxvxValidity bool

	CvVxy float64

	CvVxyValidity bool

	CvVxvy float64

	CvVxvyValidity bool

	CvVxz float64

	CvVxzValidity bool

	CvVxvz float64

	CvVxvzValidity bool

	CvY float64

	CvYValidity bool

	CvYvy float64

	CvYvyValidity bool

	CvYz float64

	CvYzValidity bool

	CvYvz float64

	CvYvzValidity bool

	CvVyvy float64

	CvVyvyValidity bool

	CvVyz float64

	CvVyzValidity bool

	CvVyvz float64

	CvVyvzValidity bool

	CvZ float64

	CvZValidity bool

	CvZvz float64

	CvZvzValidity bool

	CvVzvz float64

	CvVzvzValidity bool

	Spare string //Supposed to be Unknown

	SpareValidity bool
}

type CVECEF struct {
	BaseSentence
	CoreCVECEF
}

func NewCVECEF(sentence BaseSentence) *CVECEF {
	s := new(CVECEF)
	s.BaseSentence = sentence

	s.TrackIDValidity = false

	s.MsttIDValidity = false

	s.TrackConfidenceValidity = false

	s.SensorNamesValidity = false

	s.CvXValidity = false

	s.CvXvxValidity = false

	s.CvXYValidity = false

	s.CvXvYValidity = false

	s.CvXZValidity = false

	s.CvXvzValidity = false

	s.CvVxvxValidity = false

	s.CvVxyValidity = false

	s.CvVxvyValidity = false

	s.CvVxzValidity = false

	s.CvVxvzValidity = false

	s.CvYValidity = false

	s.CvYvyValidity = false

	s.CvYzValidity = false

	s.CvYvzValidity = false

	s.CvVyvyValidity = false

	s.CvVyzValidity = false

	s.CvVyvzValidity = false

	s.CvZValidity = false

	s.CvZvzValidity = false

	s.CvVzvzValidity = false

	s.SpareValidity = false

	return s
}

func (s *CVECEF) parse() error {
	var err error

	if s.Format != PrefixCVECEF {
		err = fmt.Errorf("%s is not a %s", s.Format, PrefixCVECEF)
		return err
	}

	if len(s.Fields) == 0 {
		return nil
	} else {
		if s.Fields[0] != "" {
			i, err := strconv.ParseUint(s.Fields[0], 10, 32)
			if err != nil {
				return fmt.Errorf("CVECEF decode variation error: %s", s.Fields[0])
			} else {
				s.CoreCVECEF.TrackID = uint32(i)
				s.CoreCVECEF.TrackIDValidity = true
			}

		}
	}

	if len(s.Fields) == 1 {
		return nil
	} else {
		if s.Fields[1] != "" {
			i, err := strconv.ParseUint(s.Fields[1], 10, 32)
			if err != nil {
				return fmt.Errorf("CVECEF decode variation error: %s", s.Fields[1])
			} else {
				s.CoreCVECEF.MsttID = uint32(i)
				s.CoreCVECEF.MsttIDValidity = true
			}

		}
	}

	if len(s.Fields) == 2 {
		return nil
	} else {
		if s.Fields[2] != "" {
			i, err := strconv.ParseFloat(s.Fields[2], 64)
			if err != nil {
				return fmt.Errorf("CVECEF decode variation error: %s", s.Fields[2])
			} else {
				s.CoreCVECEF.TrackConfidence = float64(i)
				s.CoreCVECEF.TrackConfidenceValidity = true
			}

		}
	}

	if len(s.Fields) == 3 {
		return nil
	} else {
		if s.Fields[3] != "" {
			s.SensorNames = s.Fields[3]
			s.SensorNamesValidity = true
		}
	}

	if len(s.Fields) == 4 {
		return nil
	} else {
		if s.Fields[4] != "" {
			i, err := strconv.ParseFloat(s.Fields[4], 64)
			if err != nil {
				return fmt.Errorf("CVECEF decode variation error: %s", s.Fields[4])
			} else {
				s.CoreCVECEF.CvX = float64(i)
				s.CoreCVECEF.CvXValidity = true
			}

		}
	}

	if len(s.Fields) == 5 {
		return nil
	} else {
		if s.Fields[5] != "" {
			i, err := strconv.ParseFloat(s.Fields[5], 64)
			if err != nil {
				return fmt.Errorf("CVECEF decode variation error: %s", s.Fields[5])
			} else {
				s.CoreCVECEF.CvXvx = float64(i)
				s.CoreCVECEF.CvXvxValidity = true
			}

		}
	}

	if len(s.Fields) == 6 {
		return nil
	} else {
		if s.Fields[6] != "" {
			i, err := strconv.ParseFloat(s.Fields[6], 64)
			if err != nil {
				return fmt.Errorf("CVECEF decode variation error: %s", s.Fields[6])
			} else {
				s.CoreCVECEF.CvXY = float64(i)
				s.CoreCVECEF.CvXYValidity = true
			}

		}
	}

	if len(s.Fields) == 7 {
		return nil
	} else {
		if s.Fields[7] != "" {
			i, err := strconv.ParseFloat(s.Fields[7], 64)
			if err != nil {
				return fmt.Errorf("CVECEF decode variation error: %s", s.Fields[7])
			} else {
				s.CoreCVECEF.CvXvY = float64(i)
				s.CoreCVECEF.CvXvYValidity = true
			}

		}
	}

	if len(s.Fields) == 8 {
		return nil
	} else {
		if s.Fields[8] != "" {
			i, err := strconv.ParseFloat(s.Fields[8], 64)
			if err != nil {
				return fmt.Errorf("CVECEF decode variation error: %s", s.Fields[8])
			} else {
				s.CoreCVECEF.CvXZ = float64(i)
				s.CoreCVECEF.CvXZValidity = true
			}

		}
	}

	if len(s.Fields) == 9 {
		return nil
	} else {
		if s.Fields[9] != "" {
			i, err := strconv.ParseFloat(s.Fields[9], 64)
			if err != nil {
				return fmt.Errorf("CVECEF decode variation error: %s", s.Fields[9])
			} else {
				s.CoreCVECEF.CvXvz = float64(i)
				s.CoreCVECEF.CvXvzValidity = true
			}

		}
	}

	if len(s.Fields) == 10 {
		return nil
	} else {
		if s.Fields[10] != "" {
			i, err := strconv.ParseFloat(s.Fields[10], 64)
			if err != nil {
				return fmt.Errorf("CVECEF decode variation error: %s", s.Fields[10])
			} else {
				s.CoreCVECEF.CvVxvx = float64(i)
				s.CoreCVECEF.CvVxvxValidity = true
			}

		}
	}

	if len(s.Fields) == 11 {
		return nil
	} else {
		if s.Fields[11] != "" {
			i, err := strconv.ParseFloat(s.Fields[11], 64)
			if err != nil {
				return fmt.Errorf("CVECEF decode variation error: %s", s.Fields[11])
			} else {
				s.CoreCVECEF.CvVxy = float64(i)
				s.CoreCVECEF.CvVxyValidity = true
			}

		}
	}

	if len(s.Fields) == 12 {
		return nil
	} else {
		if s.Fields[12] != "" {
			i, err := strconv.ParseFloat(s.Fields[12], 64)
			if err != nil {
				return fmt.Errorf("CVECEF decode variation error: %s", s.Fields[12])
			} else {
				s.CoreCVECEF.CvVxvy = float64(i)
				s.CoreCVECEF.CvVxvyValidity = true
			}

		}
	}

	if len(s.Fields) == 13 {
		return nil
	} else {
		if s.Fields[13] != "" {
			i, err := strconv.ParseFloat(s.Fields[13], 64)
			if err != nil {
				return fmt.Errorf("CVECEF decode variation error: %s", s.Fields[13])
			} else {
				s.CoreCVECEF.CvVxz = float64(i)
				s.CoreCVECEF.CvVxzValidity = true
			}

		}
	}

	if len(s.Fields) == 14 {
		return nil
	} else {
		if s.Fields[14] != "" {
			i, err := strconv.ParseFloat(s.Fields[14], 64)
			if err != nil {
				return fmt.Errorf("CVECEF decode variation error: %s", s.Fields[14])
			} else {
				s.CoreCVECEF.CvVxvz = float64(i)
				s.CoreCVECEF.CvVxvzValidity = true
			}

		}
	}

	if len(s.Fields) == 15 {
		return nil
	} else {
		if s.Fields[15] != "" {
			i, err := strconv.ParseFloat(s.Fields[15], 64)
			if err != nil {
				return fmt.Errorf("CVECEF decode variation error: %s", s.Fields[15])
			} else {
				s.CoreCVECEF.CvY = float64(i)
				s.CoreCVECEF.CvYValidity = true
			}

		}
	}

	if len(s.Fields) == 16 {
		return nil
	} else {
		if s.Fields[16] != "" {
			i, err := strconv.ParseFloat(s.Fields[16], 64)
			if err != nil {
				return fmt.Errorf("CVECEF decode variation error: %s", s.Fields[16])
			} else {
				s.CoreCVECEF.CvYvy = float64(i)
				s.CoreCVECEF.CvYvyValidity = true
			}

		}
	}

	if len(s.Fields) == 17 {
		return nil
	} else {
		if s.Fields[17] != "" {
			i, err := strconv.ParseFloat(s.Fields[17], 64)
			if err != nil {
				return fmt.Errorf("CVECEF decode variation error: %s", s.Fields[17])
			} else {
				s.CoreCVECEF.CvYz = float64(i)
				s.CoreCVECEF.CvYzValidity = true
			}

		}
	}

	if len(s.Fields) == 18 {
		return nil
	} else {
		if s.Fields[18] != "" {
			i, err := strconv.ParseFloat(s.Fields[18], 64)
			if err != nil {
				return fmt.Errorf("CVECEF decode variation error: %s", s.Fields[18])
			} else {
				s.CoreCVECEF.CvYvz = float64(i)
				s.CoreCVECEF.CvYvzValidity = true
			}

		}
	}

	if len(s.Fields) == 19 {
		return nil
	} else {
		if s.Fields[19] != "" {
			i, err := strconv.ParseFloat(s.Fields[19], 64)
			if err != nil {
				return fmt.Errorf("CVECEF decode variation error: %s", s.Fields[19])
			} else {
				s.CoreCVECEF.CvVyvy = float64(i)
				s.CoreCVECEF.CvVyvyValidity = true
			}

		}
	}

	if len(s.Fields) == 20 {
		return nil
	} else {
		if s.Fields[20] != "" {
			i, err := strconv.ParseFloat(s.Fields[20], 64)
			if err != nil {
				return fmt.Errorf("CVECEF decode variation error: %s", s.Fields[20])
			} else {
				s.CoreCVECEF.CvVyz = float64(i)
				s.CoreCVECEF.CvVyzValidity = true
			}

		}
	}

	if len(s.Fields) == 21 {
		return nil
	} else {
		if s.Fields[21] != "" {
			i, err := strconv.ParseFloat(s.Fields[21], 64)
			if err != nil {
				return fmt.Errorf("CVECEF decode variation error: %s", s.Fields[21])
			} else {
				s.CoreCVECEF.CvVyvz = float64(i)
				s.CoreCVECEF.CvVyvzValidity = true
			}

		}
	}

	if len(s.Fields) == 22 {
		return nil
	} else {
		if s.Fields[22] != "" {
			i, err := strconv.ParseFloat(s.Fields[22], 64)
			if err != nil {
				return fmt.Errorf("CVECEF decode variation error: %s", s.Fields[22])
			} else {
				s.CoreCVECEF.CvZ = float64(i)
				s.CoreCVECEF.CvZValidity = true
			}

		}
	}

	if len(s.Fields) == 23 {
		return nil
	} else {
		if s.Fields[23] != "" {
			i, err := strconv.ParseFloat(s.Fields[23], 64)
			if err != nil {
				return fmt.Errorf("CVECEF decode variation error: %s", s.Fields[23])
			} else {
				s.CoreCVECEF.CvZvz = float64(i)
				s.CoreCVECEF.CvZvzValidity = true
			}

		}
	}

	if len(s.Fields) == 24 {
		return nil
	} else {
		if s.Fields[24] != "" {
			i, err := strconv.ParseFloat(s.Fields[24], 64)
			if err != nil {
				return fmt.Errorf("CVECEF decode variation error: %s", s.Fields[24])
			} else {
				s.CoreCVECEF.CvVzvz = float64(i)
				s.CoreCVECEF.CvVzvzValidity = true
			}

		}
	}

	return nil
}

func (s *CVECEF) Encode() (string, error) {
	var Raw string

	if s.Format != PrefixCVECEF {
		err := fmt.Errorf("Sentence format %s is not a CVECEF sentence", s.Format)
		return "", err
	}

	Raw = s.SOS + s.Talker + s.Format

	if s.TrackIDValidity == true {

		if len(Raw) > len(strings.TrimSuffix(Raw, ",")) {

			Raw = Raw + "," + strconv.FormatUint(uint64(s.CoreCVECEF.TrackID), 10)

		} else {
			Raw = Raw + "," + strconv.FormatUint(uint64(s.CoreCVECEF.TrackID), 10)
		}

	} else if len(Raw) > len(strings.TrimSuffix(Raw, ",,")) {
		Raw = Raw + ","
	} else {
		Raw = Raw + ",,"
	}

	if s.MsttIDValidity == true {

		if len(Raw) > len(strings.TrimSuffix(Raw, ",")) {

			Raw = Raw + "," + strconv.FormatUint(uint64(s.CoreCVECEF.MsttID), 10)

		} else {
			Raw = Raw + "," + strconv.FormatUint(uint64(s.CoreCVECEF.MsttID), 10)
		}

	} else if len(Raw) > len(strings.TrimSuffix(Raw, ",,")) {
		Raw = Raw + ","
	} else {
		Raw = Raw + ",,"
	}

	if s.TrackConfidenceValidity == true {

		if len(Raw) > len(strings.TrimSuffix(Raw, ",")) {

			Raw = Raw + strconv.FormatFloat(s.CoreCVECEF.TrackConfidence, 'f', -1, 64)

		} else {
			Raw = Raw + "," + strconv.FormatFloat(s.CoreCVECEF.TrackConfidence, 'f', -1, 64)
		}

	} else if len(Raw) > len(strings.TrimSuffix(Raw, ",,")) {
		Raw = Raw + ","
	} else {
		Raw = Raw + ",,"
	}

	if s.SensorNamesValidity == true {

		if len(Raw) > len(strings.TrimSuffix(Raw, ",")) {

			Raw = Raw + s.CoreCVECEF.SensorNames

		} else {
			Raw = Raw + "," + s.CoreCVECEF.SensorNames
		}

	} else if len(Raw) > len(strings.TrimSuffix(Raw, ",,")) {
		Raw = Raw + ","
	} else {
		Raw = Raw + ",,"
	}

	if s.CvXValidity == true {

		if len(Raw) > len(strings.TrimSuffix(Raw, ",")) {

			Raw = Raw + strconv.FormatFloat(s.CoreCVECEF.CvX, 'f', -1, 64)

		} else {
			Raw = Raw + "," + strconv.FormatFloat(s.CoreCVECEF.CvX, 'f', -1, 64)
		}

	} else if len(Raw) > len(strings.TrimSuffix(Raw, ",,")) {
		Raw = Raw + ","
	} else {
		Raw = Raw + ",,"
	}

	if s.CvXvxValidity == true {

		if len(Raw) > len(strings.TrimSuffix(Raw, ",")) {

			Raw = Raw + strconv.FormatFloat(s.CoreCVECEF.CvXvx, 'f', -1, 64)

		} else {
			Raw = Raw + "," + strconv.FormatFloat(s.CoreCVECEF.CvXvx, 'f', -1, 64)
		}

	} else if len(Raw) > len(strings.TrimSuffix(Raw, ",,")) {
		Raw = Raw + ","
	} else {
		Raw = Raw + ",,"
	}

	if s.CvXYValidity == true {

		if len(Raw) > len(strings.TrimSuffix(Raw, ",")) {

			Raw = Raw + strconv.FormatFloat(s.CoreCVECEF.CvXY, 'f', -1, 64)

		} else {
			Raw = Raw + "," + strconv.FormatFloat(s.CoreCVECEF.CvXY, 'f', -1, 64)
		}

	} else if len(Raw) > len(strings.TrimSuffix(Raw, ",,")) {
		Raw = Raw + ","
	} else {
		Raw = Raw + ",,"
	}

	if s.CvXvYValidity == true {

		if len(Raw) > len(strings.TrimSuffix(Raw, ",")) {

			Raw = Raw + strconv.FormatFloat(s.CoreCVECEF.CvXvY, 'f', -1, 64)

		} else {
			Raw = Raw + "," + strconv.FormatFloat(s.CoreCVECEF.CvXvY, 'f', -1, 64)
		}

	} else if len(Raw) > len(strings.TrimSuffix(Raw, ",,")) {
		Raw = Raw + ","
	} else {
		Raw = Raw + ",,"
	}

	if s.CvXZValidity == true {

		if len(Raw) > len(strings.TrimSuffix(Raw, ",")) {

			Raw = Raw + strconv.FormatFloat(s.CoreCVECEF.CvXZ, 'f', -1, 64)

		} else {
			Raw = Raw + "," + strconv.FormatFloat(s.CoreCVECEF.CvXZ, 'f', -1, 64)
		}

	} else if len(Raw) > len(strings.TrimSuffix(Raw, ",,")) {
		Raw = Raw + ","
	} else {
		Raw = Raw + ",,"
	}

	if s.CvXvzValidity == true {

		if len(Raw) > len(strings.TrimSuffix(Raw, ",")) {

			Raw = Raw + strconv.FormatFloat(s.CoreCVECEF.CvXvz, 'f', -1, 64)

		} else {
			Raw = Raw + "," + strconv.FormatFloat(s.CoreCVECEF.CvXvz, 'f', -1, 64)
		}

	} else if len(Raw) > len(strings.TrimSuffix(Raw, ",,")) {
		Raw = Raw + ","
	} else {
		Raw = Raw + ",,"
	}

	if s.CvVxvxValidity == true {

		if len(Raw) > len(strings.TrimSuffix(Raw, ",")) {

			Raw = Raw + strconv.FormatFloat(s.CoreCVECEF.CvVxvx, 'f', -1, 64)

		} else {
			Raw = Raw + "," + strconv.FormatFloat(s.CoreCVECEF.CvVxvx, 'f', -1, 64)
		}

	} else if len(Raw) > len(strings.TrimSuffix(Raw, ",,")) {
		Raw = Raw + ","
	} else {
		Raw = Raw + ",,"
	}

	if s.CvVxyValidity == true {

		if len(Raw) > len(strings.TrimSuffix(Raw, ",")) {

			Raw = Raw + strconv.FormatFloat(s.CoreCVECEF.CvVxy, 'f', -1, 64)

		} else {
			Raw = Raw + "," + strconv.FormatFloat(s.CoreCVECEF.CvVxy, 'f', -1, 64)
		}

	} else if len(Raw) > len(strings.TrimSuffix(Raw, ",,")) {
		Raw = Raw + ","
	} else {
		Raw = Raw + ",,"
	}

	if s.CvVxvyValidity == true {

		if len(Raw) > len(strings.TrimSuffix(Raw, ",")) {

			Raw = Raw + strconv.FormatFloat(s.CoreCVECEF.CvVxvy, 'f', -1, 64)

		} else {
			Raw = Raw + "," + strconv.FormatFloat(s.CoreCVECEF.CvVxvy, 'f', -1, 64)
		}

	} else if len(Raw) > len(strings.TrimSuffix(Raw, ",,")) {
		Raw = Raw + ","
	} else {
		Raw = Raw + ",,"
	}

	if s.CvVxzValidity == true {

		if len(Raw) > len(strings.TrimSuffix(Raw, ",")) {

			Raw = Raw + strconv.FormatFloat(s.CoreCVECEF.CvVxz, 'f', -1, 64)

		} else {
			Raw = Raw + "," + strconv.FormatFloat(s.CoreCVECEF.CvVxz, 'f', -1, 64)
		}

	} else if len(Raw) > len(strings.TrimSuffix(Raw, ",,")) {
		Raw = Raw + ","
	} else {
		Raw = Raw + ",,"
	}

	if s.CvVxvzValidity == true {

		if len(Raw) > len(strings.TrimSuffix(Raw, ",")) {

			Raw = Raw + strconv.FormatFloat(s.CoreCVECEF.CvVxvz, 'f', -1, 64)

		} else {
			Raw = Raw + "," + strconv.FormatFloat(s.CoreCVECEF.CvVxvz, 'f', -1, 64)
		}

	} else if len(Raw) > len(strings.TrimSuffix(Raw, ",,")) {
		Raw = Raw + ","
	} else {
		Raw = Raw + ",,"
	}

	if s.CvYValidity == true {

		if len(Raw) > len(strings.TrimSuffix(Raw, ",")) {

			Raw = Raw + strconv.FormatFloat(s.CoreCVECEF.CvY, 'f', -1, 64)

		} else {
			Raw = Raw + "," + strconv.FormatFloat(s.CoreCVECEF.CvY, 'f', -1, 64)
		}

	} else if len(Raw) > len(strings.TrimSuffix(Raw, ",,")) {
		Raw = Raw + ","
	} else {
		Raw = Raw + ",,"
	}

	if s.CvYvyValidity == true {

		if len(Raw) > len(strings.TrimSuffix(Raw, ",")) {

			Raw = Raw + strconv.FormatFloat(s.CoreCVECEF.CvYvy, 'f', -1, 64)

		} else {
			Raw = Raw + "," + strconv.FormatFloat(s.CoreCVECEF.CvYvy, 'f', -1, 64)
		}

	} else if len(Raw) > len(strings.TrimSuffix(Raw, ",,")) {
		Raw = Raw + ","
	} else {
		Raw = Raw + ",,"
	}

	if s.CvYzValidity == true {

		if len(Raw) > len(strings.TrimSuffix(Raw, ",")) {

			Raw = Raw + strconv.FormatFloat(s.CoreCVECEF.CvYz, 'f', -1, 64)

		} else {
			Raw = Raw + "," + strconv.FormatFloat(s.CoreCVECEF.CvYz, 'f', -1, 64)
		}

	} else if len(Raw) > len(strings.TrimSuffix(Raw, ",,")) {
		Raw = Raw + ","
	} else {
		Raw = Raw + ",,"
	}

	if s.CvYvzValidity == true {

		if len(Raw) > len(strings.TrimSuffix(Raw, ",")) {

			Raw = Raw + strconv.FormatFloat(s.CoreCVECEF.CvYvz, 'f', -1, 64)

		} else {
			Raw = Raw + "," + strconv.FormatFloat(s.CoreCVECEF.CvYvz, 'f', -1, 64)
		}

	} else if len(Raw) > len(strings.TrimSuffix(Raw, ",,")) {
		Raw = Raw + ","
	} else {
		Raw = Raw + ",,"
	}

	if s.CvVyvyValidity == true {

		if len(Raw) > len(strings.TrimSuffix(Raw, ",")) {

			Raw = Raw + strconv.FormatFloat(s.CoreCVECEF.CvVyvy, 'f', -1, 64)

		} else {
			Raw = Raw + "," + strconv.FormatFloat(s.CoreCVECEF.CvVyvy, 'f', -1, 64)
		}

	} else if len(Raw) > len(strings.TrimSuffix(Raw, ",,")) {
		Raw = Raw + ","
	} else {
		Raw = Raw + ",,"
	}

	if s.CvVyzValidity == true {

		if len(Raw) > len(strings.TrimSuffix(Raw, ",")) {

			Raw = Raw + strconv.FormatFloat(s.CoreCVECEF.CvVyz, 'f', -1, 64)

		} else {
			Raw = Raw + "," + strconv.FormatFloat(s.CoreCVECEF.CvVyz, 'f', -1, 64)
		}

	} else if len(Raw) > len(strings.TrimSuffix(Raw, ",,")) {
		Raw = Raw + ","
	} else {
		Raw = Raw + ",,"
	}

	if s.CvVyvzValidity == true {

		if len(Raw) > len(strings.TrimSuffix(Raw, ",")) {

			Raw = Raw + strconv.FormatFloat(s.CoreCVECEF.CvVyvz, 'f', -1, 64)

		} else {
			Raw = Raw + "," + strconv.FormatFloat(s.CoreCVECEF.CvVyvz, 'f', -1, 64)
		}

	} else if len(Raw) > len(strings.TrimSuffix(Raw, ",,")) {
		Raw = Raw + ","
	} else {
		Raw = Raw + ",,"
	}

	if s.CvZValidity == true {

		if len(Raw) > len(strings.TrimSuffix(Raw, ",")) {

			Raw = Raw + strconv.FormatFloat(s.CoreCVECEF.CvZ, 'f', -1, 64)

		} else {
			Raw = Raw + "," + strconv.FormatFloat(s.CoreCVECEF.CvZ, 'f', -1, 64)
		}

	} else if len(Raw) > len(strings.TrimSuffix(Raw, ",,")) {
		Raw = Raw + ","
	} else {
		Raw = Raw + ",,"
	}

	if s.CvZvzValidity == true {

		if len(Raw) > len(strings.TrimSuffix(Raw, ",")) {

			Raw = Raw + strconv.FormatFloat(s.CoreCVECEF.CvZvz, 'f', -1, 64)

		} else {
			Raw = Raw + "," + strconv.FormatFloat(s.CoreCVECEF.CvZvz, 'f', -1, 64)
		}

	} else if len(Raw) > len(strings.TrimSuffix(Raw, ",,")) {
		Raw = Raw + ","
	} else {
		Raw = Raw + ",,"
	}

	if s.CvVzvzValidity == true {

		if len(Raw) > len(strings.TrimSuffix(Raw, ",")) {

			Raw = Raw + strconv.FormatFloat(s.CoreCVECEF.CvVzvz, 'f', -1, 64)

		} else {
			Raw = Raw + "," + strconv.FormatFloat(s.CoreCVECEF.CvVzvz, 'f', -1, 64)
		}

	} else if len(Raw) > len(strings.TrimSuffix(Raw, ",,")) {
		Raw = Raw + ","
	} else {
		Raw = Raw + ",,"
	}

	check := Checksum(Raw)

	Raw = Raw + check

	return Raw, nil

}