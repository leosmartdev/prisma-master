package mongo

import (
	. "prisma/tms/db"
)

var (
	dbtrack_TimeField  = ResolveName(DBTrack{}, "Time")
	//dbtrack_TargetID   = ResolveName(DBTrack{}, "Target.Id")
	//dbtrack_DeviceType = ResolveName(DBTrack{}, "Target.Type")
	//dbtrack_Sensor     = ResolveName(DBTrack{}, "Track.Producer")
	//dbtrack_Site       = ResolveName(DBTrack{}, "Track.Producer.Site")

	dbRegEntry_ID       = ResolveName(DBRegistryEntry{}, "ID")
	dbRegEntry_Position = ResolveName(DBRegistryEntry{}, "Entry.Target.Position")
	dbRegEntry_Redirect = ResolveName(DBRegistryEntry{}, "Entry.Redirect")
)
