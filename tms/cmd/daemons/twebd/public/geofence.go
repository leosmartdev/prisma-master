package public

import (
	restful "github.com/orolia/go-restful" 
	"net/http"
	"prisma/gogroup"
	"prisma/tms/db"
	"prisma/tms/security"
	"prisma/tms/log"
	"prisma/tms/moc"
	"prisma/tms/client_api"
	"github.com/globalsign/mgo/bson"
	"prisma/tms/rest"
	"github.com/go-openapi/spec"
)

const (
	OBJECT_GEOFENCE = "prisma.tms.moc.GeoFence"
)

// GeoFencePublic is used to structure a response for complete information about a geofence at geofence endpoints
type GeoFencePublic struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	*moc.GeoFence
}

// GeoFenceResponse is used to response about id of a geofence only
type GeoFenceResponse struct {
	Id string `json:"id"`
}

// GeoFenceRest is a structure to determine rest api for geofence endpoints
type GeoFenceRest struct {
	miscDb db.MiscDB
	group  gogroup.GoGroup
}

var parameterGeofenceId = spec.Parameter{
	ParamProps: spec.ParamProps{
		Name:     "geofence-id",
		In:       "path",
		Required: true,
		Schema: &spec.Schema{
			SchemaProps: spec.SchemaProps{
				MinLength: &[]int64{24}[0],
				MaxLength: &[]int64{24}[0],
				Pattern:   "[0-9a-fA-F]{24}",
			},
		},
	},
	SimpleSchema: spec.SimpleSchema{
		Type:   "string",
		Format: "hexadecimal",
	},
}

// NewGeoFenceRest returns an instance for rest api of geofence endpoints
func NewGeoFenceRest(miscDb db.MiscDB, group gogroup.GoGroup) *GeoFenceRest {
	return &GeoFenceRest{
		miscDb: miscDb,
		group:  group,
	}
}

// Save data about a geofence into the database
// Also it can update a record in database if was passed database_id
func (gfr *GeoFenceRest) save(gf *moc.GeoFence) (*client_api.UpsertResponse, error) {
	goreq := db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: OBJECT_GEOFENCE,
			Obj: &db.GoObject{
				Data: gf,
			},
		},
		Ctxt: gfr.group,
		Time: &db.TimeKeeper{},
	}
	if gf.DatabaseId != "" && !bson.IsObjectIdHex(gf.DatabaseId) {
		return nil, db.ErrorBadID
	}
	if gf.DatabaseId != "" {
		goreq.Req.Obj.ID = gf.DatabaseId
	}
	upsertResponse, err := gfr.miscDb.Upsert(goreq)
	return upsertResponse, err
}

// Get an array of geofence from the database
// To get all records pass "" for id
func (gfr *GeoFenceRest) getArray(id string) ([]*moc.GeoFence, error) {
	gfBuffer := make([]*moc.GeoFence, 0)
	if id != "" && !bson.IsObjectIdHex(id) {
		return gfBuffer, db.ErrorBadID
	}
	goreq := db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: OBJECT_GEOFENCE,
		},
		Ctxt: gfr.group,
		Time: &db.TimeKeeper{},
	}
	if id != "" {
		goreq.Req.Obj = &db.GoObject{
			ID: id,
		}
	}
	gfences, err := gfr.miscDb.Get(goreq)

	if err != nil {
		return nil, err
	}
	for _, gf := range gfences {
		gfBuffer = append(gfBuffer, gf.Contents.Data.(*moc.GeoFence))
	}
	return gfBuffer, nil
}

func (gfr *GeoFenceRest) Get(request *restful.Request, response *restful.Response) {
	const CLASSID = "Geofence"
	const ACTION = "READ"

	if !authorized(request, response, CLASSID, ACTION) {
		return
	}
	gfences, err := gfr.getArray("")
	if !errorFree(err, request, response, CLASSID, ACTION) {
		return
	}
	security.Audit(request.Request.Context(), CLASSID, ACTION, "SUCCESS")
	if err := response.WriteEntity(gfences); err != nil {
		log.Error(err.Error())
	}
}

func (gfr *GeoFenceRest) GetOne(request *restful.Request, response *restful.Response) {
	const CLASSID = "Geofence"
	const ACTION = "READ"

	if !authorized(request, response, CLASSID, ACTION) {
		return
	}
	geofenceId, errs := rest.SanitizeValidatePathParameter(request, parameterGeofenceId)
	if !valid(errs, request, response, CLASSIDDevice, ACTION) {
		return
	}
	gfences, err := gfr.getArray(geofenceId)
	if !errorFree(err, request, response, CLASSID, ACTION) {
		return
	}
	security.Audit(request.Request.Context(), CLASSID, ACTION, "SUCCESS")
	response.WriteEntity(gfences)
}

func (gfr *GeoFenceRest) Post(request *restful.Request, response *restful.Response) {
	const CLASSID = "Geofence"
	const ACTION = "CREATE"

	if !authorized(request, response, CLASSID, ACTION) {
		return
	}
	gf := new(moc.GeoFence)
	err := request.ReadEntity(&gf)
	if err != nil {
		errorFree(err, request, response, CLASSID, ACTION)
		return
	}
	ur, err := gfr.save(gf)
	if !errorFree(err, request, response, CLASSID, ACTION) {
		return
	}
	gpublic := GeoFencePublic{
		ur.Id,
		"geofence",
		gf,
	}
	security.Audit(request.Request.Context(), CLASSID, ACTION, "SUCCESS", gpublic.ID)
	if err := response.WriteHeaderAndEntity(http.StatusCreated, gpublic); err != nil {
		log.Error(err.Error())
	}

}

func (gfr *GeoFenceRest) Put(request *restful.Request, response *restful.Response) {
	const CLASSID = "Geofence"
	const ACTION = "UPDATE"

	if !authorized(request, response, CLASSID, ACTION) {
		return
	}
	gf := new(moc.GeoFence)
	err := request.ReadEntity(&gf)
	if !errorFree(err, request, response, CLASSID, ACTION) {
		return
	}
	geofenceId, errs := rest.SanitizeValidatePathParameter(request, parameterGeofenceId)
	if !valid(errs, request, response, CLASSIDDevice, ACTION) {
		return
	}
	fences, err := gfr.getArray(geofenceId)
	if !errorFree(err, request, response, CLASSID, ACTION) {
		return
	}
	if len(fences) != 1 {
		errorFree(db.ErrorNotFound, request, response, CLASSID, ACTION)
		return
	}
	gf.DatabaseId = geofenceId
	ur, err := gfr.save(gf)
	if !errorFree(err, request, response, CLASSID, ACTION) {
		return
	}
	gpublic := GeoFencePublic{
		ur.Id,
		"geofence",
		gf,
	}
	security.Audit(request.Request.Context(), CLASSID, ACTION, "SUCCESS", gpublic.ID)
	response.WriteHeaderAndEntity(http.StatusCreated, gpublic)
}

func (gfr *GeoFenceRest) Delete(request *restful.Request, response *restful.Response) {
	const CLASSID = "Geofence"
	const ACTION = "DELETE"

	if !authorized(request, response, CLASSID, ACTION) {
		return
	}
	geofenceId, errs := rest.SanitizeValidatePathParameter(request, parameterGeofenceId)
	if !valid(errs, request, response, CLASSIDDevice, ACTION) {
		return
	}
	fences, err := gfr.getArray(geofenceId)
	if !errorFree(err, request, response, CLASSID, ACTION) {
		return
	}
	if len(fences) != 1 {
		errorFree(db.ErrorNotFound, request, response, CLASSID, ACTION)
		return
	}
	_, err = gfr.miscDb.Expire(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: OBJECT_GEOFENCE,
			Obj: &db.GoObject{
				ID: geofenceId,
			},
		},
		Ctxt: gfr.group,
		Time: &db.TimeKeeper{},
	})
	if !errorFree(err, request, response, CLASSID, ACTION) {
		return
	}
	security.Audit(request.Request.Context(), CLASSID, ACTION, "SUCCESS")
	if err := response.WriteHeaderAndEntity(http.StatusOK, &GeoFenceResponse{Id: geofenceId}); err != nil {
		log.Error(err.Error())
	}
}
