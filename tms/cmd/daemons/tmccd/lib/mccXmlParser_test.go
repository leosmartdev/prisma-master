package lib

import "testing"

func TestAlertXmlMessage_Subtest(t *testing.T) {

	tt := []struct {
		name     string
		msg      string
		protocol string
	}{
		{"Unlocated Alert Message", UnlocatedAlertMessageSample, "tcp"},
		{"Located Alert Message", LocatedAlertMessageSample, "tcp"},
		{"Confirmed Alert Message", ConfirmedAlertMessageSample, "tcp"},
		{"Located Alert MeoElement Message", LocatedAlertMessageMeoElementSample, "tcp"},
		{"Free Form Message", FreeformMessageSample, "tcp"},
		{"Unlocated Alert Message with non xml headers and footers", UnlocatedAlertMessageWithHeadersAndFooters, "ftp"},
		{"Incident alert messages from Chile mcc", ChileXML, "ftp"},
		{"Resolved alert xml message", testXML, "ftp"},
		// will fail for not {"Resolved alert message from indonesia", ResolvedAlertMessageFromIndonesia, "ftp"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			//extract the xml data from the mcc file
			data := XMLExp.Find([]byte(tc.msg))

			msg, err := MccxmlParser(data, tc.protocol)
			if err != nil {
				t.Errorf("%s failed to be parsed %+v", tc.name, err)
			}
			t.Logf("%+v\n", msg)
		})
	}
}
