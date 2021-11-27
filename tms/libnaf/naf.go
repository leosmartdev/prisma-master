package libnaf

import (
	"fmt"
	"math"
	"prisma/tms/omnicom"
	"strconv"
	"strings"
)

const (

	// NafHeader should be global in a near future
	NafHeader = "//SR//TM/OVB"
	// ER is to mark the end of a naf sentence
	ER = "//ER"
)

//TOM is a map from messages hexa header to Type of message NAF notation
var TOM = map[byte]string{
	0x01: "HPR",
	0x02: "AR",
	0x03: "GP",
	0x04: "GA",
	0x06: "SPR",
	0x08: "AUP",
	0x31: "RMH",
	0x33: "RSM",
	0x35: "UG_Polygon,UG_Circle",
	0x37: "DG",
	0x32: "UIC",
	0x45: "AA",
	0x36: "GBMN",
	0x39: "UF",
	0x34: "UGP",
	0x3A: "UAUP",
	0x07: "BM_StoV",
	0x30: "TMA",
	0x38: "BM",
	0x09: "ABM",
}

func encodeID(id uint32, str string) string {

	if id != 0 {
		str = str + "//ID/"
		str = str + strconv.FormatUint(uint64(id), 10)

		return str
	}
	return str
}

func encodeGEOID(id uint32, str string) string {

	if id != 0 {
		str = str + "//GEOID/"
		str = str + strconv.FormatUint(uint64(id), 10)

		return str
	}
	return str
}

func encodeBMS(size uint32, str string) string {

	if size != 0 {
		str = str + "//BMS/"
		str = str + strconv.FormatUint(uint64(size), 10)

		return str
	}
	return str
}

func parseID(str string) (uint32, error) {

	ID, err := strconv.ParseUint(strings.Split(str, "/")[1], 10, 32)

	return uint32(ID), err
}

func encodeDataReport(DR *omnicom.DataReport, str string) string {

	str = encodeDatePosition(DR.Date_Position, str)

	str = encodeMove(DR.Move, str)

	str = str + "//VIN/" + strconv.FormatFloat(DR.Voltage.V_IN, 'f', -1, 64)

	str = str + "//RPI/" + strconv.FormatUint(uint64(DR.Period), 10)

	str = str + "//SAF/" + strconv.FormatUint(uint64(DR.Geofence.Status_Alert), 10)

	str = str + "//GEOID/" + strconv.FormatUint(uint64(DR.Geofence.GEO_ID), 10)

	return str

}

func encodeDatePosition(DatePos *omnicom.DatePosition, str string) string {

	str = str + "//DAP/"

	str = str + strconv.FormatUint(uint64(DatePos.Year), 10)

	if DatePos.Month < 10 {
		str = str + "0" + strconv.FormatUint(uint64(DatePos.Month), 10)
	} else {
		str = str + strconv.FormatUint(uint64(DatePos.Month), 10)
	}

	if DatePos.Day < 10 {
		str = str + "0" + strconv.FormatUint(uint64(DatePos.Day), 10)
	} else {
		str = str + strconv.FormatUint(uint64(DatePos.Day), 10)
	}

	str = str + "//TIP/"

	hours := uint64(DatePos.Minute / 60)

	minute := uint64(DatePos.Minute) - (hours * 60)

	if hours < 10 {
		str = str + "0" + strconv.FormatUint(hours, 10)
	} else {
		str = str + strconv.FormatUint(hours, 10)
	}

	if minute < 10 {
		str = str + "0" + strconv.FormatUint(minute, 10)
	} else {
		str = str + strconv.FormatUint(minute, 10)
	}

	str = str + "//XLT/"

	str = str + strconv.FormatFloat(float64(DatePos.Latitude), 'f', -1, 32)

	str = str + "//XLG/"

	str = str + strconv.FormatFloat(float64(DatePos.Longitude), 'f', -1, 32)

	return str
}

func encodeDateEvent(date *omnicom.DateEvent, str string) string {

	str = str + "//DAE/"

	str = str + strconv.FormatUint(uint64(date.Year), 10)

	if date.Month < 10 {
		str = str + "0" + strconv.FormatUint(uint64(date.Month), 10)
	} else {
		str = str + strconv.FormatUint(uint64(date.Month), 10)
	}

	if date.Day < 10 {
		str = str + "0" + strconv.FormatUint(uint64(date.Day), 10)
	} else {
		str = str + strconv.FormatUint(uint64(date.Day), 10)
	}

	str = str + "//TIE/"

	hours := uint64(date.Minute / 60)

	minute := uint64(date.Minute) - (hours * 60)

	if hours < 10 {
		str = str + "0" + strconv.FormatUint(hours, 10)
	} else {
		str = str + strconv.FormatUint(hours, 10)
	}

	if minute < 10 {
		str = str + "0" + strconv.FormatUint(minute, 10)
	} else {
		str = str + strconv.FormatUint(minute, 10)
	}

	return str

}

func encodeDate(date *omnicom.Dt, str string) string {

	str = str + "//DA/"

	str = str + strconv.FormatUint(uint64(date.Year), 10)

	if date.Month < 10 {
		str = str + "0" + strconv.FormatUint(uint64(date.Month), 10)
	} else {
		str = str + strconv.FormatUint(uint64(date.Month), 10)
	}

	if date.Day < 10 {
		str = str + "0" + strconv.FormatUint(uint64(date.Day), 10)
	} else {
		str = str + strconv.FormatUint(uint64(date.Day), 10)
	}

	str = str + "//TI/"

	hours := uint64(date.Minute / 60)

	minute := uint64(date.Minute) - (hours * 60)

	if hours < 10 {
		str = str + "0" + strconv.FormatUint(hours, 10)
	} else {
		str = str + strconv.FormatUint(hours, 10)
	}

	if minute < 10 {
		str = str + "0" + strconv.FormatUint(minute, 10)
	} else {
		str = str + strconv.FormatUint(minute, 10)
	}

	return str

}

func encodeMove(Mv *omnicom.MV, str string) string {

	str = str + "//SP/"

	str = str + strconv.FormatUint(uint64(Mv.Speed), 10)

	if Mv.Heading != 0 {
		str = str + "//CO/"

		str = str + strconv.FormatUint(uint64(Mv.Heading), 10)
	}

	return str
}

func parseDate(DA, TI, Tag1, Tag2 string) (*omnicom.Dt, error) {

	date := &omnicom.Dt{}

	if len(strings.Split(DA, "/")) != 2 {
		return nil, fmt.Errorf("Wrong naf syntax")
	}

	if strings.Contains(DA, Tag1) && len(strings.Split(DA, "/")[1]) == 8 {

		num, err := strconv.ParseUint(strings.Split(DA, "/")[1][:4], 10, 32)
		if err != nil {
			return nil, err
		}
		date.Year = uint32(num)

		num, err = strconv.ParseUint(strings.Split(DA, "/")[1][4:6], 10, 32)
		if err != nil {
			return nil, err
		}
		date.Month = uint32(num)

		num, err = strconv.ParseUint(strings.Split(DA, "/")[1][6:8], 10, 32)
		if err != nil {
			return nil, err
		}

		date.Day = uint32(num)

	} else {
		return nil, fmt.Errorf("Invalid or missing %s field", Tag1)
	}

	if len(strings.Split(TI, "/")) != 2 {
		return nil, fmt.Errorf("Wrong naf syntax")
	}

	if strings.Contains(TI, Tag2) && len(strings.Split(TI, "/")[1]) == 4 {
		hours, err := strconv.ParseUint(strings.Split(TI, "/")[1][:2], 10, 32)
		if err != nil {
			return nil, err
		}
		minutes, err := strconv.ParseUint(strings.Split(TI, "/")[1][2:4], 10, 32)
		if err != nil {
			return nil, err
		}
		date.Minute = uint32(hours)*60 + uint32(minutes)
	} else {
		return nil, fmt.Errorf("Invalid or missing %s field", Tag2)
	}

	return date, nil
}

func toFixed(x float32, p int) float64 {
	return float64(int64(x)) + float64(int64(float64((x-float32(int64(x))))*math.Pow10(p)))/math.Pow10(p)
}
