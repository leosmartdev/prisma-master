package db

import (
	"prisma/gogroup"
	api "prisma/tms/client_api"
	"prisma/tms/feature"
	"prisma/tms/geojson"
	"prisma/tms/log"
	"prisma/tms/moc"
)

type SitesFeatureProvider struct {
	ProviderCommon
	sitedb SiteDB
	sites  <-chan moc.Site
}

func (p *SitesFeatureProvider) Init(req *api.ViewRequest, group gogroup.GoGroup, _ TrackDB, _ MiscDB, sitedb SiteDB, _ DeviceDB) error {
	log.Debug("Starting sites provider")
	var err error
	err = p.commonInit(req, group)
	if err != nil {
		return nil
	}
	p.ctxt = group
	p.sitedb = sitedb
	return err
}

func (p *SitesFeatureProvider) DetailsStream(GoDetailRequest) (<-chan GoFeatureDetail, error) {
	// TODO
	return nil, nil
}

func (p *SitesFeatureProvider) History(GoHistoryRequest) (<-chan *feature.F, error) {
	return nil, nil
}

func (p *SitesFeatureProvider) Destroy() {
	p.ctxt.Cancel(nil)
}

func (p *SitesFeatureProvider) Service(ch chan<- FeatureUpdate) error {
	var err error
	sites, err := p.sitedb.FindAll(p.ctxt)
	for _, site := range sites {
		if site.Point == nil {
			continue
		}
		ch <- FeatureUpdate{
			Status: api.Status_Current,
			Feature: &feature.F{
				Feature: geojson.Feature{
					ID: site.Id,
					Geometry: &geojson.Point{
						Coordinates: geojson.Position{
							site.Point.Longitude,
							site.Point.Latitude,
							0},
					},
					Properties: map[string]interface{}{
						"databaseId": site.Id,
						"name":       site.Name,
						"type":       "site",
						"siteId":     site.SiteId,
						"status":     site.ConnectionStatus.String(),
					},
				},
			},
		}
	}
	return err
}
