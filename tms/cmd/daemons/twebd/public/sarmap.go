package public

import (
	"prisma/gogroup"
	"prisma/tms/db"
	"prisma/tms/db/mongo"
	"prisma/tms/log"
	"prisma/tms/moc"
	"prisma/tms/rest"

	restful "github.com/orolia/go-restful" 
	"github.com/golang/protobuf/proto"
)

type SarmapRest struct {
	sarmapDb db.SarmapDB
}

func NewSarmapRest(group gogroup.GoGroup, mongoClient *mongo.MongoClient) *SarmapRest {
	return &SarmapRest{
		sarmapDb: mongo.NewMongoSarmapDb(group, mongoClient),
	}
}

func (r *SarmapRest) ReadAll(req *restful.Request, rsp *restful.Response) {
	ACTION := moc.Incident_READ.String()
	ctx := req.Request.Context()

	geojsons, err := r.sarmapDb.FindAll(ctx)
	if err != nil {
		log.Error("Sarmap FindAll error %+v", err)
	}
	if !errorFree(err, req, rsp, CLASSIDDevice, ACTION) {
		return
	}
	log.Debug("Sarmap Geojson output %+v", geojsons)

	err = rest.WriteProtoSpliceSafely(rsp, toMessagesFromGeoJson(geojsons))
	if err != nil {
		log.Error("Sarmap Writeprotosplicacesafely Error: %+v", err)
	}
}

func toMessagesFromGeoJson(geojsons []*moc.GeoJsonFeaturePoint) []proto.Message {
	var messages []proto.Message
	for _, geoJson := range geojsons {
		messages = append(messages, geoJson)
	}
	return messages
}
