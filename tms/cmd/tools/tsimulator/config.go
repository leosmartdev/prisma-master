package main

import (
	"encoding/json"
	"os"
	"prisma/tms/cmd/tools/tsimulator/object"
	"prisma/tms/log"
)

type fileStructure struct {
	TimeConfig object.TimeConfig `json:"time_config"`
	Objects    []object.Object
}

// Returns a slice of stations, also it initializes them
func getStationFromFile(configFile string) (stations []object.Station) {
	fstations, err := os.Open(configFile)
	if err != nil {
		log.Fatal("Error to open a file: %v", err)
	}
	defer fstations.Close()
	if err = json.NewDecoder(fstations).Decode(&stations); err != nil {
		log.Fatal("Unable to parse a file of AIS targets: %v", err)
	}
	// we have a slice of stations, so init them
	for i := range stations {
		stations[i].Init()
	}
	return
}

// Returns a slice of sea objects
func getParametersFromFile(configFile string) (object.TimeConfig, []object.Object) {
	fvessels, err := os.Open(configFile)
	if err != nil {
		log.Fatal("Error to open a file: %v", err)
	}
	defer fvessels.Close()
	fData := fileStructure{}
	if err = json.NewDecoder(fvessels).Decode(&fData); err != nil {
		log.Fatal("Unable to parse a file of seaObjects: %v", err)
	}
	return fData.TimeConfig, fData.Objects
}
