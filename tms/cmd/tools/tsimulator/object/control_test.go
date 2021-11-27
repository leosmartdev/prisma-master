package object

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func CompareSeaObjects(t *testing.T, obj1, obj2 *Object) {
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

func TestSeaObjectControl_Delete(t *testing.T) {
	control := NewObjectControl([]Object{
		getVessel(), getVessel(), getVessel(),
	}, nil)
	ch := make(chan Object, 2)
	control.RegisterChannel(ch)
	assert.Len(t, control.objects, 3)
	assert.NoError(t, control.Delete(2))
	<-ch
	assert.Len(t, control.objects, 2)
	assert.Error(t, control.Delete(2))
	assert.Len(t, control.objects, 2)
}

func TestSeaObjectControl_Insert(t *testing.T) {
	control := NewObjectControl([]Object{
		getVessel(), getVessel(), getVessel(),
	}, nil)
	assert.Len(t, control.objects, 3)
	control.Insert(getVessel())
	assert.Len(t, control.objects, 4)

	// check collided objects
	v1, v2 := getVessel(), getVessel()
	v1.Pos[0].ArrivalTimeSeconds = 10
	v2.Pos[0].ArrivalTimeSeconds = 10
	control.Insert(v1)
	control.Insert(v2)
	control.detectIntersect()
	assert.Len(t, control.objects, 8)

}

func TestSeaObjectControl_GetByIndex(t *testing.T) {
	control := NewObjectControl([]Object{
		getVessel(), getVessel(), getVessel(),
	}, nil)
	obj, err := control.GetByIndex(1)
	assert.NoError(t, err)
	retObj := getVessel()
	retObj.Id = 1
	retObj.Init()

	CompareSeaObjects(t, &retObj, &obj)
	_, err = control.GetByIndex(99999)
	assert.Error(t, err)
	_, err = control.GetByIndex(-1)
	assert.Error(t, err)
}

func TestSeaObjectControl_GetList(t *testing.T) {
	control := NewObjectControl([]Object{
		getVessel(), getVessel(), getVessel(),
	}, nil)
	obj := control.GetList()
	assert.Len(t, obj, 3)
	obj[1].Name = "TestSeaObjectControl_GetList"
	assert.NotEqual(t, obj[1].Name, control.objects[1].Name)
}

func TestSeaObjectControl_Update(t *testing.T) {
	control := NewObjectControl([]Object{
		getVessel(), getVessel(), getVessel(),
	}, nil)
	obj := getVessel()
	obj.Name = "TestSeaObjectControl_Update"
	assert.NoError(t, control.Update(1, obj))
	obj.Id = 1
	obj.Init()
	CompareSeaObjects(t, control.objects[1], &obj)
	assert.Error(t, control.Update(-1, obj))
	assert.Error(t, control.Update(99999999, obj))
}

func TestSeaObjectControl_SetupNewPosition(t *testing.T) {
	object := getVessel()
	control := NewObjectControl([]Object{
		object,
	}, nil)
	assert.NoError(t, control.SetupNewPosition(1, 1))
	assert.Equal(t, control.objects[1].curPos, control.objects[1].Pos[1])
	assert.Error(t, control.SetupNewPosition(-1, 1))
	assert.Error(t, control.SetupNewPosition(9999, 1))
	assert.Error(t, control.SetupNewPosition(0, 9999))
	assert.Error(t, control.SetupNewPosition(0, -1))
}
