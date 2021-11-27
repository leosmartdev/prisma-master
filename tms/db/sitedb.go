package db

import (
	"context"

	moc "prisma/tms/moc"
	"prisma/gogroup"
	"time"
)

type SiteDB interface {
	Create(ctx context.Context, site *moc.Site) (*moc.Site, error)
	Update(ctx context.Context, site *moc.Site) (*moc.Site, error)
	// UpdateConnectionStatusBySiteId updates connection status using site.SiteId to find, site.Id is not used
	UpdateConnectionStatusBySiteId(ctx context.Context, site *moc.Site) (*moc.Site, error)
	// FindBySiteId ...
	FindBySiteId(ctx context.Context, site *moc.Site) error
	// FindById ...
	FindById(ctx context.Context, siteId string) (*moc.Site, error)
	Delete(ctx context.Context, siteId string) error
	FindAll(ctx context.Context) ([]*moc.Site, error)
	FindByMap(ctx context.Context, searchMap map[string]string, sortFields SortFields) ([]*moc.Site, error)
}

// A track request supplemented with stuff only useful to go code or needed by
// various track pipeline stages.
type GoSiteRequest struct {
	Req  *moc.Site
	Ctxt gogroup.GoGroup
	Time *TimeKeeper

	MaxHistory      time.Duration
	Stream          bool
	DisableMerge    bool
	DisableTimeouts bool
	DebugQuery      bool
}