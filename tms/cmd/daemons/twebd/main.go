// twebd discovers RestAPI to manage tms's resources.
package main

import (
	"expvar"
	"flag"
	glog "log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"prisma/gogroup"
	"prisma/tms/cmd/daemons/twebd/public"
	tmsdb "prisma/tms/db"
	"prisma/tms/db/connect"
	"prisma/tms/db/mongo"
	"prisma/tms/envelope"
	"prisma/tms/libmain"
	"prisma/tms/log"
	"prisma/tms/marker"
	"prisma/tms/moc"
	"prisma/tms/rest"
	"prisma/tms/routing"
	"prisma/tms/search/providers"
	"prisma/tms/security/certificate"
	"prisma/tms/security/service"
	"prisma/tms/security/tlsconfig"
	"prisma/tms/tmsg"
	"prisma/tms/ws"

	restful "github.com/orolia/go-restful"
	restfulspec "github.com/orolia/go-restful-openapi"

	"github.com/fsnotify/fsnotify"
	"github.com/globalsign/mgo"
	"github.com/go-openapi/spec"
)

var (
	listenAddress       = ""
	listenAddressConfig = ""
	certFile            = ""
	keyFile             = ""
	pathRest            = "/api/v2"
	pathWs              = "/ws/v2"
	goroutines          = expvar.NewInt("goroutines")
)

func init() {
	flag.StringVar(&listenAddress, "listen", ":8080", "Address:port to listen upon")
	flag.StringVar(&listenAddressConfig, "listen-config", ":8081", "Address: port for configuration")
	flag.StringVar(&certFile, "certificate", "/etc/trident/certificate.pem", "The path to the certificate file")
	flag.StringVar(&keyFile, "key", "/etc/trident/key.pem", "The path to the key file")
}

func main() {
	// MongoDb debug
	if envEnabled("MGO_DEBUG") {
		logger := glog.New(os.Stdout, "[mgo] ", glog.LUTC|glog.Lshortfile)
		mgo.SetLogger(logger)
		mgo.SetDebug(true)
		logger.Println("MGO_DEBUG activated")
	}
	// go-restful debug
	if envEnabled("REST_DEBUG") {
		logger := glog.New(os.Stdout, "[rest] ", glog.LUTC|glog.Lshortfile)
		restful.SetLogger(logger)
		restful.EnableTracing(true)
		logger.Println("REST_DEBUG activated")
	}
	// setup goroutines
	gogroup.Callback = func(state gogroup.GoState) {
		if state == gogroup.GoStarted {
			goroutines.Add(1)
		} else if state == gogroup.GoFinished {
			goroutines.Add(-1)
		}
	}
	// setup tmsg infrastructure
	libmain.Main(tmsg.APP_ID_TWEBD, func(ctxt gogroup.GoGroup) {
		var trackDb tmsdb.TrackDB
		var siteDb tmsdb.SiteDB
		var miscDb tmsdb.MiscDB
		var deviceDb tmsdb.DeviceDB
		var mongoClient *mongo.MongoClient
		var err error

		log.Notice("Connecting to MongoDB")
		trackDb, miscDb, siteDb, _, deviceDb, mongoClient, err = connect.DBConnect(ctxt, tmsg.GClient)
		if err != nil {
			log.Error("Unable to connect to MongoDB: %v", err)
			time.Sleep(2 * time.Second)
		}
		ctxt = gogroup.WithValue(ctxt, "mongodb", mongoClient.DialInfo)
		ctxt = gogroup.WithValue(ctxt, "mongodb-cred", mongoClient.Cred)
		// set dial info in filter for each request
		service.SetDialInfo(mongoClient.DialInfo)
		// set creds info in the filter for each request
		service.SetCredential(mongoClient.Cred)
		if miscDb == nil {
			log.Warn("No misc data interface specified... You probably want one of these")
		}

		// Create a featuresdb
		featDb := tmsdb.NewFeatures(ctxt, trackDb, miscDb, siteDb, deviceDb)
		// Setup
		handler := NewHandler()
		// Setup v2 restful
		container := restful.NewContainer()
		webService := new(restful.WebService)
		webService.Path(pathRest).
			Consumes(restful.MIME_JSON).
			Produces(restful.MIME_JSON)
		// Publisher-Subscriber
		publisher := ws.NewPublisher()
		envelopeStream := tmsg.GClient.Listen(ctxt, routing.Listener{
			MessageType: "prisma.tms.envelope.Envelope",
		})
		ctxt.GoRestart(func() {
			for {
				select {
				case <-ctxt.Done():
					// Do nothing
				case msg := <-envelopeStream:
					envelope, ok := msg.Body.(*envelope.Envelope)
					if ok {
						publisher.Publish("Session", *envelope)
					}
				}
			}
		})
		// Filters
		webService.Filter(restful.NoBrowserCacheFilter)
		webService.Filter(service.RequestIdContextFilter)
		webService.Filter(service.SessionIdContextFilter)
		routeAuthorizer := service.RouteAuthorizer{}
		webService.Filter(routeAuthorizer.Authorize)
		// Config service used for meta routes
		configRest := public.RegisterConfigRest(ctxt, webService, &routeAuthorizer)
		// Site service
		public.RegisterSiteRest(ctxt, webService, &routeAuthorizer)
		// View service
		viewRest := public.NewViewRest(featDb, libmain.TsiKillGroup)
		webService.Route(webService.POST("/view").To(viewRest.Create).
			Doc("create a map viewport").
			Metadata(restfulspec.KeyOpenAPITags, []string{"view"}).
			Returns(http.StatusOK, "OK", public.ViewResponse{}))
		// Track service
		trackRest := public.NewTrackRest(mongoClient, libmain.TsiKillGroup)
		webService.Route(webService.GET("/track/{track-id}").To(trackRest.GetOne).
			Doc("get information about a specific track").
			Metadata(restfulspec.KeyOpenAPITags, []string{"track"}).
			Param(webService.PathParameter("track-id", "track identifier")).
			Returns(http.StatusOK, "OK", public.InfoResponse{}))
		webService.Route(webService.GET("/track").To(trackRest.Get).
			Doc("list all tracks").
			Metadata(restfulspec.KeyOpenAPITags, []string{"track"}).
			Returns(http.StatusOK, "OK", []public.InfoResponse{}))
		webService.Route(webService.POST("/track").To(trackRest.Post).
			Doc("create or update a manual track").
			ReadsWithSchema(public.ManualTrackPublic{}, public.SCHEMA_MANUAL_TRACK).
			Writes(public.ManualTrackPublic{}).
			Metadata(restfulspec.KeyOpenAPITags, []string{"track"}).
			Returns(http.StatusOK, "OK", public.ManualTrackPublic{}))
		webService.Route(webService.DELETE("/track/{registry-id}").To(trackRest.Delete).
			Doc("remove a manual track").
			Param(webService.PathParameter("registry-id", "registry identifier")).
			Metadata(restfulspec.KeyOpenAPITags, []string{"track"}).
			Returns(http.StatusOK, "OK", public.ManualTrackPublic{}))
		// Track history service
		historyRest := public.NewHistoryRest(mongoClient, libmain.TsiKillGroup)
		webService.Route(webService.GET("/history/{registry-id}").To(historyRest.Get).
			Doc("get information about a specific track").
			Metadata(restfulspec.KeyOpenAPITags, []string{"track"}).
			Param(webService.PathParameter("registry-id", "registry identifier for a track")).
			Returns(http.StatusOK, "OK", public.InfoResponse{}))
		webService.Route(webService.GET("/history-database/{database-id}").To(historyRest.GetDatabase).
			Doc("get information about a specific track").
			Metadata(restfulspec.KeyOpenAPITags, []string{"track"}).
			Param(webService.PathParameter("database-id", "database identifier")).
			Returns(http.StatusOK, "OK", public.InfoResponse{}))
		// Registry service
		registryRest := public.NewRegistryRest(mongoClient, libmain.TsiKillGroup)
		webService.Route(webService.GET("/registry/{registry-id}").To(registryRest.Get).
			Doc("get information about a specific track").
			Metadata(restfulspec.KeyOpenAPITags, []string{"registry"}).
			Param(webService.PathParameter("registry-id", "registry identifier")).
			Returns(http.StatusOK, "OK", public.InfoResponse{}))
		webService.Route(webService.GET("/search/registry").To(registryRest.Search).
			Doc("search the registry based on certain fields").
			Metadata(restfulspec.KeyOpenAPITags, []string{"search"}).
			Param(webService.QueryParameter("query", "string to search for")).
			Param(webService.QueryParameter("limit", "maximum number of results to return")).
			Returns(http.StatusOK, "OK", []tmsdb.RegistrySearchResult{}))
		// Search engine
		searchRest := public.NewSearchRest(libmain.TsiKillGroup, providers.NewMongoSearchProvider(mongoClient))
		webService.Route(webService.GET("/search/tables/{tables}").To(searchRest.SearchFunc("")).
			Doc("search data in a database").
			Metadata(restfulspec.KeyOpenAPITags, []string{"search"}).
			Param(webService.PathParameter("tables", "separated names for tables. For separating use ,")).
			Param(webService.PathParameter("text", "text for searching")).
			Returns(http.StatusOK, "OK", public.InfoResponse{}))
		webService.Route(webService.POST("/search/tables/{tables}").To(searchRest.SearchFunc("")).
			Doc("search data in a database").
			Metadata(restfulspec.KeyOpenAPITags, []string{"search"}).
			Param(webService.PathParameter("tables", "separated names for tables. For separating use ,")).
			Param(webService.PathParameter("text", "text for searching")).
			Returns(http.StatusOK, "OK", public.InfoResponse{}))
		webService.Route(webService.GET("/search/vessel").To(searchRest.SearchFunc("vessels")).
			Doc("search vessel in a database").
			Metadata(restfulspec.KeyOpenAPITags, []string{"search"}).
			Param(webService.QueryParameter("query", "text for searching")).
			Param(webService.QueryParameter("limit", "result count maximum")).
			Returns(http.StatusOK, "OK", public.InfoResponse{}))
		webService.Route(webService.GET("/search/fleet").To(searchRest.SearchFunc("fleets")).
			Doc("search fleet in a database").
			Metadata(restfulspec.KeyOpenAPITags, []string{"search"}).
			Param(webService.QueryParameter("query", "text for searching")).
			Param(webService.QueryParameter("limit", "result count maximum")).
			Returns(http.StatusOK, "OK", public.InfoResponse{}))
		// Rules service
		ruleRest, err := public.NewRuleRest(mongoClient, libmain.TsiKillGroup)
		if err != nil {
			log.Fatal("rule.NewTmsEngine(mc): %v", err)
		}
		webService.Route(webService.GET("/rule").To(ruleRest.Get).
			Doc("get all rules").
			Metadata(restfulspec.KeyOpenAPITags, []string{"rule"}))
		webService.Route(webService.GET("/rule/meta").To(ruleRest.Meta).
			Doc("get metadata of rules").
			Metadata(restfulspec.KeyOpenAPITags, []string{"rule"}))
		webService.Route(webService.GET("/rule/{rule-id}").To(ruleRest.GetOne).
			Doc("get one rule").
			Metadata(restfulspec.KeyOpenAPITags, []string{"rule"}))
		webService.Route(webService.POST("/rule").To(ruleRest.Post).
			Doc("create a rule").
			Metadata(restfulspec.KeyOpenAPITags, []string{"rule"}))
		webService.Route(webService.DELETE("/rule/{rule-id}").To(ruleRest.Delete).
			Doc("delete a rule").
			Metadata(restfulspec.KeyOpenAPITags, []string{"rule"}))
		webService.Route(webService.PUT("/rule/{rule-id}").To(ruleRest.Put).
			Doc("update a rule").
			Metadata(restfulspec.KeyOpenAPITags, []string{"rule"}))
		webService.Route(webService.PUT("/rule/{rule-id}/state/{state-id}").To(ruleRest.UpdateState).
			Doc("update state of rule").
			Metadata(restfulspec.KeyOpenAPITags, []string{"rule"}))
		// Multicast service
		public.RegisterMulticastRest(ctxt, mongoClient, webService, &routeAuthorizer, publisher, tmsg.GClient)
		// Device service
		deviceRest := public.NewDeviceRest(libmain.TsiKillGroup, mongoClient)
		webService.Route(webService.GET("/device").To(deviceRest.GetAllDevices).
			Doc("get all devices \n link header provided \n search parameters supported").
			Param(webService.PathParameter("fleet-id", "fleet identifier")).
			Writes([]moc.Device{}).
			Returns(http.StatusOK, http.StatusText(http.StatusOK), []moc.Device{}).
			Returns(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), []rest.ErrorValidation{}).
			Returns(http.StatusForbidden, http.StatusText(http.StatusForbidden), nil).
			Returns(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), nil).
			Metadata(restfulspec.KeyOpenAPITags, []string{"device"}))
		webService.Route(webService.GET("/device/meta").To(configRest.MetaFunc("device")).
			Doc("information about device").
			Metadata(restfulspec.KeyOpenAPITags, []string{"device"}))
		webService.Route(webService.GET("/device/{id}").To(deviceRest.GetDeviceById).
			Doc("get a device").
			Param(webService.PathParameter("device-id", "device identifier")).
			Writes(moc.Device{}).
			Returns(http.StatusOK, http.StatusText(http.StatusOK), moc.Device{}).
			Returns(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), []rest.ErrorValidation{}).
			Returns(http.StatusForbidden, http.StatusText(http.StatusForbidden), nil).
			Returns(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), nil).
			Metadata(restfulspec.KeyOpenAPITags, []string{"device"}))
		webService.Route(webService.GET("/device/vessel/{vessel-id}").To(deviceRest.GetDeviceByVesselId).
			Doc("get devices by vessel id").
			Param(webService.PathParameter("vessel-id", "vessel identifier")).
			Writes([]moc.Device{}).
			Returns(http.StatusOK, http.StatusText(http.StatusOK), []moc.Device{}).
			Returns(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), []rest.ErrorValidation{}).
			Returns(http.StatusForbidden, http.StatusText(http.StatusForbidden), nil).
			Returns(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), nil).
			Metadata(restfulspec.KeyOpenAPITags, []string{"device"}))
		webService.Route(webService.POST("/device/vessel/{vessel-id}").To(deviceRest.CreateWithVessel).
			Doc("create a device").
			Writes(moc.Device{}).
			Returns(http.StatusOK, http.StatusText(http.StatusOK), moc.Device{}).
			Returns(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), []rest.ErrorValidation{}).
			Returns(http.StatusForbidden, http.StatusText(http.StatusForbidden), nil).
			Returns(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), nil).
			Metadata(restfulspec.KeyOpenAPITags, []string{"device"}))
		webService.Route(webService.POST("/device").To(deviceRest.Create).
			Doc("create a device").
			Writes(moc.Device{}).
			Returns(http.StatusOK, http.StatusText(http.StatusOK), moc.Device{}).
			Returns(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), []rest.ErrorValidation{}).
			Returns(http.StatusForbidden, http.StatusText(http.StatusForbidden), nil).
			Returns(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), nil).
			Metadata(restfulspec.KeyOpenAPITags, []string{"device"}))
		webService.Route(webService.PUT("/device/{id}").To(deviceRest.Update).
			Doc("update a device").
			Param(webService.PathParameter("device-id", "device identifier")).
			//ReadsWithSchema(public.SchemaFleetUpdate, moc.Fleet{}).
			Writes(moc.Device{}).
			Returns(http.StatusOK, http.StatusText(http.StatusOK), moc.Device{}).
			Returns(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), []rest.ErrorValidation{}).
			Returns(http.StatusForbidden, http.StatusText(http.StatusForbidden), nil).
			Returns(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), nil).
			Metadata(restfulspec.KeyOpenAPITags, []string{"device"}))
		webService.Route(webService.DELETE("/device/{id}").To(deviceRest.Delete).
			Doc("delete a device").
			Param(webService.PathParameter("device-id", "device identifier")).
			Returns(http.StatusAccepted, http.StatusText(http.StatusAccepted), nil).
			Returns(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), []rest.ErrorValidation{}).
			Returns(http.StatusForbidden, http.StatusText(http.StatusForbidden), nil).
			Returns(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), nil).
			Metadata(restfulspec.KeyOpenAPITags, []string{"device"}))
		// Fleet service
		fleetRest := public.NewFleetRest(libmain.TsiKillGroup, publisher)
		webService.Route(webService.GET("/fleet").To(fleetRest.ReadAll).
			Doc("get all fleets (w/o vessels)").
			Writes([]moc.Fleet{}).
			Returns(http.StatusOK, http.StatusText(http.StatusOK), []moc.Fleet{}).
			Returns(http.StatusForbidden, http.StatusText(http.StatusForbidden), nil).
			Returns(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), nil).
			Metadata(restfulspec.KeyOpenAPITags, []string{"fleet"}))
		webService.Route(webService.GET("/fleet/vessel").To(fleetRest.ReadAll).
			Doc("get all fleets (w/ vessels)").
			Writes([]moc.Fleet{}).
			Returns(http.StatusOK, http.StatusText(http.StatusOK), []moc.Fleet{}).
			Returns(http.StatusForbidden, http.StatusText(http.StatusForbidden), nil).
			Returns(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), nil).
			Metadata(restfulspec.KeyOpenAPITags, []string{"fleet"}))
		webService.Route(webService.GET("/fleet/meta").To(configRest.MetaFunc("fleet")).
			Doc("information about fleet").
			Metadata(restfulspec.KeyOpenAPITags, []string{"fleet"}))
		webService.Route(webService.GET("/fleet/{fleet-id}").To(fleetRest.ReadOne).
			Doc("get a fleet").
			Param(webService.PathParameter("fleet-id", "fleet identifier")).
			Writes(moc.Fleet{}).
			Returns(http.StatusOK, http.StatusText(http.StatusOK), moc.Fleet{}).
			Returns(http.StatusForbidden, http.StatusText(http.StatusForbidden), nil).
			Returns(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), nil).
			Metadata(restfulspec.KeyOpenAPITags, []string{"fleet"}))
		webService.Route(webService.POST("/fleet").To(fleetRest.Create).
			Doc("create a fleet").
			//ReadsWithSchema(public.SchemaFleetCreate, moc.Fleet{}).
			Writes(moc.Fleet{}).
			Returns(http.StatusCreated, http.StatusText(http.StatusCreated), moc.Fleet{}).
			Returns(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), []rest.ErrorValidation{}).
			Returns(http.StatusForbidden, http.StatusText(http.StatusForbidden), nil).
			Returns(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), nil).
			Metadata(restfulspec.KeyOpenAPITags, []string{"fleet"}))
		webService.Route(webService.PUT("/fleet/{fleet-id}").To(fleetRest.Update).
			Doc("update a fleet").
			Param(webService.PathParameter("fleet-id", "fleet identifier")).
			//ReadsWithSchema(public.SchemaFleetUpdate, moc.Fleet{}).
			Writes(moc.Fleet{}).
			Returns(http.StatusOK, http.StatusText(http.StatusOK), moc.Fleet{}).
			Returns(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), []rest.ErrorValidation{}).
			Returns(http.StatusForbidden, http.StatusText(http.StatusForbidden), nil).
			Returns(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), nil).
			Metadata(restfulspec.KeyOpenAPITags, []string{"fleet"}))
		webService.Route(webService.PUT("/fleet/{fleet-id}/vessel/{vessel-id}").To(fleetRest.UpdateAddVessel).
			Doc("add a vessel to a fleet").
			Param(webService.PathParameter("fleet-id", "fleet identifier")).
			Param(webService.PathParameter("vessel-id", "vessel identifier")).
			Writes(moc.Fleet{}).
			Returns(http.StatusOK, http.StatusText(http.StatusOK), moc.Fleet{}).
			Returns(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), []rest.ErrorValidation{}).
			Returns(http.StatusForbidden, http.StatusText(http.StatusForbidden), nil).
			Returns(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), nil).
			Metadata(restfulspec.KeyOpenAPITags, []string{"fleet"}))
		webService.Route(webService.DELETE("/fleet/{fleet-id}/vessel/{vessel-id}").To(fleetRest.UpdateRemoveVessel).
			Doc("remove a vessel from a fleet").
			Param(webService.PathParameter("fleet-id", "fleet identifier")).
			Param(webService.PathParameter("vessel-id", "vessel identifier")).
			Writes(moc.Fleet{}).
			Returns(http.StatusOK, http.StatusText(http.StatusOK), moc.Fleet{}).
			Returns(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), []rest.ErrorValidation{}).
			Returns(http.StatusForbidden, http.StatusText(http.StatusForbidden), nil).
			Returns(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), nil).
			Metadata(restfulspec.KeyOpenAPITags, []string{"fleet"}))
		webService.Route(webService.DELETE("/fleet/{fleet-id}").To(fleetRest.Delete).
			Doc("delete a fleet").
			Param(webService.PathParameter("fleet-id", "fleet identifier")).
			Returns(http.StatusAccepted, http.StatusText(http.StatusAccepted), nil).
			Returns(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), []rest.ErrorValidation{}).
			Returns(http.StatusForbidden, http.StatusText(http.StatusForbidden), nil).
			Returns(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), nil).
			Metadata(restfulspec.KeyOpenAPITags, []string{"fleet"}))
		// Vessel service
		vesselRest := public.NewVesselRest(libmain.TsiKillGroup, mongoClient, publisher)
		webService.Route(webService.GET("/vessel").To(vesselRest.ReadAll).
			Doc("get all registered vessels \n has-fleet=false to see vessels without a fleet").
			Writes([]moc.Vessel{}).
			Metadata(restfulspec.KeyOpenAPITags, []string{"fleet"}))
		webService.Route(webService.GET("/vessel/meta").To(configRest.MetaFunc("vessel")).
			Doc("information about vessel").
			Metadata(restfulspec.KeyOpenAPITags, []string{"fleet"}))
		webService.Route(webService.GET("/vessel/{vessel-id}").To(vesselRest.FindOne).
			Doc("get a vessel by vessel id").
			Param(webService.PathParameter("vessel-id", "vessel identifier")).
			Writes(moc.Vessel{}).
			Returns(http.StatusOK, http.StatusText(http.StatusOK), moc.Vessel{}).
			Returns(http.StatusForbidden, http.StatusText(http.StatusForbidden), nil).
			Returns(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), nil).
			Metadata(restfulspec.KeyOpenAPITags, []string{"vessel"}))
		webService.Route(webService.GET("/vessel/fleet").To(vesselRest.ReadAll).
			Doc("get all registered vessels with fleet \n unassigned search").
			Param(webService.QueryParameter("limit", "limit of vessels returned for each fleet")).
			Param(webService.QueryParameter("fleet", "fleet=NONE for all vessels not part of a fleet")).
			Writes([]moc.Vessel{}).
			Metadata(restfulspec.KeyOpenAPITags, []string{"fleet"}))
		webService.Route(webService.POST("/vessel").To(vesselRest.Create).
			Doc("create a registered vessel").
			//ReadsWithSchema(public.SchemaVesselCreate, moc.Vessel{}).
			Writes(moc.Vessel{}).
			Metadata(restfulspec.KeyOpenAPITags, []string{"fleet"}))
		webService.Route(webService.PUT("/vessel/{vessel-id}").To(vesselRest.Update).
			//ReadsWithSchema(public.SchemaVesselUpdate, moc.Vessel{}).
			Doc("update a registered vessel").
			Param(webService.PathParameter("vessel-id", "registered vessel identifier")).
			Metadata(restfulspec.KeyOpenAPITags, []string{"fleet"}))
		webService.Route(webService.DELETE("/vessel/{vessel-id}").To(vesselRest.Delete).
			Doc("delete a registered vessel").
			Param(webService.PathParameter("vessel-id", "registered vessel identifier")).
			Metadata(restfulspec.KeyOpenAPITags, []string{"fleet"}))
		// Zone service
		zoneRest := public.NewZoneRest(miscDb, libmain.TsiKillGroup)
		webService.Route(webService.GET("/zone").To(zoneRest.Get).
			Metadata(restfulspec.KeyOpenAPITags, []string{"zone"}))
		webService.Route(webService.GET("/zone/{zone-id}").To(zoneRest.GetOne).
			Metadata(restfulspec.KeyOpenAPITags, []string{"zone"}))
		webService.Route(webService.POST("/zone").To(zoneRest.Post).
			Doc("insert a zone").
			Param(webService.BodyParameter("area.radius",
				"For a proximity zone - the radius of it has a range 0 to 27999999 degrees").
				DataType("float")).
			Metadata(restfulspec.KeyOpenAPITags, []string{"zone"}))
		webService.Route(webService.DELETE("/zone/{zone-id}").To(zoneRest.Delete).
			Metadata(restfulspec.KeyOpenAPITags, []string{"zone"}))
		webService.Route(webService.GET("/zone/geo").To(zoneRest.FindAllGeoJSON).
			Doc("geojson for zones").
			Metadata(restfulspec.KeyOpenAPITags, []string{"zone"}).
			Returns(http.StatusOK, http.StatusText(http.StatusOK), []*moc.GeoJsonMixedCollection{}))
		// Geofence service
		geoFenceRest := public.NewGeoFenceRest(miscDb, libmain.TsiKillGroup)
		webService.Route(webService.GET("/geofence").To(geoFenceRest.Get).
			Metadata(restfulspec.KeyOpenAPITags, []string{"geofence"}))
		webService.Route(webService.GET("/geofence/{geofence-id}").To(geoFenceRest.GetOne).
			Metadata(restfulspec.KeyOpenAPITags, []string{"geofence"}))
		webService.Route(webService.POST("/geofence").To(geoFenceRest.Post).
			Doc("insert a geofence").
			Metadata(restfulspec.KeyOpenAPITags, []string{"geofence"}))
		webService.Route(webService.PUT("/geofence/{geofence-id}").To(geoFenceRest.Put).
			Doc("update a geofence").
			Metadata(restfulspec.KeyOpenAPITags, []string{"geofence"}))
		webService.Route(webService.DELETE("/geofence/{geofence-id}").To(geoFenceRest.Delete).
			Metadata(restfulspec.KeyOpenAPITags, []string{"geofence"}))
		// Notification service
		noticeRest := public.NewNoticeRest(mongoClient, libmain.TsiKillGroup)
		webService.Route(webService.GET("/notice/history").To(noticeRest.GetHistory).
			Doc("get all notices that are acknowledged").
			Metadata(restfulspec.KeyOpenAPITags, []string{"notice"}).
			Param(webService.PathParameter("limit", "return this number of documents")).
			Param(webService.PathParameter("before", "documents before this one")).
			Param(webService.PathParameter("after", "documents after this one")).
			Writes(moc.Notice{}).
			Returns(http.StatusOK, "OK", []moc.Notice{}))
		webService.Route(webService.GET("/notice/new").To(noticeRest.GetAllNewNotices).
			Doc("get all active notices that are acknowledged").
			Metadata(restfulspec.KeyOpenAPITags, []string{"notice"}).
			Param(webService.PathParameter("limit", "return this number of documents")).
			Param(webService.PathParameter("before", "documents before this one")).
			Param(webService.PathParameter("after", "documents after this one")).
			Writes(moc.Notice{}).
			Returns(http.StatusOK, "OK", []moc.Notice{}))
		webService.Route(webService.POST("/notice/{id}/ack").To(noticeRest.Ack).
			Doc("acknowledge a notice").
			Metadata(restfulspec.KeyOpenAPITags, []string{"notice"}).
			Writes(public.AckResponsePublic{}).
			Param(webService.PathParameter("id", "database identifier of the notice to acknowledge")).
			Returns(http.StatusOK, "OK", public.AckResponsePublic{}))
		webService.Route(webService.POST("/notice/all/ack").To(noticeRest.AckAll).
			Doc("acknowledge all notices").
			Metadata(restfulspec.KeyOpenAPITags, []string{"notice"}).
			Writes(public.AckAllResponsePublic{}).
			Returns(http.StatusOK, "OK", public.AckAllResponsePublic{}))
		webService.Route(webService.POST("/notice/timeout").To(noticeRest.Timeout).
			Doc("clear all acknowledged notices older than the specified date").
			Metadata(restfulspec.KeyOpenAPITags, []string{"notice"}).
			Writes(public.TimeoutResponsePublic{}).
			Param(webService.BodyParameter("olderThan", "date in RFC3339 format (2006-01-02T15:04:05Z07:00)")).
			Returns(http.StatusOK, "OK", public.TimeoutResponsePublic{}))
		// Incident service
		incidentRest := public.NewIncidentRest(ctxt, mongoClient, configRest, publisher)
		webService.Route(webService.GET("/incident").To(incidentRest.ReadAll).
			Doc("get all incidents").
			Metadata(restfulspec.KeyOpenAPITags, []string{"incident"}).
			Writes(moc.Incident{}).
			Returns(http.StatusOK, "OK", []moc.Incident{}))
		webService.Route(webService.GET("/incident/{incident-id}").To(incidentRest.ReadOne).
			Doc("get one incident").
			Metadata(restfulspec.KeyOpenAPITags, []string{"incident"}).
			Param(webService.PathParameter("incident-id", "incident identifier")).
			Writes(moc.Incident{}).
			Returns(http.StatusOK, "OK", moc.Incident{}))
		webService.Route(webService.POST("/incident").To(incidentRest.Create).
			Doc("create incident").
			Metadata(restfulspec.KeyOpenAPITags, []string{"incident"}).
			ReadsWithSchema(moc.Incident{}, public.SCHEMA_INCIDENT_CREATE).
			Writes(moc.Incident{}).
			Returns(http.StatusCreated, "OK", moc.Incident{}))
		webService.Route(webService.PUT("/incident/{incident-id}").To(incidentRest.Update).
			Doc("update incident").
			Metadata(restfulspec.KeyOpenAPITags, []string{"incident"}).
			Param(webService.PathParameter("incident-id", "incident identifier")).
			ReadsWithSchema(moc.Incident{}, public.SCHEMA_INCIDENT_UPDATE).
			Writes(moc.Incident{}).
			Returns(http.StatusCreated, "OK", moc.Incident{}))
		webService.Route(webService.DELETE("/incident/{incident-id}").To(incidentRest.Delete).
			Doc("archive incident").
			Metadata(restfulspec.KeyOpenAPITags, []string{"incident"}).
			Param(webService.PathParameter("incident-id", "incident identifier")).
			Returns(http.StatusOK, "OK", nil))
		webService.Route(webService.POST("/incident/{incident-id}/log").To(incidentRest.CreateLogEntry).
			Doc("create an incident log entry").
			Metadata(restfulspec.KeyOpenAPITags, []string{"incident"}).
			Writes(moc.Incident{}).
			Param(webService.PathParameter("incident-id", "")).
			ReadsWithSchema(moc.Incident{}, public.SCHEMA_INCIDENT_LOGENTRY).
			Returns(http.StatusCreated, "OK", moc.Incident{}))
		webService.Route(webService.PUT("/incident/{incident-id}/log/{log-id}").To(incidentRest.UpdateLogEntry).
			Doc("update an incident log entry").
			Metadata(restfulspec.KeyOpenAPITags, []string{"incident"}).
			Writes(moc.Incident{}).
			Param(webService.PathParameter("incident-id", "")).
			ReadsWithSchema(moc.Incident{}, public.SCHEMA_INCIDENT_LOGENTRY).
			Returns(http.StatusCreated, "OK", moc.Incident{}))
		webService.Route(webService.PUT("/incident/{incident-id}/log-detach/{log-id}").To(incidentRest.DetachLogEntry).
			Doc("detach an incident log entry").
			Metadata(restfulspec.KeyOpenAPITags, []string{"incident"}).
			Writes(moc.IncidentLogEntry{}).
			Param(webService.PathParameter("incident-id", "incident identifier")).
			Param(webService.PathParameter("log-id", "log identifier")).
			Returns(http.StatusCreated, "OK", moc.IncidentLogEntry{}))
		webService.Route(webService.DELETE("/incident/{incident-id}/log/{log-id}").To(incidentRest.DeleteLogEntry).
			Doc("delete an incident log entry").
			Metadata(restfulspec.KeyOpenAPITags, []string{"incident"}).
			Writes(moc.Incident{}).
			Param(webService.PathParameter("incident-id", "")).
			Param(webService.PathParameter("log-id", "")).
			Returns(http.StatusCreated, "OK", moc.Incident{}))
		webService.Route(webService.PUT("/incident/{incident-id}/state/{incident-state}").To(incidentRest.UpdateState).
			Doc("create incident").
			Metadata(restfulspec.KeyOpenAPITags, []string{"incident"}).
			Writes(moc.Incident{}).
			Param(webService.PathParameter("incident-id", "")).
			Param(webService.PathParameter("incident-state", "")).
			ReadsWithSchema(moc.Incident{}, public.SCHEMA_INCIDENT_UPDATE).
			Returns(http.StatusCreated, "OK", moc.Incident{}))
		webService.Route(webService.PUT("/incident/{incident-id}/assignee/{user-id}").To(incidentRest.Assign).
			Doc("assign incident").
			Metadata(restfulspec.KeyOpenAPITags, []string{"incident"}).
			Writes(moc.Incident{}).
			Param(webService.PathParameter("incident-id", "")).
			Param(webService.PathParameter("user-id", "")).
			Returns(http.StatusCreated, "OK", moc.Incident{}))
		webService.Route(webService.GET("/incident/{incident-id}/processing").To(incidentRest.CreateIncidentProcessingForm).
			Doc("create an Incident Processing Form").
			Metadata(restfulspec.KeyOpenAPITags, []string{"incident"}).
			Writes(moc.Incident{}).
			Returns(http.StatusOK, "OK", []moc.Incident{}))
		// Note service
		noteRest := public.NewNoteRest(ctxt, mongoClient, configRest, publisher)
		webService.Route(webService.POST("/note").To(noteRest.Create).
			Doc("create unassigned note").
			Metadata(restfulspec.KeyOpenAPITags, []string{"note"}).
			ReadsWithSchema(moc.IncidentLogEntry{}, public.SCHEMA_NOTE).
			Writes(moc.IncidentLogEntry{}).
			Returns(http.StatusCreated, "OK", moc.IncidentLogEntry{}))
		webService.Route(webService.GET("/note").To(noteRest.ReadAll).
			Doc("get all notes").
			Metadata(restfulspec.KeyOpenAPITags, []string{"note"}).
			Writes(moc.IncidentLogEntry{}).
			Returns(http.StatusOK, "OK", []moc.IncidentLogEntry{}))
		webService.Route(webService.GET("/note/{note-id}/{is-assigned}").To(noteRest.ReadOne).
			Doc("get one note").
			Metadata(restfulspec.KeyOpenAPITags, []string{"note"}).
			Param(webService.PathParameter("note-id", "note identifier")).
			Param(webService.PathParameter("is-assigned", "to identify if a note is assigned/unassigned")).
			Writes(moc.IncidentLogEntryResponse{}).
			Returns(http.StatusOK, "OK", moc.IncidentLogEntryResponse{}))
		webService.Route(webService.PUT("/note/{note-id}").To(noteRest.Update).
			Doc("update a note").
			Metadata(restfulspec.KeyOpenAPITags, []string{"note"}).
			Writes(moc.IncidentLogEntry{}).
			Param(webService.PathParameter("note-id", "note identifier")).
			ReadsWithSchema(moc.IncidentLogEntry{}, public.SCHEMA_NOTE).
			Returns(http.StatusCreated, "OK", moc.IncidentLogEntry{}))
		webService.Route(webService.PUT("/note/{note-id}/assignee/{incident-id}").To(noteRest.Assign).
			Doc("assign a note to an incident").
			Metadata(restfulspec.KeyOpenAPITags, []string{"note"}).
			Param(webService.PathParameter("note-id", "note identifier")).
			Param(webService.PathParameter("incident-id", "incident identifier")).
			Writes(moc.Incident{}).
			Returns(http.StatusCreated, "OK", moc.Incident{}))
		webService.Route(webService.DELETE("/note/{note-id}").To(noteRest.Delete).
			Doc("delete a note").
			Metadata(restfulspec.KeyOpenAPITags, []string{"note"}).
			Param(webService.PathParameter("note-id", "note identifier")).
			Writes(moc.IncidentLogEntry{}).
			Returns(http.StatusCreated, "OK", moc.IncidentLogEntry{}))
		// Marker service
		markerRest := public.NewMarkerRest(ctxt, mongoClient)
		webService.Route(webService.POST("/marker").To(markerRest.Create).
			Doc("create marker").
			Metadata(restfulspec.KeyOpenAPITags, []string{"marker"}).
			ReadsWithSchema(marker.Marker{}, public.SCHEMA_MARKER).
			Writes(marker.Marker{}).
			Returns(http.StatusCreated, "OK", marker.Marker{}))
		webService.Route(webService.GET("/marker/{marker-id}").To(markerRest.Get).
			Doc("get a marker").
			Metadata(restfulspec.KeyOpenAPITags, []string{"marker"}).
			Param(webService.PathParameter("marker-id", "marker identifier")).
			Writes(public.InfoResponse{}).
			Returns(http.StatusOK, "OK", public.InfoResponse{}))
		webService.Route(webService.PUT("/marker/{marker-id}").To(markerRest.Update).
			Doc("update a marker").
			Metadata(restfulspec.KeyOpenAPITags, []string{"marker"}).
			Param(webService.PathParameter("marker-id", "marker identifier")).
			ReadsWithSchema(marker.Marker{}, public.SCHEMA_MARKER).
			Writes(marker.Marker{}).
			Returns(http.StatusCreated, "OK", marker.Marker{}))
		webService.Route(webService.DELETE("/marker/{marker-id}").To(markerRest.Delete).
			Doc("delete a marker").
			Metadata(restfulspec.KeyOpenAPITags, []string{"marker"}).
			Param(webService.PathParameter("marker-id", "marker identifier")).
			Writes(marker.Marker{}).
			Returns(http.StatusCreated, "OK", marker.Marker{}))
		webService.Route(webService.POST("/marker/image").To(markerRest.CreateMarkerImage).
			Doc("create marker image").
			Metadata(restfulspec.KeyOpenAPITags, []string{"marker"}).
			ReadsWithSchema(marker.MarkerImage{}, public.SCHEMA_MARKER_IMAGE).
			Writes(marker.MarkerImage{}).
			Returns(http.StatusCreated, "OK", marker.MarkerImage{}))
		webService.Route(webService.GET("/marker/image").To(markerRest.ReadAllMarkerImage).
			Doc("get all marker images").
			Metadata(restfulspec.KeyOpenAPITags, []string{"marker"}).
			Writes([]marker.MarkerImage{}).
			Returns(http.StatusOK, "OK", []marker.MarkerImage{}))
		// Icon service
		iconRest := public.NewIconRest(ctxt, mongoClient)
		webService.Route(webService.GET("/icon").To(iconRest.Get).
			Doc("get all icons").
			Metadata(restfulspec.KeyOpenAPITags, []string{"icon"}).
			Param(webService.QueryParameter("mac_address", "MAC address")).
			Writes([]moc.Icon{}).
			Returns(http.StatusOK, "OK", []moc.Icon{}))
		webService.Route(webService.POST("/icon").To(iconRest.Create).
			Doc("create an icon").
			Metadata(restfulspec.KeyOpenAPITags, []string{"icon"}).
			ReadsWithSchema(moc.Icon{}, public.SCHEMA_ICON).
			Writes(moc.Icon{}).
			Returns(http.StatusCreated, "OK", moc.Icon{}))
		webService.Route(webService.PUT("/icon/{icon-id}").To(iconRest.Update).
			Doc("update an icon").
			Metadata(restfulspec.KeyOpenAPITags, []string{"icon"}).
			Param(webService.PathParameter("icon-id", "icon identifier")).
			ReadsWithSchema(moc.Icon{}, public.SCHEMA_ICON).
			Writes(moc.Icon{}).
			Returns(http.StatusCreated, "OK", moc.Icon{}))
		webService.Route(webService.DELETE("/icon/{icon-id}").To(iconRest.Delete).
			Doc("delete an icon").
			Metadata(restfulspec.KeyOpenAPITags, []string{"icon"}).
			Param(webService.PathParameter("icon-id", "icon identifier")).
			Writes(moc.Icon{}).
			Returns(http.StatusCreated, "OK", moc.Icon{}))
		webService.Route(webService.POST("/icon/image").To(iconRest.CreateIconImage).
			Doc("create an icon image").
			Metadata(restfulspec.KeyOpenAPITags, []string{"icon"}).
			ReadsWithSchema(moc.IconImage{}, public.SCHEMA_ICON_IMAGE).
			Writes(moc.IconImage{}).
			Returns(http.StatusCreated, "OK", moc.IconImage{}))
		webService.Route(webService.GET("/icon/image").To(iconRest.GetIconImage).
			Doc("get all icon images").
			Metadata(restfulspec.KeyOpenAPITags, []string{"icon"}).
			Param(webService.QueryParameter("mac_address", "MAC address")).
			Writes([]moc.IconImage{}).
			Returns(http.StatusOK, "OK", []moc.IconImage{}))
		// Remote Site service
		remoteSiteRest := public.NewRemoteSiteRest(ctxt, mongoClient)
		webService.Route(webService.GET("/remotesite").To(remoteSiteRest.GetAll).
			Doc("get all remote sites").
			Metadata(restfulspec.KeyOpenAPITags, []string{"remotesite"}).
			Writes([]moc.RemoteSite{}).
			Returns(http.StatusOK, "OK", []moc.RemoteSite{}))
		webService.Route(webService.GET("/remotesite/{remotesite-id}").To(remoteSiteRest.Get).
			Doc("get a remote site").
			Metadata(restfulspec.KeyOpenAPITags, []string{"remotesite"}).
			Param(webService.PathParameter("remotesite-id", "remote site identifier")).
			Writes(moc.RemoteSite{}).
			Returns(http.StatusOK, "OK", moc.RemoteSite{}))
		webService.Route(webService.POST("/remotesite").To(remoteSiteRest.Create).
			Doc("create a remote site").
			Metadata(restfulspec.KeyOpenAPITags, []string{"remotesite"}).
			ReadsWithSchema(moc.RemoteSite{}, public.SCHEMA_REMOTESITE).
			Writes(moc.RemoteSite{}).
			Returns(http.StatusCreated, "OK", moc.RemoteSite{}))
		webService.Route(webService.PUT("/remotesite/{remotesite-id}").To(remoteSiteRest.Update).
			Doc("update a remote site").
			Metadata(restfulspec.KeyOpenAPITags, []string{"remotesite"}).
			Param(webService.PathParameter("remotesite-id", "remote site identifier")).
			ReadsWithSchema(moc.RemoteSite{}, public.SCHEMA_REMOTESITE).
			Writes(moc.RemoteSite{}).
			Returns(http.StatusCreated, "OK", moc.RemoteSite{}))
		webService.Route(webService.DELETE("/remotesite/{remotesite-id}").To(remoteSiteRest.Delete).
			Doc("delete a remote site").
			Metadata(restfulspec.KeyOpenAPITags, []string{"remotesite"}).
			Param(webService.PathParameter("remotesite-id", "remote site identifier")).
			Writes(moc.RemoteSite{}).
			Returns(http.StatusCreated, "OK", moc.RemoteSite{}))
		// Sit915 service
		sit915Rest := public.NewSit915Rest(ctxt, mongoClient)
		webService.Route(webService.POST("/sit915/{comm-link-type}/{remotesite-id}").To(sit915Rest.Create).
			Doc("send a sit915 message").
			Metadata(restfulspec.KeyOpenAPITags, []string{"sit915"}).
			Param(webService.PathParameter("comm-link-type", "type of communication link")).
			Param(webService.PathParameter("remotesite-id", "remote site identifier")).
			Writes(moc.Sit915{}).
			Returns(http.StatusCreated, "OK", moc.Sit915{}))
		webService.Route(webService.PUT("/sit915/retry/{message-id}").To(sit915Rest.Retry).
			Doc("retry to send a failed sit915 message").
			Metadata(restfulspec.KeyOpenAPITags, []string{"sit915"}).
			Param(webService.PathParameter("message-id", "SIT 915 message identifier")).
			Writes(moc.Sit915{}).
			Returns(http.StatusCreated, "OK", moc.Sit915{}))
		webService.Route(webService.PUT("/sit915/ack/{message-id}").To(sit915Rest.Ack).
			Doc("acknowledge a failed sit915 message").
			Metadata(restfulspec.KeyOpenAPITags, []string{"sit915"}).
			Param(webService.PathParameter("message-id", "SIT 915 message identifier")).
			Writes(moc.Sit915{}).
			Returns(http.StatusCreated, "OK", moc.Sit915{}))
		webService.Route(webService.GET("/sit915").To(sit915Rest.GetAll).
			Doc("get all sit915 messages").
			Metadata(restfulspec.KeyOpenAPITags, []string{"sit915"}).
			Writes([]moc.Sit915{}).
			Returns(http.StatusOK, "OK", []moc.Sit915{}))
		webService.Route(webService.GET("/sit915/{message-id}").To(sit915Rest.Get).
			Doc("get a sit915 message").
			Metadata(restfulspec.KeyOpenAPITags, []string{"sit915"}).
			Param(webService.PathParameter("message-id", "SIT 915 message identifier")).
			Writes(moc.Sit915{}).
			Returns(http.StatusOK, "OK", moc.Sit915{}))
		// Message service
		messageRest := public.NewMessageRest(ctxt, mongoClient)
		webService.Route(webService.GET("/message").To(messageRest.GetAll).
			Doc("get all sit messages").
			Metadata(restfulspec.KeyOpenAPITags, []string{"message"}).
			Param(webService.QueryParameter("sit-number", "SIT Number (0:ALL | 185 | 915)")).
			Param(webService.QueryParameter("start-datetime", "Start datetime to search")).
			Param(webService.QueryParameter("end-datetime", "End datetime to search")).
			Param(webService.QueryParameter("direction", "Message direction (0:ALL | 1:SENT | 2:RECEIVED)")).
			Writes([]public.MessageResponse{}).
			Returns(http.StatusOK, "OK", []public.MessageResponse{}))
		// MapConfig service
		mapconfigRest := public.NewMapConfigRest(ctxt, mongoClient)
		webService.Route(webService.GET("/mapconfig").To(mapconfigRest.ReadAll).
			Doc("get map config").
			Metadata(restfulspec.KeyOpenAPITags, []string{"mapconfig"}).
			Writes(moc.MapConfig{}).
			Returns(http.StatusOK, "OK", []moc.MapConfig{}))
		webService.Route(webService.POST("/mapconfig/set").To(mapconfigRest.SetSetting).
			Doc("set map config setting").
			Metadata(restfulspec.KeyOpenAPITags, []string{"mapconfig"}).
			ReadsWithSchema(moc.MapConfig{}, public.SCHEMA_MAP_CONFIG).
			Writes(moc.MapConfig{}).
			Returns(http.StatusCreated, "OK", moc.MapConfig{}))

		// MapConfig service
		filterTracksRest := public.NewFilterTracksRest(ctxt, mongoClient)
		webService.Route(webService.GET("/filtertracks/get/{user-id}").To(filterTracksRest.GetFilterTracks).
			Doc("get filter tracks").
			Metadata(restfulspec.KeyOpenAPITags, []string{"filtertracks"}).
			Writes(moc.FilterTracks{}).
			Returns(http.StatusOK, "OK", []moc.FilterTracks{}))
		webService.Route(webService.POST("/filtertracks/save/{user-id}").To(filterTracksRest.SaveFilterTracks).
			Doc("save filter tracks").
			Metadata(restfulspec.KeyOpenAPITags, []string{"filtertracks"}).
			ReadsWithSchema(moc.FilterTracks{}, public.SCHEMA_FILTER_TRACKS).
			Writes(moc.FilterTracks{}).
			Returns(http.StatusCreated, "OK", []moc.FilterTracks{}))

		// add to container
		container.Add(webService)
		handler.Add(webService.RootPath(), container)
		// Upload service
		fileRest := public.NewFileRest(mongoClient, libmain.TsiKillGroup)
		wsFile := new(restful.WebService).Path(pathRest + "/file")
		wsFile.Route(wsFile.POST("/").
			Metadata(restfulspec.KeyOpenAPITags, []string{"file"}).
			Consumes("multipart/form-data").
			Produces(restful.MIME_JSON).
			To(fileRest.Create))
		wsFile.Route(wsFile.PUT("/").
			Metadata(restfulspec.KeyOpenAPITags, []string{"file"}).
			Consumes("multipart/form-data").
			Produces(restful.MIME_JSON).
			To(fileRest.Create))
		wsFile.Route(wsFile.GET("/{file-id}").
			Metadata(restfulspec.KeyOpenAPITags, []string{"file"}).
			To(fileRest.Get))
		wsFile.Route(wsFile.DELETE("/{file-id}").
			Metadata(restfulspec.KeyOpenAPITags, []string{"file"}).
			Consumes(restful.MIME_JSON).
			Produces(restful.MIME_JSON).
			To(fileRest.Delete))
		// Filters for upload service
		wsFile.Filter(restful.NoBrowserCacheFilter)
		wsFile.Filter(service.RequestIdContextFilter)
		wsFile.Filter(service.SessionIdContextFilter)
		wsFile.Filter(routeAuthorizer.Authorize)
		// add to container
		container.Add(wsFile)
		handler.Add(wsFile.RootPath(), container)
		// OpenAPI spec (swagger)
		config := restfulspec.Config{
			WebServices:                   container.RegisteredWebServices(), // you control what services are visible
			APIPath:                       pathRest + "/apidocs.json",
			PostBuildSwaggerObjectHandler: enrichSwaggerObject,
			DisableCORS:                   true,
		}
		container.Add(restfulspec.NewOpenAPIService(config))
		// Setup v2 WebSocket
		wsViewHandler := public.NewViewWS(featDb, libmain.TsiKillGroup)
		handler.Add(pathWs+"/view/stream", wsViewHandler)
		wsHandler := ws.NewHandler(ctxt)
		wsHandler.Streamer = mongo.NewStreamer(mongoClient)
		publisher.Subscribe(public.TOPIC_INCIDENT, wsHandler)
		publisher.Subscribe(public.TOPIC_FLEET, wsHandler)
		publisher.Subscribe(public.TOPIC_VESSEL, wsHandler)
		publisher.Subscribe("Session", wsHandler)
		handler.Add(pathWs, wsHandler)
		server := &http.Server{Addr: listenAddress, Handler: handler}
		ServeDebug()
		log.Info("Listening on %v", listenAddress)
		// SARMAP
		//		sarmapServer := NewSarmapServer(listenAddressConfig, webService, miscDb, libmain.TsiKillGroup)
		//		sarmapServer.Run()
		// Configuration service
		configHandler := NewHandler()
		configContainer := restful.NewContainer()
		configService := new(restful.WebService)
		configService.Path(pathRest).
			Consumes(restful.MIME_JSON).
			Produces(restful.MIME_JSON)

		configService.Route(webService.GET("/config.json").To(configRest.Get).
			Doc("c2 configuration").
			Metadata(restfulspec.KeyOpenAPITags, []string{"config"}).
			Returns(http.StatusOK, "OK", map[string]interface{}{}))
		// SARMAP
		sarmapRest := public.NewSarmapRest(libmain.TsiKillGroup, mongoClient)
		configService.Route(webService.GET("/sarmap.json").To(sarmapRest.ReadAll).
			Doc("SARMAP integration").
			Metadata(restfulspec.KeyOpenAPITags, []string{"sarmap"}).
			Returns(http.StatusOK, "OK", []moc.GeoJsonFeaturePoint{}))

		configContainer.Add(configService)
		configHandler.Add(webService.RootPath(), configContainer)
		configServer := &http.Server{
			Addr:    listenAddressConfig,
			Handler: configHandler,
		}

		go func() {
			log.Info("Config service listening on %v", listenAddressConfig)
			if err = configServer.ListenAndServe(); err != nil {
				log.Fatal("config service failed: %v", err)
			}
		}()

		// Watch for certificate and key file system changes and reload key pair on create and change
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatal("unable to watch: %v", err)
		}
		defer watcher.Close()

		// initialize certificate wrapper and and populate TLSConfig with Getcertificate()
		wrappedCert := &certificate.WrappedCertificate{}
		server.TLSConfig = tlsconfig.NewTLSConfig()
		server.TLSConfig.GetCertificate = wrappedCert.GetCertificate

		// load certificate
		if err := wrappedCert.LoadCertificate(certFile, keyFile); err != nil {
			log.Error("service failed: %+v", err)
		}

		// certificate watcher required to be in a separate go routine according to the FAQ
		ctxt.Go(func() {
			wrappedCert.WatchCertificateFiles(ctxt, watcher, certFile, keyFile)
		})

		// certFile and keyFile are going to be under the same dir
		// adding both certFile and KeyFile to the watcher
		// handles if they are not on the same directory.
		if err := watcher.Add(filepath.Dir(certFile)); err != nil {
			log.Warn("Can not watch certification file: %+v", err)

		}
		if err := watcher.Add(filepath.Dir(keyFile)); err != nil {
			log.Warn("Can not watch key file: %+v", err)
		}

		// configHasCert evaluates if cerfile and keyfile paths exist and then checks if TLSConfig holds a valid reference to the certificates.
		configHasCert := certFile != "" && keyFile != "" && (len(server.TLSConfig.Certificates) > 0 || server.TLSConfig.GetCertificate != nil)
		// Run in the background so it can be rudely killed
		if configHasCert {
			handler.Add(pathRest+"/certificate.pem", certHandler{file: certFile})
			// passing empty cert and key file because the code is using TLSConfig.GetCertificate.
			if err = server.ListenAndServeTLS("", ""); err != nil {
				log.Fatal("service failed: %v", err)
			}
		} else {
			log.Debug("Listen and Server over http: %+v", server)
			if err = server.ListenAndServe(); err != nil {
				log.Fatal("service failed: %v", err)
			}
		}
		ctxt.Wait()

	})
}

type certHandler struct {
	file string
}

func (c certHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("Content-type", "application/x-pem-file")
	http.ServeFile(response, request, c.file)
}

func enrichSwaggerObject(swo *spec.Swagger) {
	swo.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "McMurdo PRISMA C2",
			Description: "McMurdo PRISMA C2 is the trusted solution for integrated maritime and aviation incident management.",
			Contact: &spec.ContactInfo{
				Name:  "McMurdo",
				Email: "support@mcmurdogroup.com",
				URL:   "http://www.mcmurdogroup.com/",
			},
			License: &spec.License{
				Name: "Commercial",
				URL:  "http://www.mcmurdogroup.com/privacy-policy/",
			},
			Version: "1.5.0",
		},
	}
	swo.Schemes = []string{"https"}
	swo.Host = listenAddress
	// workaround for definitions not exposed
	swo.Definitions["sar.isBeacon_Protocol"] = spec.Schema{}
	swo.Definitions["moc.isMulticast_Payload"] = spec.Schema{}
}

func envEnabled(name string) bool {
	switch os.Getenv(name) {
	case "1", "true", "TRUE", "t", "T":
		return true
	}
	return false
}

func TODO(_ *restful.Request, _ *restful.Response) {}
