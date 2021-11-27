package db

import (
	"fmt"
	"prisma/gogroup"
	. "prisma/tms/client_api"
	"prisma/tms/feature"
	"prisma/tms/log"
	"prisma/tms/moc"
	"reflect"
	"errors"
)

type FeatureConverter func(*GoObject) (*feature.F, error)

var (
	/*****
	 * Hey! You! Are you looking to add support for an object to be displayed
	 * as a feature? That code lives anywhere then gets hooked in here. Just
	 * write you code and have it register itself as a FeatureConverter at
	 * init() time. Or register it here. Which you should do depends on how you
	 * want to deal with the go import cycles.
	 */
	FeatureConverters = map[reflect.Type]FeatureConverter{
		reflect.TypeOf(moc.Zone{}): ConvertZoneToFeature,
	}
)

type MiscDataFeatureProvider struct {
	ProviderCommon

	miscdb  MiscDB
	table   *TableInfo
	miscReq *GoMiscRequest
	data    <-chan GoGetResponse
}

func (p *MiscDataFeatureProvider) Init(req *ViewRequest, group gogroup.GoGroup, _ TrackDB, misc MiscDB, _ SiteDB, _ DeviceDB) error {
	log.Debug("Starting provider: %v (%p)", p.table.Name, p)
	p.commonInit(req, group)

	p.miscdb = misc

	var err error
	p.miscReq, err = p.buildMiscRequest(req)
	if err != nil {
		return err
	}

	p.data, err = misc.GetStream(*p.miscReq, nil, nil)
	return err
}

func (p *MiscDataFeatureProvider) buildMiscRequest(req *ViewRequest) (*GoMiscRequest, error) {
	ret := &GoRequest{
		ObjectType: fmt.Sprintf("%v.%v", p.table.Type.PkgPath(), p.table.Type.Name()),
	}

	return NewMiscRequest(ret, p.ctxt)
}

func (p *MiscDataFeatureProvider) DetailsStream(req GoDetailRequest) (<-chan GoFeatureDetail, error) {
	if req.Req.Type != p.table.Name {
		return nil, nil
	}

	goObject := &GoObject{
		ID: req.Req.FeatureId,
	}
	goRequest := &GoRequest{
		ObjectType: fmt.Sprintf("%v.%v", p.table.Type.PkgPath(), p.table.Type.Name()),
		Obj:        goObject,
	}
	miscRequest := GoMiscRequest{
		Req:  goRequest,
		Ctxt: req.Ctxt,
	}

	miscUpdates, err := p.miscdb.Get(miscRequest)
	if err != nil {
		return nil, err
	}

	ch := make(chan GoFeatureDetail)
	req.Ctxt.Go(func() {
		defer close(ch)

		for _, miscUpdate := range miscUpdates {
			if miscUpdate.Status == Status_Current && miscUpdate.Contents != nil {
				d := &GoFeatureDetail{
					Details: p.toFeature(miscUpdate.Contents),
				}
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

func (p *MiscDataFeatureProvider) Destroy() {
}

// Get the history for a feature
func (p *MiscDataFeatureProvider) History(GoHistoryRequest) (<-chan *feature.F, error) {
	return nil, nil
}

func (p *MiscDataFeatureProvider) Service(ch chan<- FeatureUpdate) error {
	log.Debug("Servicing %v (%p)", p.table.Name, p)
	for {
		select {
		case <-p.ctxt.Done():
			return p.ctxt.Err()
		case upd, ok := <-p.data:
			if !ok {
				return errors.New("closed channel")
			}
			log.Debug("Got misc update: %v", log.Spew(upd))
			var feat *feature.F
			if upd.Contents != nil {
				feat = p.toFeature(upd.Contents)
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
	log.Debug("Provider %v dying...", p.table.Name)
	return nil
}

func (p *MiscDataFeatureProvider) toFeature(obj *GoObject) *feature.F {
	if obj == nil {
		return nil
	}

	ty := reflect.TypeOf(obj.Data)
	conv, ok := FeatureConverters[ty]
	if !ok {
		panic(fmt.Sprintf("Could not find method to convert %v to a feature!", ty))
	}
	feat, err := conv(obj)
	if err != nil {
		log.Error("Error converting %v to feature: %v", ty, err)
	}
	return feat
}
