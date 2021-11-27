// tgwad is used as a broadcast daemon. To use it in your code see TSIClient, also take a look at tmsg.GClient.
package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"prisma/tms/libmain"
	"prisma/tms/log"
	"prisma/tms/routing"
	"prisma/tms/tmsg"

	"prisma/gogroup"

	"prisma/tms/goejdb"

	// Import all PB message dirs so we can decode them for debugging
	_ "prisma/tms"
	_ "prisma/tms/client_api"
	_ "prisma/tms/db"
	_ "prisma/tms/devices"
	_ "prisma/tms/moc"
	_ "prisma/tms/nmea"
	_ "prisma/tms/routing"
)

var (
	port               = ""
	debugPort          = ""
	localSiteNum  uint = 1
	localSiteName      = ""
	dbpath             = ""
	securePort         = ""
	ca                 = ""
	cert               = ""
	priv               = ""
	siteDefs      Sites
)

func init() {
	flag.StringVar(&port, "port", ":31228", "TCP interface:port to listen on")
	flag.StringVar(&debugPort, "debug-port", ":8083", "Debugging port")
	flag.StringVar(&localSiteName, "name", "local", "Name of the local site")
	flag.StringVar(&dbpath, "db", "/tmp/tgwad.db", "Path to tgwad database for delayed remote delivery")
	flag.UintVar(&localSiteNum, "num", 1, "Number of the local site")
	flag.Var(&siteDefs, "site", "Specify a site in this format: <name>,<num>[,gw][,<route>]*. E.g. hq,5,tcp:10.5.0.5:31228")
	flag.StringVar(&securePort, "secure-port", "", "Port for accepting TLS connections")
	flag.StringVar(&ca, "ca", "", "File which contains the certificate authority")
	flag.StringVar(&cert, "cert", "", "File which contains our certificate")
	flag.StringVar(&priv, "priv", "", "File which contains our private key")
}

func main() {
	libmain.MainOpts(tmsg.APP_ID_TGWAD, false, func(ctxt gogroup.GoGroup) {
		log.Debug("tgwad %v %v starting", localSiteNum, localSiteName)

		log.Debug("Opening database " + dbpath)

		// replace goejdb in the system by bolt or redis
		jb, err := goejdb.Open(dbpath, goejdb.JBOWRITER|goejdb.JBOCREAT)
		if err != nil {
			log.Fatal("Error opening database: %v", err)
		}
		log.Debug("Creating collection")
		msgs, jberr := jb.CreateColl("msgs", nil)
		if jberr != nil {
			log.Fatal("Error creating collection in database: %v", jberr)
		}

		log.Debug("Routing table: %v", siteDefs)

		r := NewRouter(ctxt, uint32(localSiteNum), localSiteName)
		NewNetAcceptor(ctxt, r, port)

		if securePort != "" {
			if cert == "" || priv == "" || ca == "" {
				log.Fatal("A CA, certificate and private key MUST be specified to use TLS")
			}

			NewTLSNetAcceptor(ctxt, r, securePort, ca, cert, priv)
		}

		// Do debug server
		mux := http.NewServeMux()
		mux.Handle("/router/", http.StripPrefix("/router", r))

		// Set up remote sites
		for _, site := range siteDefs {
			var routes []*routing.Route
			for _, url := range site.Routes {
				routes = append(routes, &routing.Route{
					Url: url,
				})
			}
			cfg := RemoteConfig{
				Router: r,
				DB:     msgs,
				Info: routing.SiteInfo{
					Id:      site.Num,
					Name:    site.Name,
					Gateway: site.Gateway,
					Routes:  routes,
				},
				CA:      ca,
				Cert:    cert,
				PrivKey: priv,
			}
			rs, err := NewRemoteSite(ctxt, cfg)
			if err != nil {
				log.Fatal("Error creating remote site: %v", err)
			}

			prefix := "/remote/" + site.Name
			mux.Handle(prefix+"/", http.StripPrefix(prefix, rs))
		}

		ctxt.Go(http.ListenAndServe, debugPort, mux)
	})
}

type SiteDef struct {
	Num     uint32
	Name    string
	Gateway bool
	Routes  []string
}

func NewSiteDef(val string) (SiteDef, error) {
	parts := strings.Split(val, ",")
	if len(parts) < 2 {
		return SiteDef{}, errors.New("site name,num is required for site definition")
	}

	name := parts[0]
	num, err := strconv.Atoi(parts[1])
	if err != nil {
		return SiteDef{}, errors.New(fmt.Sprintf("Error interpreting site num: %v", err))
	}
	parts = parts[2:]

	gw := false
	if len(parts) > 0 && parts[0] == "gw" {
		gw = true
		parts = parts[1:]
	}

	def := SiteDef{
		Name:    name,
		Num:     uint32(num),
		Gateway: gw,
		Routes:  parts,
	}
	return def, nil
}

func (def SiteDef) String() string {
	s := fmt.Sprintf("%v,%v", def.Name, def.Num)
	if def.Gateway {
		s = s + ",gw"
	}
	if len(def.Routes) > 0 {
		s = s + "," + strings.Join(def.Routes, ",")
	}
	return s
}

type Sites []SiteDef

func (s *Sites) String() string {
	var siteStrings []string = nil
	for _, d := range *s {
		siteStrings = append(siteStrings, d.String())
	}
	return strings.Join(siteStrings, " ")
}

func (s *Sites) Set(val string) error {
	def, err := NewSiteDef(val)
	if err != nil {
		return err
	}
	*s = append(*s, def)
	return nil
}
