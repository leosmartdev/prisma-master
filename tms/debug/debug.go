// Package debug provides a flag to provide fast timers for different function.
package debug

import "flag"

var FastTimers bool

func init() {
	flag.BoolVar(&FastTimers, "fast-timers", false, "use fast timers for testing")
}
