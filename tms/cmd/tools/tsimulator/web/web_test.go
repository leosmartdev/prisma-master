package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"prisma/tms/cmd/tools/tsimulator/object"
	"testing"

	"github.com/stretchr/testify/assert"
)

func CompareSeaObjects(t *testing.T, obj1, obj2 *object.Object) {
	assert.Equal(t, []interface{}{
		obj1.Number,
		obj1.Pos,
		obj1.Mmsi,
		obj1.BeaconId,
		obj1.Imei,
		obj1.ETA,
		obj1.DOA,
		obj1.Destination,
	}, []interface{}{
		obj2.Number,
		obj2.Pos,
		obj2.Mmsi,
		obj2.BeaconId,
		obj2.Imei,
		obj2.ETA,
		obj2.DOA,
		obj2.Destination,
	})
}

func getNewControl() *object.Control {
	return object.NewObjectControl([]object.Object{
		{
			Mmsi:        235009802,
			Device:      "ais",
			Destination: "testDestination",
			Name:        "test",
			ETA:         "02101504",
			Type:        30,
			Pos: []object.PositionArrivalTime{
				{
					PositionSpeed: object.PositionSpeed{
						Latitude:  0,
						Longitude: 0,
						Speed:     10,
					},
				},
				{
					PositionSpeed: object.PositionSpeed{
						Latitude:  20.1,
						Longitude: 20.1,
						Speed:     10,
					},
				},
				{
					PositionSpeed: object.PositionSpeed{
						Latitude:  30.1,
						Longitude: 20.1,
						Speed:     30,
					},
				},
			},
			ReportPeriod: 2,
		},
	}, nil)
}

var payLoadJSONSeaObject = `{
  "device": "radar",
  "name": "RADARB",
  "pos": [{
    "latitude": 1.173931,
    "longitude": 103.141706,
    "speed": 10
  }, {
    "latitude": 1.151931,
    "longitude": 102.903754,
    "speed": 7
  }, {
    "latitude": 1.173931,
    "longitude": 103.401706,
    "speed": 8
  }, {
    "latitude": 1.172931,
    "longitude": 102.841706,
    "speed": 6
  }]
}`
var payLoadJSONSarsat = `{
  "device": "sarsat",
  "pos": [{
    "latitude": 1,
    "longitude": 3,
    "speed": 1
  }]
}
`

var (
	clientHTTP = http.DefaultClient
)

func getRequest(t *testing.T, method, urlStr string, body io.Reader) *http.Response {
	request, err := http.NewRequest(method, urlStr, body)
	assert.NoError(t, err)
	request.Header = http.Header{"Content-Type": {"application/json"}}
	response, err := clientHTTP.Do(request)
	assert.NoError(t, err)
	return response
}

func TestServiceSeaObject_Get(t *testing.T) {
	var (
		data struct {
			Count   int
			Objects object.ContainerIdObjects
		}
		control = getNewControl()
		server  = httptest.NewServer(GetRestContainer(control))
	)
	response := getRequest(t, http.MethodGet, fmt.Sprintf("%s/v1/get/", server.URL), nil)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.NoError(t, json.NewDecoder(response.Body).Decode(&data))
	assert.Equal(t, control.Len(), data.Count)

	cobjects := control.GetList()
	cobjects[1].Id = 1
	data.Objects[1].Init()

	CompareSeaObjects(t, cobjects[1], data.Objects[1])
}

func TestServiceSeaObject_CreateTarget(t *testing.T) {
	var (
		control   = getNewControl()
		server    = httptest.NewServer(GetRestContainer(control))
		lengthOld = control.Len()
		seaobject object.Object
	)
	response := getRequest(t, http.MethodPost, fmt.Sprintf("%s/v1/target/", server.URL),
		bytes.NewBuffer([]byte(payLoadJSONSeaObject)))
	assert.Equal(t, http.StatusCreated, response.StatusCode)
	cobjects := control.GetList()
	assert.Len(t, cobjects, lengthOld+1)
	json.Unmarshal([]byte(payLoadJSONSeaObject), &seaobject)
	seaobject.Id = lengthOld + 1
	seaobject.Mmsi = cobjects[lengthOld+1].Mmsi
	seaobject.Name = cobjects[lengthOld+1].Name
	seaobject.Init()
	CompareSeaObjects(t, cobjects[lengthOld+1], &seaobject)

	response = getRequest(t, http.MethodPost, fmt.Sprintf("%s/v1/target/", server.URL),
		bytes.NewBuffer([]byte(payLoadJSONSarsat)))
	assert.Equal(t, http.StatusCreated, response.StatusCode)
	cobjects = control.GetList()
	assert.NotEmpty(t, cobjects[len(cobjects)].BeaconId)
}

func TestServiceSeaObject_UpdateTarget(t *testing.T) {
	var (
		control = getNewControl()
		server  = httptest.NewServer(GetRestContainer(control))
	)
	response := getRequest(t, http.MethodPost, fmt.Sprintf("%s/v1/target/id/1", server.URL),
		bytes.NewBuffer([]byte(`{"mmsi": 1, "name": "testingAAA"}`)))
	assert.Equal(t, http.StatusNoContent, response.StatusCode)
	cobjects := control.GetList()
	assert.Equal(t, uint(1), uint(cobjects[1].Mmsi))
	assert.Equal(t, "testingAAA", cobjects[1].Name)

	response = getRequest(t, http.MethodPost, fmt.Sprintf("%s/v1/target/id/0", server.URL),
		bytes.NewBuffer([]byte(`{"mmsi": 1, "name": "testingAAA"}`)))
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	response = getRequest(t, http.MethodPost, fmt.Sprintf("%s/v1/target/id/99999999999999999", server.URL),
		bytes.NewBuffer([]byte(`{"mmsi": 1, "name": "testingAAA"}`)))
	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func TestServiceSeaObject_CreateRoute(t *testing.T) {
	var (
		control = getNewControl()
		server  = httptest.NewServer(GetRestContainer(control))
	)
	t.Skip()
	response := getRequest(t, http.MethodPost, fmt.Sprintf("%s/v1/route/target-id/1", server.URL),
		bytes.NewBuffer([]byte(`{
    "latitude": 1.173931,
    "longitude": 103.141706,
    "speed": 10
  }`)))
	assert.Equal(t, http.StatusCreated, response.StatusCode)
	cobjects := control.GetList()
	assert.Equal(t, 1.173931, cobjects[0].Pos[len(cobjects[0].Pos)-1].Latitude)
	assert.Equal(t, 103.141706, cobjects[0].Pos[len(cobjects[0].Pos)-1].Longitude)
	assert.Equal(t, 10, int(cobjects[0].Pos[len(cobjects[0].Pos)-1].Speed))
	response = getRequest(t, http.MethodPost, fmt.Sprintf("%s/v1/route/target-id/0", server.URL),
		bytes.NewBuffer([]byte(`{
    "latitude": 1.173931,
    "longitude": 103.141706,
    "speed": 10
  }`)))
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	response = getRequest(t, http.MethodPost, fmt.Sprintf("%s/v1/route/target-id/999999999999999999", server.URL),
		bytes.NewBuffer([]byte(`{
    "latitude": 1.173931,
    "longitude": 103.141706,
    "speed": 10
  }`)))
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
}

func TestServiceSeaObject_UpdateRoute(t *testing.T) {
	var (
		control = getNewControl()
		server  = httptest.NewServer(GetRestContainer(control))
	)
	response := getRequest(t, http.MethodPost, fmt.Sprintf("%s/v1/route/id/2/target-id/1", server.URL),
		bytes.NewBuffer([]byte(`{
    "latitude": 1.173931,
    "longitude": 103.141706,
    "speed": 10
  }`)))
	assert.Equal(t, http.StatusNoContent, response.StatusCode)
	cobjects := control.GetList()
	assert.Equal(t, 1.173931, cobjects[1].Pos[1].Latitude)
	assert.Equal(t, 103.141706, cobjects[1].Pos[1].Longitude)
	assert.Equal(t, 10, int(cobjects[1].Pos[1].Speed))

	response = getRequest(t, http.MethodPost, fmt.Sprintf("%s/v1/route/id/2/target-id/999999999999", server.URL),
		bytes.NewBuffer([]byte(`{
    "latitude": 1.173931,
    "longitude": 103.141706,
    "speed": 10
  }`)))
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	response = getRequest(t, http.MethodPost, fmt.Sprintf("%s/v1/route/id/2/target-id/0", server.URL),
		bytes.NewBuffer([]byte(`{
    "latitude": 1.173931,
    "longitude": 103.141706,
    "speed": 10
  }`)))
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	response = getRequest(t, http.MethodPost, fmt.Sprintf("%s/v1/route/id/999999999999/target-id/1", server.URL),
		bytes.NewBuffer([]byte(`{
    "latitude": 1.173931,
    "longitude": 103.141706,
    "speed": 10
  }`)))
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	response = getRequest(t, http.MethodPost, fmt.Sprintf("%s/v1/route/id/0/target-id/1", server.URL),
		bytes.NewBuffer([]byte(`{
    "latitude": 1.173931,
    "longitude": 103.141706,
    "speed": 10
  }`)))
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
}

func TestServiceSeaObject_UpdateWholeRoute(t *testing.T) {
	var (
		control = getNewControl()
		server  = httptest.NewServer(GetRestContainer(control))
	)
	response := getRequest(t, http.MethodPost, fmt.Sprintf("%s/v1/route/target-id/1", server.URL),
		bytes.NewBuffer([]byte(`[{
    "latitude": 1.173931,
    "longitude": 103.141706,
    "speed": 10
  },{
    "latitude": 2.173931,
    "longitude": 104.141706,
    "speed": 20
  }]`)))
	assert.Equal(t, http.StatusNoContent, response.StatusCode)
	cobjects := control.GetList()
	assert.Equal(t, 1.173931, cobjects[1].Pos[0].Latitude)
	assert.Equal(t, 103.141706, cobjects[1].Pos[0].Longitude)
	assert.Equal(t, 10, int(cobjects[1].Pos[0].Speed))
	assert.Equal(t, 2.173931, cobjects[1].Pos[1].Latitude)
	assert.Equal(t, 104.141706, cobjects[1].Pos[1].Longitude)
	assert.Equal(t, 20, int(cobjects[1].Pos[1].Speed))
	assert.Equal(t, cobjects[1].GetCurPos(), cobjects[1].Pos[0])

	response = getRequest(t, http.MethodPost, fmt.Sprintf("%s/v1/route/target-id/9999", server.URL),
		bytes.NewBuffer([]byte(`[{
    "latitude": 1.173931,
    "longitude": 103.141706,
    "speed": 10
  },{
    "latitude": 2.173931,
    "longitude": 104.141706,
    "speed": 20
  }]`)))
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	response = getRequest(t, http.MethodPost, fmt.Sprintf("%s/v1/route/target-id/1", server.URL),
		bytes.NewBuffer([]byte(`BAD JSON`)))
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	response = getRequest(t, http.MethodPost, fmt.Sprintf("%s/v1/route/target-id/0", server.URL),
		bytes.NewBuffer([]byte(`BAD JSON`)))
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

}

func TestServiceSeaObject_DeleteRoute(t *testing.T) {
	var (
		control = getNewControl()
		server  = httptest.NewServer(GetRestContainer(control))
	)
	response := getRequest(t, http.MethodDelete, fmt.Sprintf("%s/v1/route/target-id/1", server.URL), nil)
	assert.Equal(t, http.StatusNoContent, response.StatusCode)
	cobjects := control.GetList()
	assert.Len(t, cobjects[1].Pos, 0)
	response = getRequest(t, http.MethodDelete, fmt.Sprintf("%s/v1/route/target-id/0", server.URL), nil)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	response = getRequest(t, http.MethodDelete, fmt.Sprintf("%s/v1/route/target-id/999999999", server.URL), nil)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
}

func TestServiceSeaObject_DeleteTarget(t *testing.T) {
	var (
		control = getNewControl()
		server  = httptest.NewServer(GetRestContainer(control))
	)
	response := getRequest(t, http.MethodDelete, fmt.Sprintf("%s/v1/target/id/1", server.URL), nil)
	assert.Equal(t, http.StatusNoContent, response.StatusCode)
	cobjects := control.GetList()
	assert.Len(t, cobjects, 0)
	response = getRequest(t, http.MethodDelete, fmt.Sprintf("%s/v1/target/id/1", server.URL), nil)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	response = getRequest(t, http.MethodDelete, fmt.Sprintf("%s/v1/target/id/0", server.URL), nil)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	response = getRequest(t, http.MethodDelete, fmt.Sprintf("%s/v1/target/id/999999999", server.URL), nil)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
}

func TestServiceSeaObject_StartAlerting(t *testing.T) {
	var (
		control = getNewControl()
		server  = httptest.NewServer(GetRestContainer(control))
	)
	cobjects := control.GetList()
	cobjects[1].Device = "omnicom-vms"
	cobjects[1].InitMoving(cobjects[1].ReportPeriod)
	control.Update(1, *cobjects[1])
	response := getRequest(t, http.MethodPost, fmt.Sprintf("%s/v1/alert/type/pu/target-id/1", server.URL), nil)
	assert.Equal(t, http.StatusCreated, response.StatusCode)
	response = getRequest(t, http.MethodPost, fmt.Sprintf("%s/v1/alert/type/NOTEXISTS/target-id/1", server.URL), nil)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	response = getRequest(t, http.MethodPost, fmt.Sprintf("%s/v1/alert/type/pu/target-id/0", server.URL), nil)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	response = getRequest(t, http.MethodPost, fmt.Sprintf("%s/v1/alert/type/pu/target-id/999999", server.URL), nil)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
}

func TestServiceSeaObject_StopAlerting(t *testing.T) {
	var (
		control = getNewControl()
		server  = httptest.NewServer(GetRestContainer(control))
	)
	cobjects := control.GetList()
	cobjects[1].Device = "omnicom"
	cobjects[1].InitMoving(cobjects[1].ReportPeriod)
	control.Update(1, *cobjects[1])
	response := getRequest(t, http.MethodDelete, fmt.Sprintf("%s/v1/alert/type/pu/target-id/1", server.URL), nil)
	assert.Equal(t, http.StatusNoContent, response.StatusCode)
	response = getRequest(t, http.MethodDelete, fmt.Sprintf("%s/v1/alert/type/NOTEXISTS/target-id/1", server.URL), nil)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	response = getRequest(t, http.MethodDelete, fmt.Sprintf("%s/v1/alert/type/pu/target-id/0", server.URL), nil)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	response = getRequest(t, http.MethodDelete, fmt.Sprintf("%s/v1/alert/type/pu/target-id/9999999", server.URL), nil)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
}
