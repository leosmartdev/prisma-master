package main

import (
	"prisma/tms"
	"prisma/tms/iridium"
	omni "prisma/tms/omnicom"
	"prisma/tms/tmsg"
	"reflect"
	"testing"
	"time"

	proto "github.com/golang/protobuf/proto"
)

func TestProcessMessageRequest(t *testing.T) {

	tt := []struct {
		name    string
		msg     *omni.Omni
		omnicom omni.Omnicom
	}{
		{"Request specific message",
			&omni.Omni{Omnicom: &omni.Omni_Rsm{&omni.Rsm{
				Header:    []byte{0x33},
				ID_Msg:    1992,
				MsgTo_Ask: 1,
				Date:      &omni.Dt{Year: 18, Month: 1, Day: 12, Minute: 1},
			}}},
			&omni.RSM{
				Header:     0x33,
				ID_Msg:     1992,
				Msg_to_Ask: 1,
				Date:       omni.Date{Year: 18, Month: 1, Day: 12, Minute: 1},
			},
		},
		{"Request Message History",
			&omni.Omni{Omnicom: &omni.Omni_Rmh{&omni.Rmh{
				Header: []byte{0x31},
				Date:   &omni.Dt{Year: 18, Month: 1, Day: 12, Minute: 200},
				Date_Interval: &omni.DateInterval{
					Start: &omni.Dt{Year: 18, Month: 1, Day: 12, Minute: 1},
					Stop:  &omni.Dt{Year: 18, Month: 1, Day: 12, Minute: 100},
				},
				ID_Msg: 1992,
			}}},
			&omni.RMH{
				Header: 0x31,
				Date:   omni.Date{Year: 18, Month: 1, Day: 12, Minute: 1},
				Date_Interval: omni.Date_Interval{
					Start: omni.Date{Year: 18, Month: 1, Day: 12, Minute: 1},
					Stop:  omni.Date{Year: 18, Month: 1, Day: 12, Minute: 100},
				},
				ID_Msg: 1992,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			msgReq := &tms.Multicast{
				Destinations: []*tms.EntityRelationship{
					&tms.EntityRelationship{
						Id:   "111111111111111",
						Type: "iridium",
					},
				},
			}
			cmd, err := tmsg.PackFrom(tc.msg)
			if err != nil {
				t.Errorf("%s could not pack Omnicom proto message into message req cmd", tc.name)
			}
			msgReq.Payload = cmd
			t.Logf("%+v", msgReq)

			IridiumMsgs, err := packIridiumFrom(msgReq)
			if err != nil {
				t.Errorf("%s test failed: could not pack MessageRequest into Iridium commands %v", tc.name, err)
			}
			t.Logf("%+v", IridiumMsgs)
			for _, iridiumMsg := range IridiumMsgs {
				MTH, MTP, err := iridium.PopulateProtobufToMobileTerminated(iridiumMsg)
				if err != nil {
					t.Errorf("%s has failed because: %v", tc.name, err)
				}
				t.Logf("Mobile Terminated Header: %v", MTH)
				t.Logf("Mobile Terminated Payload: %v", MTP)
				//encode Mobile Terminated header
				Hraw, err := MTH.Encode()
				if err != nil {
					t.Errorf("%s has failed because: %v", tc.name, err)
				}

				// encode Mobile terminated payload
				Praw, err := MTP.Encode()
				if err != nil {
					t.Errorf("%s has failed because: %v", tc.name, err)
				}
				t.Logf("Raw Mobile Terminated Header %v", Hraw)
				t.Logf("Raw Mobile Terminated Paylod %v", Praw)

				payload, err := iridium.ParseMOPayload(Praw)
				if err != nil {
					t.Errorf("%s failed because: %v", tc.name, err)
				}
				header, err := iridium.ParseMTHeader(Hraw)
				if err != nil {
					t.Errorf("%s failed because: %v", tc.name, err)
				}

				iri, err := iridium.PopulateMTProtobuf(*header, payload)
				if err != nil {
					t.Errorf("%s failed because: %v", tc.name, err)
				}

				//simulate going in the wire and coming out at the other side
				byteList, err := proto.Marshal(iri)
				if err != nil {
					t.Errorf("%+v", err)
				}

				var output iridium.Iridium

				err = proto.Unmarshal(byteList, &output)
				if err != nil {
					t.Errorf("%+v", err)
				}
				t.Logf("%+v", output)

				MTHValue := reflect.ValueOf(output.Mth)
				MTPValue := reflect.ValueOf(output.Payload.Omnicom)

				for i := 0; i < MTHValue.Elem().NumField(); i++ {
					if MTHValue.Elem().Field(i).String() != reflect.ValueOf(iri.Mth).Elem().Field(i).String() {
						t.Errorf("%s fails because: %s do not match ", tc.name, MTHValue.Type().Field(i).Name)
					}

				}

				for i := 0; i < MTPValue.Elem().NumField()-2; i++ {
					if MTPValue.Elem().Field(i).String() != reflect.ValueOf(tc.omnicom).Elem().Field(i).String() {
						t.Errorf("%s fails because: %+v do not match %+v  ", tc.name, MTPValue.Elem().Field(i), reflect.ValueOf(tc.omnicom).Elem().Field(i))
					}
				}
			}

		})
	}

}

func TestCalculateTime(t *testing.T) {
	t.Run("", func(t *testing.T) {
		t.Run("DatePosition", calculateTimeFromDatePosition)
		t.Run("DateEvent", calculateTimeFromDateEvent)
		t.Run("Non-DateStruct", calculateTimeFromInvalidInput)
	})
}

func calculateTimeFromDatePosition(t *testing.T) {
	datePosition := &omni.DatePosition{
		Year:      16,
		Month:     10,
		Day:       4,
		Minute:    352,
		Latitude:  47.8141,
		Longitude: -3.4793,
	}

	dateTime, err := calculateTime(datePosition)
	validateTime(dateTime, err, t)
}

func calculateTimeFromDateEvent(t *testing.T) {
	dateEvent := &omni.DateEvent{
		Year:   16,
		Month:  10,
		Day:    4,
		Minute: 352,
	}

	dateTime, err := calculateTime(dateEvent)
	validateTime(dateTime, err, t)
}

func calculateTimeFromInvalidInput(t *testing.T) {
	ar := &omni.Ar{
		Date_Event: &omni.DateEvent{
			Year:   16,
			Month:  10,
			Day:    4,
			Minute: 352,
		},
		Date_Position: &omni.DatePosition{
			Year:      16,
			Month:     10,
			Day:       4,
			Minute:    352,
			Latitude:  47.8141,
			Longitude: -3.4793,
		},
		Assistance_Alert: &omni.AssistanceAlert{},
	}
	dateTime, err := calculateTime(ar)
	if err == nil {
		t.Errorf("Expected error to be non-nil")
	}
	if !dateTime.IsZero() {
		t.Errorf("Expected time to be zero, but got %v", dateTime)
	}
}

func validateTime(dateTime time.Time, err error, t *testing.T) {
	if err != nil {
		t.Errorf("Expected error to be nil, but got %v", err)
	}
	if dateTime.Year() != 2016 {
		t.Errorf("Expected year to be 2016, but got %v", dateTime.Year())
	}
	if dateTime.Month() != time.October {
		t.Errorf("Expected month to be 10, but got %v", dateTime.Month())
	}
	if dateTime.Day() != 4 {
		t.Errorf("Expected day of the month to be 4, but got %v", dateTime.Day())
	}
	if dateTime.Hour() != 5 {
		t.Errorf("Expected hour to be 5, but got %v", dateTime.Hour())
	}
	if dateTime.Minute() != 52 {
		t.Errorf("Expected minute to be 52, but got %v", dateTime.Minute())
	}
	if dateTime.Second() != 0 {
		t.Errorf("Expected seconds to be 0, but got %v", dateTime.Second())
	}
	if dateTime.Nanosecond() != 0 {
		t.Errorf("Expected year to be 0, but got %v", dateTime.Nanosecond())
	}
}

var (
	imei = "XXX234010030456"

	msgid = "1"

	datePosition = &omni.DatePosition{
		Year:      17,
		Month:     2,
		Day:       27,
		Minute:    1326,
		Latitude:  0.8917,
		Longitude: 104.0442,
	}

	move = &omni.MV{
		Heading: 170,
		Speed:   10,
	}

	source = &tms.EndPoint{
		Site: 10,
		Eid:  5500,
	}

	targetID = &tms.TargetID{
		Producer: &tms.SensorID{
			Site: source.Site,
			Eid:  source.Eid,
		},
		SerialNumber: &tms.TargetID_TimeSerial{
			TimeSerial: &tms.TimeSerialNumber{
				Seconds: time.Now().Unix(),
				Counter: 10,
			},
		},
	}
)
