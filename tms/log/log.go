// Package log provides function to log, trace, debug work of the system.
package log

import (
	"flag"
	"fmt"
	"io"
	reallog "log"
	log "log/syslog"
	"os"
	"prisma/tms/tmsg/client"
	"runtime"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
)

var (
	dflt      Logger = nil
	useStderr bool
	useSyslog bool
	useFile   string
	useDB     string
	logLevel  string
	fileLine  bool

	spewConfig = spew.ConfigState{
		Indent:   "  ",
		SortKeys: true,
		MaxDepth: 3,
	}

	colorPriority = map[log.Priority]string{
		log.LOG_EMERG:   NC,
		log.LOG_ALERT:   LightGreen,
		log.LOG_CRIT:    LightRed,
		log.LOG_ERR:     LightRed,
		log.LOG_WARNING: Yellow,
		log.LOG_NOTICE:  NC,
		log.LOG_INFO:    Blue,
		log.LOG_DEBUG:   Green,
	}
)

const (
	LOG_TRACE = log.LOG_DEBUG + 1

	LightRed    = "\033[1;31m"
	Red         = "\033[0;31m"
	Yellow      = "\033[0;33m"
	LightYellow = "\033[1;33m"
	Blue        = "\033[0;34m"
	LightBlue   = "\033[1;34m"
	NC          = "\033[0m"
	Green       = "\033[0;32m"
	LightGreen  = "\033[1;32m"
)

func init() {
	flag.BoolVar(&useStderr, "stdlog", false, "Write log to stderr?")
	flag.BoolVar(&useSyslog, "syslog", true, "Write log to syslog?")
	flag.BoolVar(&fileLine, "srcloc", true, "Find and write file:lineno to log?")
	flag.StringVar(&useDB, "dblog", "", "Write log & relevant objects to this EJDB database")
	flag.StringVar(&useFile, "filelog", "", "Write log to this file")
	flag.StringVar(&logLevel, "log", "info", "Set the logging level")
}

// Init is used to setup a logger
// Also it set ups different sources for logging
func Init(procname string, client client.TsiClient) {
	var level log.Priority
	switch strings.ToUpper(logLevel) {
	case "ERROR":
		level = log.LOG_ERR
	case "ALERT":
		level = log.LOG_ALERT
	case "CRITICAL":
		level = log.LOG_CRIT
	case "EMERGENCY":
		level = log.LOG_EMERG
	case "INFO":
		level = log.LOG_INFO
	case "NOTICE":
		level = log.LOG_NOTICE
	case "WARNING":
		level = log.LOG_WARNING
	case "DEBUG":
		level = log.LOG_DEBUG
	case "TRACE":
		level = LOG_TRACE
	default:
		reallog.Fatalf("Unknown logging level: %v", logLevel)
	}

	logger := &tsiLogger{
		level:    level,
		fileLine: fileLine,
	}

	if useStderr {
		logger.textlogs = []io.Writer{
			os.Stderr,
		}
	}

	if useFile != "" {
		f, err := os.OpenFile(useFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)
		if err != nil {
			reallog.Fatalf("Could not open log file: %v", useFile)
		} else {
			logger.textlogs = append(logger.textlogs, f)
		}
	}

	if useSyslog {
		syslog, err := log.Dial("", "", log.LOG_LOCAL0, procname)
		if err != nil {
			reallog.Fatalf("Could not dial syslog: %v", err)
		}
		logger.syslogs = []*log.Writer{
			syslog,
		}
	}

	if useDB != "" && client != nil {
		db, err := NewDBLogger(client, useDB)
		if err != nil {
			reallog.Fatalf("Could not open log DB: %v", err)
		} else {
			logger.dblogs = []DBLogger{
				db,
			}
		}
	}

	dflt = logger
}

type Logger interface {
	Log(prio log.Priority, msgFmt string, args ...interface{})
	TraceMsg(msgFmt string, args ...interface{})
	Trace(args ...interface{})
	Audit(msgFmt string, args ...interface{})
	Fatal(msgFmt string, args ...interface{})

	Emerg(msgFmt string, args ...interface{})
	Alert(msgFmt string, args ...interface{})
	Crit(msgFmt string, args ...interface{})
	Error(msgFmt string, args ...interface{})
	Warn(msgFmt string, args ...interface{})
	Notice(msgFmt string, args ...interface{})
	Info(msgFmt string, args ...interface{})
	Debug(msgFmt string, args ...interface{})
}

type tsiLogger struct {
	level    log.Priority
	fileLine bool
	syslogs  []*log.Writer
	textlogs []io.Writer
	dblogs   []DBLogger
}

// Convenience function for debugging
func Spew(obj ...interface{}) string {
	return spewConfig.Sdump(obj...)
}

/*******
 * Core logger functionality
 */
func (l *tsiLogger) Log(prio log.Priority, msgFmt string, args ...interface{}) {
	if prio <= l.level {
		formatArgs := fmtArgs(msgFmt, args)
		msg := spewConfig.Sprintf(msgFmt, formatArgs...)
		file := ""
		line := -1
		if l.fileLine || prio == LOG_TRACE {
			file, line = logSite()
			msg = fmt.Sprintf("%s: %v (%v:%v) %v", getColoredNamedPriority(prio), time.Now(), file, line, msg)
		} else {
			msg = fmt.Sprintf("%s: %v %v", getColoredNamedPriority(prio), time.Now(), msg)
		}
		if msgFmt != "" || prio != LOG_TRACE {
			l.writeToSyslogs(prio, msg)
			l.writeToTextLogs(prio, msg)
		}

		for _, db := range l.dblogs {
			db.Write(prio, file, line, msgFmt, args)
		}
	}
}

func (l *tsiLogger) TraceMsg(msgFmt string, args ...interface{}) {
	l.Log(LOG_TRACE, msgFmt, args...)
}

func (l *tsiLogger) Trace(args ...interface{}) {
	l.Log(LOG_TRACE, "", args...)
}

func (l *tsiLogger) writeToSyslogs(prio log.Priority, msg string) {
	for _, syslog := range l.syslogs {
		var err error = nil
		switch prio {
		case log.LOG_ERR:
			err = syslog.Err(msg)
		case log.LOG_ALERT:
			err = syslog.Alert(msg)
		case log.LOG_CRIT:
			err = syslog.Crit(msg)
		case log.LOG_EMERG:
			err = syslog.Emerg(msg)
		case log.LOG_INFO:
			err = syslog.Info(msg)
		case log.LOG_NOTICE:
			err = syslog.Notice(msg)
		case log.LOG_WARNING:
			err = syslog.Warning(msg)
		case log.LOG_DEBUG:
			err = syslog.Debug(msg)
		case LOG_TRACE:
			err = syslog.Debug(msg)
		default:
			// Default to ERROR, cause it's an error, right?
			err = syslog.Err(msg)
		}
		if err != nil {
			// Ummmm what to do here?
			reallog.Printf("Error returned by syslog: %v", err)
		}
	}
}

func (l *tsiLogger) writeToTextLogs(prio log.Priority, msg string) {
	msg += "\n"

	for _, textLog := range l.textlogs {
		io.WriteString(textLog, msg)
	}
}

/****
 * Utility helper functions
 */
func fmtArgs(format string, args []interface{}) []interface{} {
	lastWasPcnt := false
	var fmtParams int = 0
	for _, r := range format {
		if r == '%' {
			if !lastWasPcnt {
				fmtParams++
			} else {
				fmtParams--
			}
			lastWasPcnt = true
		} else {
			lastWasPcnt = false
		}
	}
	if fmtParams > len(args) {
		fmtParams = len(args)
	}
	return args[0:fmtParams]
}

func shaveSrcFile(fn string) string {
	idx := strings.LastIndex(fn, "src/prisma/tms")
	if idx < 0 {
		return fn
	}
	return fn[idx+len("src/prisma/tms/"):]
}

func logSite() (string, int) {
	skip := 1
	for {
		_, file, line, ok := runtime.Caller(skip)
		if !ok {
			break
		}
		file = shaveSrcFile(file)
		if file != "log/log.go" {
			return file, line
		}
		skip++
	}
	return "", -1
}

/****
 * Convenience functions
 */
func (l *tsiLogger) Audit(msgFmt string, args ...interface{}) {
	reallog.Printf("audit %v action object", args)
}
func (l *tsiLogger) Fatal(msgFmt string, args ...interface{}) {
	l.Log(log.LOG_CRIT, msgFmt, args...)
	os.Exit(1)
}
func (l *tsiLogger) Emerg(msgFmt string, args ...interface{}) {
	l.Log(log.LOG_EMERG, msgFmt, args...)
}
func (l *tsiLogger) Alert(msgFmt string, args ...interface{}) {
	l.Log(log.LOG_ALERT, msgFmt, args...)
}
func (l *tsiLogger) Crit(msgFmt string, args ...interface{}) {
	l.Log(log.LOG_CRIT, msgFmt, args...)
}
func (l *tsiLogger) Error(msgFmt string, args ...interface{}) {
	l.Log(log.LOG_ERR, msgFmt, args...)
}
func (l *tsiLogger) Warn(msgFmt string, args ...interface{}) {
	l.Log(log.LOG_WARNING, msgFmt, args...)
}
func (l *tsiLogger) Notice(msgFmt string, args ...interface{}) {
	l.Log(log.LOG_NOTICE, msgFmt, args...)
}
func (l *tsiLogger) Info(msgFmt string, args ...interface{}) {
	l.Log(log.LOG_INFO, msgFmt, args...)
}
func (l *tsiLogger) Debug(msgFmt string, args ...interface{}) {
	l.Log(log.LOG_DEBUG, msgFmt, args...)
}

/************
 *  DEFAULT logger interface
 */
func Audit(msg string) {
	reallog.Println(msg)
}
func Log(prio log.Priority, msgFmt string, args ...interface{}) {
	if dflt != nil {
		dflt.Log(prio, msgFmt, args...)
	}
}
func TraceMsg(msgFmt string, args ...interface{}) {
	if dflt != nil {
		dflt.TraceMsg(msgFmt, args...)
	}
}
func Trace(args ...interface{}) {
	if dflt != nil {
		dflt.Trace(args...)
	}
}

func Fatal(msgFmt string, args ...interface{}) {
	if dflt != nil {
		dflt.Fatal(msgFmt, args...)
	}
}
func Emerg(msgFmt string, args ...interface{}) {
	if dflt != nil {
		dflt.Emerg(msgFmt, args...)
	}
}
func Alert(msgFmt string, args ...interface{}) {
	if dflt != nil {
		dflt.Alert(msgFmt, args...)
	}
}
func Crit(msgFmt string, args ...interface{}) {
	if dflt != nil {
		dflt.Crit(msgFmt, args...)
	}
}
func Error(msgFmt string, args ...interface{}) {
	if dflt != nil {
		dflt.Error(msgFmt, args...)
	}
}
func Warn(msgFmt string, args ...interface{}) {
	if dflt != nil {
		dflt.Warn(msgFmt, args...)
	}
}
func Notice(msgFmt string, args ...interface{}) {
	if dflt != nil {
		dflt.Notice(msgFmt, args...)
	}
}
func Info(msgFmt string, args ...interface{}) {
	if dflt != nil {
		dflt.Info(msgFmt, args...)
	}
}
func Debug(msgFmt string, args ...interface{}) {
	if dflt != nil {
		dflt.Debug(msgFmt, args...)
	}
}

func getColoredNamedPriority(prio log.Priority) string {
	return colorPriority[prio] + getNameOfPriority(prio) + NC
}

func getNameOfPriority(prio log.Priority) string {
	switch prio {
	case log.LOG_ERR:
		return "ERROR"
	case log.LOG_ALERT:
		return "ALERT"
	case log.LOG_CRIT:
		return "CRITICAL"
	case log.LOG_EMERG:
		return "EMERGENCY"
	case log.LOG_INFO:
		return "INFO"
	case log.LOG_NOTICE:
		return "NOTICE"
	case log.LOG_WARNING:
		return "WARNING"
	case log.LOG_DEBUG:
		return "DEBUG"
	case LOG_TRACE:
		return "TRACE"
	default:
		return "UNKNOWN"
	}
}
