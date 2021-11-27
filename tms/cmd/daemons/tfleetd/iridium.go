package main

import (
	"fmt"
	"prisma/tms/iridium"
	"prisma/tms/log"
	omni "prisma/tms/omnicom"
	"prisma/tms/tmsg"
	"prisma/tms/util/ident"
	"prisma/tms/util/omnicom"
	"time"
)

func calculateTime(date interface{}) (time.Time, error) {
	var year, month, day, minute uint32
	switch t := date.(type) {
	case *omni.DatePosition:
		datePosition := date.(*omni.DatePosition)
		year = datePosition.Year
		month = datePosition.Month
		day = datePosition.Day
		minute = datePosition.Minute
	case *omni.DateEvent:
		dateEvent := date.(*omni.DateEvent)
		year = dateEvent.Year
		month = dateEvent.Month
		day = dateEvent.Day
		minute = dateEvent.Minute
	default:
		log.Warn("Invalid struct passed to calculateTime : %T", t)
		return time.Time{}, fmt.Errorf("Invalid struct passed to calculateTime : %T", t)
	}

	century := (time.Now().Year() / 100) * 100
	dateTime := time.Date(century+int(year),
		time.Month(month),
		int(day),
		0,
		int(minute),
		0,
		0,
		time.UTC)
	return dateTime, nil
}

// CreateAaMessage creates a new Ack Assistance (0x45) iridium message for the given IMEI.
func CreateAaMessage(imei string, messageid uint32, messageTime time.Time) (*iridium.Iridium, error) {
	date := omnicom.CreateOmnicomDate(messageTime)
	aa := omni.AA{
		Header:  0x45,
		Date:    date,
		Padding: 0,
		CRC:     18,
	}
	return createMTMessage(imei, messageid, &aa)
}

// CreateTmaMessage creates a new Test Mode Acknowledge (0x30) iridium message for the given IMEI.
func CreateTmaMessage(imei string, messageid uint32, messageTime time.Time) (*iridium.Iridium, error) {
	date := omnicom.CreateOmnicomDate(messageTime)
	tma := omni.TMA{
		Header:  0x30,
		Date:    date,
		Padding: 0,
		CRC:     18,
	}
	return createMTMessage(imei, messageid, &tma)
}

// CreateUicMessage creates a new Unit Inetrval Change (0x32) iridium message for the given IMEI.
func CreateUicMessage(imei string, messageid uint32, reportingperiod uint32, messageTime time.Time) (*iridium.Iridium, error) {
	date := omnicom.CreateOmnicomDate(messageTime)
	uic := omni.UIC{
		Header:        0x32,
		ID_Msg:        messageid,
		Date:          date,
		New_Reporting: RoundReportingPeriod(reportingperiod),
		CRC:           18,
	}
	return createMTMessage(imei, messageid, &uic)
}

// CreateUgpMessage converts the given Update Global Parameters (0x34) message into an Iridium message.
func CreateUgpMessage(imei string, messageid uint32, ugp *omni.Ugp, messageTime time.Time) (*iridium.Iridium, error) {
	date := omnicom.CreateOmnicomDate(messageTime)
	omniugp := omni.UGP{
		Header:  0x34,
		ID_Msg:  messageid,
		Date:    date,
		Padding: 0,
		CRC:     18,
	}
	if ugp.Position_Reporting_Interval != nil {
		omniugp.Position_Reporting_Interval = omni.Position_Reporting_Interval_StoV{
			ValueInMn: RoundReportingPeriod(ugp.Position_Reporting_Interval.ValueInMn),
			To_Modify: ugp.Position_Reporting_Interval.To_Modify,
		}
	}
	if ugp.Geofencing_Enable != nil {
		omniugp.Geofencing_Enable = omni.Geofencing_Enable_StoV{
			On_Off:    ugp.Geofencing_Enable.On_Off,
			To_Modify: ugp.Geofencing_Enable.To_Modify,
		}
	}
	if ugp.Geofence_Status_Check_Interval != nil {
		omniugp.Geofence_Status_Check_Interval = omni.Geofence_Status_Check_Interval_StoV{
			ValueInMn: RoundReportingPeriod(ugp.Geofence_Status_Check_Interval.ValueInMn),
			To_Modify: ugp.Geofence_Status_Check_Interval.To_Modify,
		}
	}
	if ugp.Password != nil {
		omniugp.Password = omni.Password_StoV{
			Value:     ugp.Password.Value,
			To_Modify: ugp.Password.To_Modify,
		}
	}
	if ugp.Routing != nil {
		omniugp.Routing = omni.Routing_StoV{
			Value:     ugp.Routing.Value,
			To_Modify: ugp.Routing.To_Modify,
		}
	}

	return createMTMessage(imei, messageid, &omniugp)
}

func createMTMessage(imei string, messageid uint32, omnicom omni.Omnicom) (*iridium.Iridium, error) {
	header := iridium.MTHeader{
		IEI:  0x41,
		MTHL: 21,
		UniqueClientMessageID: fmt.Sprint(messageid),
		IMEI:   imei,
		MTflag: 0,
	}
	payload := iridium.MPayload{
		IEI: 0x42,
		Omn: omnicom,
	}
	// Get payload length
	bytes, err := payload.Encode()
	if err != nil {
		return nil, err
	}
	payload.PayloadL = uint16(len(bytes))
	return iridium.PopulateMTProtobuf(header, payload)
}

// RoundReportingPeriod returns the round(reportingperiod) value, rounding down to multiples of 5.
func RoundReportingPeriod(reportingperiod uint32) uint32 {
	reportingperiod = (reportingperiod / 5) * 5
	return reportingperiod
}

func createIDforIridiumMessage(imei string) string {
	return ident.
		With("imei", imei).
		With("site", tmsg.GClient.Local().Site).
		With("eid", tmsg.GClient.Local().Eid).
		Hash()
}

func createRegistryIDforIridiumMessage(imei string) string {
	return ident.With("imei", imei).Hash()
}
