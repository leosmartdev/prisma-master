package lib

import (
	"encoding/json"
	"io/ioutil"
	"prisma/tms/sar"
	"prisma/tms/sit185"
	"regexp"
	"testing"
)

type jtemplate struct {
	MsgNum     string `json:"msg_num"`
	Date       string `json:"date"`
	DateFormat string `json:"date_format"`
	HexaID     string `json:"hex_id"`
	Encoded    string `json:"encoded"`
	DopplerA   string `json:"doppler_a"`
	DopplerB   string `json:"doppler_b"`
	Confirmed  string `json:"confirmed"`
	Doa        string `json:"doa"`
}

func TestAlertSit185lMessage_Subtest(t *testing.T) {

	jtemp, err := jsonParse("../../../../sit185/sit185-template.json")
	if err != nil {
		t.Errorf("TestAlertSit185lMessage failed to load config because of: %v", err)
	}
	fields := map[string]string{
		"date":        jtemp.Date,
		"date_format": jtemp.DateFormat,
		"confirmed":   jtemp.Confirmed,
		"encoded":     jtemp.Encoded,
		"doa":         jtemp.Doa,
		"doppler_a":   jtemp.DopplerA,
		"doppler_b":   jtemp.DopplerB,
		"msg_num":     jtemp.MsgNum,
		"hex_id":      jtemp.HexaID,
	}

	var regularpaterns []sit185.RegularExpPattern

	for key, field := range fields {
		if key != "date_format" {
			rexp, err := regexp.Compile(field)
			if err != nil {
				t.Errorf("unable to compile regular exp failed because of: %v", err)
			}
			regularpaterns = append(regularpaterns, sit185.RegularExpPattern{key, rexp})
		}
	}

	tt := []struct {
		name     string
		data     string
		protocol string
	}{
		{"DISTRESS COSPAS-SARSAT ALERT", Sit185DCSA, "ftp"},
		{"DEFAULT HANDINLING FOR UKNOWN FORMAT", UknownMccMessageFormat, "ftp"},
		{"SAMPLE 406 MHz DOPPLER CONFIRMED ALERT", sit185.DopConfAlt, "tcp"},
		{"SAMPLE 406 MHz DOPPLER INITIAL ALERT", sit185.DopInitAlt, "ftp"},
		{"SAMPLE 406 MHz DOPPLER POSITION CONFLICT ALERT", sit185.DopPosConfAlt, "ftp"},
		{"SAMPLE 406 MHz POSITION ALERT", sit185.PosAlt, "tcp"},
		{"SAMPLE 406 MHz CONFIRMED UPDATE POSITION ALERT", sit185.ConfUpdPosAlt, "ftp"},
		{"SAMPLE 406 MHz INITIAL DOA POSITION ALERT", sit185.InitDoaPosAlt, "ftp"},
		{"Malaysia 406 MHz CONFIRMED", sit185.MalaysiaSit185CONFIRMED, "ftp"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			sarMessage, err := Sit185Parser([]byte(tc.data), tc.protocol, fields["date_format"], regularpaterns)
			if err != nil {
				if err.Error() == "MCC message format is Uknown" {
					sarMessage = DefaultParser([]byte(tc.data), tc.protocol)
				} else {
					t.Fatalf("%s failed to parse sit185 message because of: %v", tc.name, err)
				}

			}
			switch tc.name {
			case "Malaysia 406 MHz CONFIRMED":
				if sarMessage.SarsatAlert.Beacon.HexId != "C2A9D40D30330D1" {
					t.Errorf("%s should have HexID: C2A9D40D30330D1", tc.name)
				}
				if sarMessage.SarsatAlert.ResolvedAlertMessage.CompositeLocation == nil {
					t.Errorf("%s shoud have a valid composite location", tc.name)
				}
				if sarMessage.SarsatAlert.ResolvedAlertMessage.MeoElemental == nil {
					t.Errorf("%s should have a valid MeoElemental (Doa) location", tc.name)
				}
			case "DISTRESS COSPAS-SARSAT ALERT":
				if sarMessage.MessageType != sar.SarsatMessage_SIT_185 {
					t.Errorf("%s has wrong message type %v", tc.name, sarMessage.MessageType)
				}
				if sarMessage.SarsatAlert.ResolvedAlertMessage.CompositeLocation == nil {
					t.Errorf("%s shoud have a valid composite location", tc.name)
				}
				if sarMessage.SarsatAlert.ResolvedAlertMessage.MeoElemental == nil {
					t.Errorf("%s should have a valid MeoElemental (Doa) location", tc.name)
				}
				if sarMessage.SarsatAlert.Beacon.HexId != "400E77A1FCFFBFF" {
					t.Errorf("%s should have HexID: 400E77A1FCFFBFF", tc.name)
				}
			case "DEFAULT HANDINLING FOR UKNOWN FORMAT":
				if sarMessage.MessageType != sar.SarsatMessage_UNKNOWN {
					t.Errorf("%s has wrong message type %v", tc.name, sarMessage.MessageType)
				}
			case "SAMPLE 406 MHz DOPPLER CONFIRMED ALERT":
				if sarMessage.MessageType != sar.SarsatMessage_SIT_185 {
					t.Errorf("%s has wrong message type %v", tc.name, sarMessage.MessageType)
				}
				if sarMessage.SarsatAlert.ResolvedAlertMessage.CompositeLocation == nil {
					t.Errorf("%s shoud have a valid composite location", tc.name)
				}
				if sarMessage.SarsatAlert.ResolvedAlertMessage.Elemental == nil {
					t.Errorf("%s should have an elemental for doppler position.", tc.name)
				}
				if sarMessage.SarsatAlert.Beacon.HexId != "9D064BED62EAFE1" {
					t.Errorf("%s should have HexID: 9D064BED62EAFE1", tc.name)
				}
			case "SAMPLE 406 MHz DOPPLER INITIAL ALERT":
				if sarMessage.MessageType != sar.SarsatMessage_SIT_185 {
					t.Errorf("%s has wrong message type %v", tc.name, sarMessage.MessageType)
				}
				if len(sarMessage.SarsatAlert.IncidentAlertMessage.Elemental) != 2 {
					t.Errorf("%s does not have two elementals corresponding to doppler_a and doppler_b", tc.name)
				}
				if sarMessage.SarsatAlert.Beacon.HexId != "ADCE402FA80028D" {
					t.Errorf("%s should have HexID: ADCE402FA80028D", tc.name)
				}
			case "SAMPLE 406 MHz DOPPLER POSITION CONFLICT ALERT":
				if sarMessage.MessageType != sar.SarsatMessage_SIT_185 {
					t.Errorf("%s has wrong message type %v", tc.name, sarMessage.MessageType)
				}
				if len(sarMessage.SarsatAlert.IncidentAlertMessage.Elemental) != 2 {
					t.Errorf("%s does not have two elementals corresponding to doppler_a and doppler_b", tc.name)
				}
				if sarMessage.SarsatAlert.Beacon.HexId != "C1ADE28809C0185" {
					t.Errorf("%s should have HexID: C1ADE28809C0185", tc.name)
				}
			case "SAMPLE 406 MHz POSITION ALERT":
				if sarMessage.MessageType != sar.SarsatMessage_SIT_185 {
					t.Errorf("%s has wrong message type %v", tc.name, sarMessage.MessageType)
				}
				if sarMessage.SarsatAlert.ResolvedAlertMessage.CompositeLocation == nil {
					t.Errorf("%s shoud have a valid composite location", tc.name)
				}
				if sarMessage.SarsatAlert.ResolvedAlertMessage.MeoElemental == nil {
					t.Errorf("%s should have a valid MeoElemental (Doa) location", tc.name)
				}
				if sarMessage.SarsatAlert.ResolvedAlertMessage.Encoded == nil {
					t.Errorf("%s should have a valid Encoded location", tc.name)
				}
				if sarMessage.SarsatAlert.Beacon.HexId != "3266E2019CFFBFF" {
					t.Errorf("%s should have HexID: 3266E2019CFFBFF", tc.name)
				}
			case "SAMPLE 406 MHz CONFIRMED UPDATE POSITION ALERT":
				if sarMessage.MessageType != sar.SarsatMessage_SIT_185 {
					t.Errorf("%s has wrong message type %v", tc.name, sarMessage.MessageType)
				}
				if sarMessage.SarsatAlert.ResolvedAlertMessage.CompositeLocation == nil {
					t.Errorf("%s shoud have a valid composite location", tc.name)
				}
				if len(sarMessage.SarsatAlert.ResolvedAlertMessage.Elemental) != 1 {
					t.Errorf("%s does not have 1 elemental corresponding to doppler_a", tc.name)
				}
				if sarMessage.SarsatAlert.ResolvedAlertMessage.Encoded == nil {
					t.Errorf("%s should have a valid Encoded location", tc.name)
				}
				if sarMessage.SarsatAlert.Beacon.HexId != "2AB82AF800FFBFF" {
					t.Errorf("%s should have HexID: 2AB82AF800FFBFF", tc.name)
				}
			case "SAMPLE 406 MHz INITIAL DOA POSITION ALERT":
				if sarMessage.MessageType != sar.SarsatMessage_SIT_185 {
					t.Errorf("%s has wrong message type %v", tc.name, sarMessage.MessageType)
				}
				if sarMessage.SarsatAlert.IncidentAlertMessage.MeoElemental == nil {
					t.Errorf("%s does not have MeoElemental (Doa) positions value", tc.name)
				}
				if sarMessage.SarsatAlert.Beacon.HexId != "278C362E3CFFBFF" {
					t.Errorf("%s should have HexID: 278C362E3CFFBFF", tc.name)
				}
			default:
				t.Errorf("%s not covered", tc.name)
			}
			//	t.Logf("%+v", sarMessage)
		})
	}

}
func jsonParse(filename string) (jtemplate, error) {

	var jsontype jtemplate

	file, e := ioutil.ReadFile(filename)
	if e != nil {
		return jsontype, e
	}
	err := json.Unmarshal(file, &jsontype)
	if err != nil {
		return jsontype, err
	}

	return jsontype, nil
}
