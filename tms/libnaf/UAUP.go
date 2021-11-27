package libnaf

import "prisma/tms/omnicom"
import "prisma/tms/iridium"
import "strings"
import "fmt"

func parseUAUP(fields []string) (*iridium.Iridium, error) {

	var err error

	iri := &iridium.Iridium{}
	iri.Mth = &iridium.MobileTerminatedHeader{}
	iri.Mth.IMEI = fields[0]
	iri.Payload = &iridium.Payload{}
	iri.Payload.Omnicom = &omnicom.Omni{}
	iri.Payload.Omnicom.Omnicom = &omnicom.Omni_Uaup{}

	iri.Payload.Omnicom.GetUaup().Header = []byte{0x3A}

	if len(fields) >= 4 {
		iri.Payload.Omnicom.GetUaup().Date, err = parseDate(fields[2], fields[3], "DA", "TI")
		if err != nil {
			return nil, err
		}
	}
	for i := 4; i < len(fields); i++ {
		switch strings.Split(fields[i], "/")[0] {
		case "WSAUS":
			iri.Payload.Omnicom.GetUaup().Web_Service_API_URL_Sending = &omnicom.WebServiceAPIURLsendingStoV{}
			iri.Payload.Omnicom.GetUaup().Web_Service_API_URL_Sending.Value = []byte(strings.Split(fields[i], "/")[1])
		case "WSAUR":
			iri.Payload.Omnicom.GetUaup().Web_Service_API_URL_Receiving = &omnicom.WebServiceAPIURLReceivingStoV{}
			iri.Payload.Omnicom.GetUaup().Web_Service_API_URL_Receiving.Value = []byte(strings.Split(fields[i], "/")[1])
		case "AIO":
			array := strings.Split(strings.Split(fields[i], "/")[1], ",")
			iri.Payload.Omnicom.GetUaup().Array = &omnicom.ArrayStoV{}
			for _, value := range array {
				switch value {
				case "HPR":
					iri.Payload.Omnicom.GetUaup().Array.Value = append(iri.Payload.Omnicom.GetUaup().Array.Value, 0x01)
				case "SPR":
					iri.Payload.Omnicom.GetUaup().Array.Value = append(iri.Payload.Omnicom.GetUaup().Array.Value, 0x06)
				case "AR":
					iri.Payload.Omnicom.GetUaup().Array.Value = append(iri.Payload.Omnicom.GetUaup().Array.Value, 0x02)
				case "GP":
					iri.Payload.Omnicom.GetUaup().Array.Value = append(iri.Payload.Omnicom.GetUaup().Array.Value, 0x03)
				case "GA":
					iri.Payload.Omnicom.GetUaup().Array.Value = append(iri.Payload.Omnicom.GetUaup().Array.Value, 0x04)
				case "AUP":
					iri.Payload.Omnicom.GetUaup().Array.Value = append(iri.Payload.Omnicom.GetUaup().Array.Value, 0x08)
				case "RMH":
					iri.Payload.Omnicom.GetUaup().Array.Value = append(iri.Payload.Omnicom.GetUaup().Array.Value, 0x31)
				case "UG_Polygon":
					fallthrough
				case "UG_Circle":
					iri.Payload.Omnicom.GetUaup().Array.Value = append(iri.Payload.Omnicom.GetUaup().Array.Value, 0x35)
				case "DG":
					iri.Payload.Omnicom.GetUaup().Array.Value = append(iri.Payload.Omnicom.GetUaup().Array.Value, 0x37)
				case "RSM":
					iri.Payload.Omnicom.GetUaup().Array.Value = append(iri.Payload.Omnicom.GetUaup().Array.Value, 0x33)
				case "UIC":
					iri.Payload.Omnicom.GetUaup().Array.Value = append(iri.Payload.Omnicom.GetUaup().Array.Value, 0x32)
				case "AA":
					iri.Payload.Omnicom.GetUaup().Array.Value = append(iri.Payload.Omnicom.GetUaup().Array.Value, 0x45)
				case "GBMN":
					iri.Payload.Omnicom.GetUaup().Array.Value = append(iri.Payload.Omnicom.GetUaup().Array.Value, 0x36)
				case "UF":
					iri.Payload.Omnicom.GetUaup().Array.Value = append(iri.Payload.Omnicom.GetUaup().Array.Value, 0x39)
				case "UGP":
					iri.Payload.Omnicom.GetUaup().Array.Value = append(iri.Payload.Omnicom.GetUaup().Array.Value, 0x34)
				case "UAUP":
					iri.Payload.Omnicom.GetUaup().Array.Value = append(iri.Payload.Omnicom.GetUaup().Array.Value, 0x3A)
				case "BM_StoV":
					iri.Payload.Omnicom.GetUaup().Array.Value = append(iri.Payload.Omnicom.GetUaup().Array.Value, 0x07)
				case "TMA":
					iri.Payload.Omnicom.GetUaup().Array.Value = append(iri.Payload.Omnicom.GetUaup().Array.Value, 0x30)
				case "BM":
					iri.Payload.Omnicom.GetUaup().Array.Value = append(iri.Payload.Omnicom.GetUaup().Array.Value, 0x38)
				case "ABM":
					iri.Payload.Omnicom.GetUaup().Array.Value = append(iri.Payload.Omnicom.GetUaup().Array.Value, 0x09)
				default:
					return nil, fmt.Errorf("AIO tag without any value")
				}
			}

		}
	} // end of for loop

	return iri, nil
}

func encodeUAUP(UAUP *omnicom.Uaup) string {
	var str string

	str = str + "//TOM/UAUP/"

	str = encodeDate(UAUP.Date, str)
	if UAUP.Web_Service_API_URL_Sending.To_Modify == 1 {
		str = str + "//WSAUS/" + strings.Replace(string(UAUP.Web_Service_API_URL_Sending.Value), "/", "\\", -1)
	}

	if UAUP.Web_Service_API_URL_Receiving.To_Modify == 1 {
		str = str + "//WSAUR/" + strings.Replace(string(UAUP.Web_Service_API_URL_Receiving.Value), "/", "\\", -1)
	}

	if UAUP.Array.To_Modify == 1 {

		str = str + "//AIO/"
		for i, id := range UAUP.Array.Value {
			str = str + TOM[id]
			if i != (len(UAUP.Array.Value) - 1) {
				str = str + ","
			}
		}
	}
	return str + ER
}
