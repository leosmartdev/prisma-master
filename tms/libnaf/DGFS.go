package libnaf

import (
	"fmt"
	"prisma/tms/iridium"
	"prisma/tms/omnicom"
	"strings"
)

func parseDGFS(fields []string) (*iridium.Iridium, error) {

	var err error

	iri := &iridium.Iridium{}
	iri.Mth = &iridium.MobileTerminatedHeader{}
	iri.Mth.IMEI = fields[0]
	iri.Payload = &iridium.Payload{}
	iri.Payload.Omnicom = &omnicom.Omni{}
	iri.Payload.Omnicom.Omnicom = &omnicom.Omni_Dg{}

	iri.Payload.Omnicom.GetDg().Header = []byte{0x37}

	for i := 2; i < len(fields); i++ {
		switch strings.Split(fields[i], "/")[0] {
		case "ID":
			iri.Payload.Omnicom.GetDg().Msg_ID, err = parseID(fields[i])
		case "DA":
			iri.Payload.Omnicom.GetDg().Date, err = parseDate(fields[i], fields[i+1], "DA", "TI")
			i++
		case "GEOID":
			iri.Payload.Omnicom.GetDg().GEO_ID, err = parseID(fields[i])
		default:
			return nil, fmt.Errorf("uknown field in DGFS naf %s", strings.Split(fields[i], "/")[0])

		}
		if err != nil {
			return nil, err
		}
	}

	return iri, nil
}

func encodeDGFS(DGFS *omnicom.Dg) string {

	var str string

	str = str + "//TOM/DG"

	str = encodeDate(DGFS.Date, str)

	str = encodeID(DGFS.Msg_ID, str)

	str = encodeGEOID(DGFS.GEO_ID, str)

	return str + ER
}
