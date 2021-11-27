package libnaf

import "prisma/tms/omnicom"
import "strconv"

func encodeHPR(HPR *omnicom.Hpr) string {

	var str string

	str = str + "//TOM/HPR"

	str = encodeID(HPR.Msg_ID, str)

	str = str + "//CDR/" + strconv.FormatUint(uint64(HPR.Count_Data_ReportsInThis_Msg), 10)

	for i := 0; i < int(HPR.Count_Data_ReportsInThis_Msg); i++ {
		str = encodeDataReport(HPR.Data_Report[i], str)
	}

	return str + ER
}
