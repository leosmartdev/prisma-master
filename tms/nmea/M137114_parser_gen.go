// Code generated by parse_nmea; DO NOT EDIT
package nmea

import "fmt"
import "strconv"
import "strings"

// M137114 represents fix data.
type CoreM137114 struct {
	MessageID uint8

	RepeatIndicator uint32

	Mmsi uint32

	Spare uint32 //Supposed to be Unknown

	Text string
}
type M137114 struct {
	VDMO
	CoreM137114
}

func NewM137114(sentence VDMO) *M137114 {
	s := new(M137114)
	s.VDMO = sentence
	return s
}

func (s *M137114) parse() error {
	var err error

	if MessageType(s.EncapData) != 14 {
		err = fmt.Errorf("message %d is not a M137114", MessageType(s.EncapData))
		return err
	}

	data := []byte(s.EncapData)

	//if len(data)*6 > 1008 {
	//	err = fmt.Errorf("Message lenght is larger than it should be [%d!=1008]", len(data)*6)
	//	return err
	//}

	s.MessageID = MessageType(s.EncapData)

	s.CoreM137114.RepeatIndicator = BitsToInt(6, 7, data)

	s.CoreM137114.Mmsi = BitsToInt(8, 37, data)

	s.CoreM137114.Spare = BitsToInt(38, 39, data)

	s.CoreM137114.Text = BitsToString(40, 6*(len(data)-1), data)

	return nil
}

func (s *M137114) Encode() (string, error) {
	var Raw string
	var Sbinary string

	if s.MessageID != 14 {
		err := fmt.Errorf("message %d is not a M137114", s.MessageID)
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

	str := strconv.FormatInt(int64(s.CoreM137114.MessageID), 2)
	for len(str) < 6 {
		str = "0" + str
	}
	Sbinary = Sbinary + str

	str = strconv.FormatInt(int64(s.CoreM137114.RepeatIndicator), 2)
	for len(str) < 2 {
		str = "0" + str
	}
	Sbinary = Sbinary + str

	str = strconv.FormatInt(int64(s.CoreM137114.Mmsi), 2)
	for len(str) < 30 {
		str = "0" + str
	}
	Sbinary = Sbinary + str

	str = strconv.FormatInt(int64(s.CoreM137114.Spare), 2)
	for len(str) < 2 {
		str = "0" + str
	}
	Sbinary = Sbinary + str

	if s.CoreM137114.Text != "" {
		str = ""

		for len([]byte(s.CoreM137114.Text)) < 968/6 {
			s.CoreM137114.Text = s.CoreM137114.Text + "@"

		}

		for _, n := range []byte(s.CoreM137114.Text) {
			if n >= 32 {
				n = n - 64
			}
			name := strconv.FormatInt(int64(n), 2)
			for len(name) < 6 {
				name = "0" + name
			}

			if len(name) > 6 {
				if len(name) == 8 {
					n = n - 128
					name = strconv.FormatInt(int64(n), 2)
				}
				if len(name) == 7 {
					n = n - 64
					name = strconv.FormatInt(int64(n), 2)
				}
			}

			str = str + name

		}

		Sbinary = Sbinary + str

	}

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