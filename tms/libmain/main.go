// Package libmain provides common main function which does extra work
package libmain

import (
	"prisma/tms/log"
	"prisma/tms/tmsg"

	"flag"
	"fmt"
	golog "log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path"
	"runtime"

	"prisma/gogroup"

	"github.com/kardianos/osext"
)

var (
	// Background routines which much exit before we exit
	TsiBackground gogroup.GoGroup
	// Background routines which can be rudely killed when we exit
	TsiKillGroup gogroup.GoGroup

	ProfilePort  string
	PrintVersion bool
)

func init() {
	flag.StringVar(&ProfilePort, "profile", "", "Profile and listen on this port e.g. localhost:6060")
	flag.BoolVar(&PrintVersion, "version", false, "Print version then exit")
	flag.String("server_address", "", "")
}

func Main(appid uint32, realMain func(gogroup.GoGroup)) {
	MainOpts(appid, true, realMain)
}

func MainOpts(appid uint32, tmsgClient bool, realMain func(gogroup.GoGroup)) {
	flag.Parse()
	log.RegisterTracers()

	exe, err := osext.Executable()
	if err != nil {
		golog.Fatalf("Cannot find executable: %v", err)
	}

	if PrintVersion {
		fmt.Printf("TMS %v version: %v build date %v\n",
			path.Base(exe), VersionNumber, VersionDate)
		os.Exit(0)
	}

	runtime.SetBlockProfileRate(0)
	runtime.SetCPUProfileRate(0)
	if ProfilePort != "" {
		runtime.SetBlockProfileRate(10)
		runtime.SetCPUProfileRate(1000)
		go func() { golog.Println(http.ListenAndServe(ProfilePort, nil)) }()
	}

	log.Init(path.Base(exe), tmsg.GClient)
	TsiBackground = gogroup.New(nil, "background")
	TsiBackground.ErrCallback(func(err error) {
		pe, ok := err.(gogroup.PanicError)
		if ok {
			log.Error("Panic in TsiBackground goroutine: %v\n%v", pe.Msg, pe.Stack)
		} else {
			log.Error("Error in TsiBackground goroutine: %v", err)
		}
	})
	TsiKillGroup = TsiBackground.Child("killgroup")

	sigch := make(chan os.Signal)
	signal.Notify(sigch, os.Interrupt)

	go func() {
		_ = <-sigch
		log.Info("Got SIGINT, cancelling main context")
		if EnvDevelopment() {
			log.Info("TMS_ENV=development, killing program")
			os.Exit(1)
		}
		TsiBackground.Cancel(nil)

		_ = <-sigch
		log.Info("Got second SIGINT, killing program")
		os.Exit(1)
	}()

	// Setup
	if tmsgClient {
		err = tmsg.TsiClientGlobal(TsiBackground, appid)
		if err != nil {
			log.Error("Could not create global TsiClient: %v", err)
		}
	}

	// Run real main
	TsiBackground.Run(realMain)

	// Run and wait for cancel to make sure there's at least one thing in the
	// TsiBackground group.
	TsiBackground.Run(func(g gogroup.GoGroup) {
		<-g.Done()
	})

	// Wait for all background processes to exit
	TsiBackground.Wait()
	TsiBackground.Cancel(nil)
}

func EnvDevelopment() bool {
	switch os.Getenv("TMS_ENV") {
	case "development":
		return true
	}
	return false
}
