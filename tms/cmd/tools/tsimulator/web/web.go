// Package web containers handlers to manage tsimulator's resources.
package web

import (
	"errors"
	"net/http"
	"prisma/tms/cmd/tools/tsimulator/object"
	"prisma/tms/log"
	"strconv"

	restful "github.com/orolia/go-restful" 
	"prisma/tms/spidertracks"
)

var errWebBadTypeAlerting = errors.New(`Bad type alert need: PU
	PD
	BA
	IA
	NPF
	JBDA
	LMC
	DA
	AA
	TM`)

// SetupServer setups a rest service that is providing CRUD for sea objects
func SetupServer(control *object.Control, addr string) func() {
	return func() {
		log.Error("WebService of simulator has error: %v", http.ListenAndServe(addr, GetRestContainer(control)))
	}
}

// GetRestContainer returns a rest container which determines endpoints to control the simulator
func GetRestContainer(control *object.Control) *restful.Container {
	apiV1 := new(restful.WebService)
	apiV1.Path("/v1/")
	{
		serverso := new(serviceSeaObject)
		addHandler(apiV1, http.MethodGet, "/get/", serverso.Get(control))

		addHandler(apiV1, http.MethodPost, "/target/id/{target-id}", serverso.UpdateTarget(control))
		addHandler(apiV1, http.MethodPost, "/target/", serverso.CreateTarget(control))
		addHandler(apiV1, http.MethodDelete, "/target/id/{target-id}", serverso.DeleteTarget(control))
		addHandler(apiV1, http.MethodPost, "/alert/type/{type-alert}/target-id/{target-id}",
			serverso.StartAlerting(control))
		addHandler(apiV1, http.MethodDelete, "/alert/type/{type-alert}/target-id/{target-id}",
			serverso.StopAlerting(control))

		addHandler(apiV1, http.MethodPost, "/route/id/{route-id}/target-id/{target-id}",
			serverso.UpdateRoute(control))
		addHandler(apiV1, http.MethodPost, "/route/target-id/{target-id}", serverso.UpdateWholeRoute(control))
		addHandler(apiV1, http.MethodDelete, "/route/target-id/{target-id}", serverso.DeleteRoute(control))
		addHandler(apiV1, http.MethodPost, "/aff/feed", serverso.GetSpidertrackData(control))
	}
	restServer := restful.NewContainer()
	restServer.Add(apiV1)
	restServer.Filter(logging)
	return restServer
}

func addHandler(api *restful.WebService, method string, route string, handler restful.RouteFunction) {
	switch method {
	case http.MethodPost:
		api.Route(api.POST(route).To(handler))
	case http.MethodGet:
		api.Route(api.GET(route).To(handler))
	case http.MethodDelete:
		api.Route(api.DELETE(route).To(handler))
	}
}

func logging(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	log.Info("[restful] - %s %s  [http response]: %d", req.Request.Method, req.Request.URL, resp.StatusCode())
	chain.ProcessFilter(req, resp)
}

type serviceSeaObject struct{}

func (*serviceSeaObject) Get(control *object.Control) restful.RouteFunction {
	return func(r *restful.Request, w *restful.Response) {
		listObjects := control.GetList()
		if err := w.WriteAsJson(map[string]interface{}{"count": len(listObjects), "objects": listObjects}); err != nil {
			writeErrorWithLog(w, http.StatusInternalServerError, err)
		}
	}
}

func (*serviceSeaObject) CreateTarget(control *object.Control) restful.RouteFunction {
	return func(r *restful.Request, w *restful.Response) {
		var seaobject object.Object
		if err := r.ReadEntity(&seaobject); err != nil {
			writeErrorWithLog(w, http.StatusBadRequest, err)
			return
		}
		seaobject.Id = control.Insert(seaobject)
		if err := w.WriteHeaderAndJson(http.StatusCreated, seaobject, restful.MIME_JSON); err != nil {
			writeErrorWithLog(w, http.StatusInternalServerError, err)
		}
	}
}

func (*serviceSeaObject) GetSpidertrackData(control *object.Control) restful.RouteFunction {
	return func(r *restful.Request, w *restful.Response) {
		list := control.GetList()
		var spidertracksList []*object.Object
		for _, obj := range list {
			if obj.Device != "spidertracks" {
				continue
			}
			spidertracksList = append(spidertracksList, obj)
		}
		var responsePos spidertracks.Spider
		for _, obj := range spidertracksList {
			device, ok := obj.GetDevice().(*object.Spidertrack)
			if !ok {
				log.Error("unable to resolve spidertrack device from: %+v", obj.GetDevice())
				continue
			}
			responsePos.PosList = append(responsePos.PosList, device.GetACPos())
		}
		if err := w.WriteAsXml(responsePos); err != nil {
			writeErrorWithLog(w, http.StatusInternalServerError, err)
			return
		}
	}
}

func (*serviceSeaObject) StartAlerting(control *object.Control) restful.RouteFunction {
	return func(r *restful.Request, w *restful.Response) {
		var (
			seaobject object.Object
			typeAlert string
			index     int
			err       error
		)
		// getting an seaobject
		if index, err = strconv.Atoi(r.PathParameter("target-id")); err != nil || index < 1 {
			writeErrorWithLog(w, http.StatusBadRequest, errors.New("bad id"))
			return
		}
		if typeAlert = r.PathParameter("type-alert"); typeAlert == "" {
			writeErrorWithLog(w, http.StatusBadRequest, errWebBadTypeAlerting)
			return
		}

		if seaobject, err = control.GetByIndex(index); err != nil {
			writeErrorWithLog(w, http.StatusBadRequest, err)
			return
		}
		switch typeAlert {
		default:
			writeErrorWithLog(w, http.StatusBadRequest, errWebBadTypeAlerting)
			return
		case "PD", "pd":
			err = seaobject.StartAlerting(object.PD)
		case "PU", "pu":
			err = seaobject.StartAlerting(object.PU)
		case "BA", "ba":
			err = seaobject.StartAlerting(object.BA)
		case "IA", "ia":
			err = seaobject.StartAlerting(object.IA)
		case "NPF", "npf":
			err = seaobject.StartAlerting(object.NPF)
		case "JBDA", "jbda":
			err = seaobject.StartAlerting(object.JBDA)
		case "LMC", "lmc":
			err = seaobject.StartAlerting(object.LMC)
		case "DA", "da":
			err = seaobject.StartAlerting(object.DA)
		case "AA", "aa":
			err = seaobject.StartAlerting(object.AA)
		case "TM", "tm":
			err = seaobject.StartAlerting(object.TM)
		}
		if err != nil {
			log.Error(err.Error())
		}
		control.Update(index, seaobject)
		w.WriteHeader(http.StatusCreated)
	}
}

func (*serviceSeaObject) StopAlerting(control *object.Control) restful.RouteFunction {
	return func(r *restful.Request, w *restful.Response) {
		var (
			seaobject object.Object
			typeAlert string
			index     int
			err       error
		)
		// getting an seaobject
		if index, err = strconv.Atoi(r.PathParameter("target-id")); err != nil || index < 1 {
			writeErrorWithLog(w, http.StatusBadRequest, errors.New("bad id"))
			return
		}
		if typeAlert = r.PathParameter("type-alert"); typeAlert == "" {
			writeErrorWithLog(w, http.StatusBadRequest, errWebBadTypeAlerting)
			return
		}
		if seaobject, err = control.GetByIndex(index); err != nil {
			writeErrorWithLog(w, http.StatusBadRequest, err)
			return
		}
		switch typeAlert {
		default:
			writeErrorWithLog(w, http.StatusBadRequest, errWebBadTypeAlerting)
			return
		case "PD", "pd":
			err = seaobject.StopAlerting(object.PD)
		case "PU", "pu":
			err = seaobject.StopAlerting(object.PU)
		case "BA", "ba":
			err = seaobject.StopAlerting(object.BA)
		case "IA", "ia":
			err = seaobject.StopAlerting(object.IA)
		case "NPF", "npf":
			err = seaobject.StopAlerting(object.NPF)
		case "JBDA", "jbda":
			err = seaobject.StopAlerting(object.JBDA)
		case "LMC", "lmc":
			err = seaobject.StopAlerting(object.LMC)
		case "DA", "da":
			err = seaobject.StopAlerting(object.DA)
		case "AA", "aa":
			err = seaobject.StopAlerting(object.AA)
		case "TM", "tm":
			err = seaobject.StopAlerting(object.TM)
		}
		if err != nil {
			log.Error(err.Error())
		}
		control.Update(index, seaobject)
		w.WriteHeader(http.StatusNoContent)
	}
}

func (*serviceSeaObject) UpdateTarget(control *object.Control) restful.RouteFunction {
	return func(r *restful.Request, w *restful.Response) {
		var (
			seaobject object.Object
			index     int
			err       error
		)
		// getting an seaobject
		if index, err = strconv.Atoi(r.PathParameter("target-id")); err != nil || index < 1 {
			writeErrorWithLog(w, http.StatusBadRequest, errors.New("bad id"))
			return
		}
		if err := r.ReadEntity(&seaobject); err != nil {
			writeErrorWithLog(w, http.StatusBadRequest, err)
			return
		}
		if _, err := control.GetByIndex(index); err == nil {
			if err := control.Update(index, seaobject); err != nil {
				writeErrorWithLog(w, http.StatusBadRequest, err)
				return
			}
		} else {
			seaobject.Id = index
			control.Insert(seaobject)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (*serviceSeaObject) CreateRoute(control *object.Control) restful.RouteFunction {
	return func(r *restful.Request, w *restful.Response) {
		var (
			seaobject object.Object
			pos       object.PositionArrivalTime
			index     int
			err       error
		)
		if index, err = strconv.Atoi(r.PathParameter("target-id")); err != nil || index < 1 {
			writeErrorWithLog(w, http.StatusBadRequest, errors.New("bad id"))
			return
		}
		index--
		if seaobject, err = control.GetByIndex(index); err != nil {
			writeErrorWithLog(w, http.StatusBadRequest, err)
			return
		}
		if err := r.ReadEntity(&pos); err != nil {
			writeErrorWithLog(w, http.StatusBadRequest, err)
			return
		}
		seaobject.Pos = append(seaobject.Pos, pos)
		if err := control.Update(index, seaobject); err != nil {
			writeErrorWithLog(w, http.StatusBadRequest, err)
			return
		}
		if err := control.SetupNewPosition(index, 0); err != nil {
			if err := w.WriteError(http.StatusInternalServerError, err); err != nil {
				log.Error(err.Error())
			}
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func (*serviceSeaObject) UpdateRoute(control *object.Control) restful.RouteFunction {
	return func(r *restful.Request, w *restful.Response) {
		var (
			seaobject  object.Object
			pos        object.PositionArrivalTime
			index      int
			indexRoute int
			err        error
		)
		if index, err = strconv.Atoi(r.PathParameter("target-id")); err != nil || index < 1 {
			writeErrorWithLog(w, http.StatusBadRequest, errors.New("bad id"))
			return
		}
		if indexRoute, err = strconv.Atoi(r.PathParameter("route-id")); err != nil {
			writeErrorWithLog(w, http.StatusBadRequest, err)
			return
		}
		indexRoute--
		if seaobject, err = control.GetByIndex(index); err != nil {
			writeErrorWithLog(w, http.StatusBadRequest, err)
			return
		}
		if err := r.ReadEntity(&pos); err != nil {
			writeErrorWithLog(w, http.StatusBadRequest, err)
			return
		}
		if indexRoute >= len(seaobject.Pos) || indexRoute < 0 {
			writeErrorWithLog(w, http.StatusBadRequest, errors.New("bad an index of a position"))
			return
		}
		seaobject.Pos[indexRoute] = pos
		if err := control.Update(index, seaobject); err != nil {
			writeErrorWithLog(w, http.StatusBadRequest, err)
			return
		}
		if err := control.SetupNewPosition(index, indexRoute); err != nil {
			writeErrorWithLog(w, http.StatusInternalServerError, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (*serviceSeaObject) UpdateWholeRoute(control *object.Control) restful.RouteFunction {
	return func(r *restful.Request, w *restful.Response) {
		var (
			seaobject object.Object
			pos       []object.PositionArrivalTime
			index     int
			err       error
		)
		if index, err = strconv.Atoi(r.PathParameter("target-id")); err != nil || index < 1 {
			writeErrorWithLog(w, http.StatusBadRequest, errors.New("bad id"))
			return
		}
		if seaobject, err = control.GetByIndex(index); err != nil {
			writeErrorWithLog(w, http.StatusBadRequest, err)
			return
		}
		if err := r.ReadEntity(&pos); err != nil {
			writeErrorWithLog(w, http.StatusBadRequest, err)
			return
		}
		seaobject.Pos = pos
		if err := control.Update(index, seaobject); err != nil {
			writeErrorWithLog(w, http.StatusBadRequest, err)
			return
		}
		if err := control.SetupNewPosition(index, 0); err != nil {
			writeErrorWithLog(w, http.StatusInternalServerError, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (*serviceSeaObject) DeleteRoute(control *object.Control) restful.RouteFunction {
	return func(r *restful.Request, w *restful.Response) {
		var (
			seaobject object.Object
			index     int
			err       error
		)
		if index, err = strconv.Atoi(r.PathParameter("target-id")); err != nil || index < 1 {
			writeErrorWithLog(w, http.StatusBadRequest, errors.New("bad id"))
			return
		}
		if seaobject, err = control.GetByIndex(index); err != nil {
			writeErrorWithLog(w, http.StatusBadRequest, err)
			return
		}
		seaobject.Pos = make([]object.PositionArrivalTime, 0)
		seaobject.SetCurPos(object.PositionArrivalTime{})
		if err := control.Update(index, seaobject); err != nil {
			writeErrorWithLog(w, http.StatusBadRequest, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (*serviceSeaObject) DeleteTarget(control *object.Control) restful.RouteFunction {
	return func(r *restful.Request, w *restful.Response) {
		var (
			index int
			err   error
		)
		if index, err = strconv.Atoi(r.PathParameter("target-id")); err != nil || index < 1 {
			writeErrorWithLog(w, http.StatusBadRequest, errors.New("bad id"))
			return
		}
		if err = control.Delete(index); err != nil {
			writeErrorWithLog(w, http.StatusBadRequest, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func writeErrorWithLog(w *restful.Response, httpCode int, err error) {
	if err := w.WriteError(httpCode, err); err != nil {
		log.Error(err.Error())
	}
}
