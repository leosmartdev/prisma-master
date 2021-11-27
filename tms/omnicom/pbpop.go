package omnicom

import (
	"fmt"
	"reflect"

	"github.com/jinzhu/copier"
)

func PopulateOmnicom(pbOmnicomStructure *Omni) (Omnicom, error) {

	if pbOmnicomStructure == nil {
		return nil, fmt.Errorf("Protobuf structure passed to PopulateOmnicom(pbOmnicomStructure *Omni) is nil")
	}

	if pbOmnicomStructure.GetRmh() != nil {

		RequestHistory := new(RMH)

		RequestHistory.Header = pbOmnicomStructure.GetRmh().Header[0]

		RequestHistory.Date = Date{}

		err := copier.Copy(&RequestHistory.Date, pbOmnicomStructure.GetRmh().Date)

		if err != nil {
			return nil, err
		}

		RequestHistory.Date_Interval = Date_Interval{}

		RequestHistory.Date_Interval.Start = Date{}

		err = copier.Copy(&RequestHistory.Date_Interval.Start, pbOmnicomStructure.GetRmh().Date_Interval.Start)

		if err != nil {
			return nil, err
		}

		RequestHistory.Date_Interval.Stop = Date{}

		err = copier.Copy(&RequestHistory.Date_Interval.Stop, pbOmnicomStructure.GetRmh().Date_Interval.Stop)

		if err != nil {
			return nil, err
		}

		RequestHistory.ID_Msg = pbOmnicomStructure.GetRmh().ID_Msg

		RequestHistory.Padding = pbOmnicomStructure.GetRmh().Padding

		RequestHistory.CRC = pbOmnicomStructure.GetRmh().CRC

		return RequestHistory, nil
	}

	if pbOmnicomStructure.GetUic() != nil {

		UnitIntervalChange := new(UIC)

		UnitIntervalChange.Header = pbOmnicomStructure.GetUic().Header[0]

		UnitIntervalChange.ID_Msg = pbOmnicomStructure.GetUic().ID_Msg

		UnitIntervalChange.Date = Date{}

		err := copier.Copy(&UnitIntervalChange.Date, pbOmnicomStructure.GetUic().Date)

		if err != nil {
			return nil, err
		}

		UnitIntervalChange.New_Reporting = pbOmnicomStructure.GetUic().New_Reporting

		UnitIntervalChange.CRC = pbOmnicomStructure.GetUic().CRC

		return UnitIntervalChange, nil

	}

	if pbOmnicomStructure.GetRsm() != nil {

		RequestSpecificMessage := new(RSM)

		RequestSpecificMessage.Header = pbOmnicomStructure.GetRsm().Header[0]

		RequestSpecificMessage.ID_Msg = pbOmnicomStructure.GetRsm().ID_Msg

		RequestSpecificMessage.Date = Date{}

		err := copier.Copy(&RequestSpecificMessage.Date, pbOmnicomStructure.GetRsm().Date)

		if err != nil {
			return nil, err
		}

		RequestSpecificMessage.Msg_to_Ask = pbOmnicomStructure.GetRsm().MsgTo_Ask

		RequestSpecificMessage.Padding = pbOmnicomStructure.GetRsm().Padding

		RequestSpecificMessage.CRC = pbOmnicomStructure.GetRsm().CRC

		return RequestSpecificMessage, nil

	}

	if pbOmnicomStructure.GetUgp() != nil {

		UploadGlobalParam := new(UGP)

		UploadGlobalParam.Header = pbOmnicomStructure.GetUgp().Header[0]

		UploadGlobalParam.ID_Msg = pbOmnicomStructure.GetUgp().ID_Msg

		UploadGlobalParam.Date = Date{}

		err := copier.Copy(&UploadGlobalParam.Date, pbOmnicomStructure.GetUgp().Date)

		if err != nil {
			return nil, err
		}

		UploadGlobalParam.Position_Reporting_Interval = Position_Reporting_Interval_StoV{}

		err = copier.Copy(&UploadGlobalParam.Position_Reporting_Interval, pbOmnicomStructure.GetUgp().Position_Reporting_Interval)

		if err != nil {
			return nil, err
		}

		UploadGlobalParam.Geofencing_Enable = Geofencing_Enable_StoV{}

		err = copier.Copy(&UploadGlobalParam.Geofencing_Enable, pbOmnicomStructure.GetUgp().Geofencing_Enable)

		if err != nil {
			return nil, err
		}

		UploadGlobalParam.Geofence_Status_Check_Interval = Geofence_Status_Check_Interval_StoV{}

		err = copier.Copy(&UploadGlobalParam.Geofence_Status_Check_Interval, pbOmnicomStructure.GetUgp().Geofence_Status_Check_Interval)

		if err != nil {
			return nil, err
		}

		UploadGlobalParam.Password = Password_StoV{}

		err = copier.Copy(&UploadGlobalParam.Password, pbOmnicomStructure.GetUgp().Password)

		if err != nil {
			return nil, err
		}

		UploadGlobalParam.Routing = Routing_StoV{}

		err = copier.Copy(&UploadGlobalParam.Routing, pbOmnicomStructure.GetUgp().Routing)

		if err != nil {
			return nil, err
		}

		UploadGlobalParam.Padding = pbOmnicomStructure.GetUgp().Padding

		UploadGlobalParam.CRC = pbOmnicomStructure.GetUgp().CRC
	}

	if pbOmnicomStructure.GetUaup() != nil {

		UpdateAPIURLParam := new(UAUP)

		UpdateAPIURLParam.Header = pbOmnicomStructure.GetUaup().Header[0]

		UpdateAPIURLParam.ID_Msg = pbOmnicomStructure.GetUaup().ID_Msg

		UpdateAPIURLParam.Date = Date{}

		err := copier.Copy(&UpdateAPIURLParam.Date, pbOmnicomStructure.GetUaup().Date)

		if err != nil {
			return nil, err
		}

		UpdateAPIURLParam.Web_Service_API_URL_Sending = Web_Service_API_URL_sending_StoV{}

		err = copier.Copy(&UpdateAPIURLParam.Web_Service_API_URL_Sending, pbOmnicomStructure.GetUaup().Web_Service_API_URL_Sending)

		if err != nil {
			return nil, err
		}

		UpdateAPIURLParam.Web_Service_API_URL_Receiving = Web_Service_API_URL_Receiving_StoV{}

		err = copier.Copy(&UpdateAPIURLParam.Web_Service_API_URL_Receiving, pbOmnicomStructure.GetUaup().Web_Service_API_URL_Receiving)

		if err != nil {
			return nil, err
		}

		UpdateAPIURLParam.Array = Array_StoV{}

		err = copier.Copy(&UpdateAPIURLParam.Array, pbOmnicomStructure.GetUaup().Array)

		if err != nil {
			return nil, err
		}

		UpdateAPIURLParam.Padding = pbOmnicomStructure.GetUaup().Padding

		UpdateAPIURLParam.CRC = pbOmnicomStructure.GetUaup().CRC

		return UpdateAPIURLParam, nil

	}

	if pbOmnicomStructure.GetUgcircle() != nil {
		circle := pbOmnicomStructure.GetUgcircle()
		gf := new(UG_Circle)
		if len(circle.Header) > 0 {
			gf.Header = circle.Header[0]
		}
		gf.Msg_ID = circle.Msg_ID
		gf.Date = Date{}
		err := copier.Copy(&gf.Date, circle.Date)
		if err != nil {
			return nil, err
		}
		gf.GEO_ID = circle.GEO_ID
		gf.Shape = circle.Shape
		gf.NAME = circle.NAME
		gf.TYPE = circle.TYPE
		gf.Priority = circle.Priority
		gf.Activated = circle.Activated
		gf.Setting = Setting{}
		err = copier.Copy(&gf.Setting, circle.Setting)
		if err != nil {
			return nil, err
		}
		gf.Number_Point = circle.Number_Point
		gf.Position = Position_Radius{}
		err = copier.Copy(&gf.Position, circle.Position)
		if err != nil {
			return nil, err
		}
		gf.Padding = circle.Padding
		gf.CRC = circle.CRC
		return gf, nil
	}

	if pbOmnicomStructure.GetDg() != nil {
		pdg := pbOmnicomStructure.GetDg()
		dg := new(DG)
		if len(pdg.Header) > 0 {
			dg.Header = pdg.Header[0]
		}
		dg.Msg_ID = pdg.Msg_ID
		dg.Date = Date{}
		err := copier.Copy(&dg.Date, pdg.Date)
		if err != nil {
			return nil, err
		}
		dg.GEO_ID = pdg.GEO_ID
		dg.Padding = pdg.Padding
		dg.CRC = pdg.CRC
		return dg, nil

	}

	if pbOmnicomStructure.GetUgpolygon() != nil {
		gf := new(UG_Polygon)
		if len(pbOmnicomStructure.GetUgpolygon().Header) > 0 {
			gf.Header = pbOmnicomStructure.GetUgpolygon().Header[0]
		}
		gf.Msg_ID = pbOmnicomStructure.GetUgpolygon().Msg_ID
		gf.Date = Date{}
		err := copier.Copy(&gf.Date, pbOmnicomStructure.GetUgpolygon().Date)
		if err != nil {
			return nil, err
		}
		gf.GEO_ID = pbOmnicomStructure.GetUgpolygon().GEO_ID
		gf.Shape = pbOmnicomStructure.GetUgpolygon().Shape
		gf.NAME = pbOmnicomStructure.GetUgpolygon().NAME
		gf.TYPE = pbOmnicomStructure.GetUgpolygon().TYPE
		gf.Priority = pbOmnicomStructure.GetUgpolygon().Priority
		gf.Activated = pbOmnicomStructure.GetUgpolygon().Activated
		gf.Setting = Setting{}
		err = copier.Copy(&gf.Setting, pbOmnicomStructure.GetUgpolygon().Setting)
		if err != nil {
			return nil, err
		}
		gf.Number_Point = pbOmnicomStructure.GetUgpolygon().Number_Point
		for i := 0; i < len(pbOmnicomStructure.GetUgpolygon().Position); i++ {
			gf.Position = append(gf.Position, Position{})
			err = copier.Copy(&gf.Position[i], pbOmnicomStructure.GetUgpolygon().Position[i])
			if err != nil {
				return nil, err
			}

		}

		gf.Padding = pbOmnicomStructure.GetUgpolygon().Padding
		gf.CRC = pbOmnicomStructure.GetUgpolygon().CRC

		return gf, nil
	}

	if pbOmnicomStructure.GetAa() != nil {

		AlertAssistance := new(AA)

		AlertAssistance.Header = pbOmnicomStructure.GetAa().Header[0]

		AlertAssistance.Date = Date{}

		err := copier.Copy(&AlertAssistance.Date, pbOmnicomStructure.GetAa().Date)

		if err != nil {
			return nil, err
		}

		AlertAssistance.Padding = pbOmnicomStructure.GetAa().Padding

		AlertAssistance.CRC = pbOmnicomStructure.GetAa().CRC

		return AlertAssistance, nil
	}

	if pbOmnicomStructure.GetTma() != nil {

		TestModeAck := new(TMA)

		TestModeAck.Header = pbOmnicomStructure.GetTma().Header[0]

		TestModeAck.Date = Date{}

		err := copier.Copy(&TestModeAck.Date, pbOmnicomStructure.GetTma().Date)

		if err != nil {
			return nil, err
		}

		return TestModeAck, nil
	}

	return nil, fmt.Errorf("Sentence type '%s' not implemented", reflect.TypeOf(pbOmnicomStructure).Elem().String())
}

func PopulateProtobuf(sentence Omnicom) (*Omni, error) {

	sen := &Omni{}

	if sentence == nil {

		return nil, fmt.Errorf("Omnicom sentence structure passed to func PopulateProtobuf(sentence Omnicom) is nil")

	}

	if reflect.TypeOf(sentence).Elem().String() == "omnicom.SPM" {

		senstruct := sentence.(*SPM)

		pbOmnicomStructure := Spm{}

		pbOmnicomStructure.Header = append(pbOmnicomStructure.Header, senstruct.Header)

		pbOmnicomStructure.Split_Msg_ID = senstruct.Split_Msg_ID

		pbOmnicomStructure.Packet_Number = senstruct.Packet_Number

		pbOmnicomStructure.Packets_Total_Count = senstruct.Packets_Total_Count

		pbOmnicomStructure.Length_Msg_Data_PartIn_Byte = senstruct.Length_Msg_Data_Part_in_Byte

		pbOmnicomStructure.Padding1 = senstruct.Padding1

		pbOmnicomStructure.Msg_Data_Part = senstruct.Msg_Data_Part

		pbOmnicomStructure.CRC = senstruct.CRC

		sen.Omnicom = &Omni_Spm{&pbOmnicomStructure}

		return sen, nil

	}

	if reflect.TypeOf(sentence).Elem().String() == "omnicom.SDR" {

		senstruct := sentence.(*SDR) //(*nmea.ABK)

		pbOmnicomStructure := Sdr{}

		pbOmnicomStructure.Header = append(pbOmnicomStructure.Header, senstruct.Header)

		pbOmnicomStructure.Date = &Dt{}

		err := copier.Copy(pbOmnicomStructure.Date, &senstruct.Date)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Split_Msg_ID = senstruct.Split_Msg_ID

		pbOmnicomStructure.Padding = senstruct.Padding

		pbOmnicomStructure.CRC = senstruct.CRC

		sen.Omnicom = &Omni_Sdr{&pbOmnicomStructure}

		return sen, nil

	}

	if reflect.TypeOf(sentence).Elem().String() == "omnicom.SMDR" {

		senstruct := sentence.(*SMDR)

		pbOmnicomStructure := Smdr{}

		pbOmnicomStructure.Header = append(pbOmnicomStructure.Header, senstruct.Header)

		pbOmnicomStructure.Split_Msg_ID = senstruct.Split_Msg_ID

		pbOmnicomStructure.Packets_Expected_Total_Count = senstruct.Packets_Expected_Total_Count

		pbOmnicomStructure.Missing_Packets_Total_Count = senstruct.Missing_Packets_Total_Count

		pbOmnicomStructure.CRC = senstruct.CRC

		sen.Omnicom = &Omni_Smdr{&pbOmnicomStructure}

		return sen, nil

	}

	if reflect.TypeOf(sentence).Elem().String() == "omnicom.UF" {

		senstruct := sentence.(*UF) //(*nmea.ABK)

		pbOmnicomStructure := Uf{}

		pbOmnicomStructure.Header = append(pbOmnicomStructure.Header, senstruct.Header)

		pbOmnicomStructure.Date = &Dt{}

		err := copier.Copy(pbOmnicomStructure.Date, &senstruct.Date)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.ID_Msg = senstruct.ID_Msg

		pbOmnicomStructure.Firmware_Target = senstruct.Firmware_Target

		pbOmnicomStructure.Flag_Last_Packet = senstruct.Flag_Last_Packet

		pbOmnicomStructure.Data_Address = senstruct.Data_Address

		pbOmnicomStructure.Data_Size = senstruct.Data_Size

		pbOmnicomStructure.Padding1 = senstruct.Padding1

		pbOmnicomStructure.Firmware_Data = senstruct.Firmware_Data

		pbOmnicomStructure.CRC = senstruct.CRC

		sen.Omnicom = &Omni_Uf{&pbOmnicomStructure}

		return sen, nil

	}

	if reflect.TypeOf(sentence).Elem().String() == "omnicom.BM_StoV" {

		senstruct := sentence.(*BM_StoV)

		pbOmnicomStructure := BMStoV{}

		pbOmnicomStructure.Header = append(pbOmnicomStructure.Header, senstruct.Header)

		pbOmnicomStructure.Date = &Dt{}

		err := copier.Copy(pbOmnicomStructure.Date, &senstruct.Date)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.ID_Msg = senstruct.ID_Msg

		pbOmnicomStructure.Length_Msg_Content = senstruct.Length_Msg_Content

		pbOmnicomStructure.Msg_Content = senstruct.Msg_Content

		pbOmnicomStructure.CRC = senstruct.CRC

		sen.Omnicom = &Omni_Bmstov{&pbOmnicomStructure}

		return sen, nil

	}

	if reflect.TypeOf(sentence).Elem().String() == "omnicom.DG" {

		senstruct := sentence.(*DG) //(*nmea.ABK)

		pbOmnicomStructure := Dg{}

		pbOmnicomStructure.Header = append(pbOmnicomStructure.Header, senstruct.Header)

		pbOmnicomStructure.Msg_ID = senstruct.Msg_ID

		pbOmnicomStructure.Date = &Dt{}

		err := copier.Copy(pbOmnicomStructure.Date, &senstruct.Date)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.GEO_ID = senstruct.GEO_ID

		pbOmnicomStructure.Padding = senstruct.Padding

		pbOmnicomStructure.CRC = senstruct.CRC

		sen.Omnicom = &Omni_Dg{&pbOmnicomStructure}

		return sen, nil

	}

	if reflect.TypeOf(sentence).Elem().String() == "omnicom.GBMN" {

		senstruct := sentence.(*GBMN) //(*nmea.ABK)

		pbOmnicomStructure := Gbmn{}

		pbOmnicomStructure.Header = append(pbOmnicomStructure.Header, senstruct.Header)

		pbOmnicomStructure.Date = &Dt{}

		err := copier.Copy(pbOmnicomStructure.Date, &senstruct.Date)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Padding = senstruct.Padding

		pbOmnicomStructure.CRC = senstruct.CRC

		sen.Omnicom = &Omni_Gbmn{&pbOmnicomStructure}

		return sen, nil

	}

	if reflect.TypeOf(sentence).Elem().String() == "omnicom.UG_Circle" {

		senstruct := sentence.(*UG_Circle) //(*nmea.ABK)

		pbOmnicomStructure := UGCircle{}

		pbOmnicomStructure.Header = append(pbOmnicomStructure.Header, senstruct.Header)

		pbOmnicomStructure.Msg_ID = senstruct.Msg_ID

		pbOmnicomStructure.Date = &Dt{}

		err := copier.Copy(pbOmnicomStructure.Date, &senstruct.Date)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.GEO_ID = senstruct.GEO_ID

		pbOmnicomStructure.Shape = senstruct.Shape

		pbOmnicomStructure.NAME = senstruct.NAME

		pbOmnicomStructure.TYPE = senstruct.TYPE

		pbOmnicomStructure.Priority = senstruct.Priority

		pbOmnicomStructure.Activated = senstruct.Activated

		pbOmnicomStructure.Setting = &Stg{}

		err = copier.Copy(pbOmnicomStructure.Setting, &senstruct.Setting)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Number_Point = senstruct.Number_Point

		pbOmnicomStructure.Position = &PositionRadius{}

		err = copier.Copy(pbOmnicomStructure.Position, &senstruct.Position)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Padding = senstruct.Padding

		pbOmnicomStructure.CRC = senstruct.CRC

		sen.Omnicom = &Omni_Ugcircle{&pbOmnicomStructure}

		return sen, nil

	}

	if reflect.TypeOf(sentence).Elem().String() == "omnicom.UG_Polygon" {

		senstruct := sentence.(*UG_Polygon) //(*nmea.ABK)

		pbOmnicomStructure := UGPolygon{}

		pbOmnicomStructure.Header = append(pbOmnicomStructure.Header, senstruct.Header)

		pbOmnicomStructure.Msg_ID = senstruct.Msg_ID

		pbOmnicomStructure.Date = &Dt{}

		err := copier.Copy(pbOmnicomStructure.Date, &senstruct.Date)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.GEO_ID = senstruct.GEO_ID

		pbOmnicomStructure.Shape = senstruct.Shape

		pbOmnicomStructure.NAME = senstruct.NAME

		pbOmnicomStructure.TYPE = senstruct.TYPE

		pbOmnicomStructure.Priority = senstruct.Priority

		pbOmnicomStructure.Activated = senstruct.Activated

		pbOmnicomStructure.Setting = &Stg{}

		err = copier.Copy(pbOmnicomStructure.Setting, &senstruct.Setting)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Number_Point = senstruct.Number_Point

		for i := 0; i < len(senstruct.Position); i++ {

			pbOmnicomStructure.Position = append(pbOmnicomStructure.Position, &Pos{})

			err = copier.Copy(pbOmnicomStructure.Position[i], &senstruct.Position[i])

			if err != nil {
				return nil, err
			}

		}

		pbOmnicomStructure.Padding = senstruct.Padding

		pbOmnicomStructure.CRC = senstruct.CRC

		sen.Omnicom = &Omni_Ugpolygon{&pbOmnicomStructure}

		return sen, nil

	}

	if reflect.TypeOf(sentence).Elem().String() == "omnicom.UAUP" {

		senstruct := sentence.(*UAUP) //(*nmea.ABK)

		pbOmnicomStructure := Uaup{}

		pbOmnicomStructure.Header = append(pbOmnicomStructure.Header, senstruct.Header)

		pbOmnicomStructure.ID_Msg = senstruct.ID_Msg

		pbOmnicomStructure.Date = &Dt{}

		err := copier.Copy(pbOmnicomStructure.Date, &senstruct.Date)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Web_Service_API_URL_Sending = &WebServiceAPIURLsendingStoV{}

		err = copier.Copy(pbOmnicomStructure.Web_Service_API_URL_Sending, &senstruct.Web_Service_API_URL_Sending)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Web_Service_API_URL_Receiving = &WebServiceAPIURLReceivingStoV{}

		err = copier.Copy(pbOmnicomStructure.Web_Service_API_URL_Receiving, &senstruct.Web_Service_API_URL_Receiving)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Array = &ArrayStoV{}

		err = copier.Copy(pbOmnicomStructure.Array, &senstruct.Array)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Padding = senstruct.Padding

		pbOmnicomStructure.CRC = senstruct.CRC

		sen.Omnicom = &Omni_Uaup{&pbOmnicomStructure}

		return sen, nil

	}

	if reflect.TypeOf(sentence).Elem().String() == "omnicom.UGP" {

		senstruct := sentence.(*UGP) //(*nmea.ABK)

		pbOmnicomStructure := Ugp{}

		pbOmnicomStructure.Header = append(pbOmnicomStructure.Header, senstruct.Header)

		pbOmnicomStructure.ID_Msg = senstruct.ID_Msg

		pbOmnicomStructure.Date = &Dt{}

		err := copier.Copy(pbOmnicomStructure.Date, &senstruct.Date)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Position_Reporting_Interval = &PositionReportingIntervalStoV{}

		err = copier.Copy(pbOmnicomStructure.Position_Reporting_Interval, &senstruct.Position_Reporting_Interval)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Geofencing_Enable = &GeofencingEnableStoV{}

		err = copier.Copy(pbOmnicomStructure.Geofencing_Enable, &senstruct.Geofencing_Enable)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Geofence_Status_Check_Interval = &GeofenceStatusCheckIntervalStoV{}

		err = copier.Copy(pbOmnicomStructure.Geofence_Status_Check_Interval, &senstruct.Geofence_Status_Check_Interval)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Password = &PasswordStoV{}

		err = copier.Copy(pbOmnicomStructure.Password, &senstruct.Password)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Routing = &RoutingStoV{}

		err = copier.Copy(pbOmnicomStructure.Routing, &senstruct.Routing)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Padding = senstruct.Padding

		pbOmnicomStructure.CRC = senstruct.CRC

		sen.Omnicom = &Omni_Ugp{&pbOmnicomStructure}

		return sen, nil

	}

	if reflect.TypeOf(sentence).Elem().String() == "omnicom.RSM" {

		senstruct := sentence.(*RSM) //(*nmea.ABK)

		pbOmnicomStructure := Rsm{}

		pbOmnicomStructure.Header = append(pbOmnicomStructure.Header, senstruct.Header)

		pbOmnicomStructure.ID_Msg = senstruct.ID_Msg

		pbOmnicomStructure.Date = &Dt{}

		err := copier.Copy(pbOmnicomStructure.Date, &senstruct.Date)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.MsgTo_Ask = senstruct.Msg_to_Ask

		pbOmnicomStructure.Padding = senstruct.Padding

		pbOmnicomStructure.CRC = senstruct.CRC

		sen.Omnicom = &Omni_Rsm{&pbOmnicomStructure}

		return sen, nil

	}

	if reflect.TypeOf(sentence).Elem().String() == "omnicom.UIC" {

		senstruct := sentence.(*UIC) //(*nmea.ABK)

		pbOmnicomStructure := Uic{}

		pbOmnicomStructure.Header = append(pbOmnicomStructure.Header, senstruct.Header)

		pbOmnicomStructure.ID_Msg = senstruct.ID_Msg

		pbOmnicomStructure.Date = &Dt{}

		err := copier.Copy(pbOmnicomStructure.Date, &senstruct.Date)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.New_Reporting = senstruct.New_Reporting

		pbOmnicomStructure.CRC = senstruct.CRC

		sen.Omnicom = &Omni_Uic{&pbOmnicomStructure}

		return sen, nil

	}

	if reflect.TypeOf(sentence).Elem().String() == "omnicom.RMH" {

		senstruct := sentence.(*RMH) //(*nmea.ABK)

		pbOmnicomStructure := Rmh{}

		pbOmnicomStructure.Header = append(pbOmnicomStructure.Header, senstruct.Header)

		pbOmnicomStructure.Date = &Dt{}

		err := copier.Copy(pbOmnicomStructure.Date, &senstruct.Date)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Date_Interval = &DateInterval{}

		err = copier.Copy(pbOmnicomStructure.Date_Interval, &senstruct.Date_Interval)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.ID_Msg = senstruct.ID_Msg

		pbOmnicomStructure.Padding = senstruct.Padding

		pbOmnicomStructure.CRC = senstruct.CRC

		sen.Omnicom = &Omni_Rmh{&pbOmnicomStructure}

		return sen, nil

	}

	if reflect.TypeOf(sentence).Elem().String() == "omnicom.AA" {

		senstruct := sentence.(*AA) //(*nmea.ABK)

		pbOmnicomStructure := Aa{}

		pbOmnicomStructure.Header = append(pbOmnicomStructure.Header, senstruct.Header)

		pbOmnicomStructure.Date = &Dt{}

		err := copier.Copy(pbOmnicomStructure.Date, &senstruct.Date)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Padding = senstruct.Padding

		pbOmnicomStructure.CRC = senstruct.CRC

		sen.Omnicom = &Omni_Aa{&pbOmnicomStructure}

		return sen, nil

	}

	// what is the difference between TMA and AA
	if reflect.TypeOf(sentence).Elem().String() == "omnicom.TMA" {

		senstruct := sentence.(*TMA) //(*nmea.ABK)

		pbOmnicomStructure := Tma{}

		pbOmnicomStructure.Header = append(pbOmnicomStructure.Header, senstruct.Header)

		pbOmnicomStructure.Date = &Dt{}

		err := copier.Copy(pbOmnicomStructure.Date, &senstruct.Date)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Padding = senstruct.Padding

		pbOmnicomStructure.CRC = senstruct.CRC

		sen.Omnicom = &Omni_Tma{&pbOmnicomStructure}

		return sen, nil

	}

	if reflect.TypeOf(sentence).Elem().String() == "omnicom.ABM" {

		senstruct := sentence.(*ABM) //(*nmea.ABK)

		pbOmnicomStructure := Abm{}

		pbOmnicomStructure.Header = append(pbOmnicomStructure.Header, senstruct.Header)

		pbOmnicomStructure.Date = &Dt{}

		err := copier.Copy(pbOmnicomStructure.Date, &senstruct.Date)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.ID_Msg = senstruct.ID_Msg

		pbOmnicomStructure.Error_Type = senstruct.Error_Type

		pbOmnicomStructure.Padding = senstruct.Padding

		pbOmnicomStructure.CRC = senstruct.CRC

		sen.Omnicom = &Omni_Abm{&pbOmnicomStructure}

		return sen, nil

	}

	if reflect.TypeOf(sentence).Elem().String() == "omnicom.BM" {

		senstruct := sentence.(*BM)

		pbOmnicomStructure := Bm{}

		pbOmnicomStructure.Header = append(pbOmnicomStructure.Header, senstruct.Header)

		pbOmnicomStructure.Date = &Dt{}

		err := copier.Copy(pbOmnicomStructure.Date, &senstruct.Date)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.ID_Msg = senstruct.ID_Msg

		pbOmnicomStructure.Length_Msg_Content = senstruct.Length_Msg_Content

		pbOmnicomStructure.Msg_Content = senstruct.Msg_Content

		pbOmnicomStructure.CRC = senstruct.CRC

		sen.Omnicom = &Omni_Bm{&pbOmnicomStructure}

		return sen, nil

	}

	if reflect.TypeOf(sentence).Elem().String() == "omnicom.GA" {

		senstruct := sentence.(*GA) //(*nmea.ABK)

		pbOmnicomStructure := Ga{}

		pbOmnicomStructure.Header = append(pbOmnicomStructure.Header, senstruct.Header)

		pbOmnicomStructure.Msg_ID = senstruct.Msg_ID

		pbOmnicomStructure.Date_Position = &DatePosition{}

		err := copier.Copy(pbOmnicomStructure.Date_Position, &senstruct.Date_Position)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Error_Type = senstruct.Error_Type

		pbOmnicomStructure.GEO_ID = senstruct.GEO_ID

		pbOmnicomStructure.Padding = senstruct.Padding

		pbOmnicomStructure.CRC = senstruct.CRC

		sen.Omnicom = &Omni_Ga{&pbOmnicomStructure}

		return sen, nil

	}

	if reflect.TypeOf(sentence).Elem().String() == "omnicom.AUP" {

		senstruct := sentence.(*AUP) //(*nmea.ABK)

		pbOmnicomStructure := Aup{}

		pbOmnicomStructure.Header = append(pbOmnicomStructure.Header, senstruct.Header)

		pbOmnicomStructure.ID_Msg = senstruct.ID_Msg

		pbOmnicomStructure.Date_Position = &DatePosition{}

		err := copier.Copy(pbOmnicomStructure.Date_Position, &senstruct.Date_Position)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Web_Service_API_URL_Sending = &WebServiceAPIURLSending{}

		err = copier.Copy(pbOmnicomStructure.Web_Service_API_URL_Sending, &senstruct.Web_Service_API_URL_Sending)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Web_Service_API_URL_Receiving = &WebServiceAPIURLReceiving{}

		err = copier.Copy(pbOmnicomStructure.Web_Service_API_URL_Receiving, &senstruct.Web_Service_API_URL_Receiving)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Array = &Arr{}

		err = copier.Copy(pbOmnicomStructure.Array, &senstruct.Array)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Padding = senstruct.Padding

		pbOmnicomStructure.CRC = senstruct.CRC

		sen.Omnicom = &Omni_Aup{&pbOmnicomStructure}

		return sen, nil

	}

	if reflect.TypeOf(sentence).Elem().String() == "omnicom.GP" {

		senstruct := sentence.(*GP) //(*nmea.ABK)

		pbOmnicomStructure := Gp{}

		pbOmnicomStructure.Header = append(pbOmnicomStructure.Header, senstruct.Header)

		pbOmnicomStructure.Beacon_ID = senstruct.Beacon_ID

		pbOmnicomStructure.ID_Msg = senstruct.ID_Msg

		pbOmnicomStructure.Date_Position = &DatePosition{}

		err := copier.Copy(pbOmnicomStructure.Date_Position, &senstruct.Date_Position)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Position_Reporting_Interval = &PositionReportingInterval{
			ValueInMn: senstruct.Position_Reporting_Interval.Value_in_mn,
			Modified:  senstruct.Position_Reporting_Interval.Modified,
		}

		pbOmnicomStructure.Geofencing_Enable = &GeofencingEnable{}

		err = copier.Copy(pbOmnicomStructure.Geofencing_Enable, &senstruct.Geofencing_Enable)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Position_Collection_Interval = &PositionCollectionInterval{
			ValueInMn: uint64(senstruct.Position_Collection_Interval.Value_in_mn),
			Modified:  uint64(senstruct.Position_Collection_Interval.Modified),
		}

		pbOmnicomStructure.Password = &Pwd{
			ValueInMn: senstruct.Password.Value_in_mn,
			Modified:  senstruct.Password.Modified,
		}

		pbOmnicomStructure.Routing = &Rtg{
			ValueInMn: senstruct.Routing.Value_in_mn,
			Modified:  senstruct.Routing.Modified,
		}

		pbOmnicomStructure.Firmware_Dome_Version = senstruct.Firmware_Dome_Version

		pbOmnicomStructure.Junction_Box_Version = append(pbOmnicomStructure.Junction_Box_Version, senstruct.Junction_Box_Version...)

		pbOmnicomStructure.SIM_Card_ICCID = senstruct.SIM_Card_ICCID

		pbOmnicomStructure.G3_IMEI = senstruct.G3_IMEI

		pbOmnicomStructure.IRI_IMEI = senstruct.IRI_IMEI

		pbOmnicomStructure.CRC = senstruct.CRC

		sen.Omnicom = &Omni_Gp{&pbOmnicomStructure}

		return sen, nil

	}

	if reflect.TypeOf(sentence).Elem().String() == "omnicom.AR" {

		senstruct := sentence.(*AR) //(*nmea.ABK)

		pbOmnicomStructure := Ar{}

		pbOmnicomStructure.Header = append(pbOmnicomStructure.Header, senstruct.Header)

		pbOmnicomStructure.Msg_ID = senstruct.Msg_ID

		pbOmnicomStructure.Date_Position = &DatePosition{}

		err := copier.Copy(pbOmnicomStructure.Date_Position, &senstruct.Date_Position)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Date_Event = &DateEvent{}

		err = copier.Copy(pbOmnicomStructure.Date_Event, &senstruct.Date_Event)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Power_Up = &PowerUp{}

		err = copier.Copy(pbOmnicomStructure.Power_Up, &senstruct.Power_Up)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Power_Down = &PowerDown{}

		err = copier.Copy(pbOmnicomStructure.Power_Down, &senstruct.Power_Down)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Battery_Alert = &BatteryAlert{}

		err = copier.Copy(pbOmnicomStructure.Battery_Alert, &senstruct.Battery_Alert)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Intrusion_Alert = &IntrusionAlert{}

		err = copier.Copy(pbOmnicomStructure.Intrusion_Alert, &senstruct.Intrusion_Alert)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.No_Position_Fix = &NoPositionFix{}

		err = copier.Copy(pbOmnicomStructure.No_Position_Fix, &senstruct.No_Position_Fix)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.JB_Dome_Alert = &JBDomeAlert{}

		err = copier.Copy(pbOmnicomStructure.JB_Dome_Alert, &senstruct.JB_Dome_Alert)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Loss_Mobile_Com = &LossMobileCom{}

		err = copier.Copy(pbOmnicomStructure.Loss_Mobile_Com, &senstruct.Loss_Mobile_Com)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Daylight_Alert = &DaylightAlert{}

		err = copier.Copy(pbOmnicomStructure.Daylight_Alert, &senstruct.Daylight_Alert)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Assistance_Alert = &AssistanceAlert{}

		err = copier.Copy(pbOmnicomStructure.Assistance_Alert, &senstruct.Assistance_Alert)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Test_Mode = &TestMode{}

		err = copier.Copy(pbOmnicomStructure.Test_Mode, &senstruct.Test_Mode)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Extention_Bit_Beacon_ID = senstruct.Extention_Bit_Beacon_ID

		pbOmnicomStructure.Beacon_ID = senstruct.Beacon_ID

		pbOmnicomStructure.Extention_Bit_Move = senstruct.Extention_Bit_Move

		pbOmnicomStructure.Move = &MV{}

		err = copier.Copy(pbOmnicomStructure.Move, &senstruct.Move)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Padding = senstruct.Padding

		pbOmnicomStructure.CRC = senstruct.CRC

		sen.Omnicom = &Omni_Ar{&pbOmnicomStructure}

		return sen, nil

	}

	if reflect.TypeOf(sentence).Elem().String() == "omnicom.HPR" {

		senstruct := sentence.(*HPR) //(*nmea.HPR)

		pbOmnicomStructure := Hpr{}

		pbOmnicomStructure.Header = append(pbOmnicomStructure.Header, senstruct.Header)

		pbOmnicomStructure.Msg_ID = senstruct.Msg_ID

		pbOmnicomStructure.Source_Flag = senstruct.Source_Flag

		pbOmnicomStructure.Count_Total_Data_Reports = senstruct.Count_Total_Data_Reports

		pbOmnicomStructure.Count_Data_ReportsInThis_Msg = senstruct.Count_Data_Reports_in_this_Msg

		for i := 0; i < len(senstruct.Data_Report); i++ {

			pbOmnicomStructure.Data_Report = append(pbOmnicomStructure.Data_Report, &DataReport{})

			pbOmnicomStructure.Data_Report[i].Number_Data_Report = senstruct.Data_Report[i].Number_Data_Report

			pbOmnicomStructure.Data_Report[i].Date_Position = &DatePosition{}

			err := copier.Copy(pbOmnicomStructure.Data_Report[i].Date_Position, &senstruct.Data_Report[i].Date_Position)

			if err != nil {
				return nil, err
			}

			pbOmnicomStructure.Data_Report[i].Move = &MV{}

			err = copier.Copy(pbOmnicomStructure.Data_Report[i].Move, &senstruct.Data_Report[i].Move)

			if err != nil {
				return nil, err
			}
			pbOmnicomStructure.Data_Report[i].Period = senstruct.Data_Report[i].Period

			pbOmnicomStructure.Data_Report[i].Voltage = &Vltg{}

			err = copier.Copy(pbOmnicomStructure.Data_Report[i].Voltage, &senstruct.Data_Report[i].Voltage)

			if err != nil {
				return nil, err
			}

			pbOmnicomStructure.Data_Report[i].Geofence = &GF{}

			err = copier.Copy(pbOmnicomStructure.Data_Report[i].Geofence, &senstruct.Data_Report[i].Geofence)

			if err != nil {
				return nil, err
			}

		}

		pbOmnicomStructure.Extention_Bit_Beacon_ID = senstruct.Extention_Bit_Beacon_ID

		pbOmnicomStructure.Beacon_ID = senstruct.Beacon_ID

		pbOmnicomStructure.Padding = senstruct.Padding

		pbOmnicomStructure.CRC = senstruct.CRC

		sen.Omnicom = &Omni_Hpr{&pbOmnicomStructure}

		return sen, nil

	}

	if reflect.TypeOf(sentence).Elem().String() == "omnicom.SPRS" {

		senstruct := sentence.(*SPRS)

		pbOmnicomStructure := Sprs{}

		pbOmnicomStructure.Header = append(pbOmnicomStructure.Header, senstruct.Header)

		pbOmnicomStructure.Date_Position = &DatePosition{}

		err := copier.Copy(pbOmnicomStructure.Date_Position, &senstruct.Date_Position)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Move = &MV{}

		err = copier.Copy(pbOmnicomStructure.Move, &senstruct.Move)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Period = senstruct.Period

		pbOmnicomStructure.BatteryVoltage = senstruct.BatteryVoltage

		pbOmnicomStructure.BatteryCapacity = senstruct.BatteryCapacity

		pbOmnicomStructure.SolarPanelVoltage = senstruct.SolarPanelVoltage

		pbOmnicomStructure.Padding = senstruct.Padding

		pbOmnicomStructure.CRC = senstruct.CRC

		sen.Omnicom = &Omni_Sprs{&pbOmnicomStructure}

		return sen, nil
	}

	if reflect.TypeOf(sentence).Elem().String() == "omnicom.SPR" {

		senstruct := sentence.(*SPR)

		pbOmnicomStructure := Spr{}

		pbOmnicomStructure.Header = append(pbOmnicomStructure.Header, senstruct.Header)

		pbOmnicomStructure.Date_Position = &DatePosition{}

		err := copier.Copy(pbOmnicomStructure.Date_Position, &senstruct.Date_Position)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Move = &MV{}

		err = copier.Copy(pbOmnicomStructure.Move, &senstruct.Move)

		if err != nil {
			return nil, err
		}

		pbOmnicomStructure.Period = senstruct.Period

		pbOmnicomStructure.Extention_Bit_Beacon_ID = senstruct.Extention_Bit_Beacon_ID

		pbOmnicomStructure.Beacon_ID = senstruct.Beacon_ID

		pbOmnicomStructure.Padding = senstruct.Padding

		pbOmnicomStructure.CRC = senstruct.CRC

		sen.Omnicom = &Omni_Spr{&pbOmnicomStructure}

		return sen, nil
	}

	err := fmt.Errorf("Sentence type '%s' not implemented", reflect.TypeOf(sentence).Elem().String())

	return nil, err
}
