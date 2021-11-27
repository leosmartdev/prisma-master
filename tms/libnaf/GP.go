package libnaf

import (
	"prisma/tms/omnicom"
	"strconv"
	"strings"
)

func encodeGP(GP *omnicom.Gp) string {

	var str string

	str = str + "//TOM/GP"
	str = str + "//BID/" + strconv.FormatUint(uint64(GP.Beacon_ID), 10)
	GP.Date_Position.Latitude = float32(toFixed(GP.Date_Position.Latitude, 3))
	GP.Date_Position.Longitude = float32(toFixed(GP.Date_Position.Longitude, 3))
	str = encodeDatePosition(GP.Date_Position, str)
	str = str + "//RPI/" + strconv.FormatUint(uint64(GP.Position_Reporting_Interval.ValueInMn), 10)
	str = str + "//GFE/" + strconv.FormatUint(uint64(GP.Geofencing_Enable.On_Off), 10)
	str = str + "//PCI/" + strconv.FormatUint(uint64(GP.Position_Collection_Interval.ValueInMn), 10)
	str = str + "//PWD/" + strings.Split(string(GP.Password.ValueInMn), " ")[0]
	str = str + "//RTG/" + strconv.FormatUint(uint64(GP.Routing.ValueInMn), 10)

	if GP.Firmware_Dome_Version != nil {
		str = str + "//FDV/"
		for i, v := range GP.Firmware_Dome_Version {
			str = str + strconv.FormatUint(uint64(v), 10)
			if i != (len(GP.Firmware_Dome_Version) - 1) {
				str = str + "."
			}
		}
	}
	if GP.Junction_Box_Version != nil {
		str = str + "//JBV/"
		for i, v := range GP.Junction_Box_Version {
			str = str + strconv.FormatUint(uint64(v), 10)
			if i != (len(GP.Junction_Box_Version) - 1) {
				str = str + "."
			}
		}
	}
	if GP.SIM_Card_ICCID != nil {
		str = str + "//SCI/" + string(GP.SIM_Card_ICCID)
	}

	if GP.G3_IMEI != nil {
		str = str + "//3GIMEI/" + string(GP.G3_IMEI)
	}

	if GP.IRI_IMEI != nil {
		str = str + "//IMEI/" + string(GP.IRI_IMEI)
	}

	return str + ER
}
