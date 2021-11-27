// Package ais provides extra functions for AIS devices.
package ais

import "fmt"

func FormatMMSI(mmsi int) string {
	return fmt.Sprintf("%09d", mmsi)
}
