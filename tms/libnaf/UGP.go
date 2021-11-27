package libnaf

import "prisma/tms/iridium"
import "prisma/tms/omnicom"
import "strings"
import "strconv"

func parseUGP(fields []string) (*iridium.Iridium, error) {

	var err error
	var value uint64

	iri := &iridium.Iridium{}
	iri.Mth = &iridium.MobileTerminatedHeader{}
	iri.Mth.IMEI = fields[0]
	iri.Payload = &iridium.Payload{}
	iri.Payload.Omnicom = &omnicom.Omni{}
	iri.Payload.Omnicom.Omnicom = &omnicom.Omni_Ugp{}

	iri.Payload.Omnicom.GetUgp().Header = []byte{0x34}

	if len(fields) >= 4 {
		iri.Payload.Omnicom.GetUgp().Date, err = parseDate(fields[2], fields[3], "DA", "TI")
		if err != nil {
			return nil, err
		}

	}

	for i := 4; i < len(fields); i++ {
		switch strings.Split(fields[i], "/")[0] {
		case "RPI":
			value, err = strconv.ParseUint(strings.Split(fields[i], "/")[1], 10, 32)
			iri.Payload.Omnicom.GetUgp().Position_Reporting_Interval = &omnicom.PositionReportingIntervalStoV{}
			iri.Payload.Omnicom.GetUgp().Position_Reporting_Interval.ValueInMn = uint32(value)
		case "GFE":
			value, err = strconv.ParseUint(strings.Split(fields[i], "/")[1], 10, 32)
			iri.Payload.Omnicom.GetUgp().Geofencing_Enable = &omnicom.GeofencingEnableStoV{}
			iri.Payload.Omnicom.GetUgp().Geofencing_Enable.On_Off = uint32(value)
		case "GFSI":
			value, err = strconv.ParseUint(strings.Split(fields[i], "/")[1], 10, 32)
			iri.Payload.Omnicom.GetUgp().Geofence_Status_Check_Interval = &omnicom.GeofenceStatusCheckIntervalStoV{}
			iri.Payload.Omnicom.GetUgp().Geofence_Status_Check_Interval.ValueInMn = uint32(value)
		case "PSW":
			iri.Payload.Omnicom.GetUgp().Password = &omnicom.PasswordStoV{}
			iri.Payload.Omnicom.GetUgp().Password.Value = []byte(strings.Split(fields[i], "/")[1])
		case "RTG":
			value, err = strconv.ParseUint(strings.Split(fields[i], "/")[1], 10, 32)
			iri.Payload.Omnicom.GetUgp().Routing = &omnicom.RoutingStoV{}
			iri.Payload.Omnicom.GetUgp().Routing.Value = uint32(value)
		}

		if err != nil {
			return nil, err
		}
	}

	return iri, nil

}

func encodeUGP(UGP *omnicom.Ugp) string {
	var str string

	str = str + "//TOM/UGP"

	str = encodeDate(UGP.Date, str)

	if UGP.Position_Reporting_Interval.To_Modify == 1 {
		str = str + "//RPI/" + strconv.Itoa(int(UGP.Position_Reporting_Interval.ValueInMn))
	}

	if UGP.Geofencing_Enable.To_Modify == 1 {
		str = str + "//GFE/" + strconv.Itoa(int(UGP.Geofencing_Enable.On_Off))
	}

	if UGP.Geofence_Status_Check_Interval.To_Modify == 1 {
		str = str + "//GFSI/" + strconv.Itoa(int(UGP.Geofence_Status_Check_Interval.ValueInMn))
	}

	if UGP.Password.To_Modify == 1 {
		str = str + "//PSW/" + string(UGP.Password.Value)
	}

	if UGP.Routing.To_Modify == 1 {
		str = str + "//RTG/" + strconv.Itoa(int(UGP.Routing.Value))
	}

	return str + ER
}
