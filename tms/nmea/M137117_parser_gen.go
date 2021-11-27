// Code generated by parse_nmea; DO NOT EDIT
package nmea

import "fmt"
import "strconv"
import "strings"

// M137117 represents fix data.
type CoreM137117 struct {
	MessageID uint8

	RepeatIndicator uint32

	Mmsi uint32

	Spare1 uint32 //Supposed to be Unknown

	Longitude float64

	Latitude float64

	Spare2 uint32 //Supposed to be Unknown

	MessageType uint32

	StationID uint32

	Zcount uint32

	SeqNumber uint32

	N uint32

	Health uint32

	DgnssDataWords uint32 //Supposed to be Unknown
}
type M137117 struct {
	VDMO
	CoreM137117
}

func NewM137117(sentence VDMO) *M137117 {
	s := new(M137117)
	s.VDMO = sentence
	return s
}

func (s *M137117) parse() error {
	var err error

	if MessageType(s.EncapData) != 17 {
		err = fmt.Errorf("message %d is not a M137117", MessageType(s.EncapData))
		return err
	}

	data := []byte(s.EncapData)

	//if len(data)*6 > 813 {
	//	err = fmt.Errorf("Message lenght is larger than it should be [%d!=813]", len(data)*6)
	//	return err
	//}

	s.MessageID = MessageType(s.EncapData)

	s.CoreM137117.RepeatIndicator = BitsToInt(6, 7, data)

	s.CoreM137117.Mmsi = BitsToInt(8, 37, data)

	s.CoreM137117.Spare1 = BitsToInt(38, 39, data)

	s.CoreM137117.Longitude = (float64(int32(BitsToInt(40, 57, data)) << 4)) / 16

	s.CoreM137117.Latitude = (float64(int32(BitsToInt(58, 74, data)) << 5)) / 32

	s.CoreM137117.Spare2 = BitsToInt(75, 76, data)

	s.CoreM137117.MessageType = BitsToInt(77, 82, data)

	s.CoreM137117.StationID = BitsToInt(83, 92, data)

	s.CoreM137117.Zcount = BitsToInt(93, 105, data)

	s.CoreM137117.SeqNumber = BitsToInt(106, 108, data)

	s.CoreM137117.N = BitsToInt(109, 113, data)

	s.CoreM137117.Health = BitsToInt(114, 116, data)

	s.CoreM137117.DgnssDataWords = BitsToInt(117, 6*(len(data)-1), data)

	return nil
}

func (s *M137117) Encode() (string, error) {
	var Raw string
	var Sbinary string

	if s.MessageID != 17 {
		err := fmt.Errorf("message %d is not a M137117", s.MessageID)
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

	str := strconv.FormatInt(int64(s.CoreM137117.MessageID), 2)
	for len(str) < 6 {
		str = "0" + str
	}
	Sbinary = Sbinary + str

	str = strconv.FormatInt(int64(s.CoreM137117.RepeatIndicator), 2)
	for len(str) < 2 {
		str = "0" + str
	}
	Sbinary = Sbinary + str

	str = strconv.FormatInt(int64(s.CoreM137117.Mmsi), 2)
	for len(str) < 30 {
		str = "0" + str
	}
	Sbinary = Sbinary + str

	str = strconv.FormatInt(int64(s.CoreM137117.Spare1), 2)
	for len(str) < 2 {
		str = "0" + str
	}
	Sbinary = Sbinary + str

	str = strconv.FormatInt(int64(s.CoreM137117.Longitude), 2)
	for len(str) < 18 {
		str = "0" + str
	}
	Sbinary = Sbinary + str

	str = strconv.FormatInt(int64(s.CoreM137117.Latitude), 2)
	for len(str) < 17 {
		str = "0" + str
	}
	Sbinary = Sbinary + str

	str = strconv.FormatInt(int64(s.CoreM137117.Spare2), 2)
	for len(str) < 2 {
		str = "0" + str
	}
	Sbinary = Sbinary + str

	str = strconv.FormatInt(int64(s.CoreM137117.MessageType), 2)
	for len(str) < 6 {
		str = "0" + str
	}
	Sbinary = Sbinary + str

	str = strconv.FormatInt(int64(s.CoreM137117.StationID), 2)
	for len(str) < 10 {
		str = "0" + str
	}
	Sbinary = Sbinary + str

	str = strconv.FormatInt(int64(s.CoreM137117.Zcount), 2)
	for len(str) < 13 {
		str = "0" + str
	}
	Sbinary = Sbinary + str

	str = strconv.FormatInt(int64(s.CoreM137117.SeqNumber), 2)
	for len(str) < 3 {
		str = "0" + str
	}
	Sbinary = Sbinary + str

	str = strconv.FormatInt(int64(s.CoreM137117.N), 2)
	for len(str) < 5 {
		str = "0" + str
	}
	Sbinary = Sbinary + str

	str = strconv.FormatInt(int64(s.CoreM137117.Health), 2)
	for len(str) < 3 {
		str = "0" + str
	}
	Sbinary = Sbinary + str

	str = strconv.FormatInt(int64(s.CoreM137117.DgnssDataWords), 2)
	for len(str) < 696 {
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