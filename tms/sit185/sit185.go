// Package sit185 provides functions to work with sit185 format.
package sit185

import (
	"fmt"
	"regexp"
)

//Sit185 strcut is the output after parsing a sit185 message using RegularExpPattern stuct
type Sit185 struct {
	Raw    string
	Fields map[string]string
}

//RegularExpPattern strct contains fieldname and regular exp that will be used to extract field value
type RegularExpPattern struct {
	Fieldname string
	Regex     *regexp.Regexp
}

//Parse function takes the raw sit185 messages and RegularExpPattern to extract rcc relevant data from the message.
func Parse(msg string, reps []RegularExpPattern) (Sit185, error) {

	var sit Sit185

	sit.Raw = msg
	sit.Fields = make(map[string]string)

	for _, rep := range reps {
		matches := rep.Regex.FindStringSubmatch(msg)

		for key, name := range rep.Regex.SubexpNames() {
			if key > 0 && key <= len(matches) {
				sit.Fields[name] = matches[key]
			}
		}
	}

	if len(sit.Fields) == 0 {
		return sit, fmt.Errorf("MCC message format is Uknown")
	}

	return sit, nil

}
