package libnaf

import (
	"prisma/tms"
	"prisma/tms/iridium"
	"prisma/tms/omnicom"
	"reflect"
	"testing"

	google_protobuf1 "github.com/golang/protobuf/ptypes/wrappers"
)

func toFixedTest(t *testing.T) {
	var position float32 = -4.05123
	precision := 4
	expectedPosition := -4.0512
	newPosition := toFixed(position, precision)
	if expectedPosition != newPosition {
		t.Errorf("toFixed helper function failed to convert %+v to %+v: newpos = %+v", position, expectedPosition, newPosition)
	}
}

func TestNAF_Subtest(t *testing.T) {

	Date := &omnicom.Dt{Year: 17, Month: 07, Day: 10, Minute: 500}
	DatePosition := &omnicom.DatePosition{Year: 2016, Month: 03, Day: 01, Minute: 578, Longitude: 3.56123, Latitude: -4.05123}
	DatePositionAR := *DatePosition
	DatePositionGP := *DatePosition
	MV := &omnicom.MV{Speed: 20, Heading: 30}
	DateInterval := &omnicom.DateInterval{Start: Date, Stop: Date}
	DateEvent := &omnicom.DateEvent{Year: 17, Month: 07, Day: 10, Minute: 500}
	PowerUp := &omnicom.PowerUp{Power_Up_Status: 1}
	PowerDown := &omnicom.PowerDown{}
	BatteryAlert := &omnicom.BatteryAlert{}
	IntrusionAlert := &omnicom.IntrusionAlert{}
	NoPositionFix := &omnicom.NoPositionFix{Number_SatelliteIn_View: 10}
	JBDomeAlert := &omnicom.JBDomeAlert{}
	LossMobileCom := &omnicom.LossMobileCom{}
	DaylightAlert := &omnicom.DaylightAlert{}
	AssistanceAlert := &omnicom.AssistanceAlert{}
	TestMode := &omnicom.TestMode{Current_Test_Mode_Status: 18}
	str := "oroliabeacon.com"
	URL := []byte(str)
	URLSending := &omnicom.WebServiceAPIURLSending{Value: URL, Modified: 1}
	URLRecieving := &omnicom.WebServiceAPIURLReceiving{Value: URL, Modified: 1}
	URLSendStoV := &omnicom.WebServiceAPIURLsendingStoV{Value: URL, To_Modify: 1}
	URLRecStoV := &omnicom.WebServiceAPIURLReceivingStoV{Value: URL, To_Modify: 1}
	Array := &omnicom.Arr{Value: []byte{0x08}}
	ArrayStoV := &omnicom.ArrayStoV{Value: []byte{0x08}}
	PRI := &omnicom.PositionReportingInterval{}
	GE := &omnicom.GeofencingEnable{}
	PCI := &omnicom.PositionCollectionInterval{}
	PWD := &omnicom.Pwd{ValueInMn: []byte{0x56}}
	RTG := &omnicom.Rtg{}
	STG := &omnicom.Stg{}
	Position := &omnicom.PositionRadius{}
	PolygonPosition := []*omnicom.Pos{{}}
	PRIStoV := &omnicom.PositionReportingIntervalStoV{}
	GEStoV := &omnicom.GeofencingEnableStoV{}
	GSCI := &omnicom.GeofenceStatusCheckIntervalStoV{}
	PWDStoV := &omnicom.PasswordStoV{Value: []byte{0x56}}
	RTGStoV := &omnicom.RoutingStoV{}

	//NOTES: An iridium structure should not have MobileOriginatedHeader and MobileTerminated and MessageTerminatedConfirmation and MobileOriginatedLocationInformation
	// 1- MobileOriginated structure comes with omnicom data payload only with messages coming from the beacon to the server thru the iridium gateway
	// 2- MobileTerminated structure comes with omnicom data payload only with messages going from the server to the beacon
	// 3- MobileTerminatedConfirmation comes back from the iridium gateway without omnicom data payload when the server sends MobileTerminated to a beacon
	// as a confirmation that the message from the server reached the gateway and with information about any network errors.
	// 4- MobileOriginatedLocationInformation structure comes without any omnicom data payload and gives us relatively accurate infromation about the beacon location
	// it's mosetly parsed to let us know that out server can receive data and is connected to the iridium gateway properly.
	//TODO: Read the Pandore-Naf document, and populate the iridium structures with an omnicom payload && (MobileTerminatedHeader || MobileOriginatedHeader )
	//MobileOriginatedHeader: should appear for messages: beacon -> server
	//MobileTerminatedHeade: should appear for messages: server -> beacon
	// to run just unit tests on the naf library: from prisma repo run ( cd tms ; go test -v ./libnaf )
	AA := &omnicom.Omni{}
	AA.Omnicom = &omnicom.Omni_Aa{Aa: &omnicom.Aa{Header: []byte{0x45}, Date: Date}}
	AcknowledgeAssistance := &iridium.Iridium{}
	AcknowledgeAssistance.Moh = &iridium.MobileOriginatedHeader{MO_IEI: []byte{0x01}, MOHL: 28, CDR: 2578512475, IMEI: "300234010030450", SessStatus: "0", MOMSN: 15661, MTMSN: 375, TimeOfSession: 1475582020}
	AcknowledgeAssistance.Payload = &iridium.Payload{IEI: []byte{0x01}, Omnicom: AA}

	ABM := &omnicom.Omni{}
	ABM.Omnicom = &omnicom.Omni_Abm{Abm: &omnicom.Abm{Header: []byte{0x06}, Date: Date, ID_Msg: 341}}
	AcknowledgeBinaryMessage := &iridium.Iridium{}
	AcknowledgeBinaryMessage.Moh = &iridium.MobileOriginatedHeader{MO_IEI: []byte{0x01}, MOHL: 28, CDR: 2578512475, IMEI: "300234010030450", SessStatus: "0", MOMSN: 15661, MTMSN: 375, TimeOfSession: 1475582020}
	AcknowledgeBinaryMessage.Payload = &iridium.Payload{IEI: []byte{0x01}, Omnicom: ABM}

	AR := &omnicom.Omni{}
	AR.Omnicom = &omnicom.Omni_Ar{Ar: &omnicom.Ar{Header: []byte{0x02}, Msg_ID: 2570, Date_Position: &DatePositionAR, Date_Event: DateEvent, Power_Up: PowerUp,
		Power_Down: PowerDown, Battery_Alert: BatteryAlert, Intrusion_Alert: IntrusionAlert, No_Position_Fix: NoPositionFix, JB_Dome_Alert: JBDomeAlert,
		Loss_Mobile_Com: LossMobileCom, Daylight_Alert: DaylightAlert, Assistance_Alert: AssistanceAlert, Test_Mode: TestMode, Extention_Bit_Move: 1, Move: MV}}
	AlertReport := &iridium.Iridium{}
	AlertReport.Moh = &iridium.MobileOriginatedHeader{MO_IEI: []byte{0x02}, MOHL: 28, CDR: 2578512475, IMEI: "300234010030450", SessStatus: "0", MOMSN: 15661, MTMSN: 375, TimeOfSession: 1475582020}
	AlertReport.Payload = &iridium.Payload{IEI: []byte{0x02}, Omnicom: AR}

	AUP := &omnicom.Omni{}
	AUP.Omnicom = &omnicom.Omni_Aup{&omnicom.Aup{Header: []byte{0x08}, ID_Msg: 375, Date_Position: DatePosition, Web_Service_API_URL_Sending: URLSending, Web_Service_API_URL_Receiving: URLRecieving, Array: Array}}
	APIURLParameters := &iridium.Iridium{}
	APIURLParameters.Moh = &iridium.MobileOriginatedHeader{MO_IEI: []byte{0x08}, MOHL: 28, CDR: 2578512475, IMEI: "300234010030450", SessStatus: "0", MOMSN: 15661, MTMSN: 375, TimeOfSession: 1475582020}
	APIURLParameters.Payload = &iridium.Payload{IEI: []byte{0x08}, Omnicom: AUP}

	BM_StoV := &omnicom.Omni{}
	BM_StoV.Omnicom = &omnicom.Omni_Bmstov{&omnicom.BMStoV{Header: []byte{0x38}, Date: Date, Msg_Content: []byte{0x30}}}
	BinaryMessageFromServerToVessels := &iridium.Iridium{}
	BinaryMessageFromServerToVessels.Mth = &iridium.MobileTerminatedHeader{MO_IEI: []byte{0x38}, MTHL: 28, UniqueClientMessageID: "2578512475", IMEI: "300234010030450"}
	BinaryMessageFromServerToVessels.Payload = &iridium.Payload{IEI: []byte{0x38}, Omnicom: BM_StoV}

	BM := &omnicom.Omni{}
	BM.Omnicom = &omnicom.Omni_Bm{&omnicom.Bm{Header: []byte{0x07}, Date: Date, ID_Msg: 340, Length_Msg_Content: 73, Msg_Content: []byte{0x30}}}
	BinaryMessage := &iridium.Iridium{}
	BinaryMessage.Moh = &iridium.MobileOriginatedHeader{MO_IEI: []byte{0x07}, MOHL: 28, CDR: 2578512475, IMEI: "300234010030450", SessStatus: "0", MOMSN: 15661, MTMSN: 375, TimeOfSession: 1475582020}
	BinaryMessage.Payload = &iridium.Payload{IEI: []byte{0x07}, Omnicom: BM}

	DG := &omnicom.Omni{}
	DG.Omnicom = &omnicom.Omni_Dg{&omnicom.Dg{Header: []byte{0x37}, Msg_ID: 1034, Date: Date, GEO_ID: 14}}
	DeleteGeofence := &iridium.Iridium{}
	DeleteGeofence.Moh = &iridium.MobileOriginatedHeader{MO_IEI: []byte{0x07}, MOHL: 28, CDR: 2578512475, IMEI: "300234010030450", SessStatus: "0", MOMSN: 15661, MTMSN: 375, TimeOfSession: 1475582020}
	DeleteGeofence.Payload = &iridium.Payload{IEI: []byte{0x07}, Omnicom: DG}

	GA := &omnicom.Omni{}
	GA.Omnicom = &omnicom.Omni_Ga{&omnicom.Ga{Header: []byte{0x04}, Msg_ID: 1905, Date_Position: DatePosition}}
	GeofencingAcknowledge := &iridium.Iridium{}
	GeofencingAcknowledge.Moh = &iridium.MobileOriginatedHeader{MO_IEI: []byte{0x07}, MOHL: 28, CDR: 2578512475, IMEI: "300234010030450", SessStatus: "0", MOMSN: 15661, MTMSN: 375, TimeOfSession: 1475582020}
	GeofencingAcknowledge.Payload = &iridium.Payload{IEI: []byte{0x07}, Omnicom: GA}

	GBMN := &omnicom.Omni{}
	GBMN.Omnicom = &omnicom.Omni_Gbmn{&omnicom.Gbmn{Header: []byte{0x36}, Date: Date}}
	GBinaryMessageNotification := &iridium.Iridium{}
	GBinaryMessageNotification.Moh = &iridium.MobileOriginatedHeader{MO_IEI: []byte{0x36}, MOHL: 28, CDR: 2578512475, IMEI: "300234010030450", SessStatus: "0", MOMSN: 15661, MTMSN: 375, TimeOfSession: 1475582020}
	GBinaryMessageNotification.Payload = &iridium.Payload{IEI: []byte{0x36}, Omnicom: GBMN}

	GP := &omnicom.Omni{}
	GP.Omnicom = &omnicom.Omni_Gp{&omnicom.Gp{Header: []byte{0x03}, Beacon_ID: 461109, ID_Msg: 127, Date_Position: &DatePositionGP, Position_Reporting_Interval: PRI, Geofencing_Enable: GE, Position_Collection_Interval: PCI, Password: PWD, Routing: RTG, Firmware_Dome_Version: []byte{0x06, 0x07, 0x08}, Junction_Box_Version: []byte{0x02, 0x03, 0x04}, SIM_Card_ICCID: []byte("12345678901234567890"), G3_IMEI: []byte("300234010031990"), IRI_IMEI: []byte("300234010031991")}}
	GlobalParameters := &iridium.Iridium{}
	GlobalParameters.Moh = &iridium.MobileOriginatedHeader{MO_IEI: []byte{0x03}, MOHL: 28, CDR: 2578512475, IMEI: "300234010030450", SessStatus: "0", MOMSN: 15661, MTMSN: 375, TimeOfSession: 1475582020}
	GlobalParameters.Payload = &iridium.Payload{IEI: []byte{0x03}, Omnicom: GP}

	HPR := &omnicom.Omni{}
	Volt := &omnicom.Vltg{V_IN: 16.0}
	GF := &omnicom.GF{}
	DR := []*omnicom.DataReport{{Number_Data_Report: 1, Date_Position: DatePosition, Move: MV, Voltage: Volt, Geofence: GF}}
	HPR.Omnicom = &omnicom.Omni_Hpr{&omnicom.Hpr{Header: []byte{0x01}, Msg_ID: 301, Count_Total_Data_Reports: 1, Count_Data_ReportsInThis_Msg: 1, Data_Report: DR}}
	HistoryPositionReport := &iridium.Iridium{}
	HistoryPositionReport.Moh = &iridium.MobileOriginatedHeader{MO_IEI: []byte{0x01}, MOHL: 28, CDR: 2578512475, IMEI: "300234010030450", SessStatus: "0", MOMSN: 15661, MTMSN: 375, TimeOfSession: 1475582020}
	HistoryPositionReport.Payload = &iridium.Payload{IEI: []byte{0x01}, Omnicom: HPR}

	RMH := &omnicom.Omni{}
	RMH.Omnicom = &omnicom.Omni_Rmh{&omnicom.Rmh{Header: []byte{0x31}, Date: Date, Date_Interval: DateInterval, ID_Msg: 3741}}
	RequestMessageHistory := &iridium.Iridium{}
	RequestMessageHistory.Mth = &iridium.MobileTerminatedHeader{MO_IEI: []byte{0x31}, MTHL: 28, UniqueClientMessageID: "2578512475", IMEI: "300234010030450"}
	RequestMessageHistory.Payload = &iridium.Payload{IEI: []byte{0x31}, Omnicom: RMH}

	RSM := &omnicom.Omni{}
	RSM.Omnicom = &omnicom.Omni_Rsm{&omnicom.Rsm{Header: []byte{0x33}, ID_Msg: 1237, Date: Date, MsgTo_Ask: 3741}}
	RequestSpecificMessage := &iridium.Iridium{}
	RequestSpecificMessage.Mth = &iridium.MobileTerminatedHeader{MO_IEI: []byte{0x33}, MTHL: 28, UniqueClientMessageID: "2578512475", IMEI: "300234010030450"}
	RequestSpecificMessage.Payload = &iridium.Payload{IEI: []byte{0x33}, Omnicom: RSM}

	SPR := &omnicom.Omni{}
	SPR.Omnicom = &omnicom.Omni_Spr{&omnicom.Spr{Header: []byte{0x06}, Date_Position: DatePosition, Move: MV, Period: 1, CRC: 18}}
	SinglePositionReport := &iridium.Iridium{}
	SinglePositionReport.Moh = &iridium.MobileOriginatedHeader{MO_IEI: []byte{0x01}, MOHL: 28, CDR: 2578512475, IMEI: "300234010030450", SessStatus: "0", MOMSN: 15661, MTMSN: 375, TimeOfSession: 1475582020}
	SinglePositionReport.Payload = &iridium.Payload{IEI: []byte{0x01}, Omnicom: SPR}

	TMA := &omnicom.Omni{}
	TMA.Omnicom = &omnicom.Omni_Tma{&omnicom.Tma{Header: []byte{0x30}, Date: Date}}
	TestModeAcknowledge := &iridium.Iridium{}
	TestModeAcknowledge.Mth = &iridium.MobileTerminatedHeader{MO_IEI: []byte{0x30}, MTHL: 28, UniqueClientMessageID: "2578512475", IMEI: "300234010030450"}
	TestModeAcknowledge.Payload = &iridium.Payload{IEI: []byte{0x30}, Omnicom: TMA}

	UAUP := &omnicom.Omni{}
	UAUP.Omnicom = &omnicom.Omni_Uaup{&omnicom.Uaup{Header: []byte{0x3A}, ID_Msg: 92, Date: Date, Web_Service_API_URL_Sending: URLSendStoV, Web_Service_API_URL_Receiving: URLRecStoV, Array: ArrayStoV}}
	UpdateAUP := &iridium.Iridium{}
	UpdateAUP.Mth = &iridium.MobileTerminatedHeader{MO_IEI: []byte{0x3A}, MTHL: 28, UniqueClientMessageID: "2578512475", IMEI: "300234010030450"}
	UpdateAUP.Payload = &iridium.Payload{IEI: []byte{0x3A}, Omnicom: UAUP}

	UGFC := &omnicom.Omni{}
	UGFC.Omnicom = &omnicom.Omni_Ugcircle{&omnicom.UGCircle{Header: []byte{0x35}, Msg_ID: 604, Date: Date, GEO_ID: 14, Shape: 1, NAME: []byte("Shield"), Activated: 1, Setting: STG, Position: Position}}
	UpdateGFCircle := &iridium.Iridium{}
	UpdateGFCircle.Mth = &iridium.MobileTerminatedHeader{MO_IEI: []byte{0x35}, MTHL: 28, UniqueClientMessageID: "2578512475", IMEI: "300234010030450"}
	UpdateGFCircle.Payload = &iridium.Payload{IEI: []byte{0x35}, Omnicom: UGFC}

	UGFP := &omnicom.Omni{}
	UGFP.Omnicom = &omnicom.Omni_Ugpolygon{Ugpolygon: &omnicom.UGPolygon{Header: []byte{0x35}, Msg_ID: 3519, Date: Date, GEO_ID: 12, Shape: 1, NAME: []byte("Ground Zone"), TYPE: 0, Priority: 0, Activated: 1, Setting: STG, Number_Point: 0, Position: PolygonPosition, Padding: 0, CRC: 0}}
	UpdateGFPolygon := &iridium.Iridium{}
	UpdateGFPolygon.Mth = &iridium.MobileTerminatedHeader{MO_IEI: []byte{0x35}, MTHL: 28, UniqueClientMessageID: "2578512475", IMEI: "300234010030450"}
	UpdateGFPolygon.Payload = &iridium.Payload{IEI: []byte{0x35}, Omnicom: UGFP}

	UGP := &omnicom.Omni{}
	UGP.Omnicom = &omnicom.Omni_Ugp{&omnicom.Ugp{Header: []byte{0x34}, Date: Date, Position_Reporting_Interval: PRIStoV, Geofencing_Enable: GEStoV, Geofence_Status_Check_Interval: GSCI, Password: PWDStoV, Routing: RTGStoV}}
	UpdateGlobalParameters := &iridium.Iridium{}
	UpdateGlobalParameters.Mth = &iridium.MobileTerminatedHeader{MO_IEI: []byte{0x34}, MTHL: 28, UniqueClientMessageID: "2578512475", IMEI: "300234010030450"}
	UpdateGlobalParameters.Payload = &iridium.Payload{IEI: []byte{0x34}, Omnicom: UGP}

	UIC := &omnicom.Omni{}
	UIC.Omnicom = &omnicom.Omni_Uic{&omnicom.Uic{Header: []byte{0x32}, Date: Date}}
	UpdateInterval := &iridium.Iridium{}
	UpdateInterval.Mth = &iridium.MobileTerminatedHeader{MO_IEI: []byte{0x32}, MTHL: 28, UniqueClientMessageID: "2578512475", IMEI: "300234010030450"}
	UpdateInterval.Payload = &iridium.Payload{IEI: []byte{0x32}, Omnicom: UIC}

	AAcomp := "//SR//TM/OVB//IMEI/300234010030450//TOM/AA//DA/170710//TI/0820//ER"
	ABMcomp := "//SR//TM/OVB//IMEI/300234010030450//TOM/ABM//DA/170710//TI/0820//ID/341//ABMET/0//ER"
	ARcomp := "//SR//TM/OVB//IMEI/300234010030450//TOM/AR//ID/2570//DAP/20160301//TIP/0938//XLT/-4.0512//XLG/3.5612//SP/20//CO/30//DAE/170710//TIE/0820//PU/PU//BSR/0//IAS/0//FPR/0//NSV/10//JBDS/0//MCL/0//DLA/0//AAS/0//TMA/18//ER"
	AUPcomp := "//SR//TM/OVB//IMEI/300234010030450//TOM/AUP//ID/375//DAP/20160301//TIP/0938//XLT/-4.05123//XLG/3.56123//WSAUS/oroliabeacon.com//WSAUR/oroliabeacon.com//AIO/AUP//ER"
	BMcomp := "//SR//TM/OVB//IMEI/300234010030450//TOM/BM//DA/170710//TI/0820//ID/340//BMS/73//ER"
	BMSVcomp := "//SR//TM/OVB//IMEI/300234010030450//TOM/BM_Stov//DA/170710//TI/0820//ER"
	DGcomp := "//SR//TM/OVB//IMEI/300234010030450//TOM/DG//DA/170710//TI/0820//ID/1034//GEOID/14//ER"
	GBMNcomp := "//SR//TM/OVB//IMEI/300234010030450//TOM/GBMN//DA/170710//TI/0820//ER"
	GAcomp := "//SR//TM/OVB//IMEI/300234010030450//TOM/GA//ID/1905//DAP/20160301//TIP/0938//XLT/-4.05123//XLG/3.56123//GFET/0//ER"
	GPcomp := "//SR//TM/OVB//IMEI/300234010030450//TOM/GP//BID/461109//DAP/20160301//TIP/0938//XLT/-4.051//XLG/3.561//RPI/0//GFE/0//PCI/0//PWD/V//RTG/0//FDV/6.7.8//JBV/2.3.4//SCI/12345678901234567890//3GIMEI/300234010031990//IMEI/300234010031991//ER"
	HPRcomp := "//SR//TM/OVB//IMEI/300234010030450//TOM/HPR//ID/301//CDR/1//DAP/20160301//TIP/0938//XLT/-4.05123//XLG/3.56123//SP/20//CO/30//VIN/16//RPI/0//SAF/0//GEOID/0//ER"
	RMHcomp := "//SR//TM/OVB//IMEI/300234010030450//TOM/RMH//DA/170710//TI/0820//DIS/170710//TIS/0820//DIE/170710//TIE/0820//ID/3741//ER"
	RSMcomp := "//SR//TM/OVB//IMEI/300234010030450//TOM/RSM//ID/1237//DA/170710//TI/0820//MTA/3741//ER"
	SPRcomp := "//SR//TM/OVB//IMEI/300234010030450//TOM/SPR//DAP/20160301//TIP/0938//XLT/-4.05123//XLG/3.56123//SP/20//CO/30//VIN/0//RPI/1//ER"
	TMAcomp := "//SR//TM/OVB//IMEI/300234010030450//TOM/TMA//DA/170710//TI/0820//ER"
	UAUPcomp := "//SR//TM/OVB//IMEI/300234010030450//TOM/UAUP///DA/170710//TI/0820//WSAUS/oroliabeacon.com//WSAUR/oroliabeacon.com//ER"
	UGFCcomp := "//SR//TM/OVB//IMEI/300234010030450//TOM/UG_Cirlce//ID/604//DA/170710//TI/0820//GEOID/14//GZN/Shield//GEOP/0//GEOA/1//NRI/0//SPT/0//GEOC/0,0//GEOR/0//ER"
	UGFPcomp := "//SR//TM/OVB//IMEI/300234010030450//TOM/UG_Polygon//ID/3519//DA/170710//TI/0820//GEOID/12//GZN/Ground Zone//GEOP/0//GEOA/1//NRI/0//SPT/0//GEOV/0,0//ER"
	UGPcomp := "//SR//TM/OVB//IMEI/300234010030450//TOM/UGP//DA/170710//TI/0820//ER"
	UICcomp := "//SR//TM/OVB//IMEI/300234010030450//UIC///DA/170710//TI/0820//NRI/0//ER"

	tt := []struct {
		name string
		iri  *iridium.Iridium
		comp string
	}{
		{"Ack Assistance", AcknowledgeAssistance, AAcomp},
		{"Single Position Report", SinglePositionReport, SPRcomp},
		{"Ack Binary Message", AcknowledgeBinaryMessage, ABMcomp},
		{"Alert Report", AlertReport, ARcomp},
		{"History Position Report", HistoryPositionReport, HPRcomp},
		{"Binary Message from Server to Vessels", BinaryMessageFromServerToVessels, BMSVcomp},
		{"Binary Message", BinaryMessage, BMcomp},
		{"Delete Geofence", DeleteGeofence, DGcomp},
		{"Geofencing Acknowledge", GeofencingAcknowledge, GAcomp},
		{"3G Binary Message notification", GBinaryMessageNotification, GBMNcomp},
		{"Request Message History", RequestMessageHistory, RMHcomp},
		{"Request Specific Mesaage", RequestSpecificMessage, RSMcomp},
		{"Test Mode Acknowledge", TestModeAcknowledge, TMAcomp},
		{"Update Interval Change", UpdateInterval, UICcomp},
		{"API URL Parameters", APIURLParameters, AUPcomp},
		{"Global Parameters", GlobalParameters, GPcomp},
		{"Update API URL Parameters", UpdateAUP, UAUPcomp},
		{"Update Geo-Fence Circle", UpdateGFCircle, UGFCcomp},
		{"Update Geo-Fence Polygon", UpdateGFPolygon, UGFPcomp},
		{"Update Global Parameters", UpdateGlobalParameters, UGPcomp},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			data := &tms.MessageActivity{
				Imei:     &google_protobuf1.StringValue{Value: "300234010030450"},
				MetaData: &tms.MessageActivity_Omni{tc.iri.Payload.Omnicom},
			}
			msg, err := EncodeNaf(data)
			if err != nil {
				t.Errorf("error: %+v", err)
			} else {
				t.Logf("%+v\n", msg)
			}
			if reflect.TypeOf(msg) != reflect.TypeOf(tc.comp) {
				t.Errorf("%s failed because %s and %s are not the same type.", tc.name, reflect.TypeOf(msg).String(), reflect.TypeOf(tc.comp).String())
			}
			if reflect.DeepEqual(tc.comp, msg) == false {
				t.Errorf("%s failed because %+v and %+v are not deeply equal.", tc.name, tc.comp, msg)
			}
		})
	}
}
