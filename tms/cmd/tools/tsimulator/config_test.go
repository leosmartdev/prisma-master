package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/json-iterator/go/assert"
)

func TestGetStationFromFile(t *testing.T) {
	file, err := ioutil.TempFile("", "test_station_config_tsimulator")
	assert.NoError(t, err)
	defer os.Remove(file.Name())
	defer file.Close()
	_, err = file.Write([]byte(`[{
  "device": "ais",
  "latitude": 1.171931,
  "longitude": 103.131706,
  "radius": 0.5,
  "addr": ":9000"
}]`))
	assert.NoError(t, err)
	stations := getStationFromFile(file.Name())
	assert.Len(t, stations, 1)
	assert.Equal(t, "ais", stations[0].Device)
	assert.Equal(t, 1.171931, stations[0].Latitude)
	assert.Equal(t, 103.131706, stations[0].Longitude)
	assert.Equal(t, 0.5, stations[0].Radius)
	assert.Equal(t, ":9000", stations[0].Addr)
}

func TestGetSeaObjectsFromFile(t *testing.T) {

	file, err := ioutil.TempFile("", "test_objects_config_tsimulator")
	assert.NoError(t, err)
	defer os.Remove(file.Name())
	defer file.Close()
	_, err = file.Write([]byte(`{
  "Objects": [
    {
      "device": "ais",
      "mmsi": 235009802,
      "destination": "RU",
      "name": "TESTA",
      "eta": "02101504",
      "type": 58,
      "pos": []
    },{
      "device": "sarsat",
      "mmsi": 235009802,
      "destination": "RU",
      "name": "TESTA",
      "eta": "02101504",
      "type": 58,
      "pos": []
    }
   ]
}`))

	assert.NoError(t, err)
	_, objects := getParametersFromFile(file.Name())
	assert.Len(t, objects, 2)
	assert.Equal(t, "ais", objects[0].Device)
	for _, object := range objects[1:] {
		assert.Equal(t, "sarsat", object.Device)
	}
}
