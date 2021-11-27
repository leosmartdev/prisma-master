package main

import (
	"prisma/tms"
	"prisma/tms/devices"
	"prisma/tms/moc"
	"prisma/tms/util/ais"
)

func TargetInfoFromTrack(t *tms.Track) *moc.TargetInfo {
	if len(t.Targets) == 0 {
		return &moc.TargetInfo{}
	}
	tgt := t.Targets[0]

	var md *tms.TrackMetadata
	if len(t.Metadata) != 0 {
		md = t.Metadata[0]
	}

	info := &moc.TargetInfo{
		TrackId:    t.Id,
		DatabaseId: t.DatabaseId,
		RegistryId: t.RegistryId,
		Type:       tgt.Type.String(),
	}
	if md != nil {
		info.Name = md.Name
	}

	switch tgt.Type {
	case devices.DeviceType_Radar:
		info.RadarTarget = tgt.Nmea.Ttm.Number
	case devices.DeviceType_AIS, devices.DeviceType_TV32,
		devices.DeviceType_Orb, devices.DeviceType_SART:
		info.Mmsi = ais.FormatMMSI(int(tgt.Nmea.Vdm.M1371.Mmsi))
	case devices.DeviceType_OmnicomSolar:
		if tgt.Imei != nil {
			info.Imei = tgt.Imei.Value
		}
		if tgt.Nodeid != nil {
			info.IngenuNodeId = tgt.Nodeid.Value
		}
	case devices.DeviceType_SARSAT:
		info.SarsatBeacon = tgt.Sarmsg.SarsatAlert.Beacon
	}
	return info
}

func TargetInfoFromActivity(act *tms.MessageActivity) *moc.TargetInfo {

	info := &moc.TargetInfo{
		TrackId:    act.ActivityId,
		DatabaseId: act.DatabaseId,
		RegistryId: act.RegistryId,
		Type:       act.Type.String(),
	}
	if act.Imei != nil {
		info.Imei = act.Imei.Value
	} else if act.NodeId != nil {
		info.IngenuNodeId = act.NodeId.Value
	}

	return info
}
