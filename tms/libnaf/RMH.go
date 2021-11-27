package libnaf

import "prisma/tms/omnicom"
import "prisma/tms/iridium"
import "strings"

func parseRMH(fields []string) (*iridium.Iridium, error) {
	var err error

	iri := &iridium.Iridium{}
	iri.Mth = &iridium.MobileTerminatedHeader{}
	iri.Mth.IMEI = fields[0]
	iri.Payload = &iridium.Payload{}
	iri.Payload.Omnicom = &omnicom.Omni{}
	iri.Payload.Omnicom.Omnicom = &omnicom.Omni_Rmh{}
	iri.Payload.Omnicom.GetRmh().Date_Interval = &omnicom.DateInterval{}

	iri.Payload.Omnicom.GetRmh().Header = []byte{0x31}

	if len(fields) >= 4 {
		iri.Payload.Omnicom.GetRmh().Date, err = parseDate(fields[2], fields[3], "DA", "TI")
		if err != nil {
			return nil, err
		}

	}

	if len(fields) >= 6 {
		iri.Payload.Omnicom.GetRmh().Date_Interval.Start, err = parseDate(fields[4], fields[5], "DIS", "TIS")
		if err != nil {
			return nil, err
		}

	}

	if len(fields) >= 8 {
		iri.Payload.Omnicom.GetRmh().Date_Interval.Stop, err = parseDate(fields[6], fields[7], "DIE", "TIE")
		if err != nil {
			return nil, err
		}
	}

	if len(fields) >= 9 {

		iri.Payload.Omnicom.GetRmh().ID_Msg, err = parseID(fields[8])
		if err != nil {
			return nil, err
		}

	}

	return iri, nil
}

func encodeRMH(RMH *omnicom.Rmh) string {
	var str string

	str = str + "//TOM/RMH"

	str = encodeDate(RMH.Date, str)

	str = str + strings.Replace(strings.Replace(encodeDate(RMH.Date_Interval.Start, ""), "DA", "DIS", 1), "TI", "TIS", 1)

	str = str + strings.Replace(strings.Replace(encodeDate(RMH.Date_Interval.Stop, ""), "DA", "DIE", 1), "TI", "TIE", 1)

	str = encodeID(RMH.ID_Msg, str)

	return str + ER
}
