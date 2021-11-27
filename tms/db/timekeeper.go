package db

import (
	"time"
)

/**
 * When running in Replay mode, we want some way to map the current time to the
 * simulated/replay time. This object is responsible for doing exactly that.
 * When not in Replay mode, it simply passes through values.
 */
type TimeKeeper struct {
	Replay     bool
	StartTime  time.Time
	ReplayTime time.Time
	Speed      float64
}

// Convert a simulated/replay time to real/actual time
func (tk *TimeKeeper) ToReal(t time.Time) time.Time {
	if !tk.Replay {
		return t
	}

	sinceRT := t.Sub(tk.ReplayTime)
	dialated := time.Duration(float64(sinceRT) * tk.Speed)
	return tk.StartTime.Add(dialated)
}

// Convert real/actual time to simulated/replay time
func (tk *TimeKeeper) FromReal(t time.Time) time.Time {
	if !tk.Replay {
		return t
	}
	dialated := time.Duration(float64(t.Sub(tk.StartTime)) / tk.Speed)
	simTime := tk.ReplayTime.Add(dialated)
	return simTime
}

// Get the current simulated/replay time
func (tk *TimeKeeper) Now() time.Time {
	if !tk.Replay {
		return time.Now()
	}
	return tk.FromReal(time.Now())
}
