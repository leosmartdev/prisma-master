package lib

import (
	"prisma/tms"
	"prisma/tms/devices"
	"prisma/tms/sar"
	"prisma/tms/sit185"
	"regexp"
	"testing"
)

func TestProcessMCCMessagesToTracks_Subtest(t *testing.T) {

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
		msg      string
		protocol string
		me       *tms.SensorID
	}{
		{"Unlocated Alert Message", UnlocatedAlertMessageSample, "tcp", &tms.SensorID{Site: 10, Eid: 9}},
		{"Located Alert Message", LocatedAlertMessageSample, "tcp", &tms.SensorID{Site: 10, Eid: 10}},
		{"Confirmed Alert Message", ConfirmedAlertMessageSample, "tcp", &tms.SensorID{Site: 10, Eid: 11}},
		{"Located Alert MeoElement Message", LocatedAlertMessageMeoElementSample, "tcp", &tms.SensorID{Site: 10, Eid: 12}},
		{"DISTRESS COSPAS-SARSAT ALERT", Sit185DCSA, "ftp", &tms.SensorID{Site: 10, Eid: 13}},
		{"UNKNOWN MESSAGE", "Hey people I am dying here", "ftp", &tms.SensorID{Site: 10, Eid: 13}},

		//{"Free Form Message", FreeformMessageSample, "tcp", &tms.SensorID{Site: 10, Eid: 13}},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var mccmsg *sar.SarsatMessage
			var err error
			mccmsg, err = MccxmlParser([]byte(tc.msg), tc.protocol)
			if err != nil {
				mccmsg, err = Sit185Parser([]byte(tc.msg), tc.protocol, fields["date_format"], regularpaterns)
				if err != nil {
					mccmsg = DefaultParser([]byte(tc.msg), tc.protocol)
				}
			}
			t.Logf("%s is being tested", tc.name)
			track, err := Process(mccmsg, tc.me)
			if err != nil {
				t.Errorf("%s failed to be processed to tms track: %+v", tc.name, err)
			} else {
				if len(track.Targets) == 0 {
					t.Logf("%+v has no target populated inside the track", tc.name)
				} else {
					for _, target := range track.Targets {
						if target.Type != devices.DeviceType_SARSAT {
							t.Errorf("%+v has a target type %+v != devices.SARSAT", tc.name, target.Type)
						}
					}
				}

				if len(track.Id) == 0 {
					t.Errorf("%+v has no Id inside track", tc.name)
				}

				t.Logf("\nMessage type %v:  %+v \n %+v ", mccmsg.MessageType, tc.name, track)
			}

		})
	}
}
