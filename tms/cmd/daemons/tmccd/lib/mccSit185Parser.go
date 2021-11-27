package lib

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	tsi "prisma/tms"
	"prisma/tms/sar"
	sitPkg "prisma/tms/sit"
	"prisma/tms/sit185"

	"github.com/golang/protobuf/ptypes/wrappers"
)

//Sit185Parser returns SarsatMessage structure and take the raw sit185 message, message protocol, date format and regex template
func Sit185Parser(msg []byte, protocol, dateformat string, RegexpTemplate []sit185.RegularExpPattern) (*sar.SarsatMessage, error) {
	sarsatMessage := &sar.SarsatMessage{}
	var err error
	var sit sit185.Sit185
	composite := sar.Point{}
	encoded := sar.Point{}
	doa := sar.DOA{}
	var elementals []*sar.Element

	sit, err = sit185.Parse(string(msg), RegexpTemplate)
	if err != nil {
		return nil, err
	}
	sarsatMessage.MessageBody = sit.Raw
	sarsatMessage.Protocol = protocol
	sarsatMessage.Received = true
	sarsatMessage.RemoteType = sar.SarsatMessage_MCC
	sarsatMessage.MessageType = sar.SarsatMessage_SIT_185
	sarsatMessage.SarsatAlert = &sar.SarsatAlert{}

	headerParser := sitPkg.NewScanner(sit.Raw)
	headerParser.Next()
	sarsatMessage.RemoteName = headerParser.Next()

	for _, field := range RegexpTemplate {

		switch field.Fieldname {
		case "msg_num":
			if sit.Fields["msg_num"] != "" {
				MsgNum, err := strconv.Atoi(sit.Fields["msg_num"])
				if err != nil {
					return nil, fmt.Errorf("unable to process mcc msg number field: %v", err)
				}
				sarsatMessage.MessageNumber = int32(MsgNum)
			}
		case "date":
			if sit.Fields["date"] != "" {
				timestamp, err := fromSit185Date(sit.Fields["date"], dateformat)
				if err != nil {
					return nil, fmt.Errorf("unable to process date field: %v", err)
				}
				sarsatMessage.MessageTime = tsi.ToTimestamp(timestamp)
			}
		case "hex_id":
			if sit.Fields["hex_id"] != "" || len(sit.Fields["hex_id"]) != 15 {
				str := strings.TrimSpace(sit.Fields["hex_id"])
				if str == "NIL" || str == "UKNOWN" {
					continue
				}
				beacon, err := sar.DecodeHexID(str)
				if err != nil {
					return nil, err
				}
				sarsatMessage.SarsatAlert.Beacon = beacon
			}
		case "confirmed":
			if sit.Fields["confirmed"] != "" {
				str := strings.TrimSpace(sit.Fields["confirmed"])
				if str == "NIL" || str == "UNKNOWN" {
					continue
				}
				latdeg, err := strconv.Atoi(strings.TrimSpace(sit.Fields["confirmed_lat_degree"]))
				if err != nil {
					return nil, err
				}
				latmin, err := strconv.ParseFloat(strings.TrimSpace(sit.Fields["confirmed_lat_min"]), 64)
				if err != nil {
					return nil, err
				}
				londeg, err := strconv.Atoi(strings.TrimSpace(sit.Fields["confirmed_lon_degree"]))
				if err != nil {
					return nil, err
				}
				lonmin, err := strconv.ParseFloat(strings.TrimSpace(sit.Fields["confirmed_lon_min"]), 64)
				if err != nil {
					return nil, err
				}

				point, err := parseDDM(latdeg, londeg, latmin, lonmin,
					strings.TrimSpace(sit.Fields["confirmed_lat_cardinal_point"]),
					strings.TrimSpace(sit.Fields["confirmed_lon_cardinal_point"]))
				if err != nil {
					return nil, err
				}

				composite = point
			}

		case "doppler_a":

			if sit.Fields["doppler_a"] != "" {
				str := strings.TrimSpace(sit.Fields["doppler_a"])
				if str == "NIL" || str == "UNKNOWN" {
					continue
				}
				element, err := dopplerTOelement(sit.Fields["doppler_a_lat_degree"], sit.Fields["doppler_a_lat_min"],
					sit.Fields["doppler_a_lon_degree"], sit.Fields["doppler_a_lon_min"],
					sit.Fields["doppler_a_lat_cardinal_point"], sit.Fields["doppler_a_lon_cardinal_point"],
					sit.Fields["doppler_a_prob"])
				if err != nil {
					return nil, err
				}

				elementals = append(elementals, element)
			}
		case "doppler_b":

			if sit.Fields["doppler_b"] != "" {
				str := strings.TrimSpace(sit.Fields["doppler_b"])
				if str == "NIL" || str == "UNKNOWN" {
					continue
				}

				element, err := dopplerTOelement(sit.Fields["doppler_b_lat_degree"], sit.Fields["doppler_b_lat_min"],
					sit.Fields["doppler_b_lon_degree"], sit.Fields["doppler_b_lon_min"],
					sit.Fields["doppler_b_lat_cardinal_point"], sit.Fields["doppler_b_lon_cardinal_point"],
					sit.Fields["doppler_b_prob"])
				if err != nil {
					return nil, err
				}

				elementals = append(elementals, element)
			}
		case "doa":
			if sit.Fields["doa"] != "" {
				str := strings.TrimSpace(sit.Fields["doa"])
				if str == "NIL" || str == "UKNOWN" {
					continue
				}
				latdeg, err := strconv.Atoi(strings.TrimSpace(sit.Fields["doa_lat_degree"]))
				if err != nil {
					return nil, err
				}
				latmin, err := strconv.ParseFloat(strings.TrimSpace(sit.Fields["doa_lat_min"]), 64)
				if err != nil {
					return nil, err
				}
				londeg, err := strconv.Atoi(strings.TrimSpace(sit.Fields["doa_lon_degree"]))
				if err != nil {
					return nil, err
				}
				lonmin, err := strconv.ParseFloat(strings.TrimSpace(sit.Fields["doa_lon_min"]), 64)
				if err != nil {
					return nil, err
				}

				point, err := parseDDM(latdeg, londeg, latmin, lonmin,
					strings.TrimSpace(sit.Fields["doa_lat_cardinal_point"]),
					strings.TrimSpace(sit.Fields["doa_lon_cardinal_point"]))
				if err != nil {
					return nil, err
				}

				doa.DoaPosition = &point

				altitude, err := strconv.ParseFloat(strings.TrimSpace(sit.Fields["doa_elevation"]), 64)
				doa.Altitude = &wrappers.DoubleValue{Value: altitude}
			}

		case "encoded":
			if sit.Fields["encoded"] != "" {
				str := strings.TrimSpace(sit.Fields["encoded"])
				if str == "NIL" || str == "UNKNOWN" {
					continue
				}

				latdeg, err := strconv.Atoi(strings.TrimSpace(sit.Fields["encoded_lat_degree"]))
				if err != nil {
					return nil, err
				}
				latmin, err := strconv.ParseFloat(strings.TrimSpace(sit.Fields["encoded_lat_min"]), 64)
				if err != nil {
					return nil, err
				}
				londeg, err := strconv.Atoi(strings.TrimSpace(sit.Fields["encoded_lon_degree"]))
				if err != nil {
					return nil, err
				}
				lonmin, err := strconv.ParseFloat(strings.TrimSpace(sit.Fields["encoded_lon_min"]), 64)
				if err != nil {
					return nil, err
				}

				point, err := parseDDM(latdeg, londeg, latmin, lonmin,
					strings.TrimSpace(sit.Fields["encoded_lat_cardinal_point"]),
					strings.TrimSpace(sit.Fields["encoded_lon_cardinal_point"]))
				if err != nil {
					return nil, err
				}

				encoded = point
			}

		default:
			continue
		} // end of switch
	} // end of for loop

	if !isPointEmpty(composite) {
		resolvedalert := &sar.ResolvedAlert{}
		resolvedalert.CompositeLocation = &composite
		if !isDoaEmpty(doa) {
			resolvedalert.MeoElemental = &sar.Meoelemental{Doa: &doa}
		}
		for _, element := range elementals {
			resolvedalert.Elemental = append(resolvedalert.Elemental, element)
		}
		if !isPointEmpty(encoded) {
			resolvedalert.Encoded = &encoded
		}

		sarsatMessage.SarsatAlert.ResolvedAlertMessage = resolvedalert
		sarsatMessage.SarsatAlert.AlertType = sar.SarsatAlert_ResolvedAlert
		sarsatMessage.SarsatAlert.ProcessedTime = tsi.Now()
		return sarsatMessage, nil
	}
	if !isDoaEmpty(doa) || len(elementals) != 0 || !isPointEmpty(encoded) {
		unresolvedalert := &sar.IncidentAlert{}
		if !isDoaEmpty(doa) {
			unresolvedalert.MeoElemental = &sar.Meoelemental{Doa: &doa}
		}
		for _, element := range elementals {
			unresolvedalert.Elemental = append(unresolvedalert.Elemental, element)
		}
		if !isPointEmpty(encoded) {
			unresolvedalert.Encoded = &encoded
		}

		sarsatMessage.SarsatAlert.IncidentAlertMessage = unresolvedalert
		sarsatMessage.SarsatAlert.AlertType = sar.SarsatAlert_IncidentAlert
		sarsatMessage.SarsatAlert.ProcessedTime = tsi.Now()
		return sarsatMessage, nil
	}

	if sarsatMessage.SarsatAlert == nil && (sit.Fields["msg_num"] != "" || sit.Fields["date"] != "" || sit.Fields["hex_id"] != "") {
		sarsatMessage.SarsatAlert.AlertType = sar.SarsatAlert_UnlocatedAlert
		return sarsatMessage, nil
	}

	return nil, fmt.Errorf("MCC message format is Unkown")
}

//parseDDM converts degree decimal minutes to decimal degrees
func parseDDM(latDegree, lonDegree int, latMin, lonMin float64, latCardinalPoint, lonCardinalPoint string) (sar.Point, error) {

	point := sar.Point{}

	if latCardinalPoint == "S" {
		point.Latitude = -(float64(latDegree) + latMin/60)
	} else if latCardinalPoint == "N" {
		point.Latitude = (float64(latDegree) + latMin/60)
	} else {
		return point, fmt.Errorf("unable to parse latitude cardinal point %s", latCardinalPoint)
	}

	if lonCardinalPoint == "W" {
		point.Longitude = -(float64(lonDegree) + lonMin/60)
	} else if lonCardinalPoint == "E" {
		point.Longitude = (float64(lonDegree) + lonMin/60)
	} else {
		return point, fmt.Errorf("unable to parse longitude cardinal point %s", lonCardinalPoint)
	}

	if 180.0 <= point.Longitude || point.Longitude <= -180.0 {
		return point, fmt.Errorf("unable to parse longitude due to invalid range %v", point.Longitude)
	}

	if 90.0 <= point.Latitude || point.Latitude <= -90 {
		return point, fmt.Errorf("unable to parse latitude due to invalid range %v", point.Latitude)
	}

	return point, nil

}

func fromSit185Date(date, dateformat string) (time.Time, error) {
	var err error
	var t time.Time

	formats := strings.Fields(dateformat)
	fields := strings.Fields(date)

	Months := map[string]int{
		"JAN": 1,
		"FEB": 2,
		"MAR": 3,
		"APR": 4,
		"MAY": 5,
		"JUN": 6,
		"JUL": 7,
		"AUG": 8,
		"SEP": 9,
		"OCT": 10,
		"NOV": 11,
		"DEC": 12,
	}

	var TZ *time.Location
	var Year int
	var Month int
	var Day int
	Hour := -1
	Minute := -1

	for key, format := range formats {
		switch format {
		case "%Y":
			Year, err = strconv.Atoi(fields[key])
			if err != nil {
				return t, err
			}
			// This is to make sure that two digit years representation YY, get interpreted as 20YY not 19YY
			if len(fields[key]) == 2 {
				Year = Year + 2000
			}
		case "%M":
			Month = Months[fields[key]]
			if Month == 0 {
				return t, fmt.Errorf("unable to parse Month in %s", date)
			}
		case "%D":
			Day, err = strconv.Atoi(fields[key])
			if err != nil {
				return t, err
			}
		case "%H%MN":
			if len(fields[key]) < 4 {
				return t, fmt.Errorf("unable to parse Hours Minutes in %s", date)
			}
			Hour, err = strconv.Atoi(fields[key][:2])
			if err != nil {
				return t, err
			}
			Minute, err = strconv.Atoi(fields[key][2:4])
			if err != nil {
				return t, err
			}
		case "%H":
			Hour, err = strconv.Atoi(fields[key])
			if err != nil {
				return t, err
			}
		case "%MN":
			Minute, err = strconv.Atoi(fields[key])
			if err != nil {
				return t, err
			}
		case "%TZ":
			TZ, err = time.LoadLocation(fields[key])
			if err != nil {
				return t, err
			}

		default:
			continue

		} //end of switch
	}

	if Year != 0 && Month != 0 && TZ != nil && Hour != -1 && Minute != -1 {
		t = time.Date(Year, time.Month(Month), Day, Hour, Minute, 0, 0, TZ)
		return t, nil
	}

	return t, fmt.Errorf("unable to parse date: %s", date)
}

func dopplerTOelement(latDegree, latMin, lonDegree, lonMin, latCardPoint, lonCardPoint, probability string) (*sar.Element, error) {

	latdeg, err := strconv.Atoi(strings.TrimSpace(latDegree))
	if err != nil {
		return nil, err
	}
	latmin, err := strconv.ParseFloat(strings.TrimSpace(latMin), 64)
	if err != nil {
		return nil, err
	}
	londeg, err := strconv.Atoi(strings.TrimSpace(lonDegree))
	if err != nil {
		return nil, err
	}
	lonmin, err := strconv.ParseFloat(strings.TrimSpace(lonMin), 64)
	if err != nil {
		return nil, err
	}
	point, err := parseDDM(latdeg, londeg, latmin, lonmin,
		strings.TrimSpace(latCardPoint),
		strings.TrimSpace(lonCardPoint))
	if err != nil {
		return nil, err
	}
	prob, err := strconv.ParseFloat(strings.TrimSpace(probability), 64)
	// We check is prob is between 0 and 100, but we should make sure that if prob is not available we dont fail.
	if (err != nil || prob > 100 || prob < 0) && probability != "" {
		return nil, fmt.Errorf("unable to parse doppler_a_prob value: %+v", err)
	}

	doppler := []*sar.Doppler{&sar.Doppler{
		DopplerPosition:    &point,
		Type:               0,
		DopplerProbability: &wrappers.DoubleValue{Value: prob},
	}}

	return &sar.Element{Doppler: doppler}, nil

}

func isPointEmpty(pt sar.Point) bool {
	return reflect.DeepEqual(pt, sar.Point{})
}

func isDoaEmpty(doa sar.DOA) bool {
	return reflect.DeepEqual(doa, sar.DOA{})
}
