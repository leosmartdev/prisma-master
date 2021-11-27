package db

import (
	"expvar"
	"io"
	api "prisma/tms/client_api"
	"prisma/tms/feature"
	"prisma/tms/libmain"
	"prisma/tms/log"
	"prisma/tms/util/coordsys"
	"reflect"

	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"prisma/gogroup"

	"prisma/tms/geojson"
	coll "prisma/tms/geojson/collection"
)

const (
	FeatureViewTimeout = time.Duration(15) * time.Second
	SearchCutoff       = 0.8 // Arbitrary
)

var (
	sessions = expvar.NewInt("sessions")
)

type features struct {
	ctxt    gogroup.GoGroup
	tracks  TrackDB
	misc    MiscDB
	sites   SiteDB
	devices DeviceDB

	lock  sync.RWMutex
	views map[string]FeaturesView
}

func NewFeatures(ctxt gogroup.GoGroup, tracks TrackDB, misc MiscDB, sites SiteDB, devices DeviceDB) FeatureDB {
	return &features{
		ctxt:    ctxt,
		tracks:  tracks,
		misc:    misc,
		sites:   sites,
		devices: devices,
		views:   make(map[string]FeaturesView),
	}
}

func (f *features) Version() string {
	return fmt.Sprintf("Version: %v, build date: %v",
		libmain.VersionNumber,
		libmain.VersionDate)
}

func (f *features) CreateView(req *api.ViewRequest) (FeaturesView, error) {
	id := f.reserveID(req.Id)
	v := &featuresView{
		ctxt:    f.ctxt.Child(req.Id),
		id:      id,
		req:     req,
		parent:  f,
		tracks:  f.tracks,
		devices: f.devices,
		misc:    f.misc,
		sites:   f.sites,
		trace:   log.GetTracer("view"),
	}

	v.trace.Logf("created %v", v.id)
	sessions.Add(1)
	f.lock.Lock()
	f.views[id] = v
	f.lock.Unlock()

	err := v.Service()
	if err != nil {
		v.Destroy()
		return nil, err
	}
	return v, nil
}

func (f *features) Stream(ctxt gogroup.GoGroup, reqs <-chan api.StreamRequest) (<-chan FeatureUpdate, error) {
	req, ok := <-reqs
	if !ok {
		return nil, errors.New("didn't get initial request")
	}

	v, ok := f.views[req.ViewId]
	if !ok {
		return nil, errors.New("could not locate view")
	}
	subReqs := make(chan api.StreamRequest)
	log.TraceMsg("Features starting Stream()")
	ch, err := v.Stream(ctxt, subReqs)
	if err != nil {
		return nil, err
	}
	ctxt.Go(func() {
		subReqs <- req
		defer close(subReqs)
		for {
			select {
			case <-ctxt.Done():
				return
			case req := <-reqs:
				select {
				case <-ctxt.Done():
					return
				case subReqs <- req:
				}
			}
		}
	})
	return ch, nil
}

func (f *features) Snapshot(req GoStreamRequest) ([]*feature.F, error) {
	v, ok := f.views[req.Req.ViewId]
	if !ok {
		return nil, errors.New("Could not locate view")
	}
	return v.Snapshot(req)
}

func (f *features) Details(req GoDetailRequest) (<-chan GoFeatureDetail, error) {
	v, ok := f.views[req.Req.ViewId]
	if !ok {
		return nil, errors.New("Could not locate view")
	}
	return v.Details(req)
}

func (f *features) GetHistoricalTrack(req GoHistoricalTrackRequest) (*feature.F, error) {
	log.Debug("This is trying to get some tracks ? %+v", req)
	t, err := f.tracks.GetHistoricalTrack(req)
	if err != nil {
		return nil, err
	}
	full := true
	useTrackID := true
	return feature.FromTrack(t, full, useTrackID), nil
}

func (f *features) Search(req GoFeatureSearchRequest) (<-chan *feature.F, error) {
	v, ok := f.views[req.Req.ViewId]
	if !ok {
		return nil, errors.New("Could not locate view")
	}
	return v.Search(req)
}

func (f *features) reserveID(requestedID string) string {
	f.lock.Lock()
	defer f.lock.Unlock()

	if requestedID != "" {
		if _, ok := f.views[requestedID]; !ok {
			// Requested ID is not taken
			f.views[requestedID] = nil
			return requestedID
		}
	}

	// Gotta start adding numbers until we find one not taken!
	i := 0
	requestedID = "unnamed"
	for {
		i++
		id := fmt.Sprintf("%v%v", requestedID, i)
		if _, ok := f.views[id]; !ok {
			f.views[id] = nil
			return id
		}
	}
}

type featuresView struct {
	sync.Mutex

	ctxt    gogroup.GoGroup
	id      string
	req     *api.ViewRequest
	parent  *features
	tracks  TrackDB
	misc    MiscDB
	devices DeviceDB
	sites   SiteDB

	dead         bool
	listeners    sync.WaitGroup
	numListeners int
	lastTouch    time.Time

	provwg   sync.WaitGroup
	updates  chan FeatureUpdate
	features *coll.Collection

	providers []FeaturesProvider
	streams   map[*FeatureStream]struct{}

	crs               coordsys.C
	startHeatmapCount int
	stopHeatmapCount  int

	trace *log.Tracer
}

func (v *featuresView) MarshalJSON() ([]byte, error) {
	js := fmt.Sprintf("{\"viewId\": \"%v\"}", v.id)
	return []byte(js), nil
}

func (v *featuresView) Setup() error {
	v.streams = make(map[*FeatureStream]struct{})
	v.updates = make(chan FeatureUpdate, 128)
	v.features = coll.New()

	switch v.req.Projection {
	case api.ProjectionType_UnknownProjection:
		v.crs = coordsys.EPSG4326
	case api.ProjectionType_WebMercator:
		v.crs = coordsys.EPSG3857
	}

	if v.req.StartHeatmapCount > 0 {
		v.startHeatmapCount = int(v.req.StartHeatmapCount)
	} else {
		v.startHeatmapCount = DefaultStartHeatmapCount
	}
	if v.req.StopHeatmapCount > 0 {
		v.stopHeatmapCount = int(v.req.StopHeatmapCount)
	} else {
		v.stopHeatmapCount = DefaultStopHeatmapCount
	}

	providers := GetProviders(v.req.Types)
	for _, p := range providers {
		err := v.addProvider(p)
		if err != nil {
			return err
		}
	}

	return nil
}

func (v *featuresView) Service() error {
	log.Debug("Starting view: %v", v.id)

	err := v.Setup()
	if err != nil {
		return err
	}

	for _, prov := range v.providers {
		v.provwg.Add(1)
		v.ctxt.Go(
			func(p FeaturesProvider) {
				err := p.Service(v.updates)
				v.provwg.Done()
				if err != nil {
					log.Error("Error from feature provider: %v", err)
				}
			},
			prov)
	}
	v.lastTouch = time.Now()
	v.ctxt.Go(v.timeoutView)
	v.ctxt.Go(v.processUpdates)
	v.ctxt.Go(func() {
		// Close the channel when all the providers exit
		v.provwg.Wait()
		close(v.updates)
	})
	return nil
}

func (v *featuresView) processUpdates() {
	for {
		select {
		case <-v.ctxt.Done():
			return
		case upd, ok := <-v.updates:
			if !ok {
				return
			}
			// FIXME: If this feature isn't valid, it probably is a SARSAT unlocated
			// alert which cannot be mapped anyway. Drop it. A feature without
			// a geometry but with valid properites would be preferred here.
			if upd.Feature == nil || upd.Feature.Geometry == nil {
				continue
			}
			switch upd.Status {
			case api.Status_Timeout, api.Status_LeftGeoRange:
				if upd.Feature != nil {
					v.features.Remove(upd.Feature.ID.(string))
				}
			case api.Status_Current:
				if upd.Feature != nil {
					v.features.ReplaceOrInsert(
						upd.Feature.ID.(string),
						*upd.Feature,
						nil, nil)
				}
			}

			v.Lock()
			for s := range v.streams {
				if s.input != nil {
					s.input <- upd
				}
			}
			v.Unlock()
		}
	}
}

func (v *featuresView) timeoutView() {
	for {
		select {
		case <-v.ctxt.Done():
		default:
		}
		// Wait for zero listeners
		v.listeners.Wait()
		log.Debug("View has no listeners! Checking to see if we should kill it.")

		v.Lock()
		last := v.lastTouch
		timeout := last.Add(FeatureViewTimeout)
		if v.numListeners == 0 && timeout.Before(time.Now()) {
			log.Debug("View '%v' has had no listeners for %v. Destroying...",
				v.id, time.Since(last))
			v.Destroy()
			v.Unlock()
			return
		}
		log.Debug("Nope, haven't hit the timeout yet. Sleeping for a while.", v.id)
		v.Unlock()

		// Sleep until timeout
		time.Sleep(timeout.Sub(time.Now()))
	}
}

func (v *featuresView) addProvider(p FeaturesProvider) error {
	err := p.Init(v.req, v.ctxt, v.tracks, v.misc, v.sites, v.devices)
	if err != nil {
		return err
	}
	v.providers = append(v.providers, p)
	return nil
}

func (v *featuresView) Destroy() {
	v.trace.Logf("removing %v", v.id)
	sessions.Add(-1)
	v.dead = true

	// Remove myself from parent
	v.parent.lock.Lock()
	delete(v.parent.views, v.id)
	v.parent.lock.Unlock()

	// Destroy my children providers
	for _, prov := range v.providers {
		prov.Destroy()
	}
}

func (v *featuresView) ID() string {
	return v.id
}

func (v *featuresView) addListener() {
	v.Lock()
	v.listeners.Add(1)
	v.numListeners += 1
	v.lastTouch = time.Now()
	v.Unlock()
	v.trace.Logf("added listener, total: %v", v.numListeners)
}

func (v *featuresView) remListener() {
	v.Lock()
	v.lastTouch = time.Now()
	v.numListeners -= 1
	v.listeners.Done()
	v.Unlock()
	v.trace.Logf("removed listener, total: %v", v.numListeners)
}

// This function originally converted from a geo.Feature to a GoFeature,
// but that is no longer necessary. Function still around for reprojection.
func (v *featuresView) toFeature(obj geojson.Object) *feature.F {
	if obj == nil {
		return nil
	}

	gf, ok := obj.(feature.F)
	if !ok {
		panic(fmt.Sprintf("Unable to convert, type was %v", reflect.TypeOf(obj)))
	}
	gf.BBox = nil
	feat := feature.New(gf.ID, gf.Geometry, gf.Properties).FromWGS84(v.crs)
	return feat
}

func (v *featuresView) Stream(ctxt gogroup.GoGroup, reqs <-chan api.StreamRequest) (<-chan FeatureUpdate, error) {
	v.addListener()
	if v.dead {
		return nil, errors.New("View is dead!")
	}

	ch := make(chan FeatureUpdate, 512)
	stream := &FeatureStream{
		view:       v,
		ch:         ch,
		input:      make(chan FeatureUpdate, 64),
		liveStream: true,
	}

	log.Debug("View streaming requests")

	ctxt.Go(func() {
		v.Lock()
		v.streams[stream] = struct{}{}
		v.Unlock()

		stream.service(ctxt, reqs)

		v.Lock()
		delete(v.streams, stream)
		v.Unlock()

		close(ch)
		v.remListener()
	})

	return ch, nil
}

func (v *featuresView) Snapshot(GoStreamRequest) ([]*feature.F, error) {
	return nil, errors.New("Not implemented")
}

func (v *featuresView) Details(req GoDetailRequest) (<-chan GoFeatureDetail, error) {
	// We have to command our children providers to get details then
	// aggregate their output streams until they are all closed

	v.addListener()
	if v.dead {
		return nil, errors.New("View is dead!")
	}

	req.Stream = true

	var deets <-chan GoFeatureDetail
	for _, prov := range v.providers {
		s, err := prov.DetailsStream(req)
		if err != nil {
			req.Ctxt.Cancel(err)
			return nil, err
		}

		if s != nil {
			if deets == nil {
				deets = s
			} else {
				err := errors.New("Multiple providers trying to send details!")
				req.Ctxt.Cancel(err)
				return nil, err
			}
		}
	}

	if deets == nil {
		return nil, errors.New("Could not find providers to supply details")
	}

	myOut := make(chan GoFeatureDetail)
	req.Ctxt.Go(func() {
		for deet := range deets {
			if deet.Details != nil {
				deet.Details.Geometry = coordsys.FromWGS84(v.crs, deet.Details.Geometry)
			}
			myOut <- deet
		}

		close(myOut)
		v.remListener()
	})

	return myOut, nil
}

func (v *featuresView) GetHistoricalTrack(req GoHistoricalTrackRequest) (*feature.F, error) {
	t, err := v.tracks.GetHistoricalTrack(req)
	if err != nil {
		return nil, err
	}
	full := true
	useTrackID := true
	f := feature.FromTrack(t, full, useTrackID)
	return f.FromWGS84(v.crs), nil
}

func searchScore(term string, f *feature.F) float64 {
	sum := 0.0
	term = strings.ToLower(term)
	for _, v := range f.Properties {
		str := fmt.Sprintf("%v", v)
		if strings.Contains(strings.ToLower(str), term) {
			sum += 1.0
		}
	}
	log.Debug("Score %v for term: %v and feature: %v", sum, term, f)
	return sum
}

func (v *featuresView) Search(req GoFeatureSearchRequest) (<-chan *feature.F, error) {
	v.addListener()
	if v.dead {
		return nil, errors.New("View is dead!")
	}

	ch := make(chan *feature.F, 512)

	updateFunc := func(obj geojson.Object, upd *FeatureUpdate) bool {
		if upd == nil || upd.Feature == nil {
			return true
		}

		f := upd.Feature
		score := searchScore(req.Req.Search, f)
		if score >= SearchCutoff {
			// If we want to add this, gotta copy the map first
			//f.Properties["searchScore"] = score
			select {
			case <-req.Ctxt.Done():
				close(ch)
				return false
			case ch <- f:
				return true
			}
		}
		return true
	}

	stream := &FeatureStream{
		view:       v,
		update:     updateFunc,
		input:      make(chan FeatureUpdate, 64),
		liveStream: true,
	}

	req.Ctxt.Go(func() {
		v.Lock()
		v.streams[stream] = struct{}{}
		v.Unlock()

		stream.serviceReq(req.Ctxt, nil)

		v.Lock()
		delete(v.streams, stream)
		v.Unlock()

		close(ch)
		v.remListener()
	})

	return ch, nil
}

func (v *featuresView) History(req GoHistoryRequest) (<-chan *feature.F, error) {
	// We have to command our children providers to get details then
	// aggregate their output streams until they are all closed

	v.addListener()
	if v.dead {
		return nil, errors.New("View is dead!")
	}

	var history <-chan *feature.F
	for _, prov := range v.providers {
		s, err := prov.History(req)
		if err != nil {
			req.Ctxt.Cancel(err)
			return nil, err
		}

		if s != nil {
			if history == nil {
				history = s
			} else {
				err := errors.New("Multiple providers trying to send history!")
				req.Ctxt.Cancel(err)
				return nil, err
			}
		}
	}

	if history == nil {
		return nil, errors.New("Could not find providers to supply details")
	}

	myOut := make(chan *feature.F)
	req.Ctxt.Go(func() {
		for feat := range history {
			if feat != nil {
				feat.Geometry = coordsys.FromWGS84(v.crs, feat.Geometry)
			}
			myOut <- feat
		}

		close(myOut)
		v.remListener()
	})

	return myOut, nil
}

type Lookup struct {
	trackID    string
	registryID string
}

func (l Lookup) id() string {
	if l.registryID != "" {
		return l.registryID
	}
	return l.trackID
}

type FeatureStream struct {
	view       *featuresView
	update     func(geojson.Object, *FeatureUpdate) bool
	ch         chan<- FeatureUpdate
	input      chan FeatureUpdate
	req        GoStreamRequest
	liveStream bool
	viewport   *Viewport
	histories  map[Lookup]gogroup.GoGroup
}

func (s *FeatureStream) service(pctxt gogroup.GoGroup, reqs <-chan api.StreamRequest) {
	s.histories = make(map[Lookup]gogroup.GoGroup)
	ctxt := pctxt.Child("featureStream")
	s.viewport = NewViewport(ctxt, s.view, s.ch)
	ctxt.ErrCallback(nil)
	for {
		select {
		case req := <-reqs:
			if req.Bounds != nil {
				log.TraceMsg("FeatureStream got req: %v", req)
				log.Debug("FeatureStream got req: %v", req)

				// Cancel any existing req servicing
				ctxt.Cancel(nil)
				errors := ctxt.Wait()
				if len(errors) > 0 {
					panic(errors)
				}

				// Create new ctxt and begin new servicing
				ctxt = pctxt.Child("featureStream")
				s.viewport.Ctxt = ctxt
				ctxt.ErrCallback(nil)

				ctxt.Go(func() { s.serviceReq(ctxt, &req) })
			} else if req.History != nil && req.History.ClearAll {
				for _, hctxt := range s.histories {
					hctxt.Cancel(io.EOF)
				}
				s.histories = make(map[Lookup]gogroup.GoGroup)
				s.sendHistoryState(api.Status_HistoryClearAll, Lookup{})
			} else if req.History != nil {
				lookup := Lookup{
					trackID:    req.History.TrackId,
					registryID: req.History.RegistryId,
				}
				hctxt, ok := s.histories[lookup]
				if ok && req.History.History == 0 {
					s.sendHistoryState(api.Status_HistoryStop, lookup)
					hctxt.Cancel(io.EOF)
					continue
				}
				if ok {
					s.sendHistoryState(api.Status_HistoryStop, lookup)
					hctxt.Cancel(io.EOF)
				}
				hctxt = pctxt.Child("history")
				ghr := GoHistoryRequest{
					Ctxt: hctxt,
					Req:  req.History,
				}
				s.sendHistoryState(api.Status_HistoryStart, lookup)
				go s.streamHistory(ghr)
				s.histories[lookup] = hctxt
			}
		case <-pctxt.Done():
			return
		}
	}
}

func (s *FeatureStream) sendHistoryState(status api.Status, lookup Lookup) {
	props := map[string]interface{}{
		"TrackId":    lookup.trackID,
		"RegistryId": lookup.registryID,
	}
	update := FeatureUpdate{
		Status:  status,
		Feature: feature.New(lookup.id(), geojson.Point{}, props),
	}
	s.ch <- update
}

func (s *FeatureStream) streamHistory(req GoHistoryRequest) {
	stream, err := s.view.History(req)
	if err != nil {
		log.Error("unable to open history stream: %v", err)
		return
	}
	for {
		select {
		case feature, ok := <-stream:
			if !ok {
				return
			}
			update := FeatureUpdate{
				Status:  api.Status_History,
				Feature: feature,
			}
			s.ch <- update
		case <-req.Ctxt.Done():
			return
		}
	}
}

func (s *FeatureStream) serviceReq(ctxt gogroup.GoGroup, req *api.StreamRequest) {
	s.req = GoStreamRequest{
		Ctxt: ctxt,
		Req:  req,
	}
	if req != nil {
		var bounds geojson.BBox
		if req.Bounds == nil {
			bounds = s.view.crs.Bounds
		} else {
			bounds = geojson.New2DBBox(req.Bounds.Min.Longitude, req.Bounds.Min.Latitude,
				req.Bounds.Max.Longitude, req.Bounds.Max.Latitude)
		}
		s.viewport.SetBounds(bounds)
	}

	stop := false
	for !stop {
		select {
		case <-s.req.Ctxt.Done():
			stop = true
			s.viewport.Cleanup()
		case upd := <-s.input:
			if upd.Feature != nil {
				upd.Feature = s.view.toFeature(*upd.Feature)
				s.viewport.Process(&upd)
			}
		}
	}
}
