package main

import (
	"fmt"
	"prisma/gogroup"
	"prisma/tms"
	"prisma/tms/log"
	"prisma/tms/tmsg"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/serf/serf"
)

const (
	timeUpdateDC      = 10 * time.Second
	timeUpdateServers = 10 * time.Second
)

// It can share data between data centers
type Forwarder struct {
	mu                   sync.Mutex
	localServers         []serf.Member
	dataCenters          []string
	clientsTgwadServices []*clientsDef

	client *api.Client
	curDC  string
	ctx    gogroup.GoGroup
}

type clientsDef struct {
	c  *tmsg.TsiClientTcp
	dc string
	id string
}

// Return an instance of forwarder, also it runs watchers
// for updating any information about data centers
func NewForwarder(ctx gogroup.GoGroup, address, dc string) (*Forwarder, error) {
	conf := api.DefaultConfig()
	conf.Scheme = "http"
	conf.Address = address + ":9099"
	c, err := api.NewClient(conf)
	if err != nil {
		return nil, err
	}
	fw := &Forwarder{
		client:               c,
		curDC:                dc,
		ctx:                  ctx,
		localServers:         make([]serf.Member, 0),
		dataCenters:          make([]string, 0),
		clientsTgwadServices: make([]*clientsDef, 0),
	}
	if err := fw.getDataCenters(); err != nil {
		return nil, err
	}
	fw.getTgwadServices()
	go fw.watchDataCenters()
	go fw.watchServers()
	go fw.watchServices()
	return fw, nil
}

func (fw *Forwarder) watchServers() {
	tick := time.NewTicker(timeUpdateServers)
	defer tick.Stop()
	select {
	case <-fw.ctx.Done():
		return
	case <-tick.C:
		fw.getServers()
	}
}

func (fw *Forwarder) watchDataCenters() {
	tick := time.NewTicker(timeUpdateDC)
	defer tick.Stop()
	select {
	case <-fw.ctx.Done():
		return
	case <-tick.C:
		fw.getDataCenters()
	}
}

func (fw *Forwarder) watchServices() {
	tick := time.NewTicker(timeUpdateDC)
	defer tick.Stop()
	select {
	case <-fw.ctx.Done():
		return
	case <-tick.C:
		fw.getTgwadServices()
	}
}

func (fw *Forwarder) getServers() {
}

func (fw *Forwarder) getDataCenters() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	dcs, err := fw.client.Catalog().Datacenters()
	if err != nil {
		return err
	}
	fw.dataCenters = dcs
	return nil
}

// It is used for searching an id among known id of services.
// It is for getting unique services
func (fw *Forwarder) searchId(id string) bool {
	for _, service := range fw.clientsTgwadServices {
		if service.id == id {
			return true
		}
	}
	return false
}

// Scanning services of tgwad among data centers but us one
func (fw *Forwarder) getTgwadServices() {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	for _, dc := range fw.dataCenters {
		if dc == fw.curDC {
			continue
		}
		services, _, err := fw.client.Catalog().Service("tgwad", "", &api.QueryOptions{
			Datacenter: dc,
		})
		if err != nil {
			log.Error(err.Error())
			continue
		}
		for _, service := range services {
			c, err := tmsg.ConnectTsiClient(fw.ctx, fmt.Sprintf("%s:%d", service.ServiceAddress, service.ServicePort),
				tmsg.APP_ID_CONSUL)
			if err != nil {
				log.Error(err.Error())
				continue
			}
			fw.clientsTgwadServices = append(fw.clientsTgwadServices,
				&clientsDef{
					id: service.ID,
					c:  c,
					dc: service.Datacenter,
				})
		}

	}
}

// Share a track between tgwad services. It uses the services which was gotten from a service watcher
func (fw *Forwarder) Share(track tms.Track, dataCenters []string) error {
	body, err := tmsg.PackFrom(&track)
	if err != nil {
		return err
	}
	m := &tms.TsiMessage{
		WriteTime: tms.Now(),
		SendTime:  tms.Now(),
		Body:      body,
	}
	fw.mu.Lock()
	defer fw.mu.Unlock()
	for _, client := range fw.clientsTgwadServices {
		for _, dc := range dataCenters {
			if client.dc == dc {
				m.Destination = []*tms.EndPoint{{Site: client.c.Local().Site}}
				client.c.Send(fw.ctx, m)
			}
		}
	}
	return nil
}
