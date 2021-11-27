package vts

import (
	"fmt"
	"prisma/tms"
	"prisma/tms/devices"
	"prisma/tms/nmea"
	"prisma/tms/util/ident"
	"prisma/tms/vtsx"
	"strconv"
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"
)

func populateTrackAIS(a AISTrack, me *tms.SensorID) (*tms.Track, error) {
	track := &tms.Track{}
	track.Id = ident.
		With("mmsi", a.MMSI).
		With("site", me.Site).
		With("eid", me.Eid).
		Hash()
	track.RegistryId = ident.With("mmsi", a.MMSI).Hash()

	return track, nil
}

func PopulateTrackPositionAIS(a AISTrack, me *tms.SensorID) (*tms.Track, error) {
	track, err := populateTrackAIS(a, me)
	if err != nil {
		return nil, err
	}

	sn := ident.TimeSerialNumber()
	mmsi, err := strconv.ParseUint(a.MMSI, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid MMSI %v: %v", a.MMSI, err)
	}

	updateTime, err := time.Parse(time.RFC3339, a.Timestamp)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp %v: %v", a.Timestamp, err)
	}

	track.Targets = []*tms.Target{{
		Id: &tms.TargetID{
			Producer:     me,
			SerialNumber: &tms.TargetID_TimeSerial{&sn},
		},
		Type:       devices.DeviceType_VTSAIS,
		IngestTime: tms.Now(),
		Time:       tms.Now(),
		UpdateTime: tms.ToTimestamp(updateTime),
		Speed:      &wrappers.DoubleValue{Value: a.Speed},
		Course:     &wrappers.DoubleValue{Value: a.Course},
		Heading:    &wrappers.DoubleValue{Value: a.Heading},
		Position: &tms.Point{
			Latitude:  a.Lat,
			Longitude: a.Long,
		},
		Nmea: &nmea.Nmea{
			Vdm: &nmea.Vdm{
				M1371: &nmea.M1371{
					Mmsi: uint32(mmsi),
					Pos: &nmea.M1371_Position{
						NavigationalStatus: uint32(a.NavStatus),
					},
				},
			},
		},
	}}
	return track, nil
}

func PopulateTrackVesselAIS(a AISTrack, me *tms.SensorID) (*tms.Track, error) {
	track, err := populateTrackAIS(a, me)
	if err != nil {
		return nil, err
	}

	mmsi, err := strconv.ParseUint(a.MMSI, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid MMSI %v: %v", a.MMSI, err)
	}

	imoNumber, err := strconv.ParseUint(a.IMONumber, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid IMO number %v: %v", a.IMONumber, err)
	}

	// https://www.navcen.uscg.gov/?pageName=AISMessagesAStatic
	// FIXME: not sure if this is correct
	// etaData := make([]byte, 4)
	// binary.LittleEndian.PutUint32(etaData, uint32(a.ETA))
	// r := sar.NewBitReader(etaData)
	// etaMinute := r.ReadN(6)
	// etaHour := r.ReadN(5)
	// etaDay := r.ReadN(5)
	// etaMonth := r.ReadN(4)
	etaMinute := 0
	etaHour := 0
	etaDay := 0
	etaMonth := 0

	// sizeData := make([]byte, 4)
	// binary.LittleEndian.PutUint32(sizeData, uint32(a.Size))
	// r = sar.NewBitReader(sizeData)
	// dimBow := r.ReadN(9)
	// dimStern := r.ReadN(9)
	// dimPort := r.ReadN(6)
	// dimStarboard := r.ReadN(6)
	dimBow := 0
	dimStern := 0
	dimPort := 0
	dimStarboard := 0

	track.Metadata = []*tms.TrackMetadata{{
		Type:       devices.DeviceType_VTSAIS,
		IngestTime: tms.Now(),
		Time:       tms.Now(),
		Name:       a.NameOfShip,
		Nmea: &nmea.Nmea{
			Vdm: &nmea.Vdm{
				M1371: &nmea.M1371{
					Mmsi: uint32(mmsi),
					StaticVoyage: &nmea.M1371_Static{
						CallSign:         a.Callsign,
						ShipAndCargoType: uint32(a.TypeOfShip),
						Draught:          uint32(a.Draught * 10), // units are in tenths
						ImoNumber:        uint32(imoNumber),
						Destination:      a.Destination,
						EtaMonth:         uint32(etaMonth),
						EtaDay:           uint32(etaDay),
						EtaHour:          uint32(etaHour),
						EtaMinute:        uint32(etaMinute),
						DimBow:           uint32(dimBow),
						DimStern:         uint32(dimStern),
						DimPort:          uint32(dimPort),
						DimStarboard:     uint32(dimStarboard),
					},
				},
			},
		},
	}}
	return track, nil
}

func PopulateTrackRadar(t TrackerTrack, me *tms.SensorID) (*tms.Track, error) {
	track := &tms.Track{}
	track.Id = ident.
		With("trackerId", t.TrackerID).
		With("site", me.Site).
		With("eid", me.Eid).
		Hash()

	updateTime, err := time.Parse(time.RFC3339, t.Timestamp)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp %v: %v", t.Timestamp, err)
	}
	sn := ident.TimeSerialNumber()

	state := ""
	switch t.TrackState {
	case 0:
		state = "Q" // query (acquiring?)
	case 1:
		state = "T" // tracking
	case 2:
		state = "L" // lost
	case 3:
		state = "L" // lost by deletion
	}

	acqType := ""
	switch t.AcquisitionType {
	case 0:
		acqType = "M" // manual
	case 1:
		acqType = "A" // auto
	}

	track.Targets = []*tms.Target{{
		Id: &tms.TargetID{
			Producer:     me,
			SerialNumber: &tms.TargetID_TimeSerial{&sn},
		},
		Type:       devices.DeviceType_VTSRadar,
		IngestTime: tms.Now(),
		Time:       tms.Now(),
		UpdateTime: tms.ToTimestamp(updateTime),
		Speed:      &wrappers.DoubleValue{Value: t.Speed},
		Course:     &wrappers.DoubleValue{Value: t.Course},
		Heading:    &wrappers.DoubleValue{Value: t.Heading},
		Position: &tms.Point{
			Latitude:  t.Lat,
			Longitude: t.Long,
		},
		Nmea: &nmea.Nmea{
			Ttm: &nmea.Ttm{
				Number:                     uint32(t.TrackID),
				NumberValidity:             true,
				Bearing:                    t.Bearing,
				BearingValidity:            true,
				SpeedDistanceUnits:         "N", // nautical miles
				SpeedDistanceUnitsValidity: true,
				Speed:                      t.Speed,
				SpeedValidity:              true,
				Distance:                   t.Range,
				DistanceValidity:           true,
				Status:                     state,
				StatusValidity:             state != "",
				AcquisitionType:            acqType,
				AcquisitionTypeValidity:    acqType != "",
			},
		},
		VtsRadar: &vtsx.VTSRadar{
			TrackMMSI: t.TrackMMSI,
			TrackName: t.TrackName,
			Quality:   uint32(t.Quality),
		},
	}}
	return track, nil
}
