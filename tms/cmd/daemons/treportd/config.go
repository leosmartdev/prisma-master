package main

import (
	"time"
)

type ReportConf struct {
	Destination       string        // Destination site for report
	DBPath            string        // Location of the database
	QueueTime         time.Duration // Wait for this amount of time to send data so that it can be batched with other data
	MaxDBEntries      uint          // The maximum number of entries in the backlog database
	ReportSize        uint          // Maximum number of entries per-report
	ConcurrentReports uint          // Maximum number of reports to allow in the TGWAD queue at once
}
