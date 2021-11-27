package libnaf

import (
	"fmt"
	"prisma/tms/iridium"
	"prisma/tms/omnicom"
	"strconv"
	"strings"
)

func encodeBMSV(BMSV *omnicom.BMStoV) string {

	var str string

	str = str + "//TOM/BM_Stov"

	str = encodeDate(BMSV.Date, str)

	str = encodeID(BMSV.ID_Msg, str)

	str = encodeBMS(BMSV.Length_Msg_Content, str)

	return str + ER
}

func parseBMSV(fields []string) (*iridium.Iridium, error) {

	var err error

	iri := &iridium.Iridium{}
	iri.Mth = &iridium.MobileTerminatedHeader{}
	iri.Mth.IMEI = fields[0]
	iri.Payload = &iridium.Payload{}
	iri.Payload.Omnicom = &omnicom.Omni{}
	iri.Payload.Omnicom.Omnicom = &omnicom.Omni_Bmstov{}

	iri.Payload.Omnicom.GetBmstov().Header = []byte{0x38}

	for i := 2; i < len(fields); i++ {
		switch strings.Split(fields[i], "/")[0] {
		case "DA":
			iri.Payload.Omnicom.GetBmstov().Date, err = parseDate(fields[i], fields[i+1], "DA", "TI")
			i++
		case "ID":
			iri.Payload.Omnicom.GetBmstov().ID_Msg, err = parseID(fields[i])
		case "BMS":
			var num uint64
			num, err = strconv.ParseUint(strings.Split(fields[i], "/")[1], 10, 32)
			iri.Payload.Omnicom.GetBmstov().Length_Msg_Content = uint32(num)
		default:
			return nil, fmt.Errorf("uknown field in BMS naf %s", strings.Split(fields[i], "/")[0])
		}
		if err != nil {
			return nil, err
		}
	}

	return iri, nil
}
