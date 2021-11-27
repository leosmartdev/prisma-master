package db

import (
	"flag"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"prisma/gogroup"
	. "prisma/tms"
	"prisma/tms/log"
	"prisma/tms/moc"
	"prisma/tms/routing"
	ttmsg "prisma/tms/tmsg"
	. "prisma/tms/tmsg/client"
)

type DataInsertConfig struct {
	NumThreads int
}

var (
	InsertConfig DataInsertConfig
)

func init() {
	flag.IntVar(&InsertConfig.NumThreads, "insthreads", 10, "Number of goroutines to spawn for database inserts")
}

/**
 * DataInserter is responsible for listening on tgwad for insertable messages then calling the Insert() method in the
 * tracksdb to insert the data into the database.
 */
type DataInserter struct {
	config        DataInsertConfig
	tracks        TrackDB
	activities    ActivityDB
	devices       DeviceDB
	transmissions TransmissionDB
	sites         SiteDB
	misc          MiscDB
	file          FileDB

	tclient TsiClient
	ctxt    gogroup.GoGroup

	// Statistics info
	totalInserts         uint64
	totalErrors          uint64
	insertsSinceLastTick uint64
	lastTick             time.Time

	// TSN Info
	lastSecond int64
	counter    int32

	targetStream   <-chan *TMsg
	metadataStream <-chan *TMsg
	trackStream    <-chan *TMsg
	reportStream   <-chan *TMsg
	activityStream <-chan *TMsg
	deviceStream   <-chan *TMsg
	siteStream     <-chan *TMsg
	incidentStream <-chan *TMsg
	fileStream     <-chan *TMsg
	tranStream     <-chan *TMsg
	tckr           <-chan time.Time

	tracer *log.Tracer
}

func NewDataInserter(ctxt gogroup.GoGroup, client TsiClient, waits *sync.WaitGroup,
	tracks TrackDB, activities ActivityDB, devices DeviceDB, transmission TransmissionDB, sites SiteDB, misc MiscDB, file FileDB) (*DataInserter, error) {
	d := &DataInserter{
		config:        InsertConfig,
		tracks:        tracks,
		activities:    activities,
		devices:       devices,
		transmissions: transmission,
		sites:         sites,
		misc:          misc,
		file:          file,
		tclient:       client,
		ctxt:          ctxt,
	}
	waits.Add(1)
	ctxt.Go(func() {
		d.process()
		waits.Done()
	})
	d.tracer = log.GetTracer("inserter")
	return d, nil
}

// Try to read a message for insertion and then insert it
func (d *DataInserter) doInsert() {
	var err error

	select {
	case <-d.ctxt.Done():
		// Do nothing
	case <-d.tckr:
		d.logStatus()

	case tmsg := <-d.trackStream:
		track, ok := tmsg.Body.(*Track)
		if !ok {
			atomic.AddUint64(&d.totalErrors, 1)
			log.Warn("Got non-track in track stream. Got %v instead", reflect.TypeOf(tmsg.Body))
		} else {
			err = d.tracks.Insert(track)
			if err != nil {
				log.Warn("Error in insert track %+v", err)
			}
			atomic.AddUint64(&d.insertsSinceLastTick, 1)
			atomic.AddUint64(&d.totalInserts, 1)
		}
	case tmsg := <-d.tranStream:
		tran, ok := tmsg.Body.(*Transmission)
		if !ok {
			atomic.AddUint64(&d.totalErrors, 1)
			log.Warn("Got a non-transmission in transmission stream. Got %+v instead", reflect.TypeOf(tmsg.Body))
		} else {
			d.tracer.Logf("Got a transmission %+v", tmsg)
			err = d.transmissions.Create(tran)
			if err != nil {
				log.Warn("Error in insert track %+v", err)
			}
			atomic.AddUint64(&d.insertsSinceLastTick, 1)
			atomic.AddUint64(&d.totalInserts, 1)
		}
	case tmsg := <-d.activityStream:
		activity, ok := tmsg.Body.(*MessageActivity)
		if !ok {
			atomic.AddUint64(&d.totalErrors, 1)
			log.Warn("Got non-report in report stream. Got %v instead", reflect.TypeOf(tmsg.Body))
		} else {
			err = d.activities.Insert(activity)
			if err != nil {
				log.Warn("Error in insert activity %+v", err)
			}
			atomic.AddUint64(&d.insertsSinceLastTick, 1)
			atomic.AddUint64(&d.totalInserts, 1)
		}
	case tmsg := <-d.siteStream:
		site, ok := tmsg.Body.(*moc.Site)
		if ok {
			log.Info("moc.Site %v", site)
			_, err = d.sites.UpdateConnectionStatusBySiteId(d.ctxt, site)
			if err != nil {
				atomic.AddUint64(&d.totalErrors, 1)
				log.Warn("%v", err)
				return
			}
			atomic.AddUint64(&d.insertsSinceLastTick, 1)
			atomic.AddUint64(&d.totalInserts, 1)
		} else {
			atomic.AddUint64(&d.totalErrors, 1)
			log.Warn("expected moc.Site %v", reflect.TypeOf(tmsg.Body))
		}

	case tmsg := <-d.fileStream:
		file, ok := tmsg.Body.(*moc.File)
		if !ok {
			atomic.AddUint64(&d.totalErrors, 1)
			log.Warn("moc.File %v", reflect.TypeOf(tmsg.Body))
			return
		}
		log.Debug("moc.File %v %v %v", file.Metadata.Name, file.Metadata.Id, tmsg.NotifySent)
		d.file.Create(d.ctxt, file)
		if err != nil {
			atomic.AddUint64(&d.totalErrors, 1)
			log.Error(err.Error(), err)
			log.Error("NotifyId %v", tmsg.NotifySent)
			// send fail
			drAny, _ := ttmsg.PackFrom(&routing.DeliveryReport{
				NotifyId: tmsg.NotifySent,
				Status:   routing.DeliveryReport_FAILED,
			})
			d.tclient.Send(d.ctxt, &TsiMessage{
				Destination: []*EndPoint{
					tmsg.Source,
				},
				Body: drAny,
			})
			return
		}
		// send success
		drAny, _ := ttmsg.PackFrom(&routing.DeliveryReport{
			NotifyId: tmsg.NotifySent,
			Status:   routing.DeliveryReport_PROCESSED,
		})
		d.tclient.Send(d.ctxt, &TsiMessage{
			Destination: []*EndPoint{
				tmsg.Source,
			},
			Body: drAny,
		})
		log.Debug("inserted moc.File %v %v", file.Metadata.Name, file.Metadata.Id)
		atomic.AddUint64(&d.insertsSinceLastTick, 1)
		atomic.AddUint64(&d.totalInserts, 1)
	case tmsg := <-d.incidentStream:
		incident, ok := tmsg.Body.(*moc.Incident)
		if !ok {
			atomic.AddUint64(&d.totalErrors, 1)
			log.Warn("moc.Incident %v", reflect.TypeOf(tmsg.Body))
			return
		}
		log.Debug("moc.Incident %v", incident)
		if moc.Incident_Transferring != incident.State {
			return
		}
		// set tmsg notify id, used in tms/cmd/daemons/tanalyzed/incident.go
		incident.Log[len(incident.Log)-1].Note = fmt.Sprintf("%v %v", tmsg.Source.Site, tmsg.NotifySent)
		upsertResponse, err := d.misc.Upsert(GoMiscRequest{
			Req: &GoRequest{
				ObjectType: "prisma.tms.moc.Incident",
				Obj: &GoObject{
					Data: incident,
				},
			},
			Ctxt: d.ctxt,
			Time: &TimeKeeper{},
		})
		if err != nil {
			atomic.AddUint64(&d.totalErrors, 1)
			log.Error(err.Error(), err)
			log.Error("NotifyId %v", tmsg.NotifySent)
			// send fail
			drAny, _ := ttmsg.PackFrom(&routing.DeliveryReport{
				NotifyId: tmsg.NotifySent,
				Status:   routing.DeliveryReport_FAILED,
			})
			d.tclient.Send(d.ctxt, &TsiMessage{
				Destination: []*EndPoint{
					tmsg.Source,
				},
				Body: drAny,
			})
			return
		}
		// send success
		drAny, _ := ttmsg.PackFrom(&routing.DeliveryReport{
			NotifyId: tmsg.NotifySent,
			Status:   routing.DeliveryReport_PROCESSED,
		})
		d.tclient.Send(d.ctxt, &TsiMessage{
			Destination: []*EndPoint{
				tmsg.Source,
			},
			Body: drAny,
		})
		incident.Id = upsertResponse.Id
		atomic.AddUint64(&d.insertsSinceLastTick, 1)
		atomic.AddUint64(&d.totalInserts, 1)
	case tmsg := <-d.deviceStream:
		device, ok := tmsg.Body.(*moc.Device)
		if !ok {
			atomic.AddUint64(&d.totalErrors, 1)
			log.Warn("Got non-device in device stream. Got %v instead", reflect.TypeOf(tmsg.Body))
		} else {
			err = d.devices.Insert(device)
			if err != nil {
				log.Debug("Error in insert device %v", err)
			}
			atomic.AddUint64(&d.insertsSinceLastTick, 1)
			atomic.AddUint64(&d.totalInserts, 1)
		}
	case tmsg := <-d.reportStream:
		d.tracer.Logf("report")
		report, ok := tmsg.Body.(*TrackReport)
		if !ok {
			atomic.AddUint64(&d.totalErrors, 1)
			log.Warn("Got non-report in report stream. Got %v instead", reflect.TypeOf(tmsg.Body))
		} else {
			err = nil
			for _, track := range report.Tracks {
				err := d.tracks.Insert(track)
				if err != nil {
					log.Debug("Error in insert %+v", err)
				}
				atomic.AddUint64(&d.insertsSinceLastTick, 1)
				atomic.AddUint64(&d.totalInserts, 1)
				if err != nil {
					atomic.AddUint64(&d.totalErrors, 1)
				}
			}
		}
	}
	if err != nil {
		atomic.AddUint64(&d.totalErrors, 1)
	}
}

func (d *DataInserter) insertThread() {
	// This is an insertion thread
	for {
		select {
		case <-d.ctxt.Done():
			// Are we being asked to exit?
			return
		default:
			d.doInsert()
		}
	}
}

// Setup to listen for messages and run some insert threads
func (d *DataInserter) process() {
	ctxt := d.ctxt.Child("inserter")

	ctxt.ErrCallback(func(err error) {
		pe, ok := err.(gogroup.PanicError)
		if ok {
			log.Error("Panic in insert thread: %v\n%v", pe.Msg, pe.Stack)
		} else {
			log.Error("Error in insert thread: %v", err)
		}
	})

	d.tckr = time.NewTicker(time.Duration(5) * time.Second).C

	// Listen for tracks from tgwad
	d.trackStream = d.tclient.Listen(ctxt, routing.Listener{
		MessageType: "prisma.tms.Track",
		Destination: &EndPoint{
			Site: ttmsg.GClient.ResolveSite(""),
		},
	})
	// Listen for track reports from tgwad
	d.reportStream = d.tclient.Listen(ctxt, routing.Listener{
		MessageType: "prisma.tms.TrackReport",
		Destination: &EndPoint{
			Site: ttmsg.GClient.ResolveSite(""),
		},
	})

	// Listen to activity reports from tgwad
	d.activityStream = d.tclient.Listen(ctxt, routing.Listener{
		MessageType: "prisma.tms.MessageActivity",
		Destination: &EndPoint{
			Site: ttmsg.GClient.ResolveSite(""),
		},
	})

	// Listen for transmission reports from tgwad
	d.tranStream = d.tclient.Listen(ctxt, routing.Listener{
		MessageType: "prisma.tms.Transmission",
		Destination: &EndPoint{
			Site: ttmsg.GClient.ResolveSite(""),
			Aid:  ttmsg.APP_ID_TDATABASED,
		},
	})

	// Listen to device reports from tgwad
	d.deviceStream = d.tclient.Listen(ctxt, routing.Listener{
		MessageType: "prisma.tms.moc.Device",
		Destination: &EndPoint{
			Site: ttmsg.GClient.ResolveSite(""),
		},
	})

	d.siteStream = d.tclient.Listen(ctxt, routing.Listener{
		MessageType: "prisma.tms.moc.Site",
		Destination: &EndPoint{
			Site: ttmsg.GClient.ResolveSite(""),
		},
	})

	d.incidentStream = d.tclient.Listen(ctxt, routing.Listener{
		MessageType: "prisma.tms.moc.Incident",
		Destination: &EndPoint{
			Site: ttmsg.GClient.ResolveSite(""),
		},
	})

	d.fileStream = d.tclient.Listen(ctxt, routing.Listener{
		MessageType: "prisma.tms.moc.File",
		Destination: &EndPoint{
			Site: ttmsg.GClient.ResolveSite(""),
		},
	})

	// Initialize the stat counters
	d.lastTick = time.Now()
	d.totalInserts = 0
	d.insertsSinceLastTick = 0
	d.totalErrors = 0

	// Start some insert threads
	for i := 0; i < d.config.NumThreads; i++ {
		ctxt.GoRestart(func() { d.insertThread() })
	}

	// Wait for the insert threads to die
	ctxt.Wait()
}

// Push that stats to the log
func (d *DataInserter) logStatus() {
	insSince := atomic.SwapUint64(&d.insertsSinceLastTick, 0)
	log.Debug("Inserts: %v total, %v/s, %v errors",
		atomic.LoadUint64(&d.totalInserts),
		float64(insSince)/time.Since(d.lastTick).Seconds(),
		atomic.LoadUint64(&d.totalErrors))
	d.lastTick = time.Now()
}

func (d *DataInserter) generateTargetTSN(source *EndPoint) *TargetID {
	timeInSeconds := time.Now().Unix()
	if timeInSeconds == d.lastSecond {
		d.counter++
	} else {
		d.lastSecond = timeInSeconds
		d.counter = 1
	}

	tsn := &TimeSerialNumber{
		Seconds: timeInSeconds,
		Counter: d.counter,
	}

	return &TargetID{
		Producer: &SensorID{
			Site: source.Site,
			Eid:  source.Eid,
		},
		SerialNumber: &TargetID_TimeSerial{
			TimeSerial: tsn,
		},
	}
}
