package libnaf

import (
	"prisma/tms/iridium"
	"prisma/tms/omnicom"
)

func parseAA(fields []string) (*iridium.Iridium, error) {

	iri := &iridium.Iridium{}
	iri.Mth = &iridium.MobileTerminatedHeader{}
	iri.Mth.IMEI = fields[0]
	iri.Payload = &iridium.Payload{}
	iri.Payload.Omnicom = &omnicom.Omni{}
	iri.Payload.Omnicom.Omnicom = &omnicom.Omni_Aa{}

	iri.Payload.Omnicom.GetAa().Header = []byte{0x45}

	date, err := parseDate(fields[2], fields[3], "DA", "TI")
	if err != nil {
		return nil, err
	}
	iri.Payload.Omnicom.GetAa().Date = date

	return iri, nil
}

func encodeAA(AA *omnicom.Aa) string {
	var str string

	str = str + "//TOM/AA"

	str = encodeDate(AA.Date, str)

	return str + ER
}
