package libnaf

import (
	"fmt"
	"strconv"
	"strings"

	"prisma/tms/iridium"
	"prisma/tms/omnicom"
)

func parseUGFP(fields []string) (*iridium.Iridium, error) {

	var err error

	iri := &iridium.Iridium{}
	iri.Mth = &iridium.MobileTerminatedHeader{}
	iri.Mth.IMEI = fields[0]
	iri.Payload = &iridium.Payload{}
	iri.Payload.Omnicom = &omnicom.Omni{}
	iri.Payload.Omnicom.Omnicom = &omnicom.Omni_Ugpolygon{}

	iri.Payload.Omnicom.GetUgpolygon().Header = []byte{0x35}
	iri.Payload.Omnicom.GetUgpolygon().Setting = &omnicom.Stg{}

	for i := 2; i < len(fields); i++ {
		switch strings.Split(fields[i], "/")[0] {
		case "ID":
			id, err := parseID(fields[i])
			if err != nil {
				return nil, err
			}
			iri.Payload.Omnicom.GetUgpolygon().Msg_ID = uint32(id)
		case "DA":
			iri.Payload.Omnicom.GetUgpolygon().Date, err = parseDate(fields[i], fields[i+1], "DA", "TI")
			if err != nil {
				return nil, err
			}
			i++
		case "GEOID":
			id, err := parseID(fields[i])
			if err != nil {
				return nil, err
			}
			iri.Payload.Omnicom.GetUgpolygon().GEO_ID = uint32(id)
		case "GZN":
			iri.Payload.Omnicom.GetUgpolygon().NAME = []byte(strings.Split(fields[i], "/")[1])
		case "GEOP":
			pr, err := strconv.ParseUint(strings.Split(fields[i], "/")[1], 10, 32)
			if err != nil {
				return nil, err
			}
			iri.Payload.Omnicom.GetUgpolygon().Priority = uint32(pr)
		case "GEOA":
			pr, err := strconv.ParseUint(strings.Split(fields[i], "/")[1], 10, 32)
			if err != nil {
				return nil, err
			}
			iri.Payload.Omnicom.GetUgpolygon().Activated = uint32(pr)
		case "NRI":
			in, err := strconv.ParseUint(strings.Split(fields[i], "/")[1], 10, 32)
			if err != nil {
				return nil, err
			}
			iri.Payload.Omnicom.GetUgpolygon().Setting.New_Position_Report_Period = uint32(in)
		case "SPT":
			sp, err := strconv.ParseFloat(strings.Split(fields[i], "/")[1], 32)
			if err != nil {
				return nil, err
			}
			iri.Payload.Omnicom.GetUgpolygon().Setting.Speed_Threshold = float32(sp)
		case "GEOV":
			points := strings.Split(strings.Split(fields[i], "/")[1], " ")
			iri.Payload.Omnicom.GetUgpolygon().Position = []*omnicom.Pos{}
			for _, point := range points {
				lat, err := strconv.ParseFloat(strings.Split(point, ",")[0], 64)
				if err != nil {
					return nil, err
				}
				long, err := strconv.ParseFloat(strings.Split(point, ",")[1], 64)
				if err != nil {
					return nil, err
				}
				iri.Payload.Omnicom.GetUgpolygon().Position = append(iri.Payload.Omnicom.GetUgpolygon().Position, &omnicom.Pos{Longitude: long, Latitude: lat})
			}
		default:
			return nil, fmt.Errorf("uknown field in UGPolygon naf %s", strings.Split(fields[i], "/")[0])
		}
	}

	return iri, nil

}

func encodeUGFP(UGFP *omnicom.UGPolygon) string {

	var str string

	str = str + "//TOM/UG_Polygon"

	str = encodeID(UGFP.Msg_ID, str)

	str = encodeDate(UGFP.Date, str)

	str = encodeGEOID(UGFP.GEO_ID, str)

	str = str + "//GZN/" + string(UGFP.NAME)

	str = str + "//GEOP/" + strconv.Itoa(int(UGFP.Priority))

	str = str + "//GEOA/" + strconv.Itoa(int(UGFP.Activated))

	str = str + "//NRI/" + strconv.Itoa(int(UGFP.Setting.New_Position_Report_Period))

	str = str + "//SPT/" + strconv.FormatFloat(float64(UGFP.Setting.Speed_Threshold), 'f', -1, 32)

	str = str + "//GEOV/"

	if len(UGFP.Position) != 0 {
		for i, pos := range UGFP.Position {
			str = str + strconv.FormatFloat(pos.Latitude, 'f', -1, 32) + "," + strconv.FormatFloat(pos.Longitude, 'f', -1, 32)
			if i != (len(UGFP.Position) - 1) {
				str = str + " "
			}
		}
	}

	return str + ER
}
