package lib

import (
	"errors"
	"fmt"
	"reflect"

	"prisma/tms"
	"prisma/tms/devices"
	"prisma/tms/sar"
	"prisma/tms/util/ident"
)

func Process(sarmsg *sar.SarsatMessage, me *tms.SensorID) (*tms.Track, error) {

	track, err := populateSarTrack(sarmsg, me)
	if err != nil {
		return nil, err
	}
	return track, nil
}

func ProcessActivity(sarmsg *sar.SarsatMessage, me *tms.SensorID) (*tms.MessageActivity, error) {
	seconds, counter := ident.TSN()
	id := ident.
		With("seconds", seconds).
		With("counter", counter).
		Hash()
	act := &tms.MessageActivity{
		ActivityId: id,
		Time:       tms.Now(),
		MetaData:   &tms.MessageActivity_Sarsat{sarmsg},
		Type:       devices.DeviceType_SARSAT,
	}
	return act, nil
}

//
func sarAlertTrackID(beacon *sar.Beacon) string {
	return ident.With("beaconID", beacon.HexId).Hash()
}

func sarAlertDefaultTrackID(messagetype sar.SarsatMessage_MessageType, producer *tms.SensorID) string {
	return ident.
		With("MessageType", messagetype).
		With("site", producer.Site).
		With("eid", producer.Eid).
		Hash()
}

func sarAlertRegistryID(beacon *sar.Beacon) string {
	return ident.With("sarsatBeaconID", beacon.HexId).Hash()
}

func sarFreeFormTrackID(MessageNumber int32, remoteName, localName string, producer *tms.SensorID) string {
	return ident.
		With("MessageNumber", MessageNumber).
		With("RemoteName", remoteName).
		With("LocalName", localName).
		With("site", producer.Site).
		With("eid", producer.Eid).
		Hash()
}

func generateTargetID(producer *tms.SensorID) *tms.TargetID {

	sn := ident.TimeSerialNumber()

	return &tms.TargetID{
		Producer:     producer,
		SerialNumber: &tms.TargetID_TimeSerial{&sn},
	}
}

func populateTargetDefaultSarMessage(producer *tms.SensorID, sarmsg *sar.SarsatMessage) *tms.Target {

	return &tms.Target{
		Id:         generateTargetID(producer),
		Type:       devices.DeviceType_SARSAT,
		Time:       tms.Now(),
		IngestTime: tms.Now(),
		Sarmsg:     sarmsg,
	}
}

func populateTargetResolvedAlert(producer *tms.SensorID, ResolvedAltMsg *sar.ResolvedAlert) (*tms.Target, error) {
	target := &tms.Target{
		Id:         generateTargetID(producer),
		Type:       devices.DeviceType_SARSAT,
		Time:       tms.Now(),
		IngestTime: tms.Now(),
		Position:   &tms.Point{},
	}

	if ResolvedAltMsg.CompositeLocation != nil {
		target.Position.Latitude = ResolvedAltMsg.CompositeLocation.Latitude
		target.Position.Longitude = ResolvedAltMsg.CompositeLocation.Longitude
	} else {
		return nil, fmt.Errorf("error: no composite location in a sarsat resolved alert")
	}

	return target, nil
}

func populateTargetIncidentAlert(producer *tms.SensorID, IncidentAltMsg *sar.IncidentAlert) (*tms.Target, error) {

	target := &tms.Target{
		Id:         generateTargetID(producer),
		Type:       devices.DeviceType_SARSAT,
		IngestTime: tms.Now(),
		Time:       tms.Now(),
		Positions:  []*tms.Point{},
	}

	for _, elemental := range IncidentAltMsg.Elemental {
		for _, doppler := range elemental.Doppler {
			point := &tms.Point{
				Latitude:  doppler.DopplerPosition.Latitude,
				Longitude: doppler.DopplerPosition.Longitude,
			}
			target.Positions = append(target.Positions, point)
		}
	}

	if IncidentAltMsg.MeoElemental != nil {
		point := &tms.Point{
			Latitude:  IncidentAltMsg.MeoElemental.Doa.DoaPosition.Latitude,
			Longitude: IncidentAltMsg.MeoElemental.Doa.DoaPosition.Longitude,
		}
		target.Positions = append(target.Positions, point)
	}

	if IncidentAltMsg.Encoded != nil {
		point := &tms.Point{
			Latitude:  IncidentAltMsg.Encoded.Latitude,
			Longitude: IncidentAltMsg.Encoded.Longitude,
		}
		target.Positions = append(target.Positions, point)
	}

	if len(target.Positions) == 0 {
		return nil, fmt.Errorf("error: no position included in the incident alert")
	}
	if len(target.Positions) == 1 {
		target.Position = target.Positions[0]
		target.Positions = nil
	}

	return target, nil
}
func populateSarTrack(sarmsg *sar.SarsatMessage, me *tms.SensorID) (*tms.Track, error) {
	track := &tms.Track{}
	if sarmsg == nil || reflect.DeepEqual(*sarmsg, sar.SarsatMessage{}) {
		return nil, fmt.Errorf("Unable to process empty sar message structure")
	}

	if sarmsg.GetSarsatAlert() != nil {
		if sarmsg.SarsatAlert.GetBeacon() == nil {
			return nil, fmt.Errorf("sarsat message %+v has no beacon hexID", sarmsg.SarsatAlert)
		}
		// we will populate the messages here for compound position track and alert shall
		// be sent to tgwad. for the others multipoint. for freeform message just populate methadata
		track.Id = sarAlertTrackID(sarmsg.SarsatAlert.Beacon)
		track.RegistryId = sarAlertRegistryID(sarmsg.SarsatAlert.Beacon)
		mmsi := sarmsg.SarsatAlert.Beacon.GetMmsi()
		if sarmsg.SarsatAlert.ResolvedAlertMessage != nil {
			//populate target
			target, err := populateTargetResolvedAlert(me, sarmsg.SarsatAlert.ResolvedAlertMessage)
			if err != nil {
				return nil, err
			}
			target.Sarmsg = sarmsg
			target.Mmsi = mmsi
			track.Targets = append(track.Targets, target)
		} else if sarmsg.SarsatAlert.IncidentAlertMessage != nil {
			target, err := populateTargetIncidentAlert(me, sarmsg.SarsatAlert.IncidentAlertMessage)
			if err != nil {
				return nil, err
			}
			target.Sarmsg = sarmsg
			target.Mmsi = mmsi
			track.Targets = append(track.Targets, target)
		} else if sarmsg.SarsatAlert.UnlocatedAlertMessage != nil {
			target := &tms.Target{
				Type:       devices.DeviceType_SARSAT,
				IngestTime: tms.Now(),
				Time:       tms.Now(),
				Sarmsg:     sarmsg,
			}
			track.Targets = append(track.Targets, target)
		} else {
			return nil, errors.New("unknown SARSAT alert message type")
		}
	} else if sarmsg.FreeFormMessage != nil {
		return nil, errors.New("free form message not implemented")
		// FIXME: A free form message is not a track and should probably
		// be handled as a notification
		/*
			track.Id = sarFreeFormTrackID(sarmsg.MessageNumber, sarmsg.RemoteName, sarmsg.LocalName, me)

			meta := &tms.TrackMetadata{
				IngestTime: tms.Now(),
				// FIXME: Time is not supposed to be ingest time, might have to look into this in the future
				Time:   tms.Now(),
				Sarmsg: sarmsg,
			}
			track.Metadata = []*tms.TrackMetadata{meta}
		*/
	} else if sarmsg.MessageType == sar.SarsatMessage_UNKNOWN {
		track.Id = sarAlertDefaultTrackID(sarmsg.MessageType, me)
		track.Targets = append(track.Targets, populateTargetDefaultSarMessage(me, sarmsg))
	} else {
		return nil, errors.New("cannot handle SARSAT message")
	}
	return track, nil
}
