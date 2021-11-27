// omnicom_buffers is used to create buffers for different type of messages of omnicom beacons
package object

import (
	"fmt"
	"prisma/tms/iridium"
	"prisma/tms/log"
	"prisma/tms/omngen"
	"prisma/tms/omnicom"
	"time"
)

func (o *Omnicom) makeBuffSPR() ([]byte, error) {
	date := omngen.CurrentTime()
	o.lastDatePosition = date
	spr := omnicom.SPR{
		Header: 0x06,
		Date_Position: omnicom.Date_Position{
			Year:      date.Year,
			Month:     date.Month,
			Day:       date.Day,
			Minute:    date.Minute,
			Longitude: float32(o.object.GetCurPos().Longitude),
			Latitude:  float32(o.object.GetCurPos().Latitude),
		},
		Move: omnicom.Move{
			Speed:   float32(o.object.GetCurPos().Speed),
			Heading: uint32(o.object.bearing),
		},
		Period:  o.object.ReportPeriod,
		Padding: 0,
		CRC:     18,
		Extention_Bit_Beacon_ID: 1,
		Beacon_ID:               o.object.OmnicomId,
	}

	praw, err := spr.Encode()
	if err != nil {
		return nil, err
	}

	var raw []byte
	raw = append(raw, 0x01, byte(uint16(len(praw)+3+len(o.hraw)))>>8, byte(uint16(len(praw)+3+len(o.hraw))))
	raw = append(raw, o.hraw...)

	gfLength := uint16(len(praw))
	raw = append(raw, 0x02, byte(gfLength>>8), byte(gfLength))
	raw = append(raw, praw...)
	return raw, nil
}

func (o *Omnicom) makeBuffStartAlerting(mid uint32) ([]byte, error) {
	o.ar = &omnicom.AR{
		Header:                  0x02,
		Beacon_ID:               o.object.OmnicomId,
		Extention_Bit_Beacon_ID: 1,
		Extention_Bit_Move:      1,
		Move:                    omnicom.Move{Speed: float32(o.object.GetCurPos().Speed), Heading: uint32(o.object.bearing)},
		Date_Position:           o.lastDatePosition,
		Date_Event:              eventTime(),
		Msg_ID:                  mid,
	}
	err := o.assignAlertTypeValue(o.startTypeAlerting, 1, 1)
	if err != nil {
		return nil, err
	}
	Payload := iridium.MPayload{0x02, 0, o.ar}
	praw, errPraw := Payload.Encode()
	if errPraw != nil {
		return nil, errPraw
	}
	var raw []byte
	raw = append(raw, 0x01, byte(uint16(len(praw)+len(o.hraw)))>>8, byte(uint16(len(praw)+len(o.hraw))))
	raw = append(raw, o.hraw...)
	raw = append(raw, praw...)

	// This will set all of the current mode status to 0
	o.reInitializeCurrentStatus()
	return raw, nil
}

func (o *Omnicom) makeBuffStopAlerting() ([]byte, error) {
	o.ar = &omnicom.AR{
		Header:                  0x02,
		Msg_ID:                  0,
		Beacon_ID:               o.object.OmnicomId,
		Extention_Bit_Beacon_ID: 1,
		Extention_Bit_Move:      1,
		Move:                    omnicom.Move{Speed: float32(o.object.GetCurPos().Speed), Heading: uint32(o.object.bearing)},
		Date_Position:           o.lastDatePosition,
		Date_Event:              eventTime(),
	}
	err := o.assignAlertTypeValue(o.stopTypeAlerting, 0, 1)
	if err != nil {
		return nil, err
	}
	Payload := iridium.MPayload{0x02, 0, o.ar}
	Praw, errPraw := Payload.Encode()
	if errPraw != nil {
		return nil, errPraw
	}
	var raw []byte
	raw = append(raw, 0x01, byte(uint16(len(Praw)+len(o.hraw)))>>8, byte(uint16(len(Praw)+len(o.hraw))))
	raw = append(raw, o.hraw...)
	raw = append(raw, Praw...)
	// This will set all of the current mode status to 0
	o.reInitializeCurrentStatus()
	return raw, nil
}

func (o *Omnicom) makeBuffGeofenceAck() ([]byte, error) {
	gf := omnicom.GA{
		Header:     0x04,
		Error_Type: 0,
		Date_Position: omnicom.Date_Position{
			Longitude: float32(o.object.curPos.Longitude),
			Latitude:  float32(o.object.curPos.Latitude),
			Day:       uint32(time.Now().Day()),
			Month:     uint32(time.Now().Month()),
			Year:      uint32(time.Now().Year()) - 2000,
			Minute:    uint32(time.Now().Minute() + time.Now().Hour()*60),
		},
		Msg_ID: uint32(time.Now().Nanosecond()) % 4096,
	}
	b, err := gf.Encode()
	if err != nil {
		return nil, err
	}
	var raw []byte

	// length of MO global message is length of MO Header plus length of MO payload including 3 bytes of the MO payload (IEI + length)
	raw = append(raw, 0x01, byte(uint16(len(b)+3+len(o.hraw)))>>8, byte(uint16(len(b)+3+len(o.hraw))))
	raw = append(raw, o.hraw...)

	gfLength := uint16(len(b))

	raw = append(raw, 0x02, byte(gfLength>>8), byte(gfLength))
	raw = append(raw, b...)

	return raw, nil
}

func (o *Omnicom) makeBuffReportingInterval(ui, msgId uint32) ([]byte, error) {
	hpr := omnicom.HPR{
		Header: 0x01,
		Msg_ID: msgId,
		Count_Total_Data_Reports:       1,
		Count_Data_Reports_in_this_Msg: 1,
		Source_Flag:                    0x01,
		Data_Report: []omnicom.Data_Report{{
			Date_Position: omnicom.Date_Position{
				Longitude: float32(o.object.curPos.Longitude),
				Latitude:  float32(o.object.curPos.Latitude),
				Day:       uint32(time.Now().Day()),
				Month:     uint32(time.Now().Month()),
				Year:      uint32(time.Now().Year()) - 2000,
				Minute:    uint32(time.Now().Minute() + time.Now().Hour()*60),
			},
			Period: ui,
		}},
	}
	b, err := hpr.Encode()
	if err != nil {
		return nil, err
	}
	var raw []byte

	// length of MO global message is length of MO Header plus length of MO payload including 3 bytes of the MO payload (IEI + length)
	raw = append(raw, 0x01, byte(uint16(len(b)+3+len(o.hraw)))>>8, byte(uint16(len(b)+3+len(o.hraw))))
	raw = append(raw, o.hraw...)

	gfLength := uint16(len(b))

	raw = append(raw, 0x02, byte(gfLength>>8), byte(gfLength))
	raw = append(raw, b...)

	return raw, nil

}

func (o *Omnicom) makeBuffRmh(pt []*PositionSpeedTime, totalLength uint32, msgId uint32) ([]byte, error) {
	hpr := omnicom.HPR{
		Header: 0x01,
		Msg_ID: msgId,
		Count_Total_Data_Reports:       totalLength,
		Count_Data_Reports_in_this_Msg: uint32(len(pt)),
		Source_Flag:                    0x01,
		Beacon_ID:                      o.object.OmnicomId,
		Extention_Bit_Beacon_ID:        1,
	}
	for i := range pt {
		hpr.Data_Report = append(hpr.Data_Report, omnicom.Data_Report{
			Date_Position: omnicom.Date_Position{
				Longitude: float32(pt[i].p.Longitude),
				Latitude:  float32(pt[i].p.Latitude),
				Day:       uint32(pt[i].t.Day()),
				Month:     uint32(pt[i].t.Month()),
				Year:      uint32(pt[i].t.Year()) - 2000,
				Minute:    uint32(pt[i].t.Minute() + time.Now().Hour()*60),
			},
		})
	}
	b, err := hpr.Encode()
	if err != nil {
		return nil, err
	}
	var raw []byte

	// length of MO global message is length of MO Header plus length of MO payload including 3 bytes of the MO payload (IEI + length)
	raw = append(raw, 0x01, byte(uint16(len(b)+3+len(o.hraw)))>>8, byte(uint16(len(b)+3+len(o.hraw))))
	raw = append(raw, o.hraw...)

	gfLength := uint16(len(b))

	raw = append(raw, 0x02, byte(gfLength>>8), byte(gfLength))
	raw = append(raw, b...)

	return raw, nil
}

func (o *Omnicom) makeBuffGlobalParameters(msgId uint32) ([]byte, error) {
	tnow := time.Now()
	gp := omnicom.GP{
		Date_Position: omnicom.Date_Position{
			Longitude: float32(o.object.curPos.Longitude),
			Latitude:  float32(o.object.curPos.Latitude),
			Day:       uint32(tnow.Day()),
			Month:     uint32(tnow.Month()),
			Year:      uint32(tnow.Year()) - 2000,
			Minute:    uint32(tnow.Minute()),
		},
		G3_IMEI:   []byte(o.object.ImeiG3),
		Beacon_ID: o.object.OmnicomId,
		ID_Msg:    msgId,
		Position_Reporting_Interval: omnicom.Position_Reporting_Interval{
			Value_in_mn: 1,
		},
		Header: 0x03,
		Geofencing_Enable: omnicom.Geofencing_Enable{
			On_Off: 1,
		},
		IRI_IMEI: []byte(o.object.Imei),
	}

	log.Debug("Global parameters message to send is %+v", gp)
	b, err := gp.Encode()
	if err != nil {
		return nil, err
	}
	var raw []byte

	raw = append(raw, 0x01, byte(uint16(len(b)+3+len(o.hraw)))>>8, byte(uint16(len(b)+3+len(o.hraw))))
	raw = append(raw, o.hraw...)

	gpLength := uint16(len(b))

	raw = append(raw, 0x02, byte(gpLength>>8), byte(gpLength))
	raw = append(raw, b...)

	return raw, nil
}

func (o *Omnicom) reInitializeCurrentStatus() {
	//set alert current test mode status to 0 after first occurence.
	o.ar.Test_Mode.Current_Test_Mode_Status = 0
	o.ar.Assistance_Alert.Current_Assistance_Alert_Status = 0
	o.ar.Daylight_Alert.Current_Daylight_Alert = 0
	o.ar.Loss_Mobile_Com.Current_Loss_Mobile_Com_Status = 0
	o.ar.JB_Dome_Alert.Current_JB_Dome_Status = 0
	o.ar.No_Position_Fix.Current_No_Position_Fix_Status = 0
	o.ar.Intrusion_Alert.Current_Intrusion_Status = 0
	o.ar.Battery_Alert.Current_Battery_Alert_Status = 0
	o.ar.Power_Down.Power_Down_Status = 0
	o.ar.Power_Up.Power_Up_Status = 0
}

func (o *Omnicom) assignAlertTypeValue(alt uint, mst uint32, cst uint32) error {
	switch alt {
	case PU:
		o.ar.Power_Up = omnicom.Power_Up{Alert_Status: mst, Power_Up_Status: cst}
	case BA:
		o.ar.Battery_Alert = omnicom.Battery_Alert{Alert_Status: mst, Current_Battery_Alert_Status: cst}
	case PD:
		o.ar.Power_Down = omnicom.Power_Down{Alert_Status: mst, Power_Down_Status: cst}
	case IA:
		o.ar.Intrusion_Alert = omnicom.Intrusion_Alert{Alert_Status: mst, Current_Intrusion_Status: cst}
	case NPF:
		o.ar.No_Position_Fix = omnicom.No_Position_Fix{Alert_Status: mst, Current_No_Position_Fix_Status: cst, Number_Satellite_in_View: 5}
	case JBDA:
		o.ar.JB_Dome_Alert = omnicom.JB_Dome_Alert{Alert_Status: mst, Current_JB_Dome_Status: cst}
	case LMC:
		o.ar.Loss_Mobile_Com = omnicom.Loss_Mobile_Com{Alert_Status: mst, Current_Loss_Mobile_Com_Status: cst}
	case DA:
		o.ar.Daylight_Alert = omnicom.Daylight_Alert{Alert_Status: mst, Current_Daylight_Alert: cst}
	case AA:
		o.ar.Assistance_Alert = omnicom.Assistance_Alert{Alert_Status: mst, Current_Assistance_Alert_Status: cst}
	case TM:
		o.ar.Test_Mode = omnicom.Test_Mode{Alert_Status: mst, Current_Test_Mode_Status: cst}
	case LastTypeAlerting:
		break
	default:
		return fmt.Errorf("bad alert type: %+v", alt)
	}
	return nil
}
