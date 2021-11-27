package log

import (
	"bytes"
	"log/syslog"
)

type SyslogWriter struct {
	Priority syslog.Priority
	buf      bytes.Buffer
}

func NewSyslogWriter() *SyslogWriter {
	w := SyslogWriter{
		Priority: syslog.LOG_INFO,
	}
	return &w
}

func (w *SyslogWriter) Write(p []byte) (n int, err error) {
	for _, b := range p {
		w.buf.WriteByte(b)
		if b == '\n' {
			Log(w.Priority, w.buf.String())
			w.buf.Reset()
		}
	}
	return len(p), nil
}
