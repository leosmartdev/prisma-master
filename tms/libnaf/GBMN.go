package libnaf

import "prisma/tms/omnicom"
import "prisma/tms/iridium"
import "fmt"
import "strings"

func parseGBMN(fields []string) (*iridium.Iridium, error) {

	var err error

	iri := &iridium.Iridium{}
	iri.Mth = &iridium.MobileTerminatedHeader{}
	iri.Mth.IMEI = fields[0]
	iri.Payload = &iridium.Payload{}
	iri.Payload.Omnicom = &omnicom.Omni{}
	iri.Payload.Omnicom.Omnicom = &omnicom.Omni_Gbmn{}

	iri.Payload.Omnicom.GetGbmn().Header = []byte{0x36}

	for i := 2; i < len(fields); i++ {
		if strings.Split(fields[i], "/")[0] == "DA" {
			iri.Payload.Omnicom.GetGbmn().Date, err = parseDate(fields[i], fields[i+1], "DA", "TI")
			i++
		} else {
			return nil, fmt.Errorf("uknown field in GBMN naf %s", strings.Split(fields[i], "/")[0])
		}
		if err != nil {
			return nil, err
		}
	}

	return iri, nil
}

func encodeGBMN(GBMN *omnicom.Gbmn) string {
	var str string

	str = str + "//TOM/GBMN"

	str = encodeDate(GBMN.Date, str)

	return str + ER
}
