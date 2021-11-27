// tmsd is a daemon to start and stop tms processes according to a configuration file.
package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"prisma/tms/libmain"
	"prisma/tms/log"
	"prisma/tms/tmsg"

	. "prisma/tms/cmd/daemons/tmsd/lib"

	"github.com/kardianos/osext"
	"golang.org/x/net/context"
)

var (
	start   bool
	status  bool
	info    bool
	stop    bool
	restart bool
	cleanup bool
)

func main() {
	flag.BoolVar(&start, "start", false, "Start TMS services?")
	flag.BoolVar(&status, "status", false, "Show TMS status?")
	flag.BoolVar(&stop, "stop", false, "Stop TMS services?")
	flag.BoolVar(&restart, "restart", false, "Restart TMS services?")
	flag.BoolVar(&info, "info", false, "Want detailed information of running processes?")
	flag.BoolVar(&cleanup, "cleanup", false, "Cleanup tmsd.sock, tmsd.pid and unmanaged tmsd processes?")
	flag.Parse()
	if libmain.PrintVersion {
		exe, err := osext.Executable()
		if err == nil {
			fmt.Printf("TMS %v version: %v build date %v\n", path.Base(exe), libmain.VersionNumber, libmain.VersionDate)
			os.Exit(0)
		}
	}
	log.Init("tmsd", tmsg.GClient)

	_, err := ReadConfig()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: non-existent or invalid tmsd.conf\n")
		os.Exit(1)
	}

	actions := 0
	if start {
		actions++
	}
	if status {
		actions++
	}
	if info {
		actions++
	}
	if stop {
		actions++
	}
	if restart {
		actions++
	}
	if cleanup {
		actions++
	}

	if actions != 1 {
		fmt.Fprintf(os.Stderr, "Error: must specify ONE of (--start, --status, --info, --stop, --restart, --cleanup)\n")
		os.Exit(1)
	}
	tmsd := NewTmsd(&Config)

	if start {
		tmsd.Start(context.Background())
	} else if status {
		tmsd.Status()
	} else if info {
		tmsd.Info()
	} else if stop {
		tmsd.Stop()
	} else if restart {
		tmsd.Restart(context.Background())
	} else if cleanup {
		tmsd.Cleanup()
	}
}
