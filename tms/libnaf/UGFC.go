package libnaf

import (
	"fmt"
	"prisma/tms/iridium"
	"prisma/tms/omnicom"
	"strconv"
	"strings"
)

func parseUGFC(fields []string) (*iridium.Iridium, error) {

	var err error

	iri := &iridium.Iridium{}
	iri.Mth = &iridium.MobileTerminatedHeader{}
	iri.Mth.IMEI = fields[0]
	iri.Payload = &iridium.Payload{}
	iri.Payload.Omnicom = &omnicom.Omni{}
	iri.Payload.Omnicom.Omnicom = &omnicom.Omni_Ugcircle{}

	iri.Payload.Omnicom.GetUgcircle().Header = []byte{0x35}
	iri.Payload.Omnicom.GetUgcircle().Setting = &omnicom.Stg{}
	iri.Payload.Omnicom.GetUgcircle().Position = &omnicom.PositionRadius{}

	for i := 2; i < len(fields); i++ {
		switch strings.Split(fields[i], "/")[0] {
		case "ID":
			iri.Payload.Omnicom.GetUgcircle().Msg_ID, err = parseID(fields[i])
		case "DA":
			iri.Payload.Omnicom.GetUgcircle().Date, err = parseDate(fields[i], fields[i+1], "DA", "TI")
			i++
		case "GEOID":
			iri.Payload.Omnicom.GetUgcircle().GEO_ID, err = parseID(fields[i])
		case "GZN":
			iri.Payload.Omnicom.GetUgcircle().NAME = []byte(strings.Split(fields[i], "/")[1])
		case "GEOP":
			var pr uint64
			pr, err = strconv.ParseUint(strings.Split(fields[i], "/")[1], 10, 32)
			iri.Payload.Omnicom.GetUgcircle().Priority = uint32(pr)
		case "GEOA":
			var pr uint64
			pr, err = strconv.ParseUint(strings.Split(fields[i], "/")[1], 10, 32)
			iri.Payload.Omnicom.GetUgcircle().Activated = uint32(pr)
		case "NRI":
			var in uint64
			in, err = strconv.ParseUint(strings.Split(fields[i], "/")[1], 10, 32)
			iri.Payload.Omnicom.GetUgcircle().Setting.New_Position_Report_Period = uint32(in)
		case "SPT":
			var sp float64
			sp, err = strconv.ParseFloat(strings.Split(fields[i], "/")[1], 32)
			iri.Payload.Omnicom.GetUgcircle().Setting.Speed_Threshold = float32(sp)
		case "GEOC":
			var lat, long float64
			point := strings.Split(fields[i], "/")
			lat, err = strconv.ParseFloat(strings.Split(point[1], ",")[0], 32)
			iri.Payload.Omnicom.GetUgcircle().Position.Latitude = float32(lat)

			long, err = strconv.ParseFloat(strings.Split(point[1], ",")[1], 32)
			iri.Payload.Omnicom.GetUgcircle().Position.Longitude = float32(long)
		case "GEOR":
			var radius float64
			radius, err = strconv.ParseFloat(strings.Split(fields[i], "/")[1], 32)
			iri.Payload.Omnicom.GetUgcircle().Position.Radius = float32(radius)
		default:
			return nil, fmt.Errorf("uknown field in UGCircle naf %s", strings.Split(fields[i], "/")[0])
		}
		if err != nil {
			return nil, err
		}
	}

	return iri, nil
}

func encodeUGFC(UGFC *omnicom.UGCircle) string {

	var str string

	str = str + "//TOM/UG_Cirlce"

	str = encodeID(UGFC.Msg_ID, str)

	str = encodeDate(UGFC.Date, str)

	str = encodeGEOID(UGFC.GEO_ID, str)

	str = str + "//GZN/" + string(UGFC.NAME)

	str = str + "//GEOP/" + strconv.Itoa(int(UGFC.Priority))

	str = str + "//GEOA/" + strconv.Itoa(int(UGFC.Activated))

	str = str + "//NRI/" + strconv.Itoa(int(UGFC.Setting.New_Position_Report_Period))

	str = str + "//SPT/" + strconv.FormatFloat(float64(UGFC.Setting.Speed_Threshold), 'f', -1, 32)

	str = str + "//GEOC/" + strconv.FormatFloat(float64(UGFC.Position.Latitude), 'f', -1, 32) + "," + strconv.FormatFloat(float64(UGFC.Position.Longitude), 'f', -1, 32)

	str = str + "//GEOR/" + strconv.FormatFloat(float64(UGFC.Position.Latitude), 'f', -1, 32)

	return str + ER

}
