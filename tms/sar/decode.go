//Package sar implements specifications for the sarsat beacon identifier.
//The specifications can be found here:
// https://team.technosci.com:8442/browse/CONV-801
package sar

import (
	"encoding/hex"
	fmt "fmt"
	"strconv"
	"strings"
)

// Specifications for the beacon identifier can be found here:
//
// https://team.technosci.com:8442/browse/CONV-801

// 3-bit protocol codes
const (
	pcOrbitographyProtocol = 0 // 0b000
	pcAviationUser         = 1 // 0b001
	pcMaritimeUser         = 2 // 0b010
	pcSerialUser           = 3 // 0b011
	pcNationalUser         = 4 // 0b100
	pcRadioCallSignUser    = 6 // 0b110
	pcTestUser             = 7 // 0b111
)

// 4-bit protocol codes
const (
	pcMaritimeStandardLocation     = 2  // 0b0010
	pcAviationStandardLocation     = 3  // 0b0011
	pcSerialEltStandardLocation    = 4  // 0b0100
	pcSerialEltAodStandardLocation = 5  // 0b0101
	pcSerialEpribStandardLocation  = 6  // 0b0110
	pcSerialPlbStandardLocation    = 7  // 0b0111
	pcNationalLocationElt          = 8  // 0b1000
	pcNationalLocationEprib        = 10 // 0b1010
	pcNationalLocationPlb          = 11 // 0b1011
	pcShipSecurityStandardLocation = 12 // 0b1100
	pcRlsLocation                  = 13 // 0b1101
)

// DecodeHexID ...
// http://www.cospas-sarsat.int/en/beacons-pro/beacon-message-decode-program-txsep/beacon-message-decode-program
func DecodeHexID(hexID string) (*Beacon, error) {
	if len(hexID) != 15 {
		return nil, fmt.Errorf("invalid hex id length: %v", len(hexID))
	}
	// DecodeString does not like "odd" lengths, add a zero to make it even.
	bytes, err := hex.DecodeString(hexID + "0")
	if err != nil {
		return nil, err
	}

	id := NewBitReader(bytes)
	protocolFlag := id.Read()   // bit 26
	countryCode := id.ReadN(10) // bits 27-36

	b := &Beacon{
		HexId:       hexID,
		CountryCode: countryCode,
	}

	if protocolFlag == 1 {
		if err := userAndUserLocationProtocols(id, b); err != nil {
			return nil, err
		}
	} else {
		if err := standardAndNationalLocationProtocols(id, b); err != nil {
			return nil, err
		}
	}
	return b, nil
}

func userAndUserLocationProtocols(id *BitReader, b *Beacon) error {
	protocolCode := id.ReadN(3) // bits 37-39
	switch protocolCode {
	case pcAviationUser:
		return aviationUser(id, b)
	case pcMaritimeUser:
		return maritimeUser(id, b)
	case pcSerialUser:
		return serialUser(id, b)
	case pcRadioCallSignUser:
		return radioCallSignUser(id, b)
	case pcNationalUser:
		return nationalUser(id, b)
	case pcTestUser:
		return testUser(id, b)
	case pcOrbitographyProtocol:
		return orbitographyProtocol(id, b)
	default:
		return fmt.Errorf("unhandled protocol code: %v", protocolCode)
	}
}

func aviationUser(id *BitReader, b *Beacon) error {
	au := &AviationUser{}
	marking, err := id.ReadBaudotN(7) // bits 40-81
	if err != nil {
		return err
	}

	au.AircraftRegistrationMarking = strings.TrimSpace(marking)
	au.SpecificEltNumber = id.ReadN(2)            // bits 82-83
	au.AuxiliaryRadioLocatingDevice = id.ReadN(2) // bits 84-85

	b.Protocol = &Beacon_AviationUser{
		AviationUser: au,
	}
	return nil
}

func maritimeUser(id *BitReader, b *Beacon) error {
	mu := &MaritimeUser{}
	char6, err := id.ReadBaudotN(6) // bits 40-75
	if err != nil {
		return err
	}

	// If the MMSI doesn't parse into digits, it is assumed to be a
	// radio call sign
	var mmsi uint64
	var callSign string
	digits6, err := strconv.ParseUint(char6, 10, 0)
	if err != nil {
		callSign = strings.TrimSpace(char6)
	} else {
		mmsi = combineMmsi(b.CountryCode, digits6)
	}

	beaconN, err := id.ReadBaudotN(1) // bits 76-81
	if err != nil {
		return err
	}

	mu.Mmsi = mmsi
	mu.CallSign = callSign
	mu.SpecificBeaconNumber = beaconN
	id.SkipN(2)                                   // bits 82-83
	mu.AuxiliaryRadioLocatingDevice = id.ReadN(2) // bits 84-85

	b.Protocol = &Beacon_MaritimeUser{
		MaritimeUser: mu,
	}
	return nil
}

func serialUser(id *BitReader, b *Beacon) error {
	su := &SerialUser{}
	su.BeaconType = id.ReadN(3) // bits 40-42
	hasCertNumber := id.Read()  // bit 43

	switch su.BeaconType {
	case 3: // 0b011 -- A2.5.2 Aircraft 24-bit Address
		su.AircraftAddress = id.ReadN(24)  // bits 44-67
		su.SpecificEltNumber = id.ReadN(6) // bits 68-73
		if hasCertNumber == 1 {
			su.CertificateNumber = id.ReadN(10) // bits 74-83
		} else {
			su.NationalUse = id.ReadN(10) // bits 74-83
		}
	case 1: // 0b001 -- A2.5.3 Aircraft Operator Designator and Serial Number
		aod, err := id.ReadBaudotN(3) // bits 44-61
		if err != nil {
			return err
		}
		su.AircraftOperatorDesignator = aod
		su.SerialNumber = id.ReadN(12) // bits 62-73
		if hasCertNumber == 1 {
			su.CertificateNumber = id.ReadN(10) // bits 74-83
		} else {
			su.NationalUse = id.ReadN(10) // bits 74-83
		}
	default: // A2.5.1 Serial Number
		su.SerialNumber = id.ReadN(20) // bits 44-63
		if hasCertNumber == 1 {
			su.NationalUse = id.ReadN(10)       // bits 64-73
			su.CertificateNumber = id.ReadN(10) // bits 74-83
		} else {
			su.NationalUse = id.ReadN(20) // bits 64-83
		}
	}
	su.AuxiliaryRadioLocatingDevice = id.ReadN(2) // bits 84-85

	b.Protocol = &Beacon_SerialUser{
		SerialUser: su,
	}
	return nil
}

func radioCallSignUser(id *BitReader, b *Beacon) error {
	ru := &RadioCallSignUser{}

	sign1, err := id.ReadBaudotN(4) // bits 40-63
	if err != nil {
		return err
	}

	// For BCD, format to hex string and then convert any "A" characters
	// to spaces
	sign2bcd := id.ReadN(12) // bits 64-75
	sign2 := fmt.Sprintf("%03x", sign2bcd)
	sign2 = strings.TrimSpace(strings.Replace(sign2, "a", " ", -1))

	beaconN, err := id.ReadBaudotN(1) // bits 76-81
	if err != nil {
		return err
	}
	id.SkipN(2) // bits 82-83

	ru.CallSign = strings.TrimSpace(sign1 + sign2)
	ru.SpecificBeaconNumber = beaconN
	ru.AuxiliaryRadioLocatingDevice = id.ReadN(2) // bits 84-85

	b.Protocol = &Beacon_RadioCallSignUser{
		RadioCallSignUser: ru,
	}
	return nil
}

func nationalUser(id *BitReader, b *Beacon) error {
	nu := &NationalUser{}
	nu.NationalUse = id.ReadN(45) // bits 40-85

	b.Protocol = &Beacon_NationalUser{
		NationalUser: nu,
	}
	return nil
}

func testUser(id *BitReader, b *Beacon) error {
	nu := &NationalUser{}
	nu.NationalUse = id.ReadN(45) // bits 40-85

	b.Protocol = &Beacon_TestUser{
		TestUser: nu,
	}
	return nil
}

func orbitographyProtocol(id *BitReader, b *Beacon) error {
	op := &OrbitographyProtocol{}
	op.OrbitographyProtocol = id.ReadN(45) // bits 40-85

	b.Protocol = &Beacon_OrbitographyProtocol{
		OrbitographyProtocol: op,
	}
	return nil
}

func standardAndNationalLocationProtocols(id *BitReader, b *Beacon) error {
	protocolCode := id.ReadN(4) // bits 37-40
	switch protocolCode {
	case pcMaritimeStandardLocation:
		return maritimeStandardLocation(id, b)
	case pcAviationStandardLocation:
		return aviationStandardLocation(id, b)
	case pcSerialEpribStandardLocation,
		pcSerialEltStandardLocation,
		pcSerialEltAodStandardLocation,
		pcSerialPlbStandardLocation:
		return serialStandardLocation(id, b, protocolCode)
	case pcShipSecurityStandardLocation:
		return shipSecurityStandardLocation(id, b)
	case pcNationalLocationEprib,
		pcNationalLocationElt,
		pcNationalLocationPlb:
		return nationalLocation(id, b, protocolCode)
	case pcRlsLocation:
		return rlsLocation(id, b)
	default:
		return fmt.Errorf("unhandled protocol code: %v", protocolCode)
	}
}

func maritimeStandardLocation(id *BitReader, b *Beacon) error {
	ml := &MaritimeStandardLocation{}
	mmsi6 := id.ReadN(20) // bits 41-60
	mmsi := combineMmsi(b.CountryCode, mmsi6)

	ml.Mmsi = mmsi
	ml.SpecificBeaconNumber = id.ReadN(4) // bits 61-64

	b.Protocol = &Beacon_MaritimeStandardLocation{
		MaritimeStandardLocation: ml,
	}
	return nil
}

func aviationStandardLocation(id *BitReader, b *Beacon) error {
	al := &AviationStandardLocation{}
	al.AircraftAddress = id.ReadN(24) // bits 41-64
	b.Protocol = &Beacon_AviationStandardLocation{
		AviationStandardLocation: al,
	}
	return nil
}

func serialStandardLocation(id *BitReader, b *Beacon, protocolCode uint64) error {
	sl := &SerialStandardLocation{
		BeaconType: protocolCode,
	}
	switch protocolCode {
	case pcSerialEltStandardLocation,
		pcSerialEpribStandardLocation,
		pcSerialPlbStandardLocation:
		sl.CertificateNumber = id.ReadN(10) // bits 41-50
		sl.SerialNumber = id.ReadN(14)      // bits 51-64
	case pcSerialEltAodStandardLocation:
		aod, err := id.ReadBaudot5N(3) // bits 41-55
		if err != nil {
			return err
		}
		sl.AircraftOperatorDesignator = aod
		sl.SerialNumber = id.ReadN(9) // bits 56-64
	}

	b.Protocol = &Beacon_SerialStandardLocation{
		SerialStandardLocation: sl,
	}
	return nil
}

func shipSecurityStandardLocation(id *BitReader, b *Beacon) error {
	sl := &ShipSecurityLocation{}
	mmsi6 := id.ReadN(20) // bits 41-60
	sl.Mmsi = combineMmsi(b.CountryCode, mmsi6)

	b.Protocol = &Beacon_ShipSecurityLocation{
		ShipSecurityLocation: sl,
	}
	return nil
}

func nationalLocation(id *BitReader, b *Beacon, btype uint64) error {
	nl := &NationalLocation{}
	nl.BeaconType = btype
	nl.SerialNumber = id.ReadN(18) // bits 41-58

	b.Protocol = &Beacon_NationalLocation{
		NationalLocation: nl,
	}
	return nil
}

func rlsLocation(id *BitReader, b *Beacon) error {
	rls := &RLSLocation{}
	rls.BeaconType = id.ReadN(2) // bits 41 - 42
	b43to46 := id.ReadN(4)       // bits 43 - 46
	if b43to46 == 15 {           // 0b1111
		rls.HasMmsi = true
		rls.MmsiSuffix = id.ReadN(20) // bis 47 - 66
	} else {
		rls.HasMmsi = false
		// number is going to be 10 bits but 4 have already
		// been read (bits 43 - 52)
		next := id.ReadN(6)
		suffix := b43to46<<6 | next
		num := suffix
		if rls.BeaconType == 1 { // EPRIB
			num = 1000 + suffix
		} else if rls.BeaconType == 0 { // ELT
			num = 2000 + suffix
		} else if rls.BeaconType == 2 { // PLB
			num = 3000 + suffix
		}
		if suffix < 920 {
			rls.TacNumber = num
		} else {
			rls.NationalRlsNumber = num
		}
		rls.SerialNumber = id.ReadN(14) // bits 53 - 66
	}
	b.Protocol = &Beacon_RlsLocation{
		RlsLocation: rls,
	}
	return nil
}

func combineMmsi(countryCode uint64, mmsi6 uint64) uint64 {
	return countryCode*1000000 + mmsi6
}
