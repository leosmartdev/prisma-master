//Package sar implements specifications for the sarsat beacon identifier.
//The specifications can be found here:
// https://team.technosci.com:8442/browse/CONV-801
package sar

import (
	"encoding/hex"
	"reflect"
	"testing"
)

func TestGetMmsi(t *testing.T) {

	tt := []struct {
		name   string
		beacon string
		comp   string
	}{
		{"MaritimeUser", "A029C2900D97591", "257743921"},
		{"Maritime Standard Location", "2024F72524FFBFF", "257506153"},
		{"Aviation User", "A786492E70174C1", ""},
		{"Beacon Secure Locatoion", "20383C480000000", "257123456"},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			b, err := DecodeHexID(tc.beacon)
			if err != nil {
				t.Fatalf("%s unexpected failure error %+v", tc.name, err)
			}
			if b.GetMmsi() != tc.comp {
				t.Errorf("%s failed because we could not get mmmsi correctly from: %+v", tc.name, b)
			}
		})
	}
}

func TestBeaconProtocolCode(t *testing.T) {
	tt := []struct {
		name   string
		beacon string
		err    error
	}{
		{"pcTestUser", "9C7F4B013595551", nil},
		{"pcOrbitographyProtocol", "9A22BE29630F010", nil},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			b, err := DecodeHexID(tc.beacon)
			if err != nil {
				t.Fatalf("%s error: %+v", tc.name, err)
			}
			t.Logf("beacon: %+v", b)
		})
	}
}

func TestCountryCode(t *testing.T) {
	b, err := DecodeHexID("ADC667150EFE241")
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	want := uint64(366)
	have := b.CountryCode
	if have != want {
		t.Errorf("\n want: %v \n have: %v \n", want, have)
	}
}

// http://www.cospas-sarsat.int/en/user-protocols?id=278:maritime-user-protocol-details&catid=39
func TestMaritimeUser(t *testing.T) {
	beacon, err := DecodeHexID("A029C2900D97591")

	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	p, ok := beacon.Protocol.(*Beacon_MaritimeUser)
	if !ok {
		t.Fatalf("unexpected type: %v", reflect.TypeOf(beacon.Protocol))
	}
	{
		want := uint64(257743921)
		have := p.MaritimeUser.Mmsi
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := "2"
		have := p.MaritimeUser.SpecificBeaconNumber
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(1)
		have := p.MaritimeUser.AuxiliaryRadioLocatingDevice
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
}

// http://www.cospas-sarsat.int/en/user-protocols?id=279:radio-call-sign-user-protocol-details&catid=39
func TestRadioCallSignUser(t *testing.T) {
	beacon, err := DecodeHexID("9B7B2059560E9D1")

	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	p, ok := beacon.Protocol.(*Beacon_RadioCallSignUser)
	if !ok {
		t.Fatalf("unexpected type: %v", reflect.TypeOf(beacon.Protocol))
	}

	{
		want := uint64(219) // Denmark
		have := beacon.CountryCode
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := "D52683"
		have := p.RadioCallSignUser.CallSign
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := "1"
		have := p.RadioCallSignUser.SpecificBeaconNumber
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(1)
		have := p.RadioCallSignUser.AuxiliaryRadioLocatingDevice
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
}

// http://www.cospas-sarsat.int/en/user-protocols?id=280:aviation-user-protocol-details&catid=39
func TestAviationUser(t *testing.T) {
	beacon, err := DecodeHexID("A786492E70174C1")

	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	p, ok := beacon.Protocol.(*Beacon_AviationUser)
	if !ok {
		t.Fatalf("unexpected type: %v", reflect.TypeOf(beacon.Protocol))
	}
	{
		want := uint64(316) // Canada
		have := beacon.CountryCode
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := "C7518"
		have := p.AviationUser.AircraftRegistrationMarking
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(0)
		have := p.AviationUser.SpecificEltNumber
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(1)
		have := p.AviationUser.AuxiliaryRadioLocatingDevice
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
}

// http://www.cospas-sarsat.int/en/user-protocols?catid=39:beacon-coding-guide&id=282:serial-user-protocol-examples#serialeltserialid
func TestSerialUserElt(t *testing.T) {
	beacon, err := DecodeHexID("BEEC0358DC00001")

	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	p, ok := beacon.Protocol.(*Beacon_SerialUser)
	if !ok {
		t.Fatalf("unexpected type: %v", reflect.TypeOf(beacon.Protocol))
	}
	{
		want := uint64(503) // Austrailia
		have := beacon.CountryCode
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(0) // Beacon Type Aviation
		have := p.SerialUser.BeaconType
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(54839)
		have := p.SerialUser.SerialNumber
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(0)
		have := p.SerialUser.CertificateNumber
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(1)
		have := p.SerialUser.AuxiliaryRadioLocatingDevice
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
}

// http://www.cospas-sarsat.int/en/user-protocols?catid=39:beacon-coding-guide&id=282:serial-user-protocol-examples#serialairopserialnum
func TestSerialUserAircraftOperatorDesignator(t *testing.T) {
	beacon, err := DecodeHexID("ADCCB8E29D80001")

	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	p, ok := beacon.Protocol.(*Beacon_SerialUser)
	if !ok {
		t.Fatalf("unexpected type: %v", reflect.TypeOf(beacon.Protocol))
	}
	{
		want := uint64(366) // USA
		have := beacon.CountryCode
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(1) // 0b001: ELT With Aircraft Operator Designator
		have := p.SerialUser.BeaconType
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := "AAL"
		have := p.SerialUser.AircraftOperatorDesignator
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(3456)
		have := p.SerialUser.SerialNumber
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(1)
		have := p.SerialUser.AuxiliaryRadioLocatingDevice
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
}

// http://www.cospas-sarsat.int/en/user-protocols?catid=39:beacon-coding-guide&id=282:serial-user-protocol-examples#serial24bitaddress
func TestSerialUserAircraftAddress(t *testing.T) {
	beacon, err := DecodeHexID("ADCDABC3C3C0001")

	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	p, ok := beacon.Protocol.(*Beacon_SerialUser)
	if !ok {
		t.Fatalf("unexpected type: %v", reflect.TypeOf(beacon.Protocol))
	}
	{
		want := uint64(366) // USA
		have := beacon.CountryCode
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(3) // 0b011: 24-bt address
		have := p.SerialUser.BeaconType
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(0xaf0f0f)
		have := p.SerialUser.AircraftAddress
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(1)
		have := p.SerialUser.AuxiliaryRadioLocatingDevice
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
}

// http://www.cospas-sarsat.int/en/user-protocols?catid=39:beacon-coding-guide&id=282:serial-user-protocol-examples#serialepirbserialid
func TestSerialUserEprib(t *testing.T) {
	beacon, err := DecodeHexID("A22E00027000000")

	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	p, ok := beacon.Protocol.(*Beacon_SerialUser)
	if !ok {
		t.Fatalf("unexpected type: %v", reflect.TypeOf(beacon.Protocol))
	}
	{
		want := uint64(273) // Russia
		have := beacon.CountryCode
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(4) // 0b100: Non-Float Free EPIRB
		have := p.SerialUser.BeaconType
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(156)
		have := p.SerialUser.SerialNumber
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(0)
		have := p.SerialUser.AuxiliaryRadioLocatingDevice
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
}

// http://www.cospas-sarsat.int/en/user-protocols?catid=39:beacon-coding-guide&id=282:serial-user-protocol-examples#serialplbserialid
func TestSerialUserPlb(t *testing.T) {
	beacon, err := DecodeHexID("A22F03504400001")

	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	p, ok := beacon.Protocol.(*Beacon_SerialUser)
	if !ok {
		t.Fatalf("unexpected type: %v", reflect.TypeOf(beacon.Protocol))
	}
	{
		want := uint64(273) // Russia
		have := beacon.CountryCode
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(6) // 0b110: PLB
		have := p.SerialUser.BeaconType
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(54289)
		have := p.SerialUser.SerialNumber
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(1)
		have := p.SerialUser.AuxiliaryRadioLocatingDevice
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
}

// http://www.cospas-sarsat.int/en/user-location-protocols
//
// This should be the same as non-location.
func TestUserLocationProtocol(t *testing.T) {
	beacon, err := DecodeHexID("BBAD5EE4A400191")

	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	p, ok := beacon.Protocol.(*Beacon_SerialUser)
	if !ok {
		t.Fatalf("unexpected type: %v", reflect.TypeOf(beacon.Protocol))
	}
	{
		want := uint64(477) // Hong Kong
		have := beacon.CountryCode
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(2) // 0b10: Maritime (float free)
		have := p.SerialUser.BeaconType
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(506153)
		have := p.SerialUser.SerialNumber
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(100)
		have := p.SerialUser.CertificateNumber
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(1)
		have := p.SerialUser.AuxiliaryRadioLocatingDevice
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
}

// http://www.cospas-sarsat.int/en/standard-location-protocols?catid=39:beacon-coding-guide&id=286:standard-location-protocol#standardloc_elt_24bits
func TestMaritimeStandardLocation(t *testing.T) {
	beacon, err := DecodeHexID("2024F72524FFBFF")

	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	p, ok := beacon.Protocol.(*Beacon_MaritimeStandardLocation)
	if !ok {
		t.Fatalf("unexpected type: %v", reflect.TypeOf(beacon.Protocol))
	}
	{
		want := uint64(257) // Norway
		have := beacon.CountryCode
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(257506153)
		have := p.MaritimeStandardLocation.Mmsi
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(2)
		have := p.MaritimeStandardLocation.SpecificBeaconNumber
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
}

func TestAviationStandardLocation(t *testing.T) {
	b := &BitWriter{}
	b.Write(0)                              // Protocol flag
	b.WriteN(10, 257)                       // Country Code Norway
	b.WriteN(4, pcAviationStandardLocation) // Protocol Code
	b.WriteN(24, 0xabc)
	b.WriteZeros(21) // bits 65-85

	// 202600157800000
	beaconID := hex.EncodeToString(b.bytes)[:15]
	beacon, err := DecodeHexID(beaconID)

	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	p, ok := beacon.Protocol.(*Beacon_AviationStandardLocation)
	if !ok {
		t.Fatalf("unexpected type: %v", reflect.TypeOf(beacon.Protocol))
	}
	{
		want := uint64(257) // Norway
		have := beacon.CountryCode
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(0xabc)
		have := p.AviationStandardLocation.AircraftAddress
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
}

func TestSerialEltStandardLocation(t *testing.T) {
	b := &BitWriter{}
	b.Write(0)                               // Protocol flag
	b.WriteN(10, 257)                        // Country Code Norway
	b.WriteN(4, pcSerialEltStandardLocation) // Protocol Code
	b.WriteN(10, 0xa)                        // Certificate Number
	b.WriteN(14, 0xb)                        // Serial Number
	b.WriteZeros(21)                         // bits 65-85

	// 202805001600000
	beaconID := hex.EncodeToString(b.bytes)[:15]
	beacon, err := DecodeHexID(beaconID)

	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	p, ok := beacon.Protocol.(*Beacon_SerialStandardLocation)
	if !ok {
		t.Fatalf("unexpected type: %v", reflect.TypeOf(beacon.Protocol))
	}
	{
		want := uint64(257) // Norway
		have := beacon.CountryCode
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(0xa)
		have := p.SerialStandardLocation.CertificateNumber
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(0xb)
		have := p.SerialStandardLocation.SerialNumber
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
}

func TestSerialEpribStandardLocation(t *testing.T) {
	b := &BitWriter{}
	b.Write(0)                                 // Protocol flag
	b.WriteN(10, 257)                          // Country Code Norway
	b.WriteN(4, pcSerialEpribStandardLocation) // Protocol Code
	b.WriteN(10, 0xa)                          // Certificate Number
	b.WriteN(14, 0xb)                          // Serial Number
	b.WriteZeros(21)                           // bits 65-85

	// 202C05001600000
	beaconID := hex.EncodeToString(b.bytes)[:15]
	beacon, err := DecodeHexID(beaconID)

	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	p, ok := beacon.Protocol.(*Beacon_SerialStandardLocation)
	if !ok {
		t.Fatalf("unexpected type: %v", reflect.TypeOf(beacon.Protocol))
	}
	{
		want := uint64(257) // Norway
		have := beacon.CountryCode
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(0xa)
		have := p.SerialStandardLocation.CertificateNumber
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(0xb)
		have := p.SerialStandardLocation.SerialNumber
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
}

func TestSerialPlbStandardLocation(t *testing.T) {
	b := &BitWriter{}
	b.Write(0)                               // Protocol flag
	b.WriteN(10, 257)                        // Country Code Norway
	b.WriteN(4, pcSerialPlbStandardLocation) // Protocol Code
	b.WriteN(10, 0xa)                        // Certificate Number
	b.WriteN(14, 0xb)                        // Serial Number
	b.WriteZeros(21)                         // bits 65-85

	// 202E05001600000
	beaconID := hex.EncodeToString(b.bytes)[:15]
	beacon, err := DecodeHexID(beaconID)

	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	p, ok := beacon.Protocol.(*Beacon_SerialStandardLocation)
	if !ok {
		t.Fatalf("unexpected type: %v", reflect.TypeOf(beacon.Protocol))
	}
	{
		want := uint64(257) // Norway
		have := beacon.CountryCode
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(0xa)
		have := p.SerialStandardLocation.CertificateNumber
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(0xb)
		have := p.SerialStandardLocation.SerialNumber
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
}

func TestSerialEltAodStandardLocation(t *testing.T) {
	b := &BitWriter{}
	b.Write(0)                                  // Protocol flag
	b.WriteN(10, 257)                           // Country Code Norway
	b.WriteN(4, pcSerialEltAodStandardLocation) // Protocol Code
	b.WriteBaudot5N(3, "ABC")                   // Aircraft Operator Designator
	b.WriteN(9, 0xb)                            // Serial Number
	b.WriteZeros(21)                            // bits 65-85

	// 202B89B81600000
	beaconID := hex.EncodeToString(b.bytes)[:15]
	beacon, err := DecodeHexID(beaconID)

	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	p, ok := beacon.Protocol.(*Beacon_SerialStandardLocation)
	if !ok {
		t.Fatalf("unexpected type: %v", reflect.TypeOf(beacon.Protocol))
	}
	{
		want := uint64(257) // Norway
		have := beacon.CountryCode
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := "ABC"
		have := p.SerialStandardLocation.AircraftOperatorDesignator
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(0xb)
		have := p.SerialStandardLocation.SerialNumber
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
}

func TestShipSecurityStandardLocation(t *testing.T) {
	b := &BitWriter{}
	b.Write(0)                                  // Protocol flag
	b.WriteN(10, 257)                           // Country Code Norway
	b.WriteN(4, pcShipSecurityStandardLocation) // Protocol Code
	b.WriteN(20, 123456)                        // MMSI 6
	b.WriteZeros(4)                             // bits 61-64
	b.WriteZeros(21)                            // bits 65-85

	// 20383C480000000
	beaconID := hex.EncodeToString(b.bytes)[:15]
	beacon, err := DecodeHexID(beaconID)

	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	p, ok := beacon.Protocol.(*Beacon_ShipSecurityLocation)
	if !ok {
		t.Fatalf("unexpected type: %v", reflect.TypeOf(beacon.Protocol))
	}
	{
		want := uint64(257) // Norway
		have := beacon.CountryCode
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(257123456)
		have := p.ShipSecurityLocation.Mmsi
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
}

// http://www.cospas-sarsat.int/en/national-location-protocols
func TestNationalLocationEprib(t *testing.T) {
	beacon, err := DecodeHexID("20341500BF81FE0")

	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	p, ok := beacon.Protocol.(*Beacon_NationalLocation)
	if !ok {
		t.Fatalf("unexpected type: %v", reflect.TypeOf(beacon.Protocol))
	}
	{
		want := uint64(257) // Norway
		have := beacon.CountryCode
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(10753)
		have := p.NationalLocation.SerialNumber
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(pcNationalLocationEprib)
		have := p.NationalLocation.BeaconType
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
}

func TestSerialUserEltWithCert(t *testing.T) {
	beacon, err := DecodeHexID("ADCC404C8400315")

	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	p, ok := beacon.Protocol.(*Beacon_SerialUser)
	if !ok {
		t.Fatalf("unexpected type: %v", reflect.TypeOf(beacon.Protocol))
	}
	{
		want := uint64(0) // Serial Type: ELT with Serial Identification
		have := p.SerialUser.BeaconType
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(4897)
		have := p.SerialUser.SerialNumber
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(197) // C/S Number
		have := p.SerialUser.CertificateNumber
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(1)
		have := p.SerialUser.AuxiliaryRadioLocatingDevice
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
}

func TestRlsLocation(t *testing.T) {
	beacon, err := DecodeHexID("35FA80203DBFDFF")
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	p, ok := beacon.Protocol.(*Beacon_RlsLocation)
	if !ok {
		t.Fatalf("unexpected type: %v", reflect.TypeOf(beacon.Protocol))
	}
	{
		want := uint64(1)
		have := p.RlsLocation.BeaconType
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(1001)
		have := p.RlsLocation.TacNumber
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(123)
		have := p.RlsLocation.SerialNumber
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
}

func TestRlsLocation2(t *testing.T) {
	beacon, err := DecodeHexID("2DDA0F2AD9BFDFF")
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	p, ok := beacon.Protocol.(*Beacon_RlsLocation)
	if !ok {
		t.Fatalf("unexpected type: %v", reflect.TypeOf(beacon.Protocol))
	}
	{
		want := uint64(0)
		have := p.RlsLocation.BeaconType
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(2121)
		have := p.RlsLocation.TacNumber
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
	{
		want := uint64(5555)
		have := p.RlsLocation.SerialNumber
		if have != want {
			t.Errorf("\n want: %v \n have: %v \n", want, have)
		}
	}
}
