package libnaf

import (
	"prisma/tms/iridium"
	"prisma/tms/omnicom"
)

func parseTMA(fields []string) (*iridium.Iridium, error) {

	iri := &iridium.Iridium{}
	iri.Mth = &iridium.MobileTerminatedHeader{}
	iri.Mth.IMEI = fields[0]
	iri.Payload = &iridium.Payload{}
	iri.Payload.Omnicom = &omnicom.Omni{}
	iri.Payload.Omnicom.Omnicom = &omnicom.Omni_Tma{}

	iri.Payload.Omnicom.GetTma().Header = []byte{0x30}

	date, err := parseDate(fields[2], fields[3], "DA", "TI")
	if err != nil {
		return nil, err
	}
	iri.Payload.Omnicom.GetTma().Date = date
	return iri, nil
}

func encodeTMA(TMA *omnicom.Tma) string {
	var str string

	str = str + "//TOM/TMA"

	str = encodeDate(TMA.Date, str)

	return str + ER
}
