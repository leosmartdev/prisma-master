//Package sar implements specifications for the sarsat beacon identifier.
//The specifications can be found here:
// https://team.technosci.com:8442/browse/CONV-801
package sar

import "strconv"

// GetMmsi extracts mmsi value from Beacon
func (b *Beacon) GetMmsi() string {
	if b.GetMaritimeUser() != nil {
		return strconv.FormatUint(b.GetMaritimeUser().Mmsi, 10)
	}
	if b.GetMaritimeStandardLocation() != nil {
		return strconv.FormatUint(b.GetMaritimeStandardLocation().Mmsi, 10)
	}
	if b.GetShipSecurityLocation() != nil {
		return strconv.FormatUint(b.GetShipSecurityLocation().Mmsi, 10)
	}
	return ""
}
