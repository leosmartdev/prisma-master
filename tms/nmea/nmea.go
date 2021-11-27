// Package nmea contains parsers and structures to maintain nmea messages.
package nmea

import (
	"fmt"
	"strings"
)

const (
	// chars that indicates the start of a sentence.
	sentenceStart1 = "$"
	sentenceStart2 = "!"
	//  token to delimit fields of a sentence.
	fieldSep = ","
	// The token to delimit the checksum of a sentence.
	checksumSep = "*"
	// char that indicates that the sentences comes from an AIS device
)

var PrefixAIS = []string{"VDO", "VDM"}

var aisTalkers = []string{"AB", "AD", "AI", "AN", "AR", "AS", "AT", "AX", "BS", "SA"}

//SentenceI interface
type Sentence interface {
	GetSentence() BaseSentence
}

// Sentence contains general information about an  NMEA sentence
type BaseSentence struct {
	SOS      string   // the sentence start $
	Talker   string   // the sentence talker (e.g GP)
	Format   string   // The sentence format (e.g GLL)
	Fields   []string // Array of fields
	Checksum string   // Checksum
	Raw      string   // The raw NMEA sentence received
}

// GetSentence getter
func (s *BaseSentence) GetSentence() BaseSentence {
	return *s
}

func (s *BaseSentence) parse(input string) error {
	s.Raw = input

	if strings.Count(s.Raw, checksumSep) != 1 {
		return fmt.Errorf("Sentence does not contain single checksum separator")
	}

	if !strings.Contains(s.Raw, sentenceStart1) && !strings.Contains(s.Raw, sentenceStart2) {
		return fmt.Errorf("Sentence does not contain a '$' or '!'")
	}

	flag := 0
	var sentence string
	var aisSentence string
	var NmeaSentence string
	var fieldSum []string
	var fields []string

	token := strings.Split(s.Raw, ",")[0]
loop:
	for _, aisTalker := range aisTalkers {
		if strings.Contains(token, aisTalker+"VD") {
			sentence = strings.Split(s.Raw, sentenceStart2)[1]
			aisSentence = strings.SplitN(sentence, aisTalker, 2)[1]
			fieldSum = strings.Split(aisSentence, checksumSep)
			fields = strings.Split(fieldSum[0], fieldSep)

			s.SOS = sentenceStart2
			s.Talker = aisTalker
			s.Format = fields[0]
			s.Fields = fields[1:]
			s.Checksum = strings.ToUpper(fieldSum[1])

			flag = 1
			break loop

		}
	}

	if flag == 0 {
		// remove the $ or ! character
		if strings.Contains(s.Raw, sentenceStart1) {
			sentence = strings.Split(s.Raw, sentenceStart1)[1]
		}
		if strings.Contains(s.Raw, sentenceStart2) {
			sentence = strings.Split(s.Raw, sentenceStart2)[1]
		}
		// remove the talker characters
		NmeaSentence = strings.SplitN(sentence, "", 3)[2]
		fieldSum = strings.Split(NmeaSentence, checksumSep)
		fields = strings.Split(fieldSum[0], fieldSep)
		s.SOS = sentenceStart1
		s.Talker = strings.SplitN(sentence, "", 3)[0] + strings.SplitN(sentence, "", 3)[1]
		s.Format = fields[0]
		s.Fields = fields[1:]
		s.Checksum = strings.ToUpper(fieldSum[1])

	}

	if err := s.sumOk(); err != nil {
		return fmt.Errorf("Sentence checksum mismatch %s", err)
	}

	return nil
}

// SumOk returns whether the calculated checksum matches the message checksum.
func (s *BaseSentence) sumOk() error {
	var checksum uint8
	for i := 1; i < len(s.Raw) && string(s.Raw[i]) != checksumSep; i++ {
		checksum ^= s.Raw[i]
	}

	calculated := fmt.Sprintf("%X", checksum)
	if len(calculated) == 1 {
		calculated = "0" + calculated
	}
	if calculated != s.Checksum {
		return fmt.Errorf("[%s != %s]", calculated, s.Checksum)
	}
	return nil
}

func Checksum(raw string) string {
	var checksum uint8

	for i := 1; i < len(raw); i++ {
		checksum ^= raw[i]
	}

	calculated := fmt.Sprintf("%X", checksum)

	if len(calculated) == 1 {
		calculated = "0" + calculated
	}
	check := "*" + calculated

	return check
}
