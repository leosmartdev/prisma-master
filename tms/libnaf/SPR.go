package libnaf

import (
	"prisma/tms/omnicom"
	"strconv"
)

func encodeSPR(SPR *omnicom.Spr) string {

	var str string

	str = str + "//TOM/SPR"

	str = encodeDatePosition(SPR.Date_Position, str)

	str = encodeMove(SPR.Move, str)

	//TODO: having VIN field in SPR is wrong it will need to be taken down,
	//      the customer needs it because of a bug in its
	//      decoder that request VIN to be parse SPR data
	str = str + "//VIN/0"

	str = str + "//RPI/"

	str = str + strconv.FormatUint(uint64(SPR.Period), 10)

	return str + ER

}
