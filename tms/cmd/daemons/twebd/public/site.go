package public

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"prisma/gogroup"
	"prisma/tms/db"
	"prisma/tms/moc"
	"prisma/tms/rest"
	"prisma/tms/security"
	"prisma/tms/security/service"

	"prisma/tms/db/mongo"
	"prisma/tms/log"

	"github.com/go-openapi/spec"
	"github.com/golang/protobuf/proto"
	restful "github.com/orolia/go-restful"
	restfulspec "github.com/orolia/go-restful-openapi"
)

const (
	CLASSIDSite = security.CLASSIDSite
)

var (
	schemaSiteCreate = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Required:   []string{"siteId", "name"},
			Properties: map[string]spec.Schema{},
		},
	}
)

type SiteRest struct {
	siteDb db.SiteDB
	routes []string
}

func RegisterSiteRest(ctx gogroup.GoGroup, service *restful.WebService, routeAuthorizer *service.RouteAuthorizer) {
	r := &SiteRest{}
	r.siteDb = mongo.NewSiteDb(ctx)
	service.Route(service.POST(r.registerRoute("/site")).To(r.create).
		Doc("create a remote site").
		Metadata(restfulspec.KeyOpenAPITags, []string{CLASSIDSite}).
		Returns(http.StatusOK, http.StatusText(http.StatusOK), moc.Site{}))
	service.Route(service.GET(r.registerRoute("/site")).To(r.findAll).
		Doc("get remote sites").
		Metadata(restfulspec.KeyOpenAPITags, []string{CLASSIDSite}).
		Returns(http.StatusOK, http.StatusText(http.StatusOK), []moc.Site{}))
	service.Route(service.GET(r.registerRoute("/site/geo")).To(r.findAllGeoJSON).
		Doc("get remote sites").
		Metadata(restfulspec.KeyOpenAPITags, []string{CLASSIDSite}).
		Returns(http.StatusOK, http.StatusText(http.StatusOK), []moc.Site{}))
	service.Route(service.GET(r.registerRoute("/site/{id}")).To(r.findById).
		Doc("get single remote site").
		Metadata(restfulspec.KeyOpenAPITags, []string{CLASSIDSite}).
		Returns(http.StatusOK, http.StatusText(http.StatusOK), []moc.Site{}))
	routeAuthorizer.Add(r)
}

func (r *SiteRest) create(req *restful.Request, rsp *restful.Response) {
	action := moc.Site_CREATE.String()
	ctx := req.Request.Context()
	if !authorized(req, rsp, CLASSIDSite, action) {
		return
	}
	site := &moc.Site{}
	errs := rest.SanitizeValidateReadProto(req, schemaSiteCreate, site)
	if !valid(errs, req, rsp, CLASSIDSite, action) {
		return
	}
	site, err := r.siteDb.Create(ctx, site)
	if !errorFree(err, req, rsp, CLASSIDSite, action) {
		return
	}
	security.Audit(ctx, CLASSIDSite, action, security.SUCCESS)
	rest.WriteHeaderAndProtoSafely(rsp, http.StatusCreated, site)
}

func (r *SiteRest) findById(req *restful.Request, rsp *restful.Response) {
	ACTION := moc.Site_READ.String()
	ctxt := req.Request.Context()
	if !authorized(req, rsp, CLASSIDSite, ACTION) {
		return
	}
	siteId, errs := rest.SanitizeValidatePathParameter(req, parameterSiteId)
	if !valid(errs, req, rsp, CLASSIDSite, ACTION) {
		return
	}
	site, err := r.siteDb.FindById(ctxt, siteId)
	if err != nil {
		s, serr := strconv.ParseUint(siteId, 10, 32)
		if serr != nil {
			errorFree(err, req, rsp, CLASSIDSite, ACTION)
			return
		}
		site = &moc.Site{
			SiteId: uint32(s),
		}
		err = r.siteDb.FindBySiteId(ctxt, site)
		if !errorFree(err, req, rsp, CLASSIDSite, ACTION) {
			return
		}
	}
	security.Audit(ctxt, CLASSIDSite, ACTION, security.SUCCESS)
	rest.WriteProtoSafely(rsp, site)
	log.Info("retrieved site")
}

func (r *SiteRest) findAllGeoJSON(req *restful.Request, rsp *restful.Response) {
	action := moc.Site_READ.String()
	ctx := req.Request.Context()
	if !authorized(req, rsp, CLASSIDSite, action) {
		return
	}
	sites, err := r.siteDb.FindAll(ctx)
	if !errorFree(err, req, rsp, CLASSIDSite, action) {
		return
	}
	geoJsons := make([]*moc.GeoJsonFeaturePoint, 0)
	for _, s := range sites {
		gj := &moc.GeoJsonFeaturePoint{
			Type: "Feature",
			Properties: map[string]string{
				"label":      s.Name,
				"type":       CLASSIDSite,
				"status":     s.ConnectionStatus.String(),
				"deviceId":   fmt.Sprint(s.SiteId),
				"deviceType": s.Type,
			},
			Geometry: &moc.GeoJsonGeometryPoint{
				Type:        "Point",
				Coordinates: []float64{s.Point.Longitude, s.Point.Latitude},
			},
		}
		geoJsons = append(geoJsons, gj)
	}
	featureCollection := moc.GeoJsonFeatureCollectionPoint{
		Type:     "FeatureCollection",
		Features: geoJsons,
	}
	security.Audit(ctx, CLASSIDSite, action, security.SUCCESS)
	rest.WriteProtoSafely(rsp, &featureCollection)
}

func (r *SiteRest) findAll(req *restful.Request, rsp *restful.Response) {
	action := moc.Site_READ.String()
	ctx := req.Request.Context()
	if !authorized(req, rsp, CLASSIDSite, action) {
		return
	}
	sites, err := r.siteDb.FindAll(ctx)
	if !errorFree(err, req, rsp, CLASSIDSite, action) {
		return
	}
	security.Audit(ctx, CLASSIDSite, action, security.SUCCESS)
	rest.WriteProtoSpliceSafely(rsp, toMessagesFromSites(sites))
}

func toMessagesFromSites(sites []*moc.Site) []proto.Message {
	var messages []proto.Message
	for _, site := range sites {
		messages = append(messages, site)
	}
	return messages
}

func (r *SiteRest) registerRoute(route string) string {
	if r.routes == nil {
		r.routes = make([]string, 0)
	}
	r.routes = append(r.routes, route)
	return route
}

func (r *SiteRest) MatchRoute(route string) (bool, string) {
	match := false
	for _, siteRoute := range r.routes {
		match = strings.HasSuffix(route, siteRoute)
		if match {
			break
		}
	}
	return match, CLASSIDSite
}
