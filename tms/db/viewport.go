package db

import (
	. "prisma/tms"
	. "prisma/tms/client_api"
	"prisma/tms/feature"
	"prisma/tms/log"

	"prisma/gogroup"

	"fmt"
	"math"
	"reflect"
	"strings"
	"sync"
	"time"

	geo "prisma/tms/geojson"
)

const (
	DefaultStartHeatmapCount = 2500
	DefaultStopHeatmapCount  = DefaultStartHeatmapCount - 100
	TickInterval             = time.Duration(1) * time.Second
	CellCount                = 100.0
)

type Bounds struct {
	BBox           geo.BBox
	width          float64
	height         float64
	cellWidth      float64
	cellHeight     float64
	cellWidthHalf  float64
	cellHeightHalf float64
}

func (b Bounds) String() string {
	return fmt.Sprintf("(%v %v, %v %v) w=%v h=%v cw=%v ch=%v", b.BBox.Min.X,
		b.BBox.Min.Y, b.BBox.Max.X, b.BBox.Max.Y, b.width, b.height,
		b.cellWidth, b.cellHeight)
}

func NewBounds(bbox geo.BBox, cellCount int) Bounds {
	b := Bounds{BBox: bbox}
	b.width = bbox.Max.X - bbox.Min.X
	b.height = bbox.Max.Y - bbox.Min.Y
	b.cellWidth = b.width / float64(cellCount)
	b.cellHeight = b.height / float64(cellCount)
	b.cellWidthHalf = b.cellWidth / 2.0
	b.cellHeightHalf = b.cellHeight / 2.0
	return b
}

type Viewport struct {
	Count         int // Number of tracks visible in the viewport
	Total         int // Number of tracks being processed by the system
	bounds        Bounds
	f             *featuresView
	Visible       map[interface{}]bool // A feature (tracks, zones etc.) ID not in the map is not being processed
	info          *log.Tracer
	details       *log.Tracer
	out           chan<- FeatureUpdate
	Ctxt          gogroup.GoGroup
	heatmapActive bool
	timer         *time.Timer
	mutex         sync.Mutex
	prevCount     int
	prevTotal     int
}

func NewViewport(ctxt gogroup.GoGroup, f *featuresView, out chan<- FeatureUpdate) *Viewport {
	v := &Viewport{Ctxt: ctxt, f: f, out: out}
	v.Visible = make(map[interface{}]bool)
	v.info = log.GetTracer("viewport")
	v.details = log.GetTracer("viewport-details")
	v.info.Log("created")
	return v
}

func (v *Viewport) SetBounds(bounds geo.BBox) {
	v.info.Logf("setting bounds: %v", bounds)
	v.mutex.Lock()
	defer v.mutex.Unlock()
	v.bounds = NewBounds(bounds, CellCount)

	if v.heatmapActive {
		v.sendHeatmap()
	} else {
		v.revalidateAll()
	}
	v.rescheduleTimer()
}

func (v *Viewport) Process(update *FeatureUpdate) {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	v.process(update)
}

func (v *Viewport) process(update *FeatureUpdate) {
	if v.heatmapActive && inHeatmap(update.Feature) {
		return
	}
	if update.Status == Status_Current {
		v.update(update)
	} else if update.Status == Status_Timeout {
		v.remove(update)
	} else {
		log.Warn("Unknown status: %v", update.Status)
	}
}

func (v *Viewport) revalidateAll() {
	iterator := func(id string, obj geo.Object, fields []float64) bool {
		feature := v.f.toFeature(obj)
		if !v.heatmapActive {
			v.revalidate(feature)
		}
		return true
	}
	v.f.features.Scan(0, false, iterator)
	v.info.Log(v.Status())
}

func (v *Viewport) revalidate(f *feature.F) {
	was, ok := v.Visible[f.ID]
	if !ok {
		// If this feature has not been seen yet, process it like a normal
		// update
		v.process(&FeatureUpdate{
			Status:  Status_Current,
			Feature: f,
		})
		return
	}
	now := f.IntersectsBBox(v.bounds.BBox)
	if was && !now {
		if shouldCount(f) {
			v.decrement()
		}
		v.send(&FeatureUpdate{
			Status:  Status_LeftGeoRange,
			Feature: f,
		})
	} else if !was && now {
		if shouldCount(f) {
			v.increment()
		}
		v.send(&FeatureUpdate{
			Status:  Status_Current,
			Feature: f,
		})
	}
	v.Visible[f.ID] = now
}

func (v *Viewport) Status() string {
	return fmt.Sprintf("features: %v, visible: %v", v.f.features.Count(), v.Count)
}

func (v *Viewport) update(update *FeatureUpdate) {
	was, ok := v.Visible[update.Feature.ID]
	if !ok {
		v.add(update)
		return
	}
	feature := update.Feature
	now := feature.IntersectsBBox(v.bounds.BBox)
	v.Visible[update.Feature.ID] = now
	v.details.Logf("updating %v: %v => %v", update.Feature.ID, was, now)
	if was && !now {
		v.send(update)
		if shouldCount(update.Feature) {
			v.decrement()
		}
		update.Status = Status_LeftGeoRange
	} else if !was && now {
		v.send(update)
		if shouldCount(update.Feature) {
			v.increment()
		}
	} else if now {
		v.send(update)
	}
}

func (v *Viewport) add(update *FeatureUpdate) {
	feature := update.Feature
	vis := feature.IntersectsBBox(v.bounds.BBox)
	v.details.Logf("adding %v: is %v", update.Feature.ID, vis)
	v.Visible[update.Feature.ID] = vis
	if shouldCount(update.Feature) {
		v.Total++
	}
	if vis {
		v.send(update)
		if shouldCount(update.Feature) {
			v.increment()
		}
	}
}

func (v *Viewport) remove(update *FeatureUpdate) {
	was, ok := v.Visible[update.Feature.ID]
	v.details.Logf("removing %v: was %v", update.Feature.ID, was)
	if ok && was {
		v.send(update)
		if shouldCount(update.Feature) {
			v.decrement()
		}
	}
	delete(v.Visible, update.Feature.ID)
	if ok && shouldCount(update.Feature) {
		v.Total--
	}
}

func (v *Viewport) send(update *FeatureUpdate) {
	update.Counts = &FeatureCounts{
		Total:   v.Total,
		Visible: v.Count,
	}
	v.prevTotal = v.Total
	v.prevCount = v.Count
	select {
	case <-v.Ctxt.Done():
		return
	case v.out <- *update:
	}
}

func (v *Viewport) increment() {
	v.changeCount(1)
}

func (v *Viewport) decrement() {
	v.changeCount(-1)
}

func (v *Viewport) changeCount(value int) {
	v.Count += value
	if v.Count >= v.f.startHeatmapCount && !v.heatmapActive {
		v.info.Logf("start heatmap, %v >= %v", v.Count, v.f.startHeatmapCount)
		v.heatmapActive = true
		v.send(&FeatureUpdate{
			Status: Status_HeatmapStart,
		})
		v.sendHeatmap()
	}
}

func (v *Viewport) stopHeatmap() {
	v.heatmapActive = false
	for key := range v.Visible {
		v.Visible[key] = false
	}
	v.Count = 0
	v.send(&FeatureUpdate{
		Status: Status_HeatmapStop,
	})
	v.revalidateAll()
}

func (v *Viewport) sendHeatmap() {
	if !v.heatmapActive {
		return
	}
	cells, ok := v.heatmap()
	if !ok {
		v.stopHeatmap()
		return
	}
	pbcells := make([]*DensityCell, 0, len(cells))
	for cell, val := range cells {
		x, y := CellToPoint(cell, v.bounds)
		pbcell := &DensityCell{
			X:     x,
			Y:     y,
			Count: uint32(val),
		}
		pbcells = append(pbcells, pbcell)
	}
	heatmap := &Heatmap{Cells: pbcells}
	update := &FeatureUpdate{
		Status:  Status_Heatmap,
		Heatmap: heatmap,
	}
	v.send(update)
}

func (v *Viewport) sendCounts() {
	if v.Count == v.prevCount && v.Total == v.prevTotal {
		return
	}
	v.send(&FeatureUpdate{
		Status: Status_CountOnly,
	})
}

func (v *Viewport) tick() {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	if v.heatmapActive {
		v.sendHeatmap()
	} else {
		v.sendCounts()
	}
}

func (v *Viewport) rescheduleTimer() {
	if v.timer != nil {
		return
	}
	v.timer = time.NewTimer(TickInterval)
	v.Ctxt.Go(func() {
		for {
			select {
			case <-v.timer.C:
				v.tick()
			case <-v.Ctxt.Done():
				return
			}
		}
	})
}

func (v *Viewport) heatmap() (map[Cell]int, bool) {
	cells := make(map[Cell]int)
	total := 0
	iterator := func(_ string, obj geo.Object, _ []float64) bool {
		feature := v.f.toFeature(obj)
		if inHeatmap(feature) && feature.IntersectsBBox(v.bounds.BBox) {
			cell := PointToCell(feature.Geometry.(geo.Point), v.bounds)
			count := cells[cell]
			count++
			cells[cell] = count
			total++
		}
		return true
	}
	v.f.features.Scan(0, false, iterator)
	if total <= v.f.stopHeatmapCount {
		v.info.Logf("stop heatmap, %v <= %v", total, v.f.stopHeatmapCount)
		return nil, false
	}
	v.Count = total
	v.info.Logf("heatmap %v in %v cells", total, len(cells))
	return cells, true
}

func (v *Viewport) Cleanup() {
	if v.timer != nil {
		v.timer.Stop()
	}
}

type Cell struct {
	Row int
	Col int
}

func PointToCell(point geo.Point, bounds Bounds) Cell {
	x := point.Coordinates.X
	y := point.Coordinates.Y

	dx := x - bounds.BBox.Min.X
	dy := y - bounds.BBox.Min.Y

	col := int(math.Floor(dx / bounds.cellWidth))
	row := int(math.Floor(dy / bounds.cellHeight))

	return Cell{Row: row, Col: col}
}

func CellToPoint(cell Cell, bounds Bounds) (float64, float64) {
	x0 := bounds.BBox.Min.X
	y0 := bounds.BBox.Min.Y

	x := x0 + (float64(cell.Col) * bounds.cellWidth) + bounds.cellWidthHalf
	y := y0 + (float64(cell.Row) * bounds.cellHeight) + bounds.cellHeightHalf
	return x, y
}

func inHeatmap(f *feature.F) bool {
	// All AIS tracks should belong in the heatmap except for search and
	// rescue transmitters which have an MMSI that starts with 97
	value, ok := f.Properties["mmsi"]
	if ok {
		mmsi, ok := value.(string)
		if !ok {
			log.Warn("Expecting string for MMSI but got %v of type %v", value,
				reflect.TypeOf(value))
			return false
		}
		if strings.HasPrefix(mmsi, "97") {
			return false
		}
		return true
	}

	// Also aggregate all radar tracks
	value, ok = f.Properties["type"]
	if ok {
		ttype, ok := value.(string)
		if !ok {
			log.Warn("Expecting string for type but got %v of type %v", value,
				reflect.TypeOf(value))
			return false
		}
		if ttype == "track:Radar" {
			return true
		}
	}

	// Anything else doesn't belong in the heatmap
	return false
}

// Returns true if the GoFeature being processed should be counted in the heatmap/map count. Currently, only tracks
// will be counted
func shouldCount(f *feature.F) bool {
	value, ok := f.Properties["type"]
	ttype := ""
	if ok {
		ttype, ok = value.(string)
		if !ok {
			log.Warn("Expecting string for type but got %v of type %v", value,
				reflect.TypeOf(value))
			return false
		}
	} else {
		log.Warn("Expecting to find a type: %+v", f)
		return false
	}
	switch ttype {
	case "zones":
		return false
	}
	return true
}
