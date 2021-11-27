package messages

import (
	"fmt"
	"prisma/tms/omnicom"
	"strconv"
)

func selectMessageToAsk(action omnicom.OmnicomConfiguration_Action) (ask uint32, err error) {
	switch action {
	case omnicom.OmnicomConfiguration_RequestAlertReport:
		return 0x00, nil
	case omnicom.OmnicomConfiguration_RequestLastPositionRecorded:
		return 0x01, nil
	case omnicom.OmnicomConfiguration_RequestNewPositionAquisition:
		return 0x02, nil
	case omnicom.OmnicomConfiguration_RequestGlobalParameters:
		return 0x03, nil
	default:
		return ask, fmt.Errorf("Request action not supported")
	}
}

func getGeoFenceMsgID(omn *omnicom.Omni) (id string, err error) {
	if omn != nil {
		if omn.GetUgpolygon() != nil {
			return strconv.Itoa(int(omn.GetUgpolygon().Msg_ID)), nil
		}
		return strconv.Itoa(int(omn.GetUgcircle().Msg_ID)), nil
	}
	return id, fmt.Errorf("omnicom structure is nil")
}
