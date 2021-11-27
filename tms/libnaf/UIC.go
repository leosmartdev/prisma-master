package libnaf

import "prisma/tms/omnicom"
import "prisma/tms/iridium"
import "strconv"
import "strings"

func parseUIC(fields []string) (*iridium.Iridium, error) {
	var err error

	iri := &iridium.Iridium{}
	iri.Mth = &iridium.MobileTerminatedHeader{}
	iri.Mth.IMEI = fields[0]
	iri.Payload = &iridium.Payload{}
	iri.Payload.Omnicom = &omnicom.Omni{}
	iri.Payload.Omnicom.Omnicom = &omnicom.Omni_Uic{}

	iri.Payload.Omnicom.GetUic().Header = []byte{0x32}

	if len(fields) >= 3 {

		iri.Payload.Omnicom.GetUic().ID_Msg, err = parseID(fields[2])
		if err != nil {
			return nil, err
		}

	}

	if len(fields) >= 5 {
		iri.Payload.Omnicom.GetUic().Date, err = parseDate(fields[3], fields[4], "DA", "TI")
		if err != nil {
			return nil, err
		}

	}

	if len(fields) >= 6 {
		rp, err := strconv.ParseUint(strings.Split(fields[5], "/")[1], 10, 32)
		if err != nil {
			return nil, err
		}
		iri.Payload.Omnicom.GetUic().New_Reporting = uint32(rp)

	}

	return iri, nil
}

func encodeUIC(UIC *omnicom.Uic) string {
	var str string

	str = str + "//UIC/"

	str = encodeDate(UIC.Date, str)

	str = str + "//NRI/" + strconv.Itoa(int(UIC.New_Reporting))

	return str + ER
}
