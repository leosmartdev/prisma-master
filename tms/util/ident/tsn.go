package ident

import (
	"sync"
	"time"

	"prisma/tms"
)

func Now() int64 {
	return time.Now().Unix()
}

var (
	mutex   sync.Mutex
	seconds int64
	counter int32 = 1
	Clock         = Now
)

//Generate tsn serial ID
func TSN() (int64, int32) {
	now := Clock()

	mutex.Lock()
	defer mutex.Unlock()
	if seconds != 0 && seconds == now {
		counter++
	} else {
		counter = 1
	}
	seconds = now
	return seconds, counter
}

func TimeSerialNumber() tms.TimeSerialNumber {
	seconds, counter := TSN()
	return tms.TimeSerialNumber{
		Seconds: seconds,
		Counter: counter,
	}
}
