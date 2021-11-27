package mongo

import (
	"fmt"
	"prisma/gogroup"
	"prisma/tms/client_api"
	"prisma/tms/db"
	"prisma/tms/log"
	markerProto "prisma/tms/marker"
	"prisma/tms/moc"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// ObjectTypes
const (
	MarkerObjectType      = "prisma.tms.marker.Marker"
	MarkerImageObjectType = "prisma.tms.marker.MarkerImage"
	MARKER_NOTE_TYPE      = "Marker"
)

// MongoDB collections
const CollectionMarker = "markers"
const CollectionMarkerImage = "marker_images"

type MongoMarkerDb struct {
	mongo *MongoClient
	ctxt  gogroup.GoGroup
	misc  db.MiscDB
}

func NewMarkerDb(ctxt gogroup.GoGroup, client *MongoClient) db.MarkerDB {
	markerDb := &MongoMarkerDb{
		mongo: client,
		misc:  NewMongoMiscData(ctxt, client),
		ctxt:  ctxt,
	}
	return markerDb
}

func (markerDb *MongoMarkerDb) UpsertMarker(marker *markerProto.Marker) (*client_api.UpsertResponse, error) {
	res, err := markerDb.misc.Upsert(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: MarkerObjectType,
			Obj: &db.GoObject{
				ID:   marker.Id,
				Data: marker,
			},
		},
		Ctxt: markerDb.ctxt,
		Time: &db.TimeKeeper{},
	})

	return res, err
}

func (markerDb *MongoMarkerDb) FindOneMarker(markerId string, withDeleted bool) (*markerProto.Marker, error) {
	var marker *markerProto.Marker
	markerData, err := markerDb.misc.Get(db.GoMiscRequest{
		Req: &db.GoRequest{
			Obj: &db.GoObject{
				ID: markerId,
			},
			ObjectType: MarkerObjectType,
		},
		Ctxt: markerDb.ctxt,
		Time: &db.TimeKeeper{},
	})
	if err == nil {
		markers := make([]*markerProto.Marker, 0)
		for _, markerDatum := range markerData {
			if mocMarker, ok := markerDatum.Contents.Data.(*markerProto.Marker); ok {
				if withDeleted == false && mocMarker.Deleted == true {
					continue
				}

				markers = append(markers, mocMarker)
			}
		}
		if len(markers) > 0 {
			marker = markers[0]
		} else {
			err = db.ErrorNotFound
		}
	}
	return marker, err
}

func (markerDb *MongoMarkerDb) DeleteMarker(markerId string) error {
	db := markerDb.mongo.DB()
	defer markerDb.mongo.Release(db)

	// Delete from marker collection
	query := bson.M{
		"_id": bson.ObjectIdHex(markerId),
	}

	update := bson.M{
		"$set": bson.M{
			"me.deleted": true,
		},
	}

	err := db.C(CollectionMarker).Update(query, update)
	if err != nil {
		return err
	}

	// Delete from Incident's log entry list(if exists)
	query = bson.M{
		"me.log": bson.M{
			"$elemMatch": bson.M{
				"entity.id":   markerId,
				"entity.type": "marker",
				"deleted":     false,
			},
		},
	}

	update = bson.M{
		"$set": bson.M{
			"me.log.$.deleted": true,
		},
	}

	err = db.C(CollectionIncident).Update(query, update)
	if err != nil && err != mgo.ErrNotFound {
		return err
	}

	return nil
}

func (markerDb *MongoMarkerDb) UpsertMarkerImage(markerImage *markerProto.MarkerImage) (*client_api.UpsertResponse, error) {
	res, err := markerDb.misc.Upsert(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: MarkerImageObjectType,
			Obj: &db.GoObject{
				ID:   markerImage.Id,
				Data: markerImage,
			},
		},
		Ctxt: markerDb.ctxt,
		Time: &db.TimeKeeper{},
	})

	return res, err
}

func (markerDb *MongoMarkerDb) FindAllMarkerImages() ([]*markerProto.MarkerImage, error) {
	res, err := markerDb.misc.Get(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: MarkerImageObjectType,
		},
		Ctxt: markerDb.ctxt,
		Time: &db.TimeKeeper{},
	})

	markerImages := make([]*markerProto.MarkerImage, 0)
	for _, markerImageDatum := range res {
		if mocMarkerImage, ok := markerImageDatum.Contents.Data.(*markerProto.MarkerImage); ok {
			markerImages = append(markerImages, mocMarkerImage)
		}
	}

	return markerImages, err
}

func (markerDb *MongoMarkerDb) GetPersistentStream(pipeline []bson.M) *db.MarkerStream {
	downstream := db.NewMarkerStream()
	upstream := markerDb.misc.GetPersistentStream(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: MarkerObjectType,
		},
		Ctxt: markerDb.ctxt,
	}, bson.M{
		"utime": bson.M{
			"$gt": time.Now().Add(-200000 * time.Hour),
		},
	}, pipeline)
	go markerDb.handleStream(markerDb.ctxt, upstream, downstream, true)
	return downstream
}

func (markerDb *MongoMarkerDb) handleStream(ctx gogroup.GoGroup, upstream <-chan db.GoGetResponse, downstream *db.MarkerStream, persist bool) {
	for {
		select {
		case update, ok := <-upstream:
			if !ok {
				if !persist {
					log.Error("A channel was closed")
					return
				}
				log.Warn("Connection was lost, try to reconnect")
				upstream = markerDb.misc.GetPersistentStream(db.GoMiscRequest{
					Req: &db.GoRequest{
						ObjectType: MarkerObjectType,
					},
					Ctxt: ctx,
				}, nil, nil)
				continue
			}
			object := update.Contents
			if object == nil {
				log.Error("Wrong record %v", log.Spew(update))
				continue
			}
			marker := object.Data.(*markerProto.Marker)

			// Generate GeoJSON
			geojson := new(moc.GeoJsonFeaturePoint)

			properties := map[string]string{}
			properties["id"] = marker.Id
			properties["type"] = "Marker"
			properties["marker.type"] = marker.Type
			properties["description"] = marker.Description

			if marker.Deleted {
				continue
			}

			if marker.Type == "Shape" {
				properties["shape"] = marker.Shape
				properties["color.r"] = fmt.Sprint(marker.Color.R)
				properties["color.g"] = fmt.Sprint(marker.Color.G)
				properties["color.b"] = fmt.Sprint(marker.Color.B)
				properties["color.a"] = fmt.Sprint(marker.Color.A)
			} else if marker.Type == "Image" {
				properties["image.id"] = marker.ImageMetadata.Id
			}

			geojson = &moc.GeoJsonFeaturePoint{
				Type:       "Feature",
				Properties: properties,
				Geometry: &moc.GeoJsonGeometryPoint{
					Type:        "Point",
					Coordinates: []float64{marker.Position.Longitude, marker.Position.Latitude},
				},
			}

			downstream.Updates <- geojson
		case <-ctx.Done():
			return
		}
	}
}
