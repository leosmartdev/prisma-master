package object

import (
	"errors"
	"sync"
	"time"

	"prisma/tms/cmd/tools/tsimulator/task"
	"fmt"
)

// How often a watcher should check events of objects
const watchEveryTime = 1 * time.Second

// ErrObjectNotFound is used to point out a sea object is not found
var ErrObjectNotFound = errors.New("the object is not found")

// ContainerIdObjects is a type for saving objects id => object
type ContainerIdObjects map[int]*Object

type posObject struct {
	posIndex    int
	objectIndex int
}

// Control works as a storage of objects and provides updating, getting sea objects safety
type Control struct {
	mux            sync.Mutex
	objects        ContainerIdObjects
	deletedObjects ContainerIdObjects

	tconfig            *TimeConfig
	gmux               sync.Mutex
	generateId         func() int
	autoincrement      int
	chmux              sync.Mutex
	channelsObjects    []chan<- Object    // these channels are used for sending new information to other sources
	channelTaskResults chan<- task.Result // these channels are used for sending new information to other sources
}

// NewObjectControl returns a new instance of a Control.
// Also it runs a watcher for events and initializes objects
func NewObjectControl(objects []Object, tconfig *TimeConfig) *Control {
	soc := &Control{
		objects:         make(ContainerIdObjects),
		channelsObjects: make([]chan<- Object, 0),
		tconfig:         new(TimeConfig),
	}
	soc.generateId = soc.generatorID()
	if tconfig != nil {
		soc.tconfig = tconfig
	}

	tmpCConf := *soc.tconfig
	var id int
	for _, obj := range objects {
		id = soc.generateId()
		soc.objects[id] = NewObject()
		*soc.objects[id] = obj
		soc.objects[id].Id = id
		if soc.objects[id].TimeConfig == nil {
			soc.objects[id].TimeConfig = &tmpCConf
		}
		soc.objects[id].Init()
	}
	soc.detectIntersect()
	return soc
}

func (soc *Control) RunWatcher() {
	go soc.watchObjects()
}

func (soc *Control) generatorID() func() int {
	return func() int {
		soc.gmux.Lock()
		soc.gmux.Unlock()
		soc.autoincrement++
		return soc.autoincrement
	}
}

// arrivalPosString returns a string for positionArrivalTime
// why here - cause it will be used to place a position to hashmap by this string
// if someone will change ToString() at the position structure, this change can break intersection logic
func (soc *Control) arrivalPosString(pos PositionArrivalTime) string {
	return fmt.Sprintf("arrivalTime %d seconds at position long %f lat %f", pos.ArrivalTimeSeconds, pos.Longitude, pos.Latitude)
}

// getCollidedMap returns hashmap
// Where a keys is arrival time and the value of the key is a structure of object and pos keys
func (soc *Control) getCollidedMap() map[string][]posObject {
	collidedTime := make(map[string][]posObject)
	for key := range soc.objects {
		for i := range soc.objects[key].Pos {
			if soc.objects[key].Pos[i].ArrivalTimeSeconds == 0 {
				continue
			}
			collidedTime[soc.arrivalPosString(soc.objects[key].Pos[i])] =
				append(collidedTime[soc.arrivalPosString(soc.objects[key].Pos[i])], posObject{i, key})
		}
	}
	return collidedTime
}

// detectIntersect pushes sarsat alerts for collides objects
func (soc *Control) detectIntersect() {
	collidedMap := soc.getCollidedMap()
	for key := range collidedMap {
		// no collisions
		if len(collidedMap[key]) <= 1 {
			continue
		}
		for _, objKey := range collidedMap[key] {
			collidedObject := soc.objects[objKey.objectIndex]
			obj := NewObject()
			obj.Device = "sarsat"
			obj.Pos = []PositionArrivalTime{
				{
					ArrivalTimeSeconds: collidedObject.Pos[objKey.posIndex].ArrivalTimeSeconds,
					PositionSpeed:      collidedObject.Pos[objKey.posIndex].PositionSpeed,
				},
			}
			soc.Insert(*obj)
			collidedObject.Pos[objKey.posIndex].Collided = true
		}
	}
}

// RegisterChannel appends a channel into others
func (soc *Control) RegisterChannel(ch chan<- Object) {
	soc.chmux.Lock()
	defer soc.chmux.Unlock()
	soc.channelsObjects = append(soc.channelsObjects, ch)
}

// RegisterTaskChannel saves a channel for future sending results of tasks to the channel.
// It is done to make clear client side code, encapsulating,
// performance(copying objects to get channels of them)
func (soc *Control) RegisterTaskChannel(ch chan<- task.Result) {
	soc.chmux.Lock()
	defer soc.chmux.Unlock()
	soc.channelTaskResults = ch
}

func (soc *Control) setAutoIncrement(value int) {
	soc.gmux.Lock()
	soc.gmux.Unlock()
	soc.autoincrement = value
}

func (soc *Control) getAutoIncrement() int {
	soc.gmux.Lock()
	soc.gmux.Unlock()
	return soc.autoincrement
}

func (soc *Control) Len() int {
	return len(soc.objects)
}

// GetList returns a copy of objects
func (soc *Control) GetList() ContainerIdObjects {
	soc.mux.Lock()
	defer soc.mux.Unlock()
	objects := make(ContainerIdObjects)
	for key := range soc.objects {
		objects[key] = NewObject()
		*objects[key] = *soc.objects[key]
	}
	return objects
}

// Move runs moving for each objects and send to channels
func (soc *Control) Move() {
	soc.mux.Lock()
	defer soc.mux.Unlock()
	for i := range soc.objects {
		select {
		case <-soc.objects[i].TickerMoving.C:
			soc.objects[i].Move(soc.objects[i].ReportPeriod)
			soc.sendObjectChannels(*soc.objects[i])
		case <-soc.objects[i].RequestCurrentPos:
			soc.objects[i].Move(uint32(time.Since(soc.objects[i].TimeOfLastMove).Minutes()))
			soc.sendObjectChannels(*soc.objects[i])
			soc.objects[i].RequestCurrentPos <- struct{}{}
		default:
		}
	}
}

// Watch events for sending new information
func (soc *Control) watchObjects() {
	for {
		soc.mux.Lock()
		for i := range soc.objects {
			select {
			case data := <-soc.objects[i].GetChannelHandledTask():
				select {
				case soc.channelTaskResults <- data:
				default:
				}
			default:
			}
		}
		soc.mux.Unlock()
		time.Sleep(watchEveryTime)
	}
}

// SetupNewPosition setups current position on a new one
func (soc *Control) SetupNewPosition(indexObject, indexRoute int) error {
	soc.mux.Lock()
	defer soc.mux.Unlock()
	if obj, ok := soc.objects[indexObject]; !ok {
		return ErrObjectNotFound
	} else {
		obj.activePos = indexRoute
		obj.InitMoving(obj.ReportPeriod)
		soc.objects[indexObject] = obj
	}
	return nil
}

// GetByIMEI returns an object by imei
func (soc *Control) GetByIMEI(imei string) (*Object, error) {
	soc.mux.Lock()
	defer soc.mux.Unlock()
	for _, s := range soc.objects {
		if s.Imei == imei {
			r := *s
			return &r, nil
		}
	}
	return nil, ErrObjectNotFound
}

// GetByIndex returns an object using an index
func (soc *Control) GetByIndex(index int) (Object, error) {
	soc.mux.Lock()
	defer soc.mux.Unlock()
	if _, ok := soc.objects[index]; !ok {
		return Object{}, ErrObjectNotFound
	}
	return *soc.objects[index], nil
}

func (soc *Control) sendObjectChannels(obj Object) {
	soc.chmux.Lock()
	defer soc.chmux.Unlock()
	for i := range soc.channelsObjects {
		select {
		case soc.channelsObjects[i] <- obj:
		default:
		}
	}
}

// Update an object. Also it initializes the one
func (soc *Control) Update(index int, obj Object) (err error) {
	soc.mux.Lock()
	defer soc.mux.Unlock()
	if _, ok := soc.objects[index]; !ok || index < 1 {
		err = ErrObjectNotFound
	} else {
		obj.Id = index
		if obj.TimeConfig == nil {
			obj.TimeConfig = soc.tconfig
		}
		soc.objects[index] = &obj
		soc.objects[index].Init()
		soc.sendObjectChannels(obj)
	}
	if obj.Id > soc.getAutoIncrement() {
		soc.setAutoIncrement(obj.Id)
	}
	return
}

// Insert an object. Also it initializes the one.  Returns Id
func (soc *Control) Insert(obj Object) int {
	soc.mux.Lock()
	defer soc.mux.Unlock()
	if obj.Id < 1 {
		obj.Id = soc.generateId()
	} else if obj.Id > soc.getAutoIncrement() {
		soc.setAutoIncrement(obj.Id)
	}
	if obj.TimeConfig == nil {
		obj.TimeConfig = soc.tconfig
	}
	obj.InitMoving(obj.ReportPeriod)
	soc.objects[obj.Id] = &obj
	return obj.Id
}

// Delete an object by index
func (soc *Control) Delete(index int) (err error) {
	soc.mux.Lock()
	defer soc.mux.Unlock()
	if _, ok := soc.objects[index]; !ok {
		err = ErrObjectNotFound
	} else {
		// We need to send the last message, an idea is to save in a slice of deleted objects
		// and pass to a station
		soc.objects[index].curPos = PositionArrivalTime{}
		soc.sendObjectChannels(*soc.objects[index])
		delete(soc.objects, index)
	}
	return
}
