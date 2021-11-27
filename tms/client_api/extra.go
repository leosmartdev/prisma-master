package client_api

import (
	. "prisma/tms"
	"time"
)

func (s Status) MarshalJSON() ([]byte, error) {
	str, ok := Status_name[int32(s)]
	if !ok {
		str = string(s)
	}
	str = "\"" + str + "\""
	return []byte(str), nil
}

/*
func (upd TrackUpdate) GetStatus() Status {
	return upd.Status
}
*/

func (upd TrackUpdate) Time() (time.Time, bool) {
	t := time.Time{}
	if upd.Track != nil {
		for _, tgt := range upd.Track.Targets {
			tt := FromTimestamp(tgt.Time)
			if tt.After(t) {
				t = tt
			}
		}
		for _, md := range upd.Track.Metadata {
			tt := FromTimestamp(md.Time)
			if tt.After(t) {
				t = tt
			}
		}
	}

	return time.Time{}, false
}

type TrackSlice []*Track

func (t TrackSlice) Len() int {
	return len(t)
}

func (t TrackSlice) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

type TracksByName struct {
	TrackSlice
}

func (t TracksByName) Less(i, j int) bool {
	return t.TrackSlice[i].Metadata[0].Name < t.TrackSlice[j].Metadata[0].Name
}
