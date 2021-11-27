package libnaf

import "prisma/tms/omnicom"
import "strings"

func encodeAUP(AUP *omnicom.Aup) string {

	var str string

	str = str + "//TOM/AUP"

	str = encodeID(AUP.ID_Msg, str)

	str = encodeDatePosition(AUP.Date_Position, str)

	if len(AUP.Web_Service_API_URL_Sending.Value) != 0 {
		r := strings.NewReplacer("/", "\\")
		str = str + "//WSAUS/" + r.Replace(strings.Split(string(AUP.Web_Service_API_URL_Sending.Value), " ")[0])
	}

	if len(AUP.Web_Service_API_URL_Receiving.Value) != 0 {
		r := strings.NewReplacer("/", "\\")
		str = str + "//WSAUR/" + r.Replace(strings.Split(string(AUP.Web_Service_API_URL_Receiving.Value), " ")[0])
	}

	if len(AUP.Array.Value) != 0 {
		str = str + "//AIO/"
		for _, value := range AUP.Array.Value {
			if value == 0x01 {
				str = str + "HPR,"
			}
			if value == 0x02 {
				str = str + "AR,"
			}
			if value == 0x06 {
				str = str + "SPR,"
			}
			if value == 0x03 {
				str = str + "GP,"
			}
			if value == 0x08 {
				str = str + "AUP,"
			}
			if value == 0x04 {
				str = str + "GA,"
			}
			if value == 0x07 {
				str = str + "BM,"
			}
			if value == 0x09 {
				str = str + "ABM,"
			}
		}
		str = strings.TrimRight(str, ",")
	}

	return str + ER
}
