package db

import (
	"errors"
	"prisma/gogroup"
	. "prisma/tms"
	. "prisma/tms/client_api"
	"prisma/tms/devices"
	"prisma/tms/feature"
	"prisma/tms/log"
	"strings"
	"time"

	pb "github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/duration"
)

type TracksFeatureProvider struct {
	ProviderCommon

	trackdb  TrackDB
	devices  DeviceDB
	trackReq GoTrackRequest
	tracks   <-chan TrackUpdate
}

func (p *TracksFeatureProvider) Init(req *ViewRequest, group gogroup.GoGroup, tracks TrackDB, _ MiscDB, _ SiteDB, devices DeviceDB) error {
	log.Debug("Starting tracks provider")
	var err error
	err = p.commonInit(req, group)
	if err != nil {
		return nil
	}

	p.trackdb = tracks
	p.devices = devices

	p.trackReq, err = p.buildTrackRequest(req)
	if err != nil {
		return err
	}
	log.Debug("Track Request built %+v", req)

	p.tracks, err = tracks.GetTrackStream(p.trackReq)
	if err != nil {
		return err
	}
	return nil
}

func (p *TracksFeatureProvider) buildTrackRequest(req *ViewRequest) (GoTrackRequest, error) {
	ret := GoTrackRequest{
		Ctxt: p.ctxt,
		Req: &TrackRequest{
			Merge: MergeMode_TrackID,
		},
	}

	// Should we include all DeviceTypes?
	doAll := len(req.Types) == 0
	for _, ty := range req.Types {
		if ty.Type == FeatureCategory_AllFeatures ||
			(ty.Type == FeatureCategory_TrackFeature &&
				ty.TrackDevice == devices.DeviceType_AllDevices) {
			doAll = true
		}
	}

	if !doAll {
		var devs []devices.DeviceType = nil
		for _, ty := range req.Types {
			if ty.Type == FeatureCategory_TrackFeature {
				devs = append(devs, ty.TrackDevice)
			}
		}

		ret.Req.Filter = &TrackRequest_FilterSimple{
			FilterSimple: &FilterSimple{
				DeviceType: devs,
			},
		}
	}

	return ret, nil
}

func (p *TracksFeatureProvider) toFeature(t *Track, full bool, useTrackId bool) *feature.F {
	feat := feature.FromTrack(t, full, useTrackId)
	if feat != nil {
		p.extraPropertyByDeviceType(t, feat.Properties)
	}
	return feat
}

// extraPropertyByDeviceType is used to add extra information for specific devices
func (p *TracksFeatureProvider) extraPropertyByDeviceType(t *Track, prop map[string]interface{}) {
	if len(t.Targets) == 0 {
		return
	}
	target := t.Targets[0]
	deviceType := target.Type
	if deviceType == devices.DeviceType_OmnicomSolar || deviceType == devices.DeviceType_OmnicomVMS {
		device, err := p.devices.FindNet(target.Imei.Value)
		if err != nil {
			log.Error(err.Error())
			return
		}
		prop["deviceId"] = device.DeviceId
	}
}

func (p *TracksFeatureProvider) toFeatureDetail(req GoDetailRequest, t *Track) *GoFeatureDetail {
	if t == nil {
		return nil
	}

	ret := &GoFeatureDetail{
		Details: p.toFeature(t, true, true),
	}
	// We need some way to sanitize these first. For now, just ditch them
	/*for _, tgt := range t.Targets {
		ret.Sources = append(ret.Sources, tgt)
	}
	for _, md := range t.Metadata{
		ret.Sources = append(ret.Sources, md)
	}*/

	// Build a Feature history
	if req.Req.History > 0 && len(t.Targets) > 0 {
		tgts := make([]*Target, len(t.Targets))
		copy(tgts, t.Targets)
		metas := make([]*TrackMetadata, len(t.Metadata))
		copy(metas, t.Metadata)

		// Loop through all metadata and targets. Interleave them by time while
		// also pairing them and build features with each pair
		lastTgt := tgts[len(tgts)-1]
		var lastMd *TrackMetadata
		if len(metas) > 0 {
			lastMd = metas[len(metas)-1]
		}
		for len(tgts) > 0 || len(metas) > 0 {
			// Build a feature
			feat := p.toFeature(&Track{
				Targets:  []*Target{lastTgt},
				Metadata: []*TrackMetadata{lastMd},
			}, true, false)
			ret.History = append(ret.History, feat)

			// Select either a new Target or new Metadata for the next feature
			if len(tgts) > 0 && len(metas) > 0 {
				tgtHead := tgts[len(tgts)-1]
				mdHead := metas[len(metas)-1]

				tgtTime := FromTimestamp(tgtHead.Time)
				mdTime := FromTimestamp(mdHead.Time)

				if tgtTime.Before(mdTime) {
					lastTgt = tgtHead
					tgts = tgts[0 : len(tgts)-1]
				} else {
					lastMd = mdHead
					metas = metas[0 : len(metas)-1]
				}
			} else if len(tgts) > 0 {
				lastTgt = tgts[len(tgts)-1]
				tgts = tgts[0 : len(tgts)-1]
			} else if len(metas) > 0 {
				lastMd = metas[len(metas)-1]
				metas = metas[0 : len(metas)-1]
			} else {
				panic("Logical error!")
			}
		}
	}
	return ret
}

func (p *TracksFeatureProvider) Destroy() {
	p.ctxt.Cancel(nil)
}

func (p *TracksFeatureProvider) Service(ch chan<- FeatureUpdate) error {
	for {
		select {
		case <-p.ctxt.Done():
			return p.ctxt.Err()
		case upd, ok := <-p.tracks:
			if !ok {
				return errors.New("closed channel")
			}
			var feat *feature.F
			if upd.Track != nil {
				feat = p.toFeature(upd.Track, false, true)
			}

			featUpd := FeatureUpdate{
				Status:  upd.Status,
				Feature: feat,
			}

			select {
			case <-p.ctxt.Done():
				return p.ctxt.Err()
			case ch <- featUpd:
			}
		}
	}
	return nil
}

func (p *TracksFeatureProvider) DetailsStream(req GoDetailRequest) (<-chan GoFeatureDetail, error) {
	log.Info("track provider")
	if !strings.HasPrefix(req.Req.Type, "track") {
		log.Info("no")
		return nil, nil
	}
	log.Info("yes")
	trackId := req.Req.FeatureId
	if trackId == "" {
		return nil, errors.New("no track id provided")
	}
	trackReq := GoTrackRequest{
		Req:          pb.Clone(p.trackReq.Req).(*TrackRequest),
		Ctxt:         req.Ctxt,
		DisableMerge: false,
		Stream:       req.Stream,
	}
	treq := trackReq.Req

	if req.Req.History > 0 {
		treq.History = &duration.Duration{
			Seconds: int64(req.Req.History),
		}
	} else {
		treq.History = &duration.Duration{
			Seconds: 0,
		}
	}

	if treq.GetFilterSimple() == nil {
		treq.Filter = &TrackRequest_FilterSimple{
			FilterSimple: &FilterSimple{
				Tracks: []string{trackId},
			},
		}
	} else {
		treq.GetFilterSimple().Tracks = []string{trackId}
	}

	trackUpdates, err := p.trackdb.Get(trackReq)
	if err != nil {
		return nil, err
	}

	ch := make(chan GoFeatureDetail)
	req.Ctxt.Go(func() {
		defer close(ch)
		for upd := range trackUpdates {
			if upd.Status == Status_Current && upd.Track != nil {
				d := p.toFeatureDetail(req, upd.Track)
				if d == nil {
					continue
				}
				select {
				case ch <- *d:
				case <-req.Ctxt.Done():
					return
				}
			}
		}
		log.Debug("Closing detail stream")
	})
	return ch, nil
}

func (p *TracksFeatureProvider) History(req GoHistoryRequest) (<-chan *feature.F, error) {
	if req.Req.History == 0 {
		return nil, errors.New("please specify a history (in seconds) to retrieve for this target, not 0")
	}

	trackReq := GoTrackRequest{
		Req:             pb.Clone(p.trackReq.Req).(*TrackRequest),
		Ctxt:            req.Ctxt,
		Stream:          false,
		DisableMerge:    true,
		DisableTimeouts: true,
		MaxHistory:      time.Duration(req.Req.History) * time.Second,
	}

	treq := trackReq.Req

	if req.Req.RegistryId != "" {
		treq.Filter = &TrackRequest_FilterSimple{
			FilterSimple: &FilterSimple{
				Registries: []string{req.Req.RegistryId},
			},
		}
	} else if req.Req.TrackId != "" {
		treq.Filter = &TrackRequest_FilterSimple{
			FilterSimple: &FilterSimple{
				Tracks: []string{req.Req.TrackId},
			},
		}
	} else {
		return nil, errors.New("no filter criteria specified")
	}
	trackUpdates, err := p.trackdb.Get(trackReq)
	if err != nil {
		return nil, err
	}

	ch := make(chan *feature.F)
	req.Ctxt.Go(func() {
		defer close(ch)
		for upd := range trackUpdates {
			if upd.Status == Status_Current && upd.Track != nil {
				if len(upd.Track.Targets) > 0 {
					feat := p.toFeature(upd.Track, false, false)
					if feat == nil {
						continue
					}
					select {
					case ch <- feat:
					case <-req.Ctxt.Done():
						return
					}
				}
			}
		}
		log.Debug("Closing detail stream")
	})
	return ch, nil
}
