package public

import (
	"net/http"
	"prisma/gogroup"
	"prisma/tms"
	api "prisma/tms/client_api"
	"prisma/tms/db"
	"prisma/tms/log"
	"prisma/tms/moc"
	"prisma/tms/security"

	restful "github.com/orolia/go-restful" 
	"github.com/gorilla/websocket"
)

const VIEW_CLASSID = "View"

type ViewRest struct {
	db    db.FeatureDB
	group gogroup.GoGroup
}

type ViewResponse struct {
	ID string
}

func NewViewRest(db db.FeatureDB, group gogroup.GoGroup) *ViewRest {
	return &ViewRest{
		db:    db,
		group: group,
	}
}

func (r *ViewRest) Create(request *restful.Request, response *restful.Response) {
	ACTION := moc.View_CREATE.String()
	if !security.HasPermissionForAction(request.Request.Context(), VIEW_CLASSID, ACTION) {
		security.Audit(request.Request.Context(), VIEW_CLASSID, ACTION, "FAIL")
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
		return
	}

	view, err := r.db.CreateView(&api.ViewRequest{})
	if err != nil {
		security.Audit(request.Request.Context(), VIEW_CLASSID, ACTION, "FAIL_ERROR")
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	vresp := ViewResponse{ID: view.ID()}
	security.Audit(request.Request.Context(), VIEW_CLASSID, ACTION, "SUCCESS")
	response.WriteEntity(vresp)
}

type ViewWS struct {
	db    db.FeatureDB
	group gogroup.GoGroup
	trace *log.Tracer
}

type Extent struct {
	MinLat float64 `json:"minLat"`
	MinLon float64 `json:"minLon"`
	MaxLat float64 `json:"maxLat"`
	MaxLon float64 `json:"maxLon"`
}

type History struct {
	TrackID    string `json:"trackId"`
	RegistryID string `json:"registryId"`
	Duration   int    `json:"duration"`
	ClearAll   bool   `json:"clearAll"`
}

type UpdateStreamRequest struct {
	ViewID  string   `json:"viewId"`
	Extent  *Extent  `json:"extent"`
	History *History `json:"history"`
}

func NewViewWS(db db.FeatureDB, group gogroup.GoGroup) *ViewWS {
	return &ViewWS{
		db:    db,
		group: group,
		trace: log.GetTracer("ws"),
	}
}

func (w *ViewWS) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	ACTION := moc.View_STREAM.String()
	// TODO: Add auth check
	ctxt := w.group.Child("")
	wsUpgrader := websocket.Upgrader{
		ReadBufferSize:  64, // * 1024,
		WriteBufferSize: 64, // * 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := wsUpgrader.Upgrade(response, request, nil)
	if err != nil {
		security.Audit(request.Context(), VIEW_CLASSID, ACTION, "FAIL_ERROR")
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	in := make(chan api.StreamRequest, 16)
	go w.handleIn(ctxt, conn, in)
	out, err := w.db.Stream(ctxt, in)
	if err != nil {
		log.Error("unable to open stream: %v", err)
		security.Audit(request.Context(), VIEW_CLASSID, ACTION, "FAIL_ERROR")
		ctxt.Cancel(err)
		return
	}
	go w.handleOut(ctxt, conn, out)
}

func (w *ViewWS) handleIn(ctxt gogroup.GoGroup, conn *websocket.Conn, in chan<- api.StreamRequest) {
	defer func() {
		w.trace.Logf("closed input handler: %v", conn.RemoteAddr())
	}()
	for {
		var r UpdateStreamRequest
		if err := conn.ReadJSON(&r); err != nil {
			log.Error("socket read error: %v", err)
			ctxt.Cancel(err)
			return
		}
		sr := api.StreamRequest{ViewId: r.ViewID}
		if r.Extent != nil {
			sr.Bounds = &tms.BBox{
				Max: &tms.Point{
					Latitude:  r.Extent.MaxLat,
					Longitude: r.Extent.MaxLon,
				},
				Min: &tms.Point{
					Latitude:  r.Extent.MinLat,
					Longitude: r.Extent.MinLon,
				},
			}
		}
		if r.History != nil {
			sr.History = &api.HistoryRequest{
				TrackId:    r.History.TrackID,
				RegistryId: r.History.RegistryID,
				History:    uint64(r.History.Duration),
				ClearAll:   r.History.ClearAll,
			}
		}
		in <- sr
	}
}

func (w *ViewWS) handleOut(ctxt gogroup.GoGroup, conn *websocket.Conn, out <-chan db.FeatureUpdate) {
	defer func() {
		w.trace.Logf("closed output handler: %v", conn.RemoteAddr())
	}()
	for {
		select {
		case update := <-out:
			if err := conn.WriteJSON(update); err != nil {
				log.Error("socket write error: %v", err)
				ctxt.Cancel(err)
				return
			}
		case <-ctxt.Done():
			return
		}
	}
}
