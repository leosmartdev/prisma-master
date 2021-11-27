// Code generated by parse_nmea; DO NOT EDIT
package nmea

import "fmt"
import "strconv"
import "strings"

// M137127 represents fix data.
type CoreM137127 struct {
	MessageID uint8

	RepeatIndicator uint32

	Mmsi uint32

	PositionAccuracy bool

	RaimFlag bool

	NavigationalStatus uint32

	Longitude float64

	Latitude float64

	SpeedOverGround uint32

	CourseOverGround uint32

	PositionLatency bool

	Spare uint32 //Supposed to be Unknown
}
type M137127 struct {
	VDMO
	CoreM137127
}

func NewM137127(sentence VDMO) *M137127 {
	s := new(M137127)
	s.VDMO = sentence
	return s
}

func (s *M137127) parse() error {
	var err error

	if MessageType(s.EncapData) != 27 {
		err = fmt.Errorf("message %d is not a M137127", MessageType(s.EncapData))
		return err
	}

	data := []byte(s.EncapData)

	//if len(data)*6 > 96 {
	//	err = fmt.Errorf("Message lenght is larger than it should be [%d!=96]", len(data)*6)
	//	return err
	//}

	s.MessageID = MessageType(s.EncapData)

	s.CoreM137127.RepeatIndicator = BitsToInt(6, 7, data)

	s.CoreM137127.Mmsi = BitsToInt(8, 37, data)

	s.CoreM137127.PositionAccuracy = CbnBool(38, data)

	s.CoreM137127.RaimFlag = CbnBool(39, data)

	s.CoreM137127.NavigationalStatus = BitsToInt(40, 43, data)

	s.CoreM137127.Longitude = (float64(int32(BitsToInt(44, 61, data)) << 4)) / 16

	s.CoreM137127.Latitude = (float64(int32(BitsToInt(62, 78, data)) << 5)) / 32

	s.CoreM137127.SpeedOverGround = BitsToInt(79, 84, data)

	s.CoreM137127.CourseOverGround = BitsToInt(85, 93, data)

	s.CoreM137127.PositionLatency = CbnBool(94, data)

	s.CoreM137127.Spare = BitsToInt(95, 6*(len(data)-1), data)

	return nil
}

func (s *M137127) Encode() (string, error) {
	var Raw string
	var Sbinary string

	if s.MessageID != 27 {
		err := fmt.Errorf("message %d is not a M137127", s.MessageID)
		return "", err
	}

	Raw = s.SOS + s.Talker + s.Format + ","

	if s.SentenceCountValidity == true {
		Raw = Raw + strconv.FormatInt(int64(s.SentenceCount), 10) + ","
	} else {
		Raw = Raw + ","
	}

	if s.SentenceIndexValidity == true {
		Raw = Raw + strconv.FormatInt(int64(s.SentenceIndex), 10) + ","
	} else {
		Raw = Raw + ","
	}

	if s.SeqMsgIDValidity == true {
		Raw = Raw + strconv.FormatInt(int64(s.SeqMsgID), 10) + ","
	} else {
		Raw = Raw + ","
	}

	if s.ChannelValidity == true {
		Raw = Raw + s.Channel
	}

	str := strconv.FormatInt(int64(s.CoreM137127.MessageID), 2)
	for len(str) < 6 {
		str = "0" + str
	}
	Sbinary = Sbinary + str

	str = strconv.FormatInt(int64(s.CoreM137127.RepeatIndicator), 2)
	for len(str) < 2 {
		str = "0" + str
	}
	Sbinary = Sbinary + str

	str = strconv.FormatInt(int64(s.CoreM137127.Mmsi), 2)
	for len(str) < 30 {
		str = "0" + str
	}
	Sbinary = Sbinary + str

	if s.PositionAccuracy == true {
		str = "1"
	} else {
		str = "0"
	}

	Sbinary = Sbinary + str

	if s.RaimFlag == true {
		str = "1"
	} else {
		str = "0"
	}

	Sbinary = Sbinary + str

	str = strconv.FormatInt(int64(s.CoreM137127.NavigationalStatus), 2)
	for len(str) < 4 {
		str = "0" + str
	}
	Sbinary = Sbinary + str

	str = strconv.FormatInt(int64(s.CoreM137127.Longitude), 2)
	for len(str) < 18 {
		str = "0" + str
	}
	Sbinary = Sbinary + str

	str = strconv.FormatInt(int64(s.CoreM137127.Latitude), 2)
	for len(str) < 17 {
		str = "0" + str
	}
	Sbinary = Sbinary + str

	str = strconv.FormatInt(int64(s.CoreM137127.SpeedOverGround), 2)
	for len(str) < 6 {
		str = "0" + str
	}
	Sbinary = Sbinary + str

	str = strconv.FormatInt(int64(s.CoreM137127.CourseOverGround), 2)
	for len(str) < 9 {
		str = "0" + str
	}
	Sbinary = Sbinary + str

	if s.PositionLatency == true {
		str = "1"
	} else {
		str = "0"
	}

	Sbinary = Sbinary + str

	str = strconv.FormatInt(int64(s.CoreM137127.Spare), 2)
	for len(str) < 1 {
		str = "0" + str
	}
	Sbinary = Sbinary + str

	field := strings.SplitN(Sbinary, "", len(Sbinary))

	var encdata = make([]string, int((len(Sbinary)+int(s.FillBits))/6))

	j := 0
	for i := 0; i < int((len(Sbinary)+int(s.FillBits))/6); i++ {

		if i == (int((len(Sbinary)+int(s.FillBits))/6) - 1) {
			for j < len(Sbinary) {
				encdata[i] = encdata[i] + field[j]
				j = j + 1
			}
			for h := 0; h < int(s.FillBits); h++ {
				encdata[i] = encdata[i] + "0" // fill bits
			}
		} else {
			encdata[i] = field[j] + field[j+1] + field[j+2] + field[j+3] + field[j+4] + field[j+5]
			j = j + 6
		}
	}

	var data string
	for j := 0; j < int((len(Sbinary)+int(s.FillBits))/6); j++ {
		i, _ := strconv.ParseInt(encdata[j], 2, 8)
		if i < 40 {
			i = i + 48
		} else {
			i = i + 8 + 48
		}
		data = data + string(rune(i))
	}

	Raw = Raw + "," + data + ","

	if s.FillBitsValidity == true {
		Raw = Raw + strconv.FormatInt(int64(s.FillBits), 10)
	}

	check := Checksum(Raw)

	Raw = Raw + check

	return Raw, nil

}