package main

import (
	"fmt"
	"prisma/tms"
	"prisma/tms/devices"
	"prisma/tms/omnicom"
	"prisma/tms/tmsg"
	"prisma/tms/util/ident"
	"reflect"
	"strconv"
)

//ProcessOmnicom should process all omnicom messages
func ProcessOmnicom(omni *omnicom.Omni, dev devices.DeviceType) ([]*tms.Track, *tms.MessageActivity, error) {
	if omni == nil {
		return nil, nil, fmt.Errorf("No payload in iridium message")
	}
	var tracks []*tms.Track

	if omni.GetHpr() != nil {
		HprTracks, activity := ProcessOmnicomHRP(omni.GetHpr(), dev)
		if len(HprTracks) != 0 {
			tracks = append(tracks, HprTracks...)
		}
		return tracks, activity, nil
	} else if omni.GetSprs() != nil {
		track, activity := ProcessOmnicomSPRS(omni.GetSprs(), dev)
		tracks = append(tracks, track)
		return tracks, activity, nil
	} else if omni.GetSpr() != nil {
		track, activity := ProcessOmnicomSPR(omni.GetSpr(), dev)
		tracks = append(tracks, track)
		return tracks, activity, nil
	} else if omni.GetGp() != nil {
		track, activity := ProcessOmnicomGP(omni.GetGp(), dev)
		tracks = append(tracks, track)
		return tracks, activity, nil
	} else if omni.GetAr() != nil {
		track, activity := ProcessOmnicomAR(omni.GetAr(), dev)
		tracks = append(tracks, track)
		return tracks, activity, nil
	} else if omni.GetGa() != nil {
		track, activity := ProcessOmnicomGA(omni.GetGa(), dev)
		tracks = append(tracks, track)
		return tracks, activity, nil
	}
	return nil, nil, fmt.Errorf("Omnicom Message not supported: %+v", reflect.TypeOf(omni.GetOmnicom()))
}

func createActivityFromOmnicom(Omnicom *omnicom.Omni, dev devices.DeviceType) *tms.MessageActivity {

	return &tms.MessageActivity{
		Time:     tms.Now(),
		MetaData: &tms.MessageActivity_Omni{Omnicom},
		Type:     dev,
	}

}

func GenerateTargetID() *tms.TargetID {
	sec, ctr := ident.TSN()

	targetID := &tms.TargetID{
		Producer: &tms.SensorID{
			Site: tmsg.GClient.Local().Site,
			Eid:  tmsg.GClient.Local().Eid,
		},
		SerialNumber: &tms.TargetID_TimeSerial{
			TimeSerial: &tms.TimeSerialNumber{
				Seconds: sec,
				Counter: ctr,
			},
		},
	}
	return targetID
}

//ProcessOmnicomHRP processes Historical position reports into tracks and alerts
func ProcessOmnicomHRP(hpr *omnicom.Hpr, dev devices.DeviceType) ([]*tms.Track, *tms.MessageActivity) {

	var tracks []*tms.Track
	var activity *tms.MessageActivity

	for _, datareport := range hpr.Data_Report {

		track := createTrackFromOmnicom(datareport.Date_Position, datareport.Move, GenerateTargetID(), dev)

		if len(track.Targets) != 0 {
			track.Targets[0].ReportPeriod = datareport.Period
		}
		tracks = append(tracks, track)
	}

	activity = createActivityFromOmnicom(&omnicom.Omni{Omnicom: &omnicom.Omni_Hpr{hpr}}, dev)

	if hpr.Msg_ID != 0 {
		activity.RequestId = strconv.Itoa(int(hpr.Msg_ID))
	}

	return tracks, activity
}

//ProcessOmnicomSPRS processes Single position solar report message into  a track
func ProcessOmnicomSPRS(sprs *omnicom.Sprs, dev devices.DeviceType) (*tms.Track, *tms.MessageActivity) {
	track := createTrackFromOmnicom(sprs.Date_Position, sprs.Move, GenerateTargetID(), dev)

	if len(track.Targets) != 0 {
		track.Targets[0].ReportPeriod = sprs.Period
		track.Targets[0].Omnicom = &omnicom.Omni{Omnicom: &omnicom.Omni_Sprs{sprs}}
	}

	return track, createActivityFromOmnicom(&omnicom.Omni{Omnicom: &omnicom.Omni_Sprs{sprs}}, dev)
}

//ProcessOmnicomSPR processes Single position report message into a track
func ProcessOmnicomSPR(spr *omnicom.Spr, dev devices.DeviceType) (*tms.Track, *tms.MessageActivity) {
	track := createTrackFromOmnicom(spr.Date_Position, spr.Move, GenerateTargetID(), dev)
	if len(track.Targets) != 0 {
		track.Targets[0].ReportPeriod = spr.Period
		track.Targets[0].Omnicom = &omnicom.Omni{Omnicom: &omnicom.Omni_Spr{spr}}
	}

	return track, createActivityFromOmnicom(&omnicom.Omni{Omnicom: &omnicom.Omni_Spr{spr}}, dev)
}

//ProcessOmnicomGP processes Global parameters into track and activity
func ProcessOmnicomGP(gp *omnicom.Gp, dev devices.DeviceType) (*tms.Track, *tms.MessageActivity) {

	track := createTrackFromOmnicom(gp.Date_Position, nil, GenerateTargetID(), dev)

	activity := createActivityFromOmnicom(&omnicom.Omni{Omnicom: &omnicom.Omni_Gp{gp}}, dev)

	if gp.ID_Msg != 0 {
		activity.RequestId = strconv.Itoa(int(gp.ID_Msg))
	}

	return track, activity
}

//ProcessOmnicomGA processes Geofence Ack into track and activity
func ProcessOmnicomGA(ga *omnicom.Ga, dev devices.DeviceType) (*tms.Track, *tms.MessageActivity) {

	track := createTrackFromOmnicom(ga.Date_Position, nil, GenerateTargetID(), dev)

	activity := createActivityFromOmnicom(&omnicom.Omni{Omnicom: &omnicom.Omni_Ga{ga}}, dev)

	return track, activity
}

//ProcessOmnicomAR processes Alert reports into tracks and alerts
func ProcessOmnicomAR(ar *omnicom.Ar, dev devices.DeviceType) (*tms.Track, *tms.MessageActivity) {

	var track *tms.Track

	if ar.Extention_Bit_Move != 0 {
		track = createTrackFromOmnicom(ar.Date_Position, ar.Move, GenerateTargetID(), dev)
	} else {
		track = createTrackFromOmnicom(ar.Date_Position, nil, GenerateTargetID(), dev)
	}

	activity := createActivityFromOmnicom(&omnicom.Omni{Omnicom: &omnicom.Omni_Ar{ar}}, dev)

	if ar.Msg_ID != 0 {
		activity.RequestId = strconv.Itoa(int(ar.Msg_ID))
	}

	return track, activity
}
