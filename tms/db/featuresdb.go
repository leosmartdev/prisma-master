package db

import (
	. "prisma/tms"
	. "prisma/tms/client_api"
	"prisma/tms/feature"

	"prisma/gogroup"
)

type FeatureDB interface {
	// Report the version
	Version() string

	// Create a view of the world. Can adjust the time, types of features, etc.
	CreateView(*ViewRequest) (FeaturesView, error)

	Stream(gogroup.GoGroup, <-chan StreamRequest) (<-chan FeatureUpdate, error)
	Snapshot(GoStreamRequest) ([]*feature.F, error)

	Details(GoDetailRequest) (<-chan GoFeatureDetail, error)
	GetHistoricalTrack(GoHistoricalTrackRequest) (*feature.F, error)
	Search(GoFeatureSearchRequest) (<-chan *feature.F, error)
}

type FeaturesView interface {
	ID() string
	Stream(gogroup.GoGroup, <-chan StreamRequest) (<-chan FeatureUpdate, error)
	Snapshot(GoStreamRequest) ([]*feature.F, error)

	Details(GoDetailRequest) (<-chan GoFeatureDetail, error)
	Search(GoFeatureSearchRequest) (<-chan *feature.F, error)
	History(GoHistoryRequest) (<-chan *feature.F, error)
}

type GoStreamRequest struct {
	Ctxt gogroup.GoGroup
	Req  *StreamRequest
}

type FeatureCounts struct {
	Total   int `json:"total"`
	Visible int `json:"visible"`
}

type FeatureUpdate struct {
	Status  Status         `json:"status"`
	Feature *feature.F     `json:"feature,omitempty"`
	Heatmap *Heatmap       `json:"heatmap,omitempty"`
	Counts  *FeatureCounts `json:"counts"`
}

type GoDetailRequest struct {
	Ctxt gogroup.GoGroup
	Req  *DetailRequest
	// Should we continue streaming?
	Stream bool
}

type GoHistoryRequest struct {
	Ctxt gogroup.GoGroup
	Req  *HistoryRequest
}

type GoFeatureSearchRequest struct {
	Ctxt gogroup.GoGroup
	Req  *FeatureSearchRequest
	// Should we continue streaming?
	Stream bool
}

type GoFeatureDetail struct {
	ViewID    string `json:"viewId"`
	FeatureID string `json:"featureId"`

	Details *feature.F   `json:"details"`
	History []*feature.F `json:"history,omitempty"`
}
