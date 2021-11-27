package omnicom

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestOmnicomParseEncode(t *testing.T) {
	byteList := []byte{0x40, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x01, 0x01, 0x01, 0x00, 0x12, 0x23, 0x0C}

	omn, err := Parse(byteList)
	if err != nil {
		t.Error(err)
	}
	byteList, err = json.Marshal(omn)
	if err != nil {
		t.Error(err)
	}
	t.Logf("%+v\n", byteList)

	var spm SPM
	err = json.Unmarshal(byteList, &spm)
	if err != nil {
		t.Error(err)
	}
	t.Logf("%+v\n", spm)

	blist, err := omn.Encode()
	if err != nil {
		t.Error(err)
	}
	for _, b := range blist {
		t.Logf("%x\n", b)
	}

	date := Date{1, 1, 1, 1}
	gbmn := GBMN{0x36, date, 0, 0}
	blist, err = gbmn.Encode()
	if err != nil {
		t.Error(err)
	}
	for _, b := range blist {
		t.Logf("%x\n", b)
	}
}

func TestLibOmnicom_Subtest(t *testing.T) {

	GBinaryMessageNotification := &GBMN{0x36, Date{17, 04, 29, 450}, 0, 0}

	ar := &AR{
		Header:    0x02,
		Msg_ID:    0x03,
		Test_Mode: Test_Mode{1, 1},
	}

	GlobalParameters := &GP{0x03, 10, 1992, Date_Position{17, 10, 10, 450, 2.0, 1.0}, Position_Reporting_Interval{10, 1}, Geofencing_Enable{1, 0}, Position_Collection_Interval{15, 0}, Password{[]byte("123456789012345"), 0}, Routing{1, 0}, []byte("1.2"), []byte("1.3"), []byte("12345678901234567890"), []byte("123456789012345"), []byte("123456789012345"), 26}
	ugP := &UG_Polygon{
		Header:    0x35,
		Msg_ID:    1992,
		GEO_ID:    4294967290,
		Shape:     0,
		NAME:      []byte("12345678901234567890"),
		TYPE:      1,
		Priority:  0,
		Activated: 0,

		Position: []Position{{1.0, 1.0}, {1.0, 1.0}, {1.0, 1.0}, {1.0, 1.0}, {1.0, 1.0}},
		Setting: Setting{
			New_Position_Report_Period: 5,
			Speed_Threshold:            1.2,
		},
		Date:         Date{17, 04, 29, 450},
		Number_Point: 5,
		Padding:      0,
		CRC:          0,
	}

	tt := []struct {
		name    string
		omnicom Omnicom
	}{
		{"3G binary message notification(0x36), Iridium", GBinaryMessageNotification},
		{"Alert Report(0x02), Iridium", ar},
		{"Global parameter 0x03, Iridium", GlobalParameters},
		{"Upload Geofence (0x35), Iridium", ugP},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var omn Omnicom
			raw, err := tc.omnicom.Encode()
			if err != nil {
				t.Errorf("%s failed to encode %+v due to: %+v", tc.name, tc.omnicom, err)
			} else {
				str, err := encFromByte(raw)

				err = checkCRC(str)
				if err != nil {
					t.Errorf("%s wrong encoded message %+v due to: %+v", tc.name, tc.omnicom, err)
				}
				omn, err = Parse(raw)
				if err != nil {
					t.Errorf("%s failed to parse %+v due to: %+v", tc.name, tc.omnicom, err)
				} else {
					if reflect.TypeOf(tc.omnicom) != reflect.TypeOf(omn) {
						t.Errorf("%s failed %s and %s are not the same type", tc.name, reflect.TypeOf(tc.omnicom).String(), reflect.TypeOf(omn).String())
					}

					tc.omnicom.setCRC(omn.getCRC())
					if reflect.DeepEqual(tc.omnicom, omn) == false {
						t.Errorf("%s failed because %+v and %+v are not deeply equal", tc.name, tc.omnicom, omn)
					}

					//	t.Logf("%s is ok because %+v and %+v are deeply equal", tc.name, tc.omnicom, omn)
				}
			}
		})
	}
}
