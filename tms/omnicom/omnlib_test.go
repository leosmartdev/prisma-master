package omnicom

import (
	"reflect"
	"testing"
)

func TestOmnicom_Subtest(t *testing.T) {

	GBinaryMessageNotification := &GBMN{0x36, Date{17, 07, 10, 500}, 0, 0}
	SinglePositionReport := &SPR{0x06, Date_Position{17, 07, 10, 500, 0, 0}, Move{0, 0}, 10, 1, 25, 0, 0}
	//APIURL := &AUP {0x08, 0, Date_Position{17, 07, 10, 500, 0, 0}, Web_Service_API_URL_Sending{[]byte("oroliabeacon.com"), 1}, Web_Service_API_URL_Receiving{[]byte("oroliabeacon.com"), 1}, Array{[]byte("oroliabeacon.com"), 1}, 0, 0}
	AcknowledgeBinaryMessage := &ABM{0x09, Date{17, 07, 10, 500}, 0, 0, 0, 0}
	AcknowledgeAssistance := &AA{0x45, Date{17, 07, 10, 500}, 0, 0}
	//BinaryMessageFromServerToVessels := &BM_StoV {}
	DeleteGeofence := &DG{0x37, 0, Date{17, 07, 10, 500}, 0, 0, 0}
	GeofencingAcknowledge := &GA{0x04, 0, Date_Position{17, 07, 10, 500, 0, 0}, 0, 20, 0, 0}
	RequestMessageHistory := &RMH{0x31, Date{17, 07, 10, 500}, Date_Interval{Date{17, 07, 10, 500}, Date{17, 07, 10, 501}}, 0, 0, 0}
	RequestSpecificMessage := &RSM{0x33, 0, Date{17, 07, 10, 500}, 0, 0, 0}
	SplitDiagnosticRequest := &SDR{0x41, Date{17, 07, 10, 500}, 0, 0, 0}
	SinglePositionReportSolar := &SPRS{0x0a, Date_Position{17, 07, 10, 500, 0, 0}, Move{0, 0}, 0, 0.0, 0, 0, 0, 0}
	TestModeAcknowledge := &TMA{0x30, Date{17, 07, 10, 500}, 0, 0}
	//UpdateGlobalParameters := &UGP {0x34, 0, }
	UnitIntervalChange := &UIC{0x32, 0, Date{17, 07, 10, 500}, 0, 0}

	tt := []struct {
		name    string
		omnicom Omnicom
	}{
		{"3G binary message notification(0x36), Iridium", GBinaryMessageNotification},
		{"Single Position report(0x06), Iridium/3G", SinglePositionReport},
		//{"API url parameters(0x08), Iridium", APIURL},
		{"Ack binary message(0x09), Iridium/3G", AcknowledgeBinaryMessage},
		{"Ack assistance(0x45), Iridium/3G", AcknowledgeAssistance},
		{"Delete Geofence(0x37), Iridium/3G", DeleteGeofence},
		{"Geofencing ack(0x04), Iridium/3G", GeofencingAcknowledge},
		{"Request Message History(0x31), Iridium/3G", RequestMessageHistory},
		{"Request a specific message(0x33), Iridium/3G", RequestSpecificMessage},
		{"Split diagonostic request(0x41), Iridium", SplitDiagnosticRequest},
		{"Single Position report Solar (0x0a), Iridium/3G/RPMA", SinglePositionReportSolar},
		{"Test Mode Ack(0x30), Iridium/3G", TestModeAcknowledge},
		{"Unit interval change(0x32), Iridium/3G", UnitIntervalChange},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := tc.omnicom.Encode()
			if err != nil {
				t.Errorf("%s failed to be encode %+v due to %+v", tc.name, tc.omnicom, err)
			} else {
				omn, err := Parse(raw)
				if err != nil {
					t.Errorf("%s failed to parse due to %+v", tc.name, err)
				} else {
					if reflect.TypeOf(tc.omnicom) != reflect.TypeOf(omn) {
						t.Errorf("%s failed because %s and %s are not the same type.", tc.name, reflect.TypeOf(tc.omnicom).String(), reflect.TypeOf(omn).String())
					}
					tc.omnicom.setCRC(omn.getCRC())
					if reflect.DeepEqual(tc.omnicom, omn) == false {
						t.Errorf("%s failed because %+v and %+v are not deeply equal.", tc.name, tc.omnicom, omn)
					}
				}
			}

		})

	}

}
