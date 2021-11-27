package libnaf

import "strconv"

import "prisma/tms/omnicom"

func encodeGFA(GA *omnicom.Ga) string {

	var str string

	str = str + "//TOM/GA"

	str = encodeID(GA.Msg_ID, str)

	str = encodeDatePosition(GA.Date_Position, str)

	str = str + "//GFET/" + strconv.FormatUint(uint64(GA.Error_Type), 10)

	return str + ER

}
