// Package libnaf provides functions to maintain payload for MO MT messages.
package libnaf

import "fmt"
import "prisma/tms/iridium"

import "strings"

//ParseNaf messages to iridium structures
func ParseNaf(str string) (*iridium.Iridium, error) {

	var err error

	if strings.Contains(str, "//SR//TM/OVB//IMEI/") {
		str = strings.TrimLeft(str, "//SR//TM/OVB//IMEI/")
	} else {
		return nil, fmt.Errorf("Naf string does not contain proper header //SR//TM/OVB//IMEI/")
	}
	if strings.Contains(str, ER) {
		str = string(str[:len(str)-4])
	} else {
		return nil, fmt.Errorf("Naf string does not contain proper ending //ER")
	}

	fields := strings.Split(str, "//")

	if len(fields) < 3 {
		return nil, fmt.Errorf("invalid Naf structure")
	}

	if len(fields[0]) != 15 {
		return nil, fmt.Errorf("invalid emei %s", fields[0])
	}

	iri := &iridium.Iridium{}

	switch strings.Split(fields[1], "/")[1] {
	case "TMA":
		iri, err = parseTMA(fields)
	case "AA":
		iri, err = parseAA(fields)
	case "RMH":
		iri, err = parseRMH(fields)
	case "UIC":
		iri, err = parseUIC(fields)
	case "RSM":
		iri, err = parseRSM(fields)
	case "UGP":
		iri, err = parseUGP(fields)
	case "UAUP":
		iri, err = parseUAUP(fields)
	case "UG_Polygon":
		iri, err = parseUGFP(fields)
	case "UG_Circle":
		iri, err = parseUGFC(fields)
	case "GBMN":
		iri, err = parseGBMN(fields)
	case "DG":
		iri, err = parseDGFS(fields)
	case "BM":
		iri, err = parseBMSV(fields)
	default:
		return nil, fmt.Errorf("Type %s is not handled yet in the naf parser", strings.Split(fields[1], "/")[1])
	}

	if err != nil {
		return nil, err
	}

	return iri, nil
}
