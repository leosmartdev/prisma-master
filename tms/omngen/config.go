package omngen

import "encoding/json"
import "prisma/tms/log"
import "io/ioutil"

type Pos struct {
	Lat  float64
	Long float64
}

type Jbeacon struct {
	Imei         string
	Pos          []Pos
	Reportperiod uint32
}

func JsonParse(filename string) []Jbeacon {

	file, e := ioutil.ReadFile(filename)
	if e != nil {
		log.Fatal("File error: %v\n", e)
	}

	var beacons []Jbeacon
	json.Unmarshal(file, &beacons)

	return beacons
}
