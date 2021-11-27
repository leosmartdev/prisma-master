package messages

import (
	"fmt"
	"strconv"

	"prisma/tms/db"
	"prisma/tms/omnicom"
	omn "prisma/tms/util/omnicom"

	"prisma/tms/db/mongo"

	"github.com/globalsign/mgo/bson"
)

// OmnicomConfigMulticast ...
type OmnicomConfigMulticast struct {
	val *omnicom.OmnicomConfiguration
}

// NewOmnicomConfigMulticast is an OmnicomConfigMulticast constructor
func NewOmnicomConfigMulticast(val *omnicom.OmnicomConfiguration) *OmnicomConfigMulticast {
	return &OmnicomConfigMulticast{
		val: val,
	}
}

// GetMessage returns Omni structure, omnicom message id, omnicom message type description, and error
func (o *OmnicomConfigMulticast) GetMessage(misc db.MiscDB) (*omnicom.Omni, string, string, error) {
	action := omnicom.OmnicomConfiguration_Action_name[int32(o.val.Action)]
	switch o.val.Action {
	case omnicom.OmnicomConfiguration_UnitIntervalChange:
		uic := omn.NewPbUic(o.val.PositionReportingInterval)
		return uic, strconv.Itoa(int(uic.GetUic().ID_Msg)), action, nil
	case omnicom.OmnicomConfiguration_RequestGlobalParameters,
		omnicom.OmnicomConfiguration_RequestLastPositionRecorded,
		omnicom.OmnicomConfiguration_RequestNewPositionAquisition,
		omnicom.OmnicomConfiguration_RequestAlertReport:
		ask, err := selectMessageToAsk(o.val.Action)
		if err != nil {
			return nil, "", action, err
		}
		rsm := omn.NewPbRsm(ask)
		return rsm, strconv.Itoa(int(rsm.GetRsm().ID_Msg)), action, nil
	case omnicom.OmnicomConfiguration_AckAssistanceAlert:
		return omn.NewPbAa(), bson.NewObjectId().Hex(), action, nil
	case omnicom.OmnicomConfiguration_UploadGeofence:
		zdb := mongo.NewMongoZoneMiscData(misc)
		if zdb != nil {
			zid, err := strconv.Atoi(o.val.ZoneId)
			if err != nil {
				return nil, "", action, err
			}
			mz, err := zdb.GetOne(uint32(zid))
			if err != nil {
				return nil, "", action, err
			}
			stg := omnicom.Stg{
				New_Position_Report_Period: o.val.PositionReportingInterval,
				Speed_Threshold:            o.val.SpeedThreshold,
			}
			geof := omn.NewPbUgf(mz, uint32(o.val.Priority), o.val.Activated, stg)
			id, err := getGeoFenceMsgID(geof)
			if err != nil {
				return nil, "", action, err
			}
			return geof, id, action, nil
		}
		return nil, "", action, fmt.Errorf("Could not recover zone mongo db client")
	case omnicom.OmnicomConfiguration_DeleteGeofence:
		zid, err := strconv.Atoi(o.val.ZoneId)
		if err != nil {
			return nil, "", action, err
		}
		dgf := omn.NewPbDgf(uint32(zid))
		return dgf, strconv.Itoa(int(dgf.GetDg().Msg_ID)), action, nil
	default:
		return nil, "", action, fmt.Errorf("Action %+v in not implemented", o.val.Action)
	}
}
