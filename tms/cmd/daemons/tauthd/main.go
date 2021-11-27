// tauthtd is used to manage accesses of tms system.
package main

import (
	"expvar"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"prisma/gogroup"
	"prisma/tms/cmd/daemons/tauthd/public"
	"prisma/tms/cmd/daemons/tauthd/ticker"
	"prisma/tms/db/connect"
	"prisma/tms/libmain"
	plog "prisma/tms/log"
	auth "prisma/tms/public"
	"prisma/tms/security/certificate"
	"prisma/tms/security/database"
	"prisma/tms/security/message"
	"prisma/tms/security/policy"
	"prisma/tms/security/service"
	"prisma/tms/security/tlsconfig"
	"prisma/tms/tmsg"
	"prisma/tms/ws"

	// tlog tag is used until we get rid of "log" dependency
	tlog "prisma/tms/log"

	"github.com/fsnotify/fsnotify"
	"github.com/globalsign/mgo"
	"github.com/go-openapi/spec"
	restful "github.com/orolia/go-restful"
	restfulspec "github.com/orolia/go-restful-openapi"
)

var (
	listenAddress = ""
	certFile      = ""
	keyFile       = ""
	pathRest      = "/api/v2/auth"
	goroutines    = expvar.NewInt("goroutines")
)

func init() {
	flag.StringVar(&listenAddress, "listen", ":8181", "Listen on host and port <address:port>")
	flag.StringVar(&certFile, "certificate", "/etc/trident/certificate.pem", "The path to the certificate file")
	flag.StringVar(&keyFile, "key", "/etc/trident/key.pem", "The path to the key file")
}

func main() {
	// MongoDb debug
	if envEnabled("MGO_DEBUG") {
		logger := log.New(os.Stdout, "[mgo] ", log.LUTC|log.Lshortfile)
		mgo.SetLogger(logger)
		mgo.SetDebug(true)
		logger.Println("MGO_DEBUG activated")
	}
	// go-restful debug
	if envEnabled("REST_DEBUG") {
		logger := log.New(os.Stdout, "[rest] ", log.LUTC|log.Lshortfile)
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
	libmain.Main(tmsg.APP_ID_TAAAD, func(ctxt gogroup.GoGroup) {
		plog.Notice("Connecting to database...")
		//mongodb test connection
		mongoClient, err := connect.GetMongoClient(ctxt, tmsg.GClient)
		if err != nil {
			panic(err)
		}
		ctxt = gogroup.WithValue(ctxt, "mongodb", mongoClient.DialInfo)
		ctxt = gogroup.WithValue(ctxt, "mongodb-cred", mongoClient.Cred)
		// set dial info in filter for each request
		service.SetDialInfo(mongoClient.DialInfo)
		// set creds info in the filter for each request
		service.SetCredential(mongoClient.Cred)
		// setup web service
		webService := new(restful.WebService)
		webService.Consumes(restful.MIME_JSON)
		webService.Produces(restful.MIME_JSON)
		webService.Path(pathRest)
		// setup container
		container := restful.NewContainer()
		// Publisher-Subscriber
		publisher := ws.NewPublisher()
		publisher.Subscribe("Session", tmsg.GClient)
		// Filters
		cors := restful.CrossOriginResourceSharing{
			AllowedDomains: []string{},
			AllowedHeaders: []string{restful.HEADER_ContentType, restful.HEADER_Accept, "Link", "X-Request-Id"},
			ExposeHeaders:  []string{"Link", "X-Request-Id"},
			CookiesAllowed: true,
			MaxAge:         3600,
			Container:      container}
		container.Filter(cors.Filter)
		webService.Filter(restful.NoBrowserCacheFilter)
		webService.Filter(service.RequestIdContextFilter)
		webService.Filter(service.SessionIdContextFilter)
		routeAuthorizer := &service.RouteAuthorizer{}
		webService.Filter(routeAuthorizer.Authorize)
		// User service
		userRest := public.NewUserRest(ctxt)
		webService.Route(webService.POST("/user").To(userRest.Post).
			Doc("create user").
			Metadata(restfulspec.KeyOpenAPITags, []string{"user"}).
			Writes(database.User{}).
			Returns(http.StatusCreated, "OK", database.User{}))
		webService.Route(webService.GET("/user").To(userRest.Get).
			Doc("get all users").
			Metadata(restfulspec.KeyOpenAPITags, []string{"user"}))
		webService.Route(webService.GET("/user/{user-id}").To(userRest.GetOne).
			Doc("get user").
			Metadata(restfulspec.KeyOpenAPITags, []string{"user"}))
		webService.Route(webService.PUT("/user/{user-id}").To(userRest.Put).
			Doc("update user").
			Metadata(restfulspec.KeyOpenAPITags, []string{"user"}))
		webService.Route(webService.PUT("/user/{user-id}/state/{user-state}").To(userRest.UpdateState).
			Doc("update state of user").
			Metadata(restfulspec.KeyOpenAPITags, []string{"user"}))
		webService.Route(webService.DELETE("/user/{user-id}").To(userRest.Delete).
			Doc("deactivate user").
			Metadata(restfulspec.KeyOpenAPITags, []string{"user"}))
		// Profile service
		profileRest := new(public.ProfileRest)
		webService.Route(webService.PUT("/profile/{user-id}").To(profileRest.Put).
			Doc("update profile, password").
			Metadata(restfulspec.KeyOpenAPITags, []string{"profile"}))
		// Session service
		sessionRest := public.NewSessionRest(ctxt, publisher)
		webService.Route(webService.POST("/session").To(sessionRest.Post).
			Doc("create session, login, signin").
			Metadata(restfulspec.KeyOpenAPITags, []string{"session"}).
			Returns(http.StatusCreated, "created", message.Session{}).
			ReadsWithSchema(auth.LoginRequest{}, public.SchemaSessionCreate(ctxt)).
			Writes(message.Session{}))
		webService.Route(webService.GET("/session").To(sessionRest.Get).
			Doc("get session").
			Metadata(restfulspec.KeyOpenAPITags, []string{"session"}))
		webService.Route(webService.DELETE("/session").To(sessionRest.Delete).
			Doc("delete session, logout, signout").
			Metadata(restfulspec.KeyOpenAPITags, []string{"session"}))
		// Object service
		objectRest := new(public.ObjectRest)
		webService.Route(webService.GET("/object").To(objectRest.Get).
			Doc("objects in the system and the actions that can be performed").
			Metadata(restfulspec.KeyOpenAPITags, []string{"object"}))
		// Role service
		roleRest := new(public.RoleRest)
		webService.Route(webService.GET("/role").To(roleRest.Get).
			Doc("list roles").
			Metadata(restfulspec.KeyOpenAPITags, []string{"role"}))
		// UserRole service
		userRoleRest := new(public.UserRoleRest)
		webService.Route(webService.GET("/user/{user-id}/role").To(userRoleRest.GetUserRole).
			Doc("roles of the user").
			Metadata(restfulspec.KeyOpenAPITags, []string{"role"}))
		webService.Route(webService.GET("/role/{role-id}/user").To(userRoleRest.GetRoleUser).
			Doc("users with a role").
			Param(webService.PathParameter("role-id", "role identifier")).
			Writes([]database.User{}).
			Metadata(restfulspec.KeyOpenAPITags, []string{"role"}))
		webService.Route(webService.PUT("/user/{user-id}/role/{role-id}").To(userRoleRest.Put).
			Doc("assign role to a user").
			Metadata(restfulspec.KeyOpenAPITags, []string{"role"}))
		webService.Route(webService.DELETE("/user/{user-id}/role/{role-id}").To(userRoleRest.Delete).
			Doc("remove role from a user").
			Metadata(restfulspec.KeyOpenAPITags, []string{"role"}))
		// Audit service
		auditRest := public.NewAuditRest(ctxt, mongoClient)
		webService.Route(webService.GET("/audit").To(auditRest.Get).
			Doc("list activity in the system").
			Metadata(restfulspec.KeyOpenAPITags, []string{"audit"}).
			Param(webService.QueryParameter("query", "string to search for")))
		webService.Route(webService.GET("/audit/session/{session-id}").To(auditRest.GetBySessionId).
			Doc("list activity for the session").
			Metadata(restfulspec.KeyOpenAPITags, []string{"audit"}))
		webService.Route(webService.GET("/audit/user/{user-id}").To(auditRest.GetByUserId).
			Doc("list activity for the user").
			Metadata(restfulspec.KeyOpenAPITags, []string{"audit"}))
		webService.Route(webService.GET("/audit/incident").To(auditRest.GetIncident).
			Doc("list activity for the incident").
			Metadata(restfulspec.KeyOpenAPITags, []string{"audit"}).
			Param(webService.QueryParameter("query", "string to search for")))
		// add to container
		container.Add(webService)
		// swagger
		config := restfulspec.Config{
			WebServicesURL:                "http://" + listenAddress + pathRest,
			WebServices:                   container.RegisteredWebServices(), // you control what services are visible
			APIPath:                       pathRest + "/apidocs.json",
			PostBuildSwaggerObjectHandler: enrichSwaggerObject,
			DisableCORS:                   false,
		}
		// Policy service
		policyRest := public.NewPolicyRest(func() {
			container.Remove(restfulspec.NewOpenAPIService(config))
			container.Remove(webService)
			// update
			container.Add(webService)
			// update swagger
			config := restfulspec.Config{
				WebServicesURL:                "http://" + listenAddress + pathRest,
				WebServices:                   container.RegisteredWebServices(), // you control what services are visible
				APIPath:                       pathRest + "/apidocs.json",
				PostBuildSwaggerObjectHandler: enrichSwaggerObject,
				DisableCORS:                   false,
			}
			container.Add(restfulspec.NewOpenAPIService(config))
		})
		webService.Route(webService.GET("/policy").To(policyRest.Read).
			Doc("get policies").
			Metadata(restfulspec.KeyOpenAPITags, []string{"policy"}).
			Writes(policy.Policy{}).
			Returns(http.StatusOK, "OK", policy.Policy{}))
		webService.Route(webService.PUT("/policy").To(policyRest.Update).
			Doc("set policies").
			Metadata(restfulspec.KeyOpenAPITags, []string{"policy"}).
			ReadsWithSchema(policy.Policy{}, public.SCHEMA_POLICY_UPDATE).
			Writes(policy.Policy{}).
			Returns(http.StatusOK, "OK", policy.Policy{}))
		container.Add(restfulspec.NewOpenAPIService(config))
		// run ticker
		go ticker.UserTicker(ctxt)
		// run server
		server := &http.Server{Addr: listenAddress, Handler: container}
		tlog.Debug("Listening on %+s", listenAddress)

		// Watch for certificate and key file system changes and reload key pair on create and change
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			tlog.Fatal("unable to watch: %+v", err)
		}
		defer watcher.Close()

		// initialize certificate wrapper and and populate TLSConfig with Getcertificate()
		wrappedCert := &certificate.WrappedCertificate{}
		server.TLSConfig = tlsconfig.NewTLSConfig()
		server.TLSConfig.GetCertificate = wrappedCert.GetCertificate

		// load certificate
		if err := wrappedCert.LoadCertificate(certFile, keyFile); err != nil {
			tlog.Error("service failed: %+v", err)
		}

		// certificate watcher required to be in a separate go routine according to the FAQ
		ctxt.Go(func() {
			wrappedCert.WatchCertificateFiles(ctxt, watcher, certFile, keyFile)
		})

		// certFile and keyFile are going to be under the same dir
		// adding both certFile and KeyFile to the watcher
		// handles if they are not on the same directory.
		if err := watcher.Add(filepath.Dir(certFile)); err != nil {
			tlog.Warn("Can not watch certification file: %+v", err)

		}
		if err := watcher.Add(filepath.Dir(keyFile)); err != nil {
			tlog.Warn("Can not watch key file: %+v", err)
		}
		// configHasCert evaluates if cerfile and keyfile paths exist and then checks if TLSConfig holds a valid reference to the certificates.
		configHasCert := certFile != "" && keyFile != "" && (len(server.TLSConfig.Certificates) > 0 || server.TLSConfig.GetCertificate != nil)

		// Run in the background so it can be rudely killed
		if configHasCert {
			// passing empty cert and key file because the code is using TLSConfig.GetCertificate.
			if err := server.ListenAndServeTLS("", ""); err != nil {
				log.Fatalf("service failed: %v", err)
			}
		} else {
			tlog.Debug("Listen and Server over http: %+v", server)
			if err := server.ListenAndServe(); err != nil {
				log.Fatalf("service failed: %v", err)
			}
		}
	})
}

func envEnabled(name string) bool {
	switch os.Getenv(name) {
	case "1", "true", "TRUE", "t", "T":
		return true
	}
	return false
}

func enrichSwaggerObject(swo *spec.Swagger) {
	swo.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "McMurdo PRISMA C2 Auth",
			Description: "McMurdo PRISMA C2 is the trusted solution for integrated maritime incident management. ",
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
	swo.BasePath = pathRest
	swo.Tags = []spec.Tag{
		{TagProps: spec.TagProps{Name: "audit", Description: "Audit log viewing"}},
		{TagProps: spec.TagProps{Name: "object", Description: "List of objects and actions"}},
		{TagProps: spec.TagProps{Name: "user", Description: "User management"}},
		{TagProps: spec.TagProps{Name: "policy", Description: "Policy management"}},
		{TagProps: spec.TagProps{Name: "profile", Description: "User profile"}},
		{TagProps: spec.TagProps{Name: "role", Description: "Role management"}},
		{TagProps: spec.TagProps{Name: "session", Description: "Session management - login, logout"}},
	}
	swo.Definitions["message.Session_State"] = spec.Schema{}
	swo.Definitions["message.User_State"] = spec.Schema{}
}
