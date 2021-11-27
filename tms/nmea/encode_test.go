package nmea

import (
	"fmt"
	"testing"
)

func TestEncode(t *testing.T) {
	//3554.4456,N,00528.9195,W,135138,
	s := BaseSentence{}
	s.SOS = "$"
	s.Talker = "GP"
	s.Format = "GLL"

	coregll := CoreGLL{3554.4456, true, "N", true, 00528.9195, true, "W", true, 135138, false, "", false, "V", false}

	GLL := GLL{s, coregll}

	str, err := GLL.Encode()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(str)

	sen := BaseSentence{}
	sen.SOS = "!"
	sen.Talker = "AI"
	sen.Format = "VDM"

	corevdmo := CoreVDMO{1, true, 1, true, 0, true, "A", true, "", false, 0, true}
	VDM := VDMO{sen, corevdmo}

	corem137110 := CoreM137110{10, 0, 265547250, 0, 2500912, 0}

	M1371 := M137110{VDM, corem137110}

	str, err = M1371.Encode()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(str)

	sen = BaseSentence{}
	sen.SOS = "!"
	sen.Talker = "AI"
	sen.Format = "VDM"

	corevdmo = CoreVDMO{1, true, 1, true, 0, false, "B", false, "", false, 4, false}
	VDM = VDMO{sen, corevdmo}

	corem13711 := CoreM137121{21, 0, 992429100, 14, "TANGER MED 2", false, 265115292, 21528796, 1, 1, 1, 1, 1, 46, false, 226, true, false, false, 0, "MEHDI", 1}
	M13711 := M137121{VDM, corem13711}

	str, err = M13711.Encode()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(str)

}
