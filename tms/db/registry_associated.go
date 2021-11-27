package db

import (
	. "prisma/tms/client_api"
	"prisma/tms/log"

	"prisma/gogroup"
	"sync"
)

type AssociatedDataProvider struct {
	tables []TableInfo
	miscdb MiscDB
}

type AssociatedReq struct {
	Ctxt       gogroup.GoGroup
	RegistryId string
	IgnoreTime bool
}

func NewAssociatedDataProvider(miscdb MiscDB) *AssociatedDataProvider {
	var tables []TableInfo
	for _, ti := range DefaultTables.Info {
		if ti.ContainsAssociated {
			tables = append(tables, *ti)
		}
	}
	return &AssociatedDataProvider{
		miscdb: miscdb,
		tables: tables,
	}
}

func (a *AssociatedDataProvider) Get(req AssociatedReq) (<-chan GoGetResponse, error) {
	ch := make(chan GoGetResponse, 32)
	backlogDone := sync.WaitGroup{}
	allClosed := sync.WaitGroup{}

	for _, ti := range a.tables {
		req, err := NewMiscRequest(&GoRequest{
			ObjectType: ti.TypeName(),
			Obj: &GoObject{
				RegistryId: req.RegistryId,
			},
			IgnoreTime: true,
		}, req.Ctxt)
		if err != nil {
			return nil, err
		}

		mch, err := a.miscdb.GetStream(*req, nil, nil)
		if err != nil {
			req.Ctxt.Cancel(err)
			return nil, err
		}

		backlogDone.Add(1)
		allClosed.Add(1)
		req.Ctxt.Go(func() {
			gotBacklogDone := false
			for resp := range mch {
				log.Debug("Got misc resp: %v", log.Spew(resp))
				switch resp.Status {
				case Status_InitialLoadDone:
					// Signal our backlog is done
					backlogDone.Done()
					// Wait for everyone else to finish their backlog (barrier)
					backlogDone.Wait()
					gotBacklogDone = true
				default:
					ch <- resp
				}
			}
			if !gotBacklogDone {
				backlogDone.Done()
			}
			allClosed.Done()
		})
	}

	req.Ctxt.Go(func() {
		backlogDone.Wait()
		ch <- GoGetResponse{
			Status: Status_InitialLoadDone,
		}
		allClosed.Wait()
		close(ch)
	})

	return ch, nil
}
