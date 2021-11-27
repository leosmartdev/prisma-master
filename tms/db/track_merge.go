package db

import (
	"errors"
	"fmt"
	"prisma/gogroup"
	. "prisma/tms"
	. "prisma/tms/client_api"
	"prisma/tms/devices"
	"prisma/tms/log"
	. "prisma/tms/nmea"
	"time"

	"github.com/davecgh/go-spew/spew"
)

const (
	StatusInterval = time.Duration(1) * time.Second
)

var (
	canceled   = errors.New("canceled")
	mergeStats = log.GetTracer("merge-stats")
)

type MergeOptions struct {
	Mode       MergeMode
	Historical bool
}

func DefaultMergeOptions() MergeOptions {
	return MergeOptions{
		Mode:       MergeMode_None, // FIXME TrackID,
		Historical: false,
	}
}

/**
 * This stage is responsible for taking any number of target/metadata/track
 * streams and merging them into a cogent track stream. It can/will convert
 * TrackIDs into universal IDs (e.g. removes the site/sensor id from things
 * with MMSIs) and thus merges AIS data from multiple sources. It can also
 * merge data from multiple track streams. If a history was requested, it is
 * responsible for building the list of targets and list of metadata and
 * properly ordering said lists. If no history is specified, it is responsible
 * for keeping only the latest data and ensuring that old data isn't sent down
 * the pipeline. Also allows for special cases like multiple metadata items of
 * different types (e.g. different AIS static reports) to be included instead
 * of overridden. Also includes the option to batch all data until the backlog
 * is complete.
 */
type TrackMerger struct {
	opts    MergeOptions
	ctxt    gogroup.GoGroup
	Disable bool
	backlog chan TrackUpdate // Updates coming from streams in backlog mode
	live    chan TrackUpdate // Updates coming from streams in live mode
	req     *GoTrackRequest

	backlogCount  int  // How many streams are in backlog mode?
	liveCount     int  // How many streams are in live mode?
	setupFinished bool // Have all the streams which are going to join joined?
	streamOut     bool // Should we start streaming out results?

	ch     chan TrackUpdate  // Output track updates
	tracks map[string]*Track // In memory database of tracks

	idmap func(string) string // Function to make trackIDs. Used to create universal mappings?

	// Statistics
	count uint64
	total uint64
}

// track ID merge which doesn't merge different tracks
func simpleMerge(t string) string {
	return t
}

// Set up this TrackMerger for a particular request
//func (m *TrackMerger) Configure(req *GoTrackRequest) error {

func NewTrackMerger(ctxt gogroup.GoGroup, opts MergeOptions) *TrackMerger {
	idmap := simpleMerge
	disable := false
	switch opts.Mode {
	case MergeMode_None:
		disable = true
	case MergeMode_TrackID:
		idmap = simpleMerge
	default:
		panic(fmt.Sprintf("unknown merge mode: %v", opts.Mode))
	}

	spewConfig := spew.ConfigState{
		Indent:         "  ",
		SortKeys:       true,
		MaxDepth:       3,
		DisableMethods: true,
	}
	log.TraceMsg("Configuring trackMerger for req: %v", spewConfig.Sdump(opts))

	m := &TrackMerger{
		Disable: disable,
		ctxt:    ctxt,
		opts:    opts,
		// We require zero-buffered channels to ensure message ordering
		backlog: make(chan TrackUpdate, 0),
		live:    make(chan TrackUpdate, 0),

		backlogCount:  0,
		liveCount:     0,
		setupFinished: false,
		streamOut:     !opts.Historical,

		ch:     make(chan TrackUpdate, 128),
		tracks: make(map[string]*Track),

		idmap: idmap,

		count: 0,
		total: 0,
	}

	return m
}

// When running as a pipeline stage, just use this one track update stream
func (m *TrackMerger) Start(inputch <-chan TrackUpdate) (<-chan TrackUpdate, error) {
	m.Service()
	ok := m.AddStream(inputch, "tracks")
	if !ok {
		return nil, canceled
	}
	ok = m.SetupFinished()
	if !ok {
		return nil, canceled
	}
	return m.ch, nil
}

// Tell the merge thread that it's done being set up
func (m *TrackMerger) SetupFinished() bool {
	select {
	case <-m.ctxt.Done():
		return false
	case m.backlog <- TrackUpdate{Status: Status_SetupDone}:
	}
	return true
}

// Start feeding this merger with this update stream
func (m *TrackMerger) AddStream(ch <-chan TrackUpdate, name string) bool {
	// Tell the merger thread that we're starting a stream in backlog mode
	select {
	case <-m.ctxt.Done():
		return false
	case m.backlog <- TrackUpdate{Status: Status_Starting}:
	}

	m.ctxt.Go(func() {
		doBacklog := true
		sends := 0
		for doBacklog {
			// As long as this stream is in backlog mode, feed data down the backlog channel
			select {
			case <-m.ctxt.Done():
				return
			case obj, ok := <-ch:
				if !ok || obj.Status == Status_Closing {
					// Stream is closing, tell the merge thread, then die
					m.backlog <- TrackUpdate{Status: Status_Closing}
					return
				}
				if obj.Status == Status_InitialLoadDone {
					m.ch <- obj
					// Stream is transitioning from backlog mode to live mode. Tell the merger thread
					m.live <- TrackUpdate{Status: Status_Starting}
					log.TraceMsg("Switching stream '%v' from backlog to live after %v", name, sends)
					doBacklog = false // Stop reading in backlog
				}
				m.backlog <- obj
				sends += 1
			}
		}

		for {
			// In live mode, pump data down the live channel until the stream closes
			select {
			case <-m.ctxt.Done():
				return
			case obj, ok := <-ch:
				if !ok || obj.Status == Status_Closing {
					log.TraceMsg("Closing stream '%v' after %v.", name, sends)
					m.live <- TrackUpdate{Status: Status_Closing}
					return
				}
				m.live <- obj
				sends += 1
			}
		}
	})
	return true
}

// Run a thread which reads from the backlog and live channels and processes
// their data
func (m *TrackMerger) Service() {
	m.ctxt.Go(func() {
		defer close(m.ch)
		m.count = 0
		m.total = 0
		statusTicker := time.NewTicker(StatusInterval)
		defer statusTicker.Stop()
		var die bool = false
		for !die && (!m.setupFinished ||
			m.backlogCount > 0 ||
			m.liveCount > 0) {
			// As long as we're running normally, drain data from the input channels
			die = m.drain(statusTicker)
		}

		log.TraceMsg("Closing down merge service. %v tracks served so far.", m.total)

		// Close the channels for sending
		close(m.backlog)
		close(m.live)

		log.TraceMsg("Draining final contents...")

		// Drain their contents
		for obj := range m.backlog {
			m.object(obj, "backlog drain")
		}
		for {
			select {
			case <-m.ctxt.Done():
				return
			case obj, ok := <-m.live:
				if !ok {
					log.Warn("closed channel")
					return
				}
				m.object(obj, "live drain")
			}
		}

		log.TraceMsg("Exiting merge service. %v tracks served upon death.", m.total)
	})
}

// Read data from the input streams and process it
func (m *TrackMerger) drain(statusTicker *time.Ticker) bool {
	select {
	case <-m.ctxt.Done():
		return true
	case <-statusTicker.C:
		mergeStats.Logf("Tracks: %v/sec, %v total",
			float64(m.count)/StatusInterval.Seconds(), m.total)
		m.count = 0

	case obj := <-m.backlog:
		// We got something from the backlog stream
		switch obj.Status {
		case Status_Starting:
			m.backlogCount += 1
		case Status_Closing, Status_InitialLoadDone:
			m.backlogCount -= 1
		case Status_SetupDone:
			m.setupFinished = true
		default:
			m.object(obj, "backlog")
		}
		if obj.Status != Status_Current {
			log.TraceMsg("Got status update: %v", obj.Status)
		}

	case obj := <-m.live:
		// We got something from the live stream
		switch obj.Status {
		case Status_Starting:
			m.liveCount += 1
		case Status_Closing, Status_InitialLoadDone:
			m.liveCount -= 1
		case Status_SetupDone:
			m.setupFinished = true
		default:
			m.object(obj, "live")
		}
		if obj.Status != Status_Current {
			log.TraceMsg("Got status update: %v", obj.Status)
		}
	}

	if m.streamOut == false && m.setupFinished && m.backlogCount == 0 {
		// All backlogs have completed, time to start sending data
		log.TraceMsg("Turning on stream out")
		m.streamOut = true
		m.sendAll()
	}
	return false
}

// Process an object 'obj' from stream 'src'
func (m *TrackMerger) object(obj TrackUpdate, src string) {
	if obj.Track != nil {
		if m.Disable {
			m.track(obj.Track)
		} else {
			// *** Get the existing track in memory, create if necessary
			memTrack := m.getTrack(obj.Track.Id)

			// *** Update the track in memory
			send := false
			for _, tgt := range obj.Track.Targets {
				if m.insertTarget(tgt, memTrack) {
					send = true
				}
			}
			for _, md := range obj.Track.Metadata {
				if m.insertMetadata(md, memTrack) {
					send = true
				}
			}
			// *** Send the updated track to the client if it's actually been
			// updated
			if send {
				m.track(memTrack)
			}
		}
	}
}

// Get the existing track in memory, create if necessary
func (m *TrackMerger) getTrack(id string) *Track {
	if id == "" {
		log.Warn("Object didn't have a track_id!")
		return nil
	}

	// Merged (mapped) ID
	mid := m.idmap(id)
	// Lookup ID
	lid := mid
	track, ok := m.tracks[lid]
	if !ok {
		// Doesn't exist. Create one
		n := &Track{
			Id:       mid,
			Targets:  make([]*Target, 0, 8),
			Metadata: make([]*TrackMetadata, 0, 8),
		}
		m.tracks[lid] = n
		return n
	}
	return track
}

// Should 'n' replace 'o'?
func (m *TrackMerger) targetReplaces(n *Target, o *Target) bool {
	// FIXME: This probably should be removed
	return true
	//return n.TrackId == o.TrackId
}

// Update track 'track' with potentially new target information 'tgt'
func (m *TrackMerger) insertTarget(tgt *Target, track *Track) bool {
	// Find the correct insert position
	ipos := -1
	// If someone is replaced by me, what position is it?
	rpos := -1

	// Iterate through the existing targets and figure out where to put the new
	// target and what existing target it should replace, if any
	for i := 0; i < len(track.Targets); i++ {
		itgt := track.Targets[i]

		if ipos < 0 {
			if (tgt.Type == devices.DeviceType_Fusion &&
				itgt.Type != devices.DeviceType_Fusion) ||
				tgt.Time.Seconds > itgt.Time.Seconds ||
				(tgt.Time.Seconds == itgt.Time.Seconds &&
					tgt.Time.Nanos >= itgt.Time.Nanos) {
				ipos = i
			}
		}

		if rpos == -1 && m.targetReplaces(tgt, itgt) {
			rpos = i
		}
	}

	if m.opts.Historical {
		// Don't replace anything if the client wants histories
		rpos = -1
	}

	if rpos >= 0 {
		// Check to see if the current data is newer
		rpTime := FromTimestamp(track.Targets[rpos].Time)
		tgtTime := FromTimestamp(tgt.Time)
		if !tgtTime.After(rpTime) {
			// Trying to replace something newer? Bad call, yo
			return false
		}
	}

	if ipos == -1 {
		// By default, prepend the data
		ipos = 0
	}

	// OK, so now we have to insert 'tgt' at 'ipos' and if 'rpos' was filled,
	// we have to remove it. We can obviously do this in two separate steps,
	// but if both are specified, it's often more efficient to move things
	// around a bit and use the space from 'rpos' for tgt at 'ipos'. I wouldn't
	// ordinarily write this sort of code, but I think the code in this merge
	// may be performance sensitive and this ain't _that_ bad.

	if rpos == -1 {
		// Adding the data only, nothing to replace
		track.Targets = append(track.Targets, nil)
		copy(track.Targets[ipos+1:], track.Targets[ipos:])
		track.Targets[ipos] = tgt
	} else if ipos == rpos {
		// Excellent! Simple replacement
		track.Targets[ipos] = tgt
	} else if rpos < ipos {
		// We can shift elements to the left, overwriting rpos
		copy(track.Targets[rpos:], track.Targets[rpos+1:ipos])
		track.Targets[ipos] = tgt
	} else if rpos > ipos {
		// We can shift elements to the right, overwriting rpos
		copy(track.Targets[ipos+1:], track.Targets[ipos:rpos])
		track.Targets[ipos] = tgt
	} else {
		panic("Internal error: this case should be a logical impossibility")
	}

	return true
}

// If it exists, get M1371 data from a metadata
func getM1371(meta *TrackMetadata) *M1371 {
	if meta.Nmea == nil {
		return nil
	}
	if meta.Nmea.Vdm != nil &&
		meta.Nmea.Vdm.M1371 != nil {
		return meta.Nmea.Vdm.M1371
	}
	if meta.Nmea.Vdo != nil &&
		meta.Nmea.Vdo.M1371 != nil {
		return meta.Nmea.Vdo.M1371
	}
	return nil
}

// Should metadata 'n' replace metadata 'o' in the list of metadata?
func (m *TrackMerger) metadataReplaces(n *TrackMetadata, o *TrackMetadata) bool {
	// FIXME: This probably should be removed
	return true

	/*
		if n.TrackId != o.TrackId {
			return false
		}
		// If we get here, they have the same TrackID
		if n.Type == devices.DeviceType_AIS && o.Type == devices.DeviceType_AIS {
			// Special case for AIS...
			nm := getM1371(n)
			om := getM1371(o)
			// If they both have different report types, don't replace
			if nm != nil && om != nil &&
				nm.MessageId != om.MessageId {
				return false
			}
		}
		// Usually they replace each other if they have the same trackid
		return true
	*/
}

// Update the metadata list in 'track' with the potentially new metadata
// 'meta'. Yes, this function is very, very similar to 'insertTarget' above.
// Consider making a more generic function which can handle both in the future.
// Unfortunately, it would be a bit ugly since Go doesn't have generics or
// metaprogramming of any sort.
func (m *TrackMerger) insertMetadata(meta *TrackMetadata, track *Track) bool {
	// Find the correct insert position
	ipos := -1
	// If someone is replaced by me, what position is it?
	rpos := -1

	// Iterate through the existing metadata and figure out where to put the
	// new metadata and what existing metadata it should replace, if any
	for i := 0; i < len(track.Metadata); i++ {
		imeta := track.Metadata[i]

		if ipos < 0 {
			if (meta.Type == devices.DeviceType_Fusion &&
				imeta.Type != devices.DeviceType_Fusion) ||
				meta.Time.Seconds > imeta.Time.Seconds ||
				(meta.Time.Seconds == imeta.Time.Seconds &&
					meta.Time.Nanos >= imeta.Time.Nanos) {
				ipos = i
			}
		}

		if rpos == -1 && m.metadataReplaces(meta, imeta) {
			rpos = i
		}
	}

	if m.opts.Historical {
		// Don't replace anything if the client wants histories
		rpos = -1
	}

	if rpos >= 0 {
		// Check to see if the current data is newer
		rpTime := FromTimestamp(track.Metadata[rpos].Time)
		mdTime := FromTimestamp(meta.Time)
		if !mdTime.After(rpTime) {
			// Trying to replace something newer? Bad call, yo
			return false
		}
	}

	if ipos == -1 {
		// By default, prepend the data
		ipos = 0
	}

	// OK, so now we have to insert 'meta' at 'ipos' and if 'rpos' was filled,
	// we have to remove it. We can obviously do this in two separate steps,
	// but if both are specified, it's often more efficient to move things
	// around a bit and use the space from 'rpos' for meta at 'ipos'. I wouldn't
	// ordinarily write this sort of code, but I think the code in this merge
	// may be performance sensitive and this ain't _that_ bad.

	if rpos == -1 {
		// Adding the data only, nothing to replace
		track.Metadata = append(track.Metadata, nil)
		copy(track.Metadata[ipos+1:], track.Metadata[ipos:])
		track.Metadata[ipos] = meta
	} else if ipos == rpos {
		// Excellent! Simple replacement
		track.Metadata[ipos] = meta
	} else if rpos < ipos {
		// We can shift elements to the left, overwriting rpos
		copy(track.Metadata[rpos:], track.Metadata[rpos+1:ipos])
		track.Metadata[ipos] = meta
	} else if rpos > ipos {
		// We can shift elements to the right, overwriting rpos
		copy(track.Metadata[ipos+1:], track.Metadata[ipos:rpos])
		track.Metadata[ipos] = meta
	} else {
		panic("Internal error: this case should be a logical impossibility")
	}

	return true
}

// Send everything currently in memory to the client
func (m *TrackMerger) sendAll() {
	if !m.Disable {
		for _, track := range m.tracks {
			m.ch <- TrackUpdate{
				Status: Status_Current,
				Track:  track,
			}
			m.count++
			m.total++
		}
		log.TraceMsg("Total: %v after sendAll()", m.total)
	}
	m.ch <- TrackUpdate{
		Status: Status_InitialLoadDone,
		Track:  nil,
	}
}

// Send a track to the client
func (m *TrackMerger) track(track *Track) {
	if m.Disable {
		id := track.Id
		last, ok := m.tracks[id]
		if ok && track.Time().Before(last.Time()) {
			// Trying to send old data? Bad call, bro
			return
		}
		m.tracks[id] = track
		m.ch <- TrackUpdate{
			Status: Status_Current,
			Track:  track,
		}
		m.count++
		m.total++
	} else if len(track.Targets) > 0 && track != nil && m.streamOut {
		update := TrackUpdate{
			Status: Status_Current,
			Track:  track,
		}
		m.ch <- update
		m.count++
		m.total++
	}
}
