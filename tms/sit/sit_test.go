package sit

import (
	"reflect"
	"testing"
)

func TestSplitFunc(t *testing.T) {
	tests := []struct {
		name   string
		msg    string
		fields []string
	}{
		{
			"header",
			`/00030 00015/3660/80 160 1550`,
			[]string{"00030 00015", "3660", "80 160 1550"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := NewScanner(test.msg)
			have := make([]string, 3)
			have[0] = s.Next()
			have[1] = s.Next()
			have[2] = s.Next()

			if s.Err != nil {
				t.Fatalf("unexpected error: %v", s.Err)
			}
			if !reflect.DeepEqual(have, test.fields) {
				t.Errorf("\n have: %v \n want: %v", have, test.fields)
			}
		})
	}
}

func TestNarrative(t *testing.T) {
	msg := `
/xxxxx
/This is a test.
This is only a test.
QQQQ
/yyyyy
`
	want := []string{
		"xxxxx",
		"This is a test.\nThis is only a test.",
		"yyyyy",
	}
	have := make([]string, 3)
	s := NewScanner(msg)
	have[0] = s.Next()
	have[1] = s.NarrativeText()
	have[2] = s.Next()

	if s.Err != nil {
		t.Fatalf("unexpected error: %v", s.Err)
	}
	if !reflect.DeepEqual(have, want) {
		t.Errorf("\n have: %v \n want: %v", have, want)
	}
}

func TestNarrativeNoEndMarker(t *testing.T) {
	msg := `
/This is a test.
This is only a test.
/yyyyy
`
	s := NewScanner(msg)
	_ = s.NarrativeText()
	if s.Err == nil {
		t.Errorf("expected error")
	}
}

func TestSit185(t *testing.T) {
	msg, err := Parse(sampleCSA002)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	have := []string{
		msg.MessageNumber,
		msg.ReportingFacility,
		msg.MessageTransmitTime,
		msg.Sit,
		msg.DestinationMCC,
		msg.NarrativeText,
	}
	want := []string{
		"00030 00015",
		"3660",
		"80 160 1550",
		"915",
		"3160",
		`THE NARRATIVE TEXT IN PRINTABLE CHARACTERS
IS PLACED HERE, WITH NO MORE THAN 69
CHARACTERS PER LINE.`,
	}
	if !reflect.DeepEqual(have, want) {
		t.Errorf("\n have:\n%v \n want:\n%v", have, want)
	}
}

func TestUnknownSit(t *testing.T) {
	_, err := Parse(sampleUnknown)
	if err == nil {
		t.Fatalf("expected error")
	}
	have := err.Error()
	want := "expecting 915 message but got '999'"
	if have != want {
		t.Errorf("\n have: %v \n want: %v", have, want)
	}
}
