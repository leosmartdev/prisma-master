package tms

import (
	"time"

	pbts "github.com/golang/protobuf/ptypes/timestamp"
)

func ToTimestamp(t time.Time) *pbts.Timestamp {
	return &pbts.Timestamp{
		Seconds: t.Unix(),
		Nanos:   int32(t.UnixNano() % 1000000000),
	}
}

func FromTimestamp(t *pbts.Timestamp) time.Time {
	if t == nil {
		return time.Time{}
	}
	return time.Unix(t.Seconds, int64(t.Nanos))
}

// Return the number of milliseconds since Jan 1, 1970
func ToMilli(t time.Time) int64 {
	return (int64(t.Unix()) * 1000) + (int64(t.Nanosecond()) / 1000000)
}

// Return the time given the number of milliseconds in Jan 1, 1970
func FromMilli(i int64) time.Time {
	return time.Unix(i/1000, (i%1000)*1000000)
}

func Now() *pbts.Timestamp {
	return ToTimestamp(time.Now())
}
