package public

import (
	"errors"
	"net/http"
	"time"

	"prisma/gogroup"
	"prisma/tms"
	api "prisma/tms/client_api"
	"prisma/tms/db"
	"prisma/tms/db/mongo"
	"prisma/tms/devices"
	"prisma/tms/log"
	"prisma/tms/moc"
	"prisma/tms/rest"
	"prisma/tms/sar"
	"prisma/tms/security"
	"prisma/tms/tmsg"
	"prisma/tms/util/ident"

	"github.com/globalsign/mgo/bson"
	restful "github.com/orolia/go-restful"

	"github.com/go-openapi/spec"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/wrappers"
)

const (
	HeadingNa = 511
	RotNa     = -128
	CourseNa  = 360.0
	SpeedNa   = -1
	HexIDNa   = ""
)

type TrackRest struct {
	trackdb db.TrackDB
	regdb   db.RegistryDB
	group   gogroup.GoGroup
}

type ManualTrackPublic struct {
	RegistryID string  `json:"registryId"`
	Name       string  `json:"name"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Speed      float64 `json:"speed"`
	Course     float64 `json:"course"`
	Heading    float64 `json:"heading"`
	Rot        float64 `json:"rateOfTurn"`
	HexID      string  `json:"hexid"`
}

var (
	ClassIDTrack       = "Track"
	PARAMETER_TRACK_ID = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "track-id",
			In:       "path",
			Required: true,
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					MinLength: &[]int64{32}[0],
					MaxLength: &[]int64{32}[0],
					Pattern:   "[0-9a-fA-F]{32}",
				},
			},
		},
	}

	PARAMETER_FIRST_TARGET = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "first",
			In:       "query",
			Required: false,
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Pattern: "^(true|false)$",
				},
			},
		},
	}

	SCHEMA_MANUAL_TRACK = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Required: []string{"name", "latitude", "longitude"},
			Properties: map[string]spec.Schema{
				"registryId": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{32}[0],
						MaxLength: &[]int64{32}[0],
						Pattern:   "[0-9a-fA-F]{32}",
					},
				},
				"name": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{1}[0],
						MaxLength: &[]int64{255}[0],
					},
				},
				"latitude": {
					SchemaProps: spec.SchemaProps{
						Minimum:   &[]float64{-90}[0],
						Maximum:   &[]float64{90}[0],
						MinLength: &[]int64{1}[0],
						MaxLength: &[]int64{18}[0],
						Pattern:   "[0-9\\.\\-]{1,18}",
					},
				},
				"longitude": {
					SchemaProps: spec.SchemaProps{
						Minimum:   &[]float64{-180}[0],
						Maximum:   &[]float64{180}[0],
						MinLength: &[]int64{1}[0],
						MaxLength: &[]int64{18}[0],
						Pattern:   "[0-9\\.\\-]{1,18}",
					},
				},
				"speed": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{1}[0],
						MaxLength: &[]int64{18}[0],
						Pattern:   "[0-9\\.\\-]{1,10}",
					},
				},
				"course": {
					SchemaProps: spec.SchemaProps{
						Minimum:   &[]float64{0}[0],
						Maximum:   &[]float64{360}[0],
						MinLength: &[]int64{1}[0],
						MaxLength: &[]int64{18}[0],
						Pattern:   "[0-9\\.\\-]{1,18}",
					},
				},
				"heading": {
					SchemaProps: spec.SchemaProps{
						Minimum:   &[]float64{0}[0],
						Maximum:   &[]float64{360}[0],
						MinLength: &[]int64{1}[0],
						MaxLength: &[]int64{18}[0],
						Pattern:   "[0-9\\.\\-]{1,18}",
					},
				},
				"rateOfTurn": {
					SchemaProps: spec.SchemaProps{
						Minimum:   &[]float64{-128}[0],
						Maximum:   &[]float64{127}[0],
						MinLength: &[]int64{1}[0],
						MaxLength: &[]int64{18}[0],
						Pattern:   "[0-9\\.\\-]{1,18}",
					},
				},
				"hexid": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{15}[0],
						MaxLength: &[]int64{15}[0],
						Pattern:   "[0-9a-fA-F]{15}",
					},
				},
			},
		},
	}
)

func NewTrackRest(client *mongo.MongoClient, group gogroup.GoGroup) *TrackRest {
	return &TrackRest{
		trackdb: mongo.NewMongoTracks(group, client),
		regdb:   mongo.NewMongoRegistry(group, client),
		group:   group,
	}
}

func (r *TrackRest) GetOne(request *restful.Request, response *restful.Response) {
	ACTION := moc.Track_GET.String()
	if !authorized(request, response, ClassIDTrack, ACTION) {
		return
	}

	trackID, errs := rest.SanitizeValidatePathParameter(request, PARAMETER_TRACK_ID)
	if !valid(errs, request, response, ClassIDTrack, ACTION) {
		return
	}
	first, errs := rest.SanitizeValidateQueryParameter(request, PARAMETER_FIRST_TARGET)
	if !valid(errs, request, response, ClassIDTrack, ACTION) {
		return
	}

	var track *tms.Track
	var err error
	if first == "true" {
		track, err = r.trackdb.GetFirstTrack(bson.M{"track_id": trackID})
	} else {
		track, err = r.trackdb.GetLastTrack(bson.M{"track_id": trackID})
	}
	if !errorFree(err, request, response, ClassIDTrack, ACTION) {
		return
	}
	if len(track.Targets) == 0 {
		errorFree(errors.New("track does not have a target"), request, response, ClassIDTrack, ACTION)
		return
	}
	target := track.Targets[0]
	var metadata *tms.TrackMetadata
	if len(track.Metadata) > 0 {
		metadata = track.Metadata[0]
	}
	infoResponse := &InfoResponse{
		ID:         trackID,
		TrackID:    track.Id,
		RegistryID: track.RegistryId,
		LookupID:   track.LookupID(),
		Target:     target,
		Metadata:   metadata,
	}

	security.Audit(request.Request.Context(), CLASSID, ACTION, "SUCCESS")
	rest.WriteHeaderAndEntitySafely(response, http.StatusOK, infoResponse)
}

func (r *TrackRest) Get(request *restful.Request, response *restful.Response) {
	ACTION := moc.Track_GET.String()
	if !authorized(request, response, ClassIDTrack, ACTION) {
		return
	}

	result, err := r.trackdb.GetTracks(db.GoTrackRequest{
		Req:  &api.TrackRequest{},
		Ctxt: r.group,
	})
	if !errorFree(err, request, response, ClassIDTrack, ACTION) {
		return
	}

	infos := make([]*InfoResponse, 0, len(result.Tracks))
	for _, track := range result.Tracks {
		target := track.Targets[0]
		var metadata *tms.TrackMetadata
		if len(track.Metadata) > 0 {
			metadata = track.Metadata[0]
		}
		infoResponse := &InfoResponse{
			TrackID:    track.Id,
			RegistryID: track.RegistryId,
			LookupID:   track.LookupID(),
			Target:     target,
			Metadata:   metadata,
		}
		infos = append(infos, infoResponse)
	}

	security.Audit(request.Request.Context(), CLASSID, ACTION, "SUCCESS")
	rest.WriteHeaderAndEntitySafely(response, http.StatusOK, infos)
}

func (r *TrackRest) Post(request *restful.Request, response *restful.Response) {
	ACTION := moc.Track_UPDATE.String()
	if !authorized(request, response, ClassIDTrack, ACTION) {
		return
	}
	track := &ManualTrackPublic{
		Speed:   SpeedNa,
		Heading: HeadingNa,
		Rot:     RotNa,
		Course:  CourseNa,
		HexID:   HexIDNa,
	}
	errs := rest.SanitizeValidateReadEntity(request, SCHEMA_MANUAL_TRACK, track)
	if !valid(errs, request, response, ClassIDTrack, ACTION) {
		return
	}
	id := ident.With("manual", bson.NewObjectId().Hex()).Hash()
	if len(track.RegistryID) == 0 {
		track.RegistryID = id
	}
	pnow, err := ptypes.TimestampProto(time.Now())
	if !errorFree(err, request, response, ClassIDTrack, ACTION) {
		return
	}
	target := &tms.Track{
		Id:         track.RegistryID,
		RegistryId: track.RegistryID,
		Targets: []*tms.Target{
			{
				Time:       pnow,
				IngestTime: pnow,
				UpdateTime: pnow,
				Type:       devices.DeviceType_Manual,
				Position: &tms.Point{
					Latitude:  track.Latitude,
					Longitude: track.Longitude,
				},
				Manual: &devices.ManualDevice{},
			},
		},
		Metadata: []*tms.TrackMetadata{
			{
				Time:       pnow,
				IngestTime: pnow,
				Type:       devices.DeviceType_Manual,
				Name:       track.Name,
			},
		},
	}
	if track.Speed != SpeedNa {
		target.Targets[0].Speed = &wrappers.DoubleValue{Value: track.Speed}
	}
	if track.Heading != HeadingNa {
		target.Targets[0].Heading = &wrappers.DoubleValue{Value: track.Heading}
	}
	if track.Rot != RotNa {
		target.Targets[0].RateOfTurn = &wrappers.DoubleValue{Value: track.Rot}
	}
	if track.Course != CourseNa {
		target.Targets[0].Course = &wrappers.DoubleValue{Value: track.Course}
	}
	if track.HexID != HexIDNa {
		target.Targets[0].Sarmsg = &sar.SarsatMessage{
			SarsatAlert: &sar.SarsatAlert{
				Beacon: &sar.Beacon{
					HexId: track.HexID,
				},
			},
		}
	}
	err = sendTrack(r.group, target)
	if !errorFree(err, request, response, ClassIDTrack, ACTION) {
		log.Error("%+v", err)
		return
	}
	security.Audit(request.Request.Context(), ClassIDTrack, ACTION, "SUCCESS")
	rest.WriteHeaderAndEntitySafely(response, http.StatusOK, track)
}

func (r *TrackRest) Delete(request *restful.Request, response *restful.Response) {
	ACTION := moc.Track_DELETE.String()

	if !authorized(request, response, ClassIDTrack, ACTION) {
		return
	}

	registryID, errs := rest.SanitizeValidatePathParameter(request, PARAMETER_REGISTRY_ID)
	if !valid(errs, request, response, ClassIDTrack, ACTION) {
		return
	}

	entry, err := r.regdb.Get(registryID)
	if !errorFree(err, request, response, ClassIDTrack, ACTION) {
		return
	}

	if entry.DeviceType != devices.DeviceType_Manual {
		security.Audit(request.Request.Context(), CLASSID, ACTION, "FAIL")
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.New("not a manual track"))
		return
	}
	if entry.Target == nil || entry.Target.Position == nil {
		security.Audit(request.Request.Context(), CLASSID, ACTION, "FAIL")
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.New("no valid previous position"))
		return
	}
	previousLocation := entry.Target.Position
	pnow, err := ptypes.TimestampProto(time.Now())
	if !errorFree(err, request, response, ClassIDTrack, ACTION) {
		return
	}

	target := &tms.Track{
		Id:         registryID,
		RegistryId: registryID,
		Targets: []*tms.Target{
			{
				Time:       pnow,
				IngestTime: pnow,
				Type:       devices.DeviceType_Manual,
				Position: &tms.Point{
					Latitude:  previousLocation.Latitude,
					Longitude: previousLocation.Longitude,
				},
				Manual: &devices.ManualDevice{
					IssueTimeout: true,
				},
			},
		},
	}
	err = sendTrack(r.group, target)
	if !errorFree(err, request, response, ClassIDTrack, ACTION) {
		return
	}

	security.Audit(request.Request.Context(), CLASSID, ACTION, "SUCCESS")
	response.WriteHeader(http.StatusOK)
}

func sendTrack(ctxt gogroup.GoGroup, t *tms.Track) error {
	pnow, err := ptypes.TimestampProto(time.Now())
	if err != nil {
		return err
	}
	body, err := tmsg.PackFrom(t)
	if err != nil {
		return err
	}
	m := &tms.TsiMessage{
		Destination: []*tms.EndPoint{
			{
				Site: tmsg.GClient.ResolveSite(""),
			},
		},
		WriteTime: pnow,
		SendTime:  pnow,
		Body:      body,
	}
	tmsg.GClient.Send(ctxt, m)
	return nil
}
