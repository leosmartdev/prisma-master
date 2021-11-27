package libnaf

import "prisma/tms/omnicom"
import "strconv"

func encodeBM(BM *omnicom.Bm) string {

	var str string

	str = str + "//TOM/BM"

	str = encodeDate(BM.Date, str)

	str = encodeID(BM.ID_Msg, str)

	str = str + "//BMS/" + strconv.FormatUint(uint64(BM.Length_Msg_Content), 10)

	return str + ER
}
