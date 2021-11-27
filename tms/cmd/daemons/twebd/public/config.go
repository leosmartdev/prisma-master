package public

import (
	"context"
	"net/http"
	"reflect"
	"strings"

	"prisma/gogroup"
	"prisma/tms/db/mongo"
	"prisma/tms/libmain"
	"prisma/tms/log"
	"prisma/tms/moc"
	"prisma/tms/rest"
	"prisma/tms/security"
	"prisma/tms/security/policy"
	"prisma/tms/security/service"
	"prisma/tms/util/resolver"

	restful "github.com/orolia/go-restful"
	restfulspec "github.com/orolia/go-restful-openapi"
)

const (
	CLASSIDConfig = security.CLASSIDConfig
)

type ConfigRest struct {
	ctx    context.Context
	config *mongo.Configuration
	routes []string
}

func RegisterConfigRest(ctx gogroup.GoGroup, service *restful.WebService, routeAuthorizer *service.RouteAuthorizer) *ConfigRest {
	configDb := mongo.ConfigDb{}
	config, err := configDb.Read(ctx)
	if err != nil {
		config = defaultConfig(ctx)
	}
	// TODO: This line overrides any policy changes as a security measure
	//       The correct way to address this security concern is to not allow
	//		 any REST call to modify the polity/config
	//config.Policy = policy.GetStore(ctx).Get()
	r := ConfigRest{
		ctx:    ctx,
		config: config,
	}
	service.Route(service.PUT(r.registerRoute("/config")).To(r.update).
		Doc("update configuration").
		Metadata(restfulspec.KeyOpenAPITags, []string{CLASSIDSite}).
		Returns(http.StatusOK, http.StatusText(http.StatusOK), nil))
	routeAuthorizer.Add(&r)
	return &r
}

func (r *ConfigRest) update(req *restful.Request, rsp *restful.Response) {
	action := moc.Config_UPDATE.String()
	ctx := req.Request.Context()
	if !authorized(req, rsp, CLASSIDConfig, action) {
		return
	}
	config := mongo.Configuration{}
	err := req.ReadEntity(&config)
	var errs []rest.ErrorValidation
	if err != nil {
		errs = append(errs, rest.ErrorValidation{
			Property: reflect.TypeOf(&config).Elem().String(),
			Rule:     "Unmarshal",
			Message:  err.Error(),
		})
	}
	if !valid(errs, req, rsp, CLASSIDConfig, action) {
		return
	}
	configDb := mongo.ConfigDb{}
	err = configDb.Update(ctx, &config)
	if !errorFree(err, req, rsp, CLASSIDConfig, action) {
		return
	}
	r.config = &config
	security.Audit(ctx, CLASSIDConfig, action, security.SUCCESS)
	rest.WriteEntitySafely(rsp, config)
}

func (r *ConfigRest) Prefix() string {
	var prefix string
	if nil != r.config.Site {
		prefix = r.config.Site.IncidentIdPrefix
	}
	return prefix
}

func (r *ConfigRest) MetaFunc(objectId string) func(request *restful.Request, response *restful.Response) {
	meta := r.config.Meta
	return func(request *restful.Request, response *restful.Response) {
		rest.WriteEntitySafely(response, meta)
	}
}

// Get config.json is the public configuration needed by everyone.
func (r *ConfigRest) Get(request *restful.Request, response *restful.Response) {
	rest.WriteEntitySafely(response, r.config)
}

func (r *ConfigRest) registerRoute(route string) string {
	if r.routes == nil {
		r.routes = make([]string, 0)
	}
	r.routes = append(r.routes, route)
	return route
}

func (r *ConfigRest) MatchRoute(route string) (bool, string) {
	match := false
	for _, siteRoute := range r.routes {
		match = strings.HasSuffix(route, siteRoute)
		if match {
			break
		}
	}
	return match, CLASSIDSite
}

func defaultConfig(ctx context.Context) *mongo.Configuration {
	// meta
	metaMap := make(map[string]map[string]interface{})

	// meta vessel
	metaMap["vessel"] = make(map[string]interface{})
	metaMap["vessel"]["type"] = []string{"ship-fishing", "ship-passenger", "ship-cargo", "ship-tanker", "ship-pleasure", "ship-supply", "ship-utility", "ship-research", "ship-military"}
	metaMap["vessel"]["actions"] = moc.Vessel_Action_value
	// meta fleet
	metaMap["fleet"] = make(map[string]interface{})
	metaMap["fleet"]["actions"] = moc.Fleet_Action_value
	// meta device
	metaMap["device"] = make(map[string]interface{})
	metaMap["device"]["type"] = []string{"phone", "email", "ais", "omnicom-vms", "omnicom-solar", "epirb", "sart-radar", "sart-ais", "elt", "plb", "mob-ais"}
	metaMap["device"]["actions"] = moc.Device_Action_value
	metaMap["device"]["network"] = make(map[string]map[string]interface{})
	networkType := make(map[string]interface{})
	networkType["type"] = []string{"iridium"}
	metaMap["device"]["network"] = networkType
	// client
	clientConfig := rest.Client{
		Locale:           "en-US",
		Distance:         "nauticalMiles",
		ShortDistance:    "meters",
		Speed:            "knots",
		CoordinateFormat: "degreesMinutes",
		TimeZone:         "UTC",
	}
	// server
	host, err := resolver.ResolveHostIP()
	if err != nil {
		log.Error(err.Error())
	}
	serviceConfig := rest.Service{
		Ws: &rest.Service_WS{
			Map: "wss://" + host + ":8080/ws/v2/view/stream",
			Tms: "wss://" + host + ":8080/ws/v2/",
		},
		Tms: &rest.Service_TMS{
			Headers:       nil,
			Base:          "https://" + host + ":8080/api/v2",
			Device:        "https://" + host + ":8080/api/v2/device",
			Fleet:         "https://" + host + ":8080/api/v2/fleet",
			Track:         "https://" + host + ":8080/api/v2/track",
			Incident:      "https://" + host + ":8080/api/v2/incident",
			Communication: "https://" + host + ":8080/api/v2/communication",
			Notification:  "https://" + host + ":8080/api/v2/notification",
			Vessel:        "https://" + host + ":8080/api/v2/vessel",
			Registry:      "https://" + host + ":8080/api/v2/registry",
			Rule:          "https://" + host + ":8080/api/v2/rule",
			Map:           "https://" + host + ":8080/api/v2/view",
			Zone:          "https://" + host + ":8080/api/v2/zone",
			File:          "https://" + host + ":8080/api/v2/file",
			Pagination:    "https://" + host + ":8080",
			Swagger:       "https://" + host + ":8080/api/v2/apidocs.json",
			Activity:      "http://" + host + ":7077/activity",
			Request:       "http://" + host + ":7077/request",
		},
		Aaa: &rest.Service_AAA{
			Headers:    nil,
			Base:       "https://" + host + ":8181/api/v2",
			Session:    "https://" + host + ":8181/api/v2/auth/session",
			User:       "https://" + host + ":8181/api/v2/auth/user",
			Role:       "https://" + host + ":8181/api/v2/auth/role",
			Policy:     "https://" + host + ":8181/api/v2/auth/policy",
			Profile:    "https://" + host + ":8181/api/v2/auth/profile",
			Audit:      "https://" + host + ":8181/api/v2/auth/audit",
			Pagination: "https://" + host + ":8181",
			Swagger:    "https://" + host + ":8181/api/v2/auth/apidocs.json",
		},
		Sim: &rest.Service_SIM{
			Headers: nil,
			Alert:   "http://" + host + ":8089/v1/alert",
			Target:  "http://" + host + ":8089/v1/target",
			Route:   "http://" + host + ":8089/v1/route",
		},
	}
	site := moc.Site{
		SiteId:      1,
		Type:        "RCC",
		Name:        "Site Name",
		Description: "Site description",
		Address:     "123",
		Country:     "USA",
		Point:       nil,
		ParentId:    "",
		Devices:     nil,
		Cscode:      "RCC1",
		Csname:      "MYRCC",
		Capability: &moc.Site_Capability{
			InputIncident:  true,
			OutputIncident: true,
		},
	}
	c := mongo.Configuration{
		Meta:    metaMap,
		Site:    &site,
		Service: &serviceConfig,
		Client:  &clientConfig,
		Brand: mongo.Brand{
			Name:        "PRISMA",
			Version:     libmain.VersionNumber,
			ReleaseDate: libmain.VersionDate,
		},
		Policy: policy.GetStore(ctx).Get(),
	}
	configDb := mongo.ConfigDb{}
	err = configDb.Create(ctx, &c)
	if err != nil {
		log.Error(err.Error(), err)
	}
	return &c
}
