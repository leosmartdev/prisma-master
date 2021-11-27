// Package moc contains common structures for tms system.
package moc

import (
	"prisma/tms/util/ident"
	"fmt"
)

// GetRegistryIdByType is used to compute registryId from network data by network type
func GetRegistryIdByType(ntype, subscriberId string) (registryId string) {
	switch ntype {
	case "omnicom-vms", "omnicom-solar":
		fmt.Println("here!!!")
		registryId = ident.With("imei", subscriberId).Hash()
	case "ais", "mob-ais", "sart-ais":
		registryId = ident.With("mmsi", subscriberId).Hash()
	}
	return
}
