// tsimulator provides information about objects and stations.
package main

import (
	"context"
	"flag"
	"net"
	"time"

	"prisma/gogroup"
	"prisma/tms/cmd/tools/tsimulator/object"
	"prisma/tms/cmd/tools/tsimulator/task"
	"prisma/tms/cmd/tools/tsimulator/web"
	"prisma/tms/log"
)

var (
	fileStations   = flag.String("fstations", "stations_data.json", "A file which contains data of station")
	fileVessels    = flag.String("fvessels", "vessels_data.json", "A file which contains data of seaObjects")
	tfleetdAddress = flag.String("tfaddr", ":7777", "An address of tfleetd for sending information about seaobjects, for example: omnicom beacons")
	tmccdAddress   = flag.String("tmccdaddr", ":9999", "An address of tmccd for sending information about sending a alert for rescuing, it can be Sarsat beacons")
	webAddr        = flag.String("webaddr", ":8089", "An address for listening to connections of REST requests")
	gssAddr        = flag.String("gssAddr", ":10800", "Address for simulating gss")
	tFleetdTCPAddr *net.TCPAddr
	tmccdTCPAddr   *net.TCPAddr
	timeConfig     object.TimeConfig
)

const sleepCPUSafe = 500 * time.Millisecond
const chanTasksResultSize = 128

func init() {
	log.Init("tsimulator", nil)
}

// InitMoving common data also group of stations and seaObjects
func main() {
	flag.Parse()
	stations := getStationFromFile(*fileStations)
	var objects []object.Object
	timeConfig, objects = getParametersFromFile(*fileVessels)
	seaObjectControl := object.NewObjectControl(objects, &timeConfig)
	resolveGlobalAddresses()

	// Register channels at an observer
	chanBeacons := make(chan object.Object, len(objects))
	seaObjectChan := make([]chan object.Object, len(stations))
	chanTasks := make(chan task.Result, chanTasksResultSize)

	seaObjectControl.RegisterChannel(chanBeacons)
	seaObjectControl.RegisterTaskChannel(chanTasks)
	for i := range stations {
		seaObjectChan[i] = make(chan object.Object, len(objects))
		go stations[i].See(seaObjectChan[i])
		seaObjectControl.RegisterChannel(seaObjectChan[i])
	}
	seaObjectControl.RunWatcher()
	// setup a listener for mt messages
	dpmt := web.NewDirectIPMT(seaObjectControl)

	goSimulator := gogroup.New(nil, "simulator")
	goSimulator.Go(lifeCycle(goSimulator, seaObjectControl))
	goSimulator.Go(handleResultTasks(goSimulator, chanTasks))
	goSimulator.Go(handleBeacons(goSimulator, chanBeacons))
	goSimulator.Go(listenMT(dpmt, goSimulator))
	//Run a life of station targets and seaObjects
	for i := range stations {
		goSimulator.Go(listenClients(goSimulator, &stations[i]))
	}
	goSimulator.Go(web.SetupServer(seaObjectControl, *webAddr))
	goSimulator.Wait()
}

func listenMT(dpmt *web.DirectIPMT, ctxt gogroup.GoGroup) func() {
	return func() {
		if err := dpmt.Listen(ctxt, *gssAddr); err != nil {
			log.Crit(err.Error())
		}
	}
}

// wait and listen to port for clients and send information from AIS targets
func listenClients(ctx context.Context, station *object.Station) func() {
	return func() {
		listener, err := net.Listen("tcp", station.Addr)
		if err != nil {
			log.Fatal("Error listen to port: %v", err)
		}
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Warn("Error to accept a connection: %v", err)
				continue
			}
			log.Debug("got a connection %v", conn)
			go handleConnection(conn, ctx, station)
		}
	}
}

func resolveGlobalAddresses() {
	var err error
	tFleetdTCPAddr, err = net.ResolveTCPAddr("tcp4", *tfleetdAddress)
	if err != nil {
		log.Fatal("Unable to resolve webAddr of tfleetd: %v", err)
	}
	tmccdTCPAddr, err = net.ResolveTCPAddr("tcp4", *tmccdAddress)
	if err != nil {
		log.Fatal("Unable to resolve webAddr of tfleetd: %v", err)
	}
}

// It's just a loop for moving objects on the map
func lifeCycle(ctx context.Context, seaObjectControl *object.Control) func() {
	return func() {
		for {
			select {
			case <-ctx.Done():
			default:
				seaObjectControl.Move()
			}
			time.Sleep(sleepCPUSafe)
		}
	}
}
