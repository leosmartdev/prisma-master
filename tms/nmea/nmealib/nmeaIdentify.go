package nmealib

import (
	"errors"
	"fmt"
	. "prisma/tms"
	. "prisma/tms/nmea"
	"prisma/tms/util/ident"
)

type NmeaIdentify struct {
	sensorId *SensorID
	ttm_ids  map[uint32]TimeSerialNumber
}

func NewNmeaIdentify(sensor *SensorID) *NmeaIdentify {
	ret := &NmeaIdentify{
		sensorId: sensor,
		ttm_ids:  make(map[uint32]TimeSerialNumber),
	}
	return ret
}

//Nmea message is not used when generating target ID
func (nmeaIdentify *NmeaIdentify) NmeaGenerateTargetID(id *TargetID, nmea *Nmea) {

	nmeaIdentify.generateTargetID(id)

}

//Generate target ID for a target using the sensor ID and tsn
func (nmeaIdentify *NmeaIdentify) generateTargetID(id *TargetID) {

	id.Producer = nmeaIdentify.sensorId
	sn := ident.TimeSerialNumber()
	id.SerialNumber = &TargetID_TimeSerial{&sn}
}

func aisTrackID(mmsi uint32, producer *SensorID) string {
	return ident.
		With("mmsi", mmsi).
		With("site", producer.Site).
		With("eid", producer.Eid).
		Hash()
}

func genericTrackID(tsn TimeSerialNumber, producer *SensorID) string {
	return ident.
		With("seconds", tsn.Seconds).
		With("counter", tsn.Counter).
		With("site", producer.Site).
		With("eid", producer.Eid).
		Hash()
}

//Generate track ID for a track or target using nmea proto message
func (nmeaIdentify *NmeaIdentify) setIdentifiers(track *Track, nmea *Nmea) error {
	track.Producer = nmeaIdentify.sensorId
	if (nmea.Vdm != nil) && (nmea.GetVdm().M1371 != nil) {
		mmsi := nmea.GetVdm().GetM1371().Mmsi
		track.Id = aisTrackID(mmsi, track.Producer)
		track.RegistryId = ident.With("mmsi", mmsi).Hash()
		return nil
	} else if (nmea.Vdo != nil) && (nmea.GetVdo().M1371 != nil) {
		mmsi := nmea.GetVdo().GetM1371().Mmsi
		track.Id = aisTrackID(mmsi, track.Producer)
		track.RegistryId = ident.With("mmsi", mmsi).Hash()
		return nil
	} else if nmea.Ttm != nil {
		id, err := nmeaIdentify.getTtmID(track, nmea.GetTtm())
		if err != nil {
			return fmt.Errorf("%v can not be assigned a track id because: %v", nmea.OriginalString, err)
		}

		track.Id = id
		return nil
	}
	return errors.New("Could not determine track id")
}

func (nmeaIdentify *NmeaIdentify) getTtmID(track *Track, ttm *Ttm) (string, error) {
	var err error
	if !ttm.NumberValidity {
		err = errors.New("TTM with no number found")
		return "", err
	}

	lost := ttm.StatusValidity && ttm.Status == "L"
	tsn, exist := nmeaIdentify.ttm_ids[ttm.Number]

	// If we are not tracking the target that has been lost, don't assign
	// a track ID
	if !exist && lost {
		return "", errors.New("lost a target that was not being tracked")
	}
	if !exist {
		tsn = ident.TimeSerialNumber()
		nmeaIdentify.ttm_ids[ttm.Number] = tsn
	}
	// Stop tracking this target if it has been lost
	if lost {
		delete(nmeaIdentify.ttm_ids, ttm.Number)
	}
	id := genericTrackID(tsn, nmeaIdentify.sensorId)
	return id, nil
}
