package db

import (
	"prisma/gogroup"
	. "prisma/tms/client_api"
	"prisma/tms/log"
)

/**
 * A TrackPipelineStage is something which processes track updates and can send
 * them further down the pipeline
 */
type TrackPipelineStage interface {
	Start(<-chan TrackUpdate) (<-chan TrackUpdate, error)
}

/**
 * TrackPipeline is a series of stages which are used to process a stream up
 * tracks and track updates. This simple struct constructs the pipeline,
 * configures it, and starts it.
 */
type TrackPipeline struct {
	tracks TrackDB
	reg    RegistryDB
	misc   MiscDB

	stages []TrackPipelineStage
}

func NewTrackPipeline(tracks TrackDB, reg RegistryDB) *TrackPipeline {
	p := &TrackPipeline{
		tracks: tracks,
		reg:    reg,
	}
	return p
}

func (p *TrackPipeline) Append(s TrackPipelineStage) {
	p.stages = append(p.stages, s)
}

func (p *TrackPipeline) Start(inputch <-chan TrackUpdate) (<-chan TrackUpdate, error) {
	var ch <-chan TrackUpdate = inputch
	var err error
	for _, stage := range p.stages {
		ch, err = stage.Start(ch)
		if err != nil {
			return nil, err
		}
	}
	return ch, nil
}

// A pipeline stage which does nothing. It's a pass-through
type NullTrackPipelineStage struct{}

func (_ NullTrackPipelineStage) Start(inputch <-chan TrackUpdate) (<-chan TrackUpdate, error) {
	return inputch, nil
}

// A pipeline stange that echos everything to the log.
type LogPipelineStage struct {
	ctxt   gogroup.GoGroup
	tracer *log.Tracer
}

func NewLogStage(ctxt gogroup.GoGroup) *LogPipelineStage {
	return &LogPipelineStage {
		ctxt:   ctxt,
		tracer: log.GetTracer("pipeline"),
	}
}

func (p *LogPipelineStage) Start(in <-chan TrackUpdate) (<-chan TrackUpdate, error) {
	out := make(chan TrackUpdate, 128)
	p.ctxt.Go(func() {
		defer close(out)
		for {
			select {
			case <-p.ctxt.Done():
				return
			case update, ok := <-in:
				if !ok {
					p.tracer.Logf("end of tracks")
					return
				}
				p.tracer.Logf("update: %v", update.Status)
				out <- update
			}
		}
	})
	return out, nil
}
