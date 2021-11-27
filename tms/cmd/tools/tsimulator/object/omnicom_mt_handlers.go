// omnicom_mt_handlers is used for placing functions to handle MT message tasks
package object

import (
	"prisma/tms/cmd/tools/tsimulator/task"
	"prisma/tms/log"
	"prisma/tms/omnicom"
	"prisma/tms/tmsg"
	"time"
)

func (o *Omnicom) requestGlobalParameters(t *omnicom.RSM) {
	raw, err := o.makeBuffGlobalParameters(t.ID_Msg)
	if err != nil {
		log.Error(err.Error())
		return
	}
	o.handledCh <- task.Result{
		Data: raw,
		To:   tmsg.APP_ID_TFLEETD,
	}
}

func (o *Omnicom) requestAlertReport(t *omnicom.RSM) {
	// we set start type alerting to last type alert in order to get the last alert status of the beacon in the simulator
	o.startTypeAlerting = LastTypeAlerting
	raw, err := o.makeBuffStartAlerting(t.ID_Msg)
	if err != nil {
		log.Error(err.Error())
		return
	}
	o.handledCh <- task.Result{
		Data: raw,
		To:   tmsg.APP_ID_TFLEETD,
	}
}

func (o *Omnicom) sendLastPositionReport(t *omnicom.RSM) {
	s, ok := o.object.historyPosition.Prev().Value.(*PositionSpeedTime)
	if !ok {
		log.Error("could not assert %+v to PositionSpeedTime", o.object.historyPosition.Prev)
		return
	}
	raw, err := o.makeBuffRmh([]*PositionSpeedTime{s}, 1, t.ID_Msg)
	if err != nil {
		log.Error(err.Error())
		return
	}
	o.handledCh <- task.Result{
		Data: raw,
		To:   tmsg.APP_ID_TFLEETD,
	}
}

func (o *Omnicom) sendCurrentPositionReport(t *omnicom.RSM) {
	s, ok := o.object.historyPosition.Value.(*PositionSpeedTime)
	if !ok {
		log.Error("could not assert %+v to PositionSpeedTime", o.object.historyPosition.Prev)
		return
	}
	raw, err := o.makeBuffRmh([]*PositionSpeedTime{s}, 1, t.ID_Msg)
	if err != nil {
		log.Error(err.Error())
		return
	}
	o.handledCh <- task.Result{
		Data: raw,
		To:   tmsg.APP_ID_TFLEETD,
	}
}

func (o *Omnicom) globalParametersMessage(t *omnicom.UGP) {
	raw, err := o.makeBuffGlobalParameters(t.ID_Msg)
	if err != nil {
		log.Error(err.Error())
		return
	}
	o.handledCh <- task.Result{
		Data: raw,
		To:   tmsg.APP_ID_TFLEETD,
	}
}

func (o *Omnicom) unitIntervalChangeMessage(t *omnicom.UIC) {
	raw, err := o.makeBuffReportingInterval(t.New_Reporting, t.ID_Msg)
	if err != nil {
		log.Error(err.Error())
		return
	}
	o.handledCh <- task.Result{
		Data: raw,
		To:   tmsg.APP_ID_TFLEETD,
	}
}

// Get geofence request. Function should answer geofencing ACK 0x04(from the documentation)
func (o *Omnicom) geofenceMessage(t *omnicom.UG_Polygon) {
	o.mgfence.Lock()
	defer o.mgfence.Unlock()
	var tpg timerPositionGeofence
	for _, pos := range t.Position {
		tpg.p = append(tpg.p, Position{
			Latitude:  float64(pos.Latitude),
			Longitude: float64(pos.Longitude),
		})
	}
	if t.Setting.New_Position_Report_Period > 0 {
		tpg.t = time.NewTicker(time.Duration(t.Setting.New_Position_Report_Period) * unitForGeoFenceReportTime)
	} else {
		tpg.t = time.NewTicker(o.defaultTimeTimer)
	}
	// each messages have GEO_ID we should replace data with the same ID
	o.geoFences[t.GEO_ID] = tpg
	raw, err := o.makeBuffGeofenceAck()
	if err != nil {
		log.Error(err.Error())
	}
	o.handledCh <- task.Result{
		Data: raw,
		To:   tmsg.APP_ID_TFLEETD,
	}
}

func (o *Omnicom) rmhMessage(t *omnicom.RMH) {
	dt := t.Date_Interval
	dstart := time.Date(int(dt.Start.Year)+2000, time.Month(dt.Start.Month),
		int(dt.Start.Day), 0, int(dt.Start.Minute), 0, 0, time.UTC)
	dstop := time.Date(int(dt.Stop.Year)+2000, time.Month(dt.Stop.Month),
		int(dt.Stop.Day), 0, int(dt.Stop.Minute), 0, 0, time.UTC)
	pt := o.object.GetHistoryFromTo(dstart, dstop)
	msgID := time.Now().Nanosecond() % 4096
	var queue []*PositionSpeedTime
	for i := range pt {
		queue = append(queue, pt[i])
		if !(i == len(pt)-1 || i%sizePositionPerSocket == 0 && i != 0) {
			continue
		}
		raw, err := o.makeBuffRmh(queue, uint32(len(pt)), uint32(msgID))
		if err != nil {
			log.Error(err.Error())
			continue
		}
		o.handledCh <- task.Result{
			Data: raw,
			To:   tmsg.APP_ID_TFLEETD,
		}
		queue = queue[:0]
	}
}
