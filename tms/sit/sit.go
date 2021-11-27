package sit

import (
	"errors"
	"fmt"
	"strings"
)

type Sit915 struct {
	Raw                 string
	MessageNumber       string
	ReportingFacility   string
	MessageTransmitTime string
	Sit                 string
	DestinationMCC      string
	NarrativeText       string
}

func Parse(text string) (Sit915, error) {
	var msg Sit915
	s := NewScanner(text)
	msg.Raw = text
	msg.MessageNumber = s.Next()
	msg.ReportingFacility = s.Next()
	msg.MessageTransmitTime = s.Next()
	msg.Sit = s.Next()

	if msg.Sit != "915" {
		return msg, fmt.Errorf("expecting 915 message but got '%v'", msg.Sit)
	}

	msg.DestinationMCC = s.Next()
	msg.NarrativeText = s.NarrativeText()
	return msg, s.Err
}

type Scanner struct {
	data string
	tok  strings.Builder
	ch   byte
	pos  int
	EOF  bool
	Err  error
}

var eof = errors.New("eof")

func NewScanner(data string) *Scanner {
	s := &Scanner{data: data}
	s.scan()      // init ch field
	s.scanField() // discard up to first slash
	return s
}

func (s *Scanner) Next() string {
	if s.EOF {
		return ""
	}
	s.tok.Reset()
	s.scanField()
	tok := strings.TrimSpace(s.tok.String())
	return tok
}

func (s *Scanner) NarrativeText() string {
	s.tok.Reset()
	// Pos marks the next position to read from, so when
	// parsing a narrative block, start at the previous
	pos := s.pos - 1
	idx := strings.Index(s.data[pos:], "QQQQ")
	if idx < 0 {
		s.pos = len(s.data)
		s.Err = errors.New("end marker for narrative text not found")
		return ""
	}
	s.tok.WriteString(s.data[pos : pos+idx])
	tok := strings.TrimSpace(s.tok.String())

	// Advance to the next field
	s.pos = s.pos + idx + 4 // 4 = len of QQQQ
	s.scan()                // re-init ch field
	s.scanField()           // discard up to next slash

	return tok
}

func (s *Scanner) scanField() {
	for s.ch != '/' && !s.EOF {
		s.tok.WriteByte(s.ch)
		s.scan()
	}
	s.scan()
}

func (s *Scanner) scan() {
	if s.pos >= len(s.data) {
		s.EOF = true
		return
	}
	s.ch = s.data[s.pos]
	s.pos++
}
