package db

import (
	. "prisma/tms"
	. "prisma/tms/client_api"
	"prisma/tms/log"

	"errors"

	"github.com/StefanSchroeder/Golang-Ellipsoid/ellipsoid"
)

var (
	Globe = ellipsoid.Init(
		"WGS84",
		ellipsoid.Degrees,
		ellipsoid.Nm,
		ellipsoid.LongitudeIsSymmetric,
		ellipsoid.BearingIsSymmetric)
)

/**
 * Filter a track stream based on geographic zones. When a track is inside of a
 * zone then leaves it, send a LeftGeoZone status to inform the client.
 */
type GeoFilter struct {
	req   *GoTrackRequest
	geodb *GeoDB
}

// Set up a list of GeoRanges from a requested filter
func NewGeoFilter(req *GoTrackRequest) (*GeoFilter, error) {
	gf := &GeoFilter{}
	gf.req = req

	f := req.Req.Filter
	if f != nil {
		switch x := f.(type) {
		case *TrackRequest_FilterSimple:
			gf.geodb = &GeoDB{
				inRange: make(map[string]bool),
				ranges:  make([]GeoRange, 0),
			}
			if x.FilterSimple != nil {
				for _, cr := range x.FilterSimple.CircularRange {
					if cr.Center == nil {
						return nil, errors.New("Must specify r.Center!")
					}
					gf.geodb.ranges = append(gf.geodb.ranges, &CircularGeoRange{
						*cr,
					})
				}

				for _, lr := range x.FilterSimple.LinearRange {
					gf.geodb.ranges = append(gf.geodb.ranges, &LinearGeoRange{
						*lr,
					})
				}
			}

		default:
			log.Error("Could not decipher filter: %v", f, req)
			return nil, UnknownOption
		}
	}
	return gf, nil
}

// If there's any geo ranges specified, start a geo filtering thread
func (gf *GeoFilter) Start(tracks <-chan TrackUpdate) (<-chan TrackUpdate, error) {
	if gf.geodb != nil && len(gf.geodb.ranges) > 0 {
		// Only do something if we have some ranges specified
		out := make(chan TrackUpdate, 128)
		gf.req.Ctxt.Go(func() {
			// The geo filtering thread
			defer close(out)
			for {
				select {
				case <-gf.req.Ctxt.Done():
					return
				case update, ok := <-tracks:
					if !ok {
						return
					}
					update = gf.geodb.Eval(update)
					if update.Status != Status_Unknown {
						out <- update
					}
				}
			}
		})
		return out, nil
	}
	return tracks, nil
}

/**
 * GeoDB has a list of allow GeoRanges and a set of trackIDs which are
 * currently inside those ranges. This is kept separate from the GeoFilter
 * struct above so that other code can use it.
 */
type GeoDB struct {
	inRange map[string]bool
	ranges  []GeoRange
}

// Update a track update and our current list of in range tracks depending on
// whether or not it's in range
func (db *GeoDB) Eval(update TrackUpdate) TrackUpdate {
	if update.Status == Status_Current {
		// This update is telling is about a new track position
		id := update.Track.Id
		inrange := db.IsInRange(update.Track) // Is the new position in our range?
		_, ok := db.inRange[id]
		if !ok {
			// New sighting!
			if inrange {
				db.inRange[id] = true
				return update
			} else {
				// But it's out of range. Ditch this guy
				return TrackUpdate{
					Status: Status_Unknown,
				}
			}
		} else {
			// Old sighting
			if inrange {
				// And continues to be in range -- pass along!
				return update
			} else {
				// Target has left range. Inform client!
				delete(db.inRange, id)
				return TrackUpdate{
					Status: Status_LeftGeoRange,
					Track:  update.Track,
				}
			}
		}
	} else if update.Status == Status_Timeout {
		id := update.Track.Id
		_, ok := db.inRange[id]
		if !ok {
			// We are not tracking this guy. Ditch the timeout
			return TrackUpdate{
				Status: Status_Unknown,
			}
		}
	}
	return update
}

// Is this track position within our geographic range
func (db *GeoDB) IsInRange(track *Track) bool {
	for _, r := range db.ranges {
		if r.IsInRange(track) {
			return true
		}
	}
	return false
}

// Something that can determine if a track is within a range
type GeoRange interface {
	IsInRange(*Track) bool
}

// A circular geo range
type CircularGeoRange struct {
	CircularRange
}

func (r *CircularGeoRange) IsInRange(track *Track) bool {
	if len(track.Targets) == 0 {
		return false
	}
	tgt := track.Targets[0]
	if tgt.Position == nil {
		return false
	}
	pos := tgt.Position

	if r.Center == nil {
		log.Warn("IsInRange'ing on nil center!")
		return false
	}
	dist, _ := Globe.To(r.Center.Latitude, r.Center.Longitude,
		pos.Latitude, pos.Longitude)
	log.TraceMsg("Circ eval -- pos: %v, center %v, dist: %v",
		pos, r.Center, dist)
	return dist <= r.Radius
}

// A linear geo range
type LinearGeoRange struct {
	LinearRange
}

func (r *LinearGeoRange) IsInRange(track *Track) bool {
	if len(track.Targets) == 0 {
		return false
	}
	tgt := track.Targets[0]
	if tgt.Position == nil {
		return false
	}
	pos := tgt.Position
	ret := (r.MinLatitude == nil || pos.Latitude >= r.MinLatitude.Value) &&
		(r.MaxLatitude == nil || pos.Latitude <= r.MaxLatitude.Value) &&
		(r.MinLongitude == nil || pos.Longitude >= r.MinLongitude.Value) &&
		(r.MaxLongitude == nil || pos.Longitude <= r.MaxLongitude.Value)
	log.TraceMsg("MatchesLinear: %v, %v, %v", pos, r, ret)
	return ret
}
