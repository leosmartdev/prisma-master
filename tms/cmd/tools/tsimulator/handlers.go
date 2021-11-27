package main

import (
	"context"
	"net"
	"prisma/tms/cmd/tools/tsimulator/object"
	"prisma/tms/cmd/tools/tsimulator/task"
	"prisma/tms/log"
	"time"
)

// watch results of tasks
func handleResultTasks(ctx context.Context, tResult <-chan task.Result) func() {
	return func() {
		for {
			select {
			case <-ctx.Done():
				return
			case tr := <-tResult:
				if err := sendDataToAppID(tr.Data, tr.To); err != nil {
					log.Error(err.Error())
				}
			}
		}
	}
}

func handleBeacons(ctx context.Context, objects <-chan object.Object) func() {
	return func() {
		for sObj := range objects {

			select {
			case <-ctx.Done():
				log.Error("closed handling beacons")
				return
			default:
			}

			if err := sendDataToTFleetd(sObj); err != nil {
				log.Error(err.Error())
				return
			}

			if err := sendDataStartAlerting(sObj); err != nil {
				log.Error(err.Error())
				return
			}
			if err := sendDataStopAlerting(sObj); err != nil {
				log.Error(err.Error())
				return
			}
			if err := sendDataToTmccd(sObj); err != nil {
				log.Error(err.Error())
				return
			}
			select {
			case <-ctx.Done():
				return
			case <-sObj.ReportTimer:
				if err := sendDataToTFleetd(sObj); err != nil {
					log.Error(err.Error())
					return
				}
			default:
			}
		}
	}
}

// Write to a connection messages from sea Objects in a permanent loop
func handleConnection(conn net.Conn, ctx context.Context, station *object.Station) {
	defer conn.Close()

	// Before to do something with seaObjects we will have described about this station
	if err := sendStaticInformation(conn, station); err != nil {
		return
	}
	if err := sendPositionInformation(conn, station); err != nil {
		return
	}
	// Also we send information about vessel before the first tick
	for _, seaObject := range station.IterateVisibleSeaObjects() {
		if err := sendStaticInformation(conn, seaObject); err != nil {
			return
		}
	}

	// So make permanent loop and watch seaObjects
	for {
		select {
		case <-ctx.Done():
			return
		default:
			for _, sObject := range station.IterateVisibleSeaObjects() {
				select {
				case <-sObject.TickerTTM.C:
					if err := sendTrackedTargetMessage(conn, sObject,
						station.Latitude, station.Longitude); err != nil {
						return
					}
				case <-sObject.TickerPosition.C:
					if err := sendPositionInformation(conn, sObject); err != nil {
						return
					}
				case <-sObject.TickerStaticInformation.C:
					if err := sendStaticInformation(conn, sObject); err != nil {
						return
					}
				default:
				}
			}
			time.Sleep(sleepCPUSafe)
		}
	}
}
