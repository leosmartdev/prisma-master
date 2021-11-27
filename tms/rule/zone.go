package rule

import (
	"prisma/tms/moc"
	"prisma/tms/db"
	"prisma/tms/db/mongo"
	"prisma/gogroup"
	"context"
	"sync"
	"prisma/tms/client_api"
	api "prisma/tms/client_api"
	"fmt"
	"prisma/tms/log"
	"github.com/globalsign/mgo/bson"
	"time"
)

var replayZoneFilters = bson.M{"utime": bson.M{"$gt": time.Now().Add(-24*time.Hour)}}

// ZoneStore is an interface for getting actual zones
type ZoneStore interface {
	// get zones as a hashmap for searching using O(1)
	GetZones() map[string]*moc.Zone
	GetByName(name string) (*moc.Zone, error)
}

// TmsZoneStorage is a storage for keeping zones and updates of that
type TmsZoneStorage struct {
	mu         sync.Mutex
	zones      map[string]*moc.Zone
	zoneStream <-chan db.GoGetResponse
	ctxt       context.Context
}

// NewTmsZoneStorage returns an instance of the storage for zones
func NewTmsZoneStorage(client *mongo.MongoClient, ctxt gogroup.GoGroup) (*TmsZoneStorage, error) {
	zs := TmsZoneStorage{
		zones: make(map[string]*moc.Zone),
		ctxt:  ctxt,
	}
	if client == nil {
		log.Warn("using without mongo")
		return &zs, nil
	}
	miscDB := mongo.NewMongoMiscData(ctxt, client)
	goRequest := db.GoRequest{
		ObjectType: "prisma.tms.moc.Zone",
	}
	miscRequest := db.GoMiscRequest{
		Req:  &goRequest,
		Ctxt: ctxt,
		Time: &db.TimeKeeper{},
	}
	stream := miscDB.GetPersistentStream(miscRequest, replayZoneFilters, nil)
	zs.zoneStream = stream
outerLoop:
	for {
		select {
		case update, ok := <-zs.zoneStream:
			if !ok || update.Status == api.Status_InitialLoadDone {
				break outerLoop
			}
			zs.zones[update.Contents.Data.(*moc.Zone).Name] = update.Contents.Data.(*moc.Zone)
		case <-ctxt.Done():
			return &zs, nil
		}
	}
	go zs.watch()
	return &zs, nil
}

func (zs *TmsZoneStorage) GetByName(name string) (*moc.Zone, error) {
	zs.mu.Lock()
	defer zs.mu.Unlock()
	if zone, ok := zs.zones[name]; !ok {
		return nil, fmt.Errorf("The zone %s not found", name)
	} else {
		v := *zone
		return &v, nil
	}
}

func (zs *TmsZoneStorage) GetZones() map[string]*moc.Zone {
	zs.mu.Lock()
	defer zs.mu.Unlock()
	zones := make(map[string]*moc.Zone)
	for name, val := range zs.zones {
		curZone := *val
		zones[name] = &curZone
	}
	return zones
}

func (zs *TmsZoneStorage) watch() {
	for {
		select {
		case update, ok := <-zs.zoneStream:
			if !ok {
				log.Error("A channel was closed")
				return
			}
			if update.Status == api.Status_InitialLoadDone {
				continue
			}
			zs.mu.Lock()
			if update.Status == client_api.Status_Current {
				zs.zones[update.Contents.Data.(*moc.Zone).Name] = update.Contents.Data.(*moc.Zone)
			} else {
				if zone, ok := update.Contents.Data.(*moc.Zone); ok {
					delete(zs.zones, zone.Name)
				}
			}
			zs.mu.Unlock()
		case <-zs.ctxt.Done():
			return
		}
	}
}
