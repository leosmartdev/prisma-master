package nmealib

import (
	"errors"
	"flag"
	"fmt"
	"math"
	"strconv"
	"strings"

	. "prisma/tms"
	. "prisma/tms/devices"
	. "prisma/tms/geo"
	"prisma/tms/log"
	. "prisma/tms/nmea"
	"prisma/tms/util/ais"

	"github.com/StefanSchroeder/Golang-Ellipsoid/ellipsoid"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/golang/protobuf/ptypes/wrappers"
)

var (
	Conf_Radar_Latitude_Count  = flag.Int("radar_latitude_count", 1, "count of radar latitudes")
	Conf_Radar_Longitude_Count = flag.Int("radar_longitude_count", 1, "count of radar longitudes")
	Conf_Radar_Latitude        = flag.Float64("radar_latitude", 0.0, "radar latitude")
	Conf_Radar_Longitude       = flag.Float64("radar_longitude", 0.0, "radar longitude")
)

var geo = ellipsoid.Init("WGS84", ellipsoid.Degrees, ellipsoid.Nm, ellipsoid.LongitudeIsSymmetric, ellipsoid.BearingNotSymmetric)

type NmeaCopy struct {
	identify       *NmeaIdentify
	forced_coord   bool
	lat_long_valid bool
	device_lat     float64
	device_lon     float64
}

func NewNmeaCopy(ident *NmeaIdentify) *NmeaCopy {
	ret := &NmeaCopy{
		identify:       ident,
		forced_coord:   false,
		lat_long_valid: false,
	}

	if (*Conf_Radar_Latitude_Count > 0) || (*Conf_Radar_Longitude_Count > 0) {
		if (*Conf_Radar_Latitude_Count != 1) || (*Conf_Radar_Longitude_Count != 1) {
			log.Fatal("radar_latitude_count or radar_longitude_count is not 1")
		}
		if (*Conf_Radar_Latitude < -90) || (*Conf_Radar_Latitude > 90) {
			log.Fatal("Invalid radar latitude specified")
		}
		if (*Conf_Radar_Latitude < -180) || (*Conf_Radar_Latitude > 180) {
			log.Fatal("Invalid radar latitude specified")
		}
		ret.forced_coord = true
		ret.lat_long_valid = true
		ret.device_lat = *Conf_Radar_Latitude
		ret.device_lon = *Conf_Radar_Longitude
	}
	return ret
}

func (nmeaCopy *NmeaCopy) IngestPosition(nmea *Nmea) error {
	var err error
	if nmeaCopy.forced_coord {
		return nil
	}

	if nmea.GetRmc() != nil {
		rmc := nmea.GetRmc()
		if (rmc.LatitudeValidity) && (rmc.LatitudeDirectionValidity) && (rmc.LongitudeValidity) && (rmc.LongitudeDirectionValidity) {
			lat, err := Arcmin2Decimal(rmc.Latitude, rmc.LatitudeDirection)
			if err != nil {
				return err
			}
			lon, err := Arcmin2Decimal(rmc.Longitude, rmc.LongitudeDirection)
			if err != nil {
				return err
			}
			if (lat >= Latitude_Min) && (lat <= Latitude_Max) && (lon >= Longitude_Min) && (lon <= Longitude_Max) {
				nmeaCopy.device_lat = lat
				nmeaCopy.device_lon = lon
				nmeaCopy.lat_long_valid = true
			} else {
				err = errors.New("Rmc with lat/long but invalid")
			}
		}
	}
	return err
}

//Determine data type of this nmea
func (nmeaCopy *NmeaCopy) DetermineType(nmea *Nmea) DeviceType {
	if nmea.GetRmc() != nil {
		return DeviceType_GPS
	}
	if (nmea.GetVdm() != nil) || (nmea.GetVdo() != nil) {
		// Search and rescue transponders have an MMSI that starts with 97
		mmsi := ""
		if nmea.GetVdm() != nil {
			mmsi = strconv.Itoa(int(nmea.GetVdm().GetM1371().GetMmsi()))
		} else if nmea.GetVdo() != nil {
			mmsi = strconv.Itoa(int(nmea.GetVdo().GetM1371().GetMmsi()))
		}
		if strings.HasPrefix(mmsi, "97") {
			return DeviceType_SART
		}
		return DeviceType_AIS
	}
	if nmea.GetTtm() != nil {
		return DeviceType_Radar
	}
	return DeviceType_Unknown
}

func (nmeaCopy *NmeaCopy) PopulateTrack(track *Track, nmea *Nmea) error {

	err := nmeaCopy.identify.setIdentifiers(track, nmea)
	if err != nil {
		return err
	}

	target := &Target{
		Id:         &TargetID{},
		IngestTime: &timestamp.Timestamp{},
		Position:   &Point{},
		Course:     &wrappers.DoubleValue{},
		Heading:    &wrappers.DoubleValue{},
		Speed:      &wrappers.DoubleValue{},
		RateOfTurn: &wrappers.DoubleValue{},
		Nmea:       &Nmea{},
	}
	err_target := nmeaCopy.PopulateTarget(target, nmea)

	if err_target == nil {
		target.Nmea = nmea
		target.Type = nmeaCopy.DetermineType(nmea)
		track.Targets = append(track.Targets, target)
	}

	meta := &TrackMetadata{
		Time:       &timestamp.Timestamp{},
		IngestTime: &timestamp.Timestamp{},
		Nmea:       &Nmea{},
	}
	err_meta := nmeaCopy.PopulateMeta(meta, nmea)

	if err_meta == nil {
		meta.Nmea = nmea
		meta.Type = nmeaCopy.DetermineType(nmea)
		track.Metadata = append(track.Metadata, meta)
	}

	return nil
}

//Populate the target field for the track
func (nmeaCopy *NmeaCopy) PopulateTarget(target *Target, nmea *Nmea) error {

	target.Time = Now()
	target.IngestTime = Now()
	nmeaCopy.identify.NmeaGenerateTargetID(target.Id, nmea)

	if (nmea.GetVdm() != nil) && (nmea.GetVdm().GetM1371() != nil) {
		return nmeaCopy.targetPopulateM1371(target, nmea.GetVdm().GetM1371())
	} else if (nmea.GetVdo() != nil) && (nmea.GetVdo().GetM1371() != nil) {
		return nmeaCopy.targetPopulateM1371(target, nmea.GetVdo().GetM1371())
	} else if nmea.GetRmc() != nil {
		nmeaCopy.IngestPosition(nmea)
		return nmeaCopy.PopulateRMC(target, nmea.GetRmc())
	} else if nmea.GetTtm() != nil {
		return nmeaCopy.targetPopulateTTM(target, nmea.GetTtm())
	}

	return errors.New("No target info is in nmea")
}

//Populate the metadata field for the track
func (nmeaCopy *NmeaCopy) PopulateMeta(meta *TrackMetadata, nmea *Nmea) error {

	meta.Time = Now()
	meta.IngestTime = Now()

	if (nmea.GetVdm() != nil) && (nmea.GetVdm().GetM1371() != nil) {

		return nmeaCopy.metaPopulateM1371(meta, nmea.GetVdm().GetM1371())

	} else if (nmea.GetVdo() != nil) && (nmea.GetVdo().GetM1371() != nil) {

		return nmeaCopy.metaPopulateM1371(meta, nmea.GetVdo().GetM1371())

	} else if nmea.GetTtm() != nil {

		return nmeaCopy.metaPopulateTTM(meta, nmea.GetTtm())

	}

	err := errors.New("No metadata info is in nmea")
	return err
}

//Populate the mob field for the track
func (nmeaCopy *NmeaCopy) PopulateSafetyBcast(safetyBcast *SafetyBroadcast, nmea *Nmea) error {

	safetyBcast.Time = Now()
	safetyBcast.IngestTime = Now()

	if (nmea.GetVdm() != nil) && (nmea.GetVdm().GetM1371() != nil) && (nmea.GetVdm().GetM1371().MessageId == 14) {

		if nmea.GetVdm().GetM1371().GetSafetyBcast() != nil {
			safetyBcast.Mmsi = ais.FormatMMSI(int(nmea.GetVdm().GetM1371().Mmsi))
			safetyBcast.Text = nmea.GetVdm().GetM1371().GetSafetyBcast().Text
			return nil
		}

	} else if (nmea.GetVdo() != nil) && (nmea.GetVdo().GetM1371() != nil) && (nmea.GetVdo().GetM1371().MessageId == 14) {

		if nmea.GetVdo().GetM1371().GetSafetyBcast() != nil {
			safetyBcast.Mmsi = ais.FormatMMSI(int(nmea.GetVdo().GetM1371().Mmsi))
			safetyBcast.Text = nmea.GetVdo().GetM1371().GetSafetyBcast().Text
			return nil
		}

	}

	err := errors.New("No safety broadcast is in nmea")
	return err
}

//Populate ttm for metadata
func (nmeaCopy *NmeaCopy) metaPopulateTTM(meta *TrackMetadata, ttm *Ttm) error {

	if ttm.NameValidity {
		meta.Name = ttm.Name
		return nil
	}

	err := errors.New("No name info in ttm")
	return err

}

//Populate ttm for target
func (nmeaCopy *NmeaCopy) targetPopulateTTM(target *Target, ttm *Ttm) error {

	var err error

	if !ttm.NumberValidity {
		err = errors.New("No number in ttm")
		return err
	}

	var distance float64

	if !ttm.SpeedDistanceUnitsValidity {

		target.Speed.Value = ttm.Speed * 0.539956803
		distance = ttm.Distance * 0.539956803

	} else if ttm.SpeedDistanceUnits == "K" {

		target.Speed.Value = ttm.Speed * 0.539956803
		distance = ttm.Distance * 0.539956803

	} else if ttm.SpeedDistanceUnits == "N" {

		target.Speed.Value = ttm.Speed
		distance = ttm.Distance

	} else if ttm.SpeedDistanceUnits == "S" {

		target.Speed.Value = ttm.Speed * 0.868976242
		distance = ttm.Distance * 0.868976242

	} else {

		log.Warn("TTM message received with invalid init specifier %s. Default to knots", ttm.SpeedDistanceUnits)
		target.Speed.Value = ttm.Speed * 0.539956803
		distance = ttm.Distance * 0.539956803

	}

	if (ttm.CourseValidity) && (ttm.CourseRelative == "T") {

		target.Course.Value = ttm.Course

	}

	if (nmeaCopy.lat_long_valid) && (ttm.BearingValidity) {
		target.Position.Latitude, target.Position.Longitude = geo.At(nmeaCopy.device_lat, nmeaCopy.device_lon, distance, ttm.Bearing)
	}
	return nil
}

//PopulateRMC for target
func (nmeaCopy *NmeaCopy) PopulateRMC(target *Target, rmc *Rmc) error {
	var err error

	if (rmc.LatitudeValidity) && (rmc.LatitudeDirectionValidity) && (rmc.LongitudeValidity) && (rmc.LongitudeDirectionValidity) {

		lat, errLat := Arcmin2Decimal(rmc.Latitude, rmc.LatitudeDirection)

		if errLat == nil {
			target.Position.Latitude = lat
		} else {
			err = errLat
		}

		lon, errLon := Arcmin2Decimal(rmc.Longitude, rmc.LongitudeDirection)

		if errLon == nil {
			target.Position.Longitude = lon
		} else {
			err = errLon
		}
	}

	if rmc.SpeedOverGroundValidity {
		target.Speed.Value = rmc.SpeedOverGround
	}

	if rmc.CourseOverGroundValidity {
		target.Course.Value = rmc.CourseOverGround
	}
	return err

}

//Populate m1371 for target
func (nmeaCopy *NmeaCopy) targetPopulateM1371(target *Target, m1371 *M1371) error {

	target.Mmsi = strconv.FormatUint(uint64(m1371.GetMmsi()), 10)

	switch m1371.MessageId {

	case 1, 2, 3:

		m := m1371.GetPos()

		if m.Latitude != 0x3412140 {

			lat := CoordinatesMin2Deg(float64(m.Latitude))

			if (lat != Latitude_Na) && (lat > Latitude_Min) && (lat < Latitude_Max) {
				target.Position.Latitude = lat
			} else {
				return fmt.Errorf("Invalid latitude %+v in nmea sentence %+v", lat, m1371)
			}
		}

		if m.Longitude != 0x6791AC0 {

			lon := CoordinatesMin2Deg(float64(m.Longitude))

			if (lon != Longitude_Na) && (lon >= Longitude_Min) && (lon <= Longitude_Max) {
				target.Position.Longitude = lon
			} else {
				return fmt.Errorf("Invalid longitude %+v in nmea sentence %+v", lon, m1371)
			}
		}

		if m.CourseOverGround != 3600 {

			cog := float64(m.CourseOverGround) / 10.0

			if (cog != Course_Na) && (cog >= Course_Min) && (cog <= Course_Max) {
				target.Course.Value = cog
			}

		}

		if m.SpeedOverGround != 1023 {

			sog := float64(m.SpeedOverGround) / 10.0

			if m.SpeedOverGround == 1022 {
				target.Speed.Value = math.MaxFloat64
			} else if (sog != Speed_Na) && (sog >= Speed_Min) && (sog <= Speed_Max) {
				target.Speed.Value = sog
			}

		}

		if m.RateOfTurn != -128 {

			rot := float64(m.RateOfTurn)

			if m.RateOfTurn == 127 {
				target.RateOfTurn.Value = 720
			} else if m.RateOfTurn == -127 {
				target.RateOfTurn.Value = -720
			} else if (rot != Rot_Na) && (rot >= Rot_Min) && (rot <= Rot_Max) {
				rotSensor := rot / 4.733
				if rotSensor < 0 {
					rot = -(rotSensor * rotSensor)
				} else {
					rot = rotSensor * rotSensor
				}
				target.RateOfTurn.Value = rot
			}

		}

		if (m.TrueHeading != Heading_Na) && (m.TrueHeading <= Heading_Max) {
			target.Heading.Value = float64(m.TrueHeading)
		}

		return nil

	case 18:

		m := m1371.GetBPos()

		if m.Latitude != 0x3412140 {

			lat := CoordinatesMin2Deg(float64(m.Latitude))

			if (lat != Latitude_Na) && (lat >= Latitude_Min) && (lat <= Latitude_Max) {
				target.Position.Latitude = lat
			} else {
				return fmt.Errorf("Invalid latitude %+v in nmea sentence %+v", lat, m1371)
			}

		}

		if m.Longitude != 0x6791AC0 {

			lon := CoordinatesMin2Deg(float64(m.Longitude))

			if (lon != Longitude_Na) && (lon >= Longitude_Min) && (lon <= Longitude_Max) {
				target.Position.Longitude = lon
			} else {
				return fmt.Errorf("Invalid longitude %+v in nmea sentence %+v", lon, m1371)
			}
		}

		if m.CourseOverGround != 3600 {

			cog := float64(m.CourseOverGround) / 10.0

			if (cog != Course_Na) && (cog >= Course_Min) && (cog <= Course_Max) {
				target.Course.Value = cog
			}

		}

		if m.SpeedOverGround != 1023 {

			sog := float64(m.SpeedOverGround) / 10.0

			if m.SpeedOverGround == 1022 {
				target.Speed.Value = math.MaxFloat64
			} else if (sog != Speed_Na) && (sog >= Speed_Min) && (sog <= Speed_Max) {
				target.Speed.Value = sog
			}

		}

		if (m.TrueHeading != Heading_Na) && (m.TrueHeading <= Heading_Max) {
			target.Heading.Value = float64(m.TrueHeading)
		}

		return nil

	case 19:

		m := m1371.GetBExtPos()

		if m.Latitude != 0x3412140 {

			lat := CoordinatesMin2Deg(float64(m.Latitude))

			if (lat != Latitude_Na) && (lat >= Latitude_Min) && (lat <= Latitude_Max) {
				target.Position.Latitude = lat
			} else {
				return fmt.Errorf("Invalid latitude %+v in nmea sentence %+v", lat, m1371)
			}
		}

		if m.Longitude != 0x6791AC0 {

			lon := CoordinatesMin2Deg(float64(m.Longitude))

			if (lon != Longitude_Na) && (lon >= Longitude_Min) && (lon <= Longitude_Max) {
				target.Position.Longitude = lon
			} else {
				return fmt.Errorf("Invalid longitude %+v in nmea sentence %+v", lon, m1371)
			}
		}

		if m.CourseOverGround != 3600 {

			cog := float64(m.CourseOverGround) / 10.0

			if (cog != Course_Na) && (cog >= Course_Min) && (cog <= Course_Max) {
				target.Course.Value = cog
			}
		}

		if m.SpeedOverGround != 1023 {

			sog := float64(m.SpeedOverGround) / 10.0

			if m.SpeedOverGround == 1022 {
				target.Speed.Value = math.MaxFloat64
			} else if (sog != Speed_Na) && (sog >= Speed_Min) && (sog <= Speed_Max) {
				target.Speed.Value = sog
			}
		}

		if (m.TrueHeading != Heading_Na) && (m.TrueHeading <= Heading_Max) {
			target.Heading.Value = float64(m.TrueHeading)
		}

		return nil

	case 27:

		m := m1371.GetLongRangePos()

		lat := float64(m.Latitude) / (60.0 * 10.0)

		if (lat != Latitude_Na) && (lat >= Latitude_Min) && (lat <= Latitude_Max) {
			target.Position.Latitude = lat
		} else {
			return fmt.Errorf("Invalid latitude %+v in nmea sentence %+v", lat, m1371)
		}

		lon := float64(m.Longitude) / (60.0 * 10.0)

		if (lon != Longitude_Na) && (lon >= Longitude_Min) && (lon <= Longitude_Max) {
			target.Position.Longitude = lon
		} else {
			return fmt.Errorf("Invalid longitude %+v in nmea sentence %+v", lat, m1371)
		}

		if m.CourseOverGround != 360 {

			cog := float64(m.CourseOverGround)

			if (cog != Course_Na) && (cog >= Course_Min) && (cog <= Course_Max) {
				target.Course.Value = cog
			}
		}

		if m.SpeedOverGround != 63 {

			sog := float64(m.SpeedOverGround)

			if (sog != Speed_Na) && (sog >= Speed_Min) && (sog <= Speed_Max) {
				target.Speed.Value = sog
			}
		}
		return nil
	}

	return fmt.Errorf("No matched message id in m1371")
}

//Populate m1371 for metadata
func (nmeaCopy *NmeaCopy) metaPopulateM1371(meta *TrackMetadata, m1371 *M1371) error {

	switch m1371.MessageId {

	case 5:
		m := m1371.GetStaticVoyage()

		meta.Name = m.Name

		meta.CallSign = m.CallSign

		meta.Destination = m.Destination

		return nil

	case 19:
		m := m1371.GetBExtPos()

		if m.Name != "" {

			meta.Name = m.Name
			return nil

		}

		err := errors.New("No name in m1371")
		return err

	case 24:

		err := errors.New("No static data a and static data b in m1371")

		if m1371.StaticDataA != nil {

			m := m1371.GetStaticDataA()

			if m.Name != "" {
				meta.Name = m.Name
				err = nil
			}

		}

		if m1371.StaticDataB != nil {

			m := m1371.GetStaticDataB()

			if m.CallSign != "" {
				meta.CallSign = m.CallSign
				err = nil
			}

		}

		return err
	}
	err := errors.New("No matched message id in m1371")
	return err
}

//CoordinatesMin2Deg is a utility function that helps convert lat/lon from min to deg
func CoordinatesMin2Deg(min float64) float64 {
	Sign := 1.0

	if math.Signbit(min) {
		min = -min
		Sign = -1
	}

	degrees := float64(int(min / 600000))
	minutes := float64(min-600000*degrees) / 10000
	deg := degrees + minutes/60

	return Sign * deg
}
