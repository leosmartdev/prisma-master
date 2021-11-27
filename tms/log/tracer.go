package log

import (
	"flag"
	"fmt"
	"strings"
	"sync"
)

type tracerFlags []string

var (
	enabledTracers tracerFlags
	tracerMutex    sync.Mutex
	tracers        map[string]*Tracer
)

type Tracer struct {
	Enabled bool
	Prefix  string
}

func (t Tracer) trace(message string) {
	if t.Enabled {
		Alert(message)
	}
}

func (t Tracer) Log(message string) {
	t.trace(t.Prefix + message)
}

func (t Tracer) Logf(format string, args ...interface{}) {
	t.trace(t.Prefix + fmt.Sprintf(format, args...))
}

func (t *Tracer) Start(prefix string) {
	t.Enabled = true
	t.Prefix = prefix
}

func GetTracer(name string) *Tracer {
	tracerMutex.Lock()
	defer tracerMutex.Unlock()
	tracer := tracers[name]
	if tracer == nil {
		prefix := fmt.Sprintf("[%v] ", name)
		tracer = &Tracer{Prefix: prefix}
		tracers[name] = tracer
	}
	return tracer
}

func (t tracerFlags) String() string {
	return strings.Join(t, ",")
}

func (t *tracerFlags) Set(value string) error {
	for _, name := range strings.Split(value, ",") {
		*t = append(*t, name)
	}
	return nil
}

func init() {
	tracers = make(map[string]*Tracer, 0)
	flag.Var(&enabledTracers, "trace",
		"comma-separated list of tracers to enable")
}

func RegisterTracers() {
	for _, name := range enabledTracers {
		GetTracer(name).Enabled = true
	}
}
