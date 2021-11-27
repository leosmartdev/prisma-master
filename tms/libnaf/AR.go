package libnaf

import "prisma/tms/omnicom"
import "strconv"

func encodeAR(AR *omnicom.Ar) string {

	var str string

	str = str + "//TOM/AR"

	str = encodeID(AR.Msg_ID, str)

	AR.Date_Position.Latitude = float32(toFixed(AR.Date_Position.Latitude, 4))
	AR.Date_Position.Longitude = float32(toFixed(AR.Date_Position.Longitude, 4))

	str = encodeDatePosition(AR.Date_Position, str)

	if AR.Extention_Bit_Move == 1 {
		str = encodeMove(AR.Move, str)
	}

	str = encodeDateEvent(AR.Date_Event, str)

	str = str + "//PU/"

	if AR.Power_Up.Power_Up_Status == 1 {
		str = str + "PU"
	} else if AR.Power_Up.Power_Up_Status == 0 {
		str = str + "PD"
	}

	str = str + "//BSR/" + strconv.FormatUint(uint64(AR.Battery_Alert.Current_Battery_Alert_Status), 10)

	str = str + "//IAS/" + strconv.FormatUint(uint64(AR.Intrusion_Alert.Current_Intrusion_Status), 10)

	str = str + "//FPR/" + strconv.FormatUint(uint64(AR.No_Position_Fix.Current_No_Position_Fix_Status), 10)

	str = str + "//NSV/" + strconv.FormatUint(uint64(AR.No_Position_Fix.Number_SatelliteIn_View), 10)

	str = str + "//JBDS/" + strconv.FormatUint(uint64(AR.JB_Dome_Alert.Current_JB_Dome_Status), 10)

	str = str + "//MCL/" + strconv.FormatUint(uint64(AR.Loss_Mobile_Com.Current_Loss_Mobile_Com_Status), 10)

	str = str + "//DLA/" + strconv.FormatUint(uint64(AR.Daylight_Alert.Current_Daylight_Alert), 10)

	str = str + "//AAS/" + strconv.FormatUint(uint64(AR.Assistance_Alert.Current_Assistance_Alert_Status), 10)

	str = str + "//TMA/" + strconv.FormatUint(uint64(AR.Test_Mode.Current_Test_Mode_Status), 10)

	return str + ER
}
