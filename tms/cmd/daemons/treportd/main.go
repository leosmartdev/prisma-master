package main

import (
	"prisma/tms/libmain"
	"prisma/tms/log"
	"prisma/tms/tmsg"

	"flag"
	"time"

	"prisma/gogroup"
)

var (
	reportConf ReportConf
)

func init() {
	flag.StringVar(&reportConf.DBPath, "db", "/tmp/treportd.db", "Path to backlog database for delayed delivery")
	flag.StringVar(&reportConf.Destination, "dest", "hq", "Name of destination site")
	flag.DurationVar(&reportConf.QueueTime, "queue-time", time.Duration(5)*time.Second, "Wait for this amount of time to send data so that it can be batched with other data")
	flag.UintVar(&reportConf.MaxDBEntries, "max-db-size", 50000, "The maximum number of entries in the backlog database")
	flag.UintVar(&reportConf.ReportSize, "report-size", 32, "Maximum number of entries per report")
	flag.UintVar(&reportConf.ConcurrentReports, "concurrent-reports", 2, "Maximum number of reports to allow in the TGWAD queue at once")
}

func main() {
	libmain.Main(tmsg.APP_ID_TREPORTD, func(ctxt gogroup.GoGroup) {
		log.Debug("TReportD starting")

		r := NewReporter(ctxt, reportConf, tmsg.GClient)
		r.Start()
	})
}
