package libnaf

import "strconv"
import "prisma/tms/omnicom"

func encodeABM(ABM *omnicom.Abm) string {

	var str string

	str = str + "//TOM/ABM"

	str = encodeDate(ABM.Date, str)

	str = encodeID(ABM.ID_Msg, str)

	str = str + "//ABMET/" + strconv.FormatUint(uint64(ABM.Error_Type), 10)

	return str + ER
}
