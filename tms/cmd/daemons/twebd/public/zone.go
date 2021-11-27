package public

import (
	"math/rand"
	"net/http"
	"prisma/gogroup"
	"prisma/tms/db"
	"prisma/tms/moc"
	"prisma/tms/security"
	"time"

	"fmt"
	"prisma/tms/log"
	"prisma/tms/rest"

	"github.com/golang/protobuf/jsonpb"
	restful "github.com/orolia/go-restful"
	"github.com/pkg/errors"
)

const (
	OBJECT_ZONE  = "prisma.tms.moc.Zone"
	ZONE_CLASSID = "Zone"
	maxRadius    = 251.56
)

type ZonePublic struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	*moc.Zone
}

type ZoneIdResponse struct {
	Id string `json:"id"`
}

type ZoneRest struct {
	miscDb db.MiscDB
	group  gogroup.GoGroup
}

func NewZoneRest(miscDb db.MiscDB, group gogroup.GoGroup) *ZoneRest {
	return &ZoneRest{
		miscDb: miscDb,
		group:  group,
	}
}

func (r *ZoneRest) FindAllGeoJSON(req *restful.Request, rsp *restful.Response) {
	action := moc.Zone_READ.String()
	ctx := req.Request.Context()

	if !authorized(req, rsp, ZONE_CLASSID, action) {
		return
	}

	arr, err := r.miscDb.Get(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: OBJECT_ZONE,
		},
		Ctxt: r.group,
		Time: &db.TimeKeeper{},
	})
	if !errorFree(err, req, rsp, ZONE_CLASSID, action) {
		return
	}

	geoJsonArea := make([]*moc.GeoJsonFeaturePoint, 0, len(arr))
	geoJsonPolygon := make([]*moc.GeoJsonFeaturePolygon, 0, len(arr))
	for _, z := range arr {
		zone, ok := z.Contents.Data.(*moc.Zone)
		if !ok {
			continue
		}

		if zone.Area == nil && zone.Poly == nil {
			continue
		}

		if zone.Area != nil {
			gj := &moc.GeoJsonFeaturePoint{
				Type: "Feature",
				Properties: map[string]string{
					"radius":                fmt.Sprint(zone.Area.Radius),
					"id":                    z.Contents.ID,
					"type":                  "zone",
					"name":                  zone.Name,
					"create_alert_on_enter": fmt.Sprint(zone.CreateAlertOnEnter),
					"create_alert_on_exit":  fmt.Sprint(zone.CreateAlertOnExit),
					"zone_id":               fmt.Sprint(zone.ZoneId),
					"fill_color.r":          fmt.Sprint(zone.FillColor.R),
					"fill_color.g":          fmt.Sprint(zone.FillColor.G),
					"fill_color.b":          fmt.Sprint(zone.FillColor.B),
					"fill_color.a":          fmt.Sprint(zone.FillColor.A),
					"fill_pattern":          zone.FillPattern,
					"stroke_color.r":        fmt.Sprint(zone.StrokeColor.R),
					"stroke_color.g":        fmt.Sprint(zone.StrokeColor.G),
					"stroke_color.b":        fmt.Sprint(zone.StrokeColor.B),
					"stroke_color.a":        fmt.Sprint(zone.StrokeColor.A),
				},
				Geometry: &moc.GeoJsonGeometryPoint{
					Type:        "Point",
					Coordinates: []float64{zone.Area.Center.Longitude, zone.Area.Center.Latitude},
				},
			}
			if zone.Entities != nil {
				marshaler := jsonpb.Marshaler{}
				entities := moc.Entities{
					Entities: zone.Entities,
				}
				estr, err := marshaler.MarshalToString(&entities)
				if err == nil {
					gj.Properties["entities"] = estr
				}
			}

			geoJsonArea = append(geoJsonArea, gj)
		}

		if zone.Poly != nil {
			gj := &moc.GeoJsonFeaturePolygon{
				Type: "Feature",
				Properties: map[string]string{
					"id":                    z.Contents.ID,
					"type":                  "zone",
					"name":                  zone.Name,
					"create_alert_on_enter": fmt.Sprint(zone.CreateAlertOnEnter),
					"create_alert_on_exit":  fmt.Sprint(zone.CreateAlertOnExit),
					"zone_id":               fmt.Sprint(zone.ZoneId),
					"fill_color.r":          fmt.Sprint(zone.FillColor.R),
					"fill_color.g":          fmt.Sprint(zone.FillColor.G),
					"fill_color.b":          fmt.Sprint(zone.FillColor.B),
					"fill_color.a":          fmt.Sprint(zone.FillColor.A),
					"fill_pattern":          zone.FillPattern,
					"stroke_color.r":        fmt.Sprint(zone.StrokeColor.R),
					"stroke_color.g":        fmt.Sprint(zone.StrokeColor.G),
					"stroke_color.b":        fmt.Sprint(zone.StrokeColor.B),
					"stroke_color.a":        fmt.Sprint(zone.StrokeColor.A),
				},
				Geometry: &moc.GeoJsonGeometryPolygon{
					Type:        "Polygon",
					Coordinates: []*moc.GeoJsonCoordinates{},
				},
			}

			for _, v := range zone.Poly.Lines {
				for _, p := range v.Points {
					gj.Geometry.Coordinates = append(gj.Geometry.Coordinates, &moc.GeoJsonCoordinates{
						Latitude:  p.Latitude,
						Longitude: p.Longitude,
					})
				}
			}
			if zone.Entities != nil {
				marshaler := jsonpb.Marshaler{}
				entities := moc.Entities{
					Entities: zone.Entities,
				}
				estr, err := marshaler.MarshalToString(&entities)
				if err == nil {
					gj.Properties["entities"] = estr
				}
			}
			geoJsonPolygon = append(geoJsonPolygon, gj)
		}
	}

	mixedCollection := &moc.GeoJsonMixedCollection{
		Type: "Mixed",
		Points: &moc.GeoJsonFeatureCollectionPoint{
			Type:     "FeatureCollection",
			Features: geoJsonArea,
		},
		Polygons: &moc.GeoJsonFeatureCollectionPolygon{
			Type:     "FeatureCollection",
			Features: geoJsonPolygon,
		},
	}

	security.Audit(ctx, CLASSIDSite, action, security.SUCCESS)
	rest.WriteProtoSafely(rsp, mixedCollection)
}

func (zoneRest *ZoneRest) Get(request *restful.Request, response *restful.Response) {
	const ACTION = "READ"
	if security.HasPermissionForAction(request.Request.Context(), ZONE_CLASSID, ACTION) {
		zoneData, err := zoneRest.miscDb.Get(db.GoMiscRequest{
			Req: &db.GoRequest{
				ObjectType: OBJECT_ZONE,
			},
			Ctxt: zoneRest.group,
			Time: &db.TimeKeeper{},
		})
		if err != nil {
			security.Audit(request.Request.Context(), ZONE_CLASSID, ACTION, "FAIL_ERROR")
			if err := response.WriteError(http.StatusBadRequest, err); err != nil {
				log.Error(err.Error())
			}
			return
		}
		zones := make([]ZonePublic, 0)
		for _, zoneDatum := range zoneData {
			if mocZone, ok := zoneDatum.Contents.Data.(*moc.Zone); ok {
				zones = append(zones, ZonePublic{
					zoneDatum.Contents.ID,
					"zone",
					mocZone,
				})
			}
		}
		security.Audit(request.Request.Context(), ZONE_CLASSID, ACTION, "SUCCESS")
		response.WriteEntity(zones)
	} else {
		security.Audit(request.Request.Context(), ZONE_CLASSID, ACTION, "FAIL")
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}

func (zoneRest *ZoneRest) GetOne(request *restful.Request, response *restful.Response) {
	const ACTION = "READ"
	if security.HasPermissionForAction(request.Request.Context(), ZONE_CLASSID, ACTION) {
		zoneId := request.PathParameter("zone-id")
		zoneData, err := zoneRest.miscDb.Get(db.GoMiscRequest{
			Req: &db.GoRequest{
				Obj: &db.GoObject{
					ID: zoneId,
				},
				ObjectType: OBJECT_ZONE,
			},
			Ctxt: zoneRest.group,
			Time: &db.TimeKeeper{},
		})
		if err != nil {
			security.Audit(request.Request.Context(), ZONE_CLASSID, ACTION, "FAIL_ERROR")
			response.WriteError(http.StatusBadRequest, err)
			return
		}
		zones := make([]ZonePublic, 0)
		for _, zoneDatum := range zoneData {
			if mocZone, ok := zoneDatum.Contents.Data.(*moc.Zone); ok {
				zones = append(zones, ZonePublic{
					zoneDatum.Contents.ID,
					"zone",
					mocZone,
				})
			}
		}
		if len(zones) > 0 {
			security.Audit(request.Request.Context(), ZONE_CLASSID, ACTION, "SUCCESS")
			response.WriteEntity(zones[0])
		} else {
			security.Audit(request.Request.Context(), ZONE_CLASSID, ACTION, "FAIL_NOTFOUND")
			response.WriteHeaderAndEntity(http.StatusNotFound, "")
		}
	} else {
		security.Audit(request.Request.Context(), ZONE_CLASSID, ACTION, "FAIL")
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}

func (zoneRest *ZoneRest) Post(request *restful.Request, response *restful.Response) {
	const ACTION = "CREATE"

	if security.HasPermissionForAction(request.Request.Context(), ZONE_CLASSID, ACTION) {
		mocZone := new(moc.Zone)
		err := jsonpb.Unmarshal(request.Request.Body, mocZone)

		if !errorFree(err, request, response, ZONE_CLASSID, ACTION) {
			return
		}
		if mocZone.Area != nil && mocZone.Area.Radius > maxRadius {
			security.Audit(request.Request.Context(), ZONE_CLASSID, ACTION, "FAIL_ERROR")
			response.WriteError(http.StatusBadRequest, errors.New("Too big radius"))
			return
		}
		// Assign an omnicom GEO_ID to the C2 Zone. ZoneID is unique for the zones
		// There is an infinetely small change of collision.
		// The probablity of collision increaes with the number of zones we store in C2
		// When we will have 100k zones in C2 the probability of collision will be 0.23%
		// Given that 100k zones is a lot, no code shall be written to check for the availability of an id before inserting
		// If the id already exist the operation will fail, and the work arround is to try again.
		rand.Seed(time.Now().Unix())
		mocZone.ZoneId = uint32(rand.Int63n(4294967295))

		goreq := db.GoMiscRequest{
			Req: &db.GoRequest{
				ObjectType: OBJECT_ZONE,
				Obj: &db.GoObject{
					Data: mocZone,
				},
			},
			Ctxt: zoneRest.group,
			Time: &db.TimeKeeper{},
		}
		if mocZone.DatabaseId != "" {
			goreq.Req.Obj.ID = mocZone.DatabaseId
		}
		upsertResponse, err := zoneRest.miscDb.Upsert(goreq)
		if err != nil {
			security.Audit(request.Request.Context(), ZONE_CLASSID, ACTION, "FAIL_ERROR")
			response.WriteError(http.StatusBadRequest, err)
			return
		}
		zone := ZonePublic{
			upsertResponse.Id,
			"zone",
			mocZone,
		}
		security.Audit(request.Request.Context(), ZONE_CLASSID, ACTION, "SUCCESS", zone.ID)
		response.WriteHeaderAndEntity(http.StatusCreated, zone)
	} else {
		security.Audit(request.Request.Context(), ZONE_CLASSID, ACTION, "FAIL")
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}

func (zoneRest *ZoneRest) Put(request *restful.Request, response *restful.Response) {
	const ACTION = "UPDATE"
	if security.HasPermissionForAction(request.Request.Context(), ZONE_CLASSID, ACTION) {
		security.Audit(request.Request.Context(), ZONE_CLASSID, ACTION, "SUCCESS")
		response.WriteEntity(map[string]string{
			"service": "ZoneRest.PUT",
		})
	} else {
		security.Audit(request.Request.Context(), ZONE_CLASSID, ACTION, "FAIL")
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}

func (zoneRest *ZoneRest) Delete(request *restful.Request, response *restful.Response) {
	const ACTION = "DELETE"
	if security.HasPermissionForAction(request.Request.Context(), ZONE_CLASSID, ACTION) {
		zoneId := request.PathParameter("zone-id")
		_, err := zoneRest.miscDb.Expire(db.GoMiscRequest{
			Req: &db.GoRequest{
				ObjectType: OBJECT_ZONE,
				Obj: &db.GoObject{
					ID: zoneId,
				},
			},
			Ctxt: zoneRest.group,
			Time: &db.TimeKeeper{},
		})
		if err != nil {
			security.Audit(request.Request.Context(), ZONE_CLASSID, ACTION, "FAIL_ERROR")
			response.WriteError(http.StatusBadRequest, err)
			return
		} else {
			security.Audit(request.Request.Context(), ZONE_CLASSID, ACTION, "SUCCESS")
			response.WriteHeaderAndEntity(http.StatusOK, &ZoneIdResponse{Id: zoneId})
		}
	} else {
		security.Audit(request.Request.Context(), ZONE_CLASSID, ACTION, "FAIL")
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
}
