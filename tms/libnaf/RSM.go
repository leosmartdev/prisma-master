package libnaf

import "prisma/tms/omnicom"
import "prisma/tms/iridium"
import "strconv"
import "strings"

func parseRSM(fields []string) (*iridium.Iridium, error) {
	var err error

	iri := &iridium.Iridium{}
	iri.Mth = &iridium.MobileTerminatedHeader{}
	iri.Mth.IMEI = fields[0]
	iri.Payload = &iridium.Payload{}
	iri.Payload.Omnicom = &omnicom.Omni{}
	iri.Payload.Omnicom.Omnicom = &omnicom.Omni_Rsm{}

	iri.Payload.Omnicom.GetRsm().Header = []byte{0x33}

	if len(fields) >= 3 {

		iri.Payload.Omnicom.GetRsm().ID_Msg, err = parseID(fields[2])
		if err != nil {
			return nil, err
		}
	}

	if len(fields) >= 5 {

		iri.Payload.Omnicom.GetRsm().Date, err = parseDate(fields[3], fields[4], "DA", "TI")
		if err != nil {
			return nil, err
		}
	}

	if len(fields) >= 6 {
		msg, err := strconv.ParseUint(strings.Split(fields[5], "/")[1], 16, 32)
		if err != nil {
			return nil, err
		}
		iri.Payload.Omnicom.GetRsm().MsgTo_Ask = uint32(msg)
	}

	return iri, nil
}

func encodeRSM(RSM *omnicom.Rsm) string {
	var str string

	str = str + "//TOM/RSM"

	str = encodeID(RSM.ID_Msg, str)

	str = encodeDate(RSM.Date, str)

	str = str + "//MTA/" + strconv.FormatUint(uint64(RSM.MsgTo_Ask), 10)

	return str + ER
}
