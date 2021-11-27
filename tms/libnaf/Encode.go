package libnaf

import (
	"fmt"
	"prisma/tms"
	"reflect"
)

//EncodeNaf iridium proto structures into strings
func EncodeNaf(activity *tms.MessageActivity) (string, error) {

	str := NafHeader
	omni := activity.GetOmni()

	if activity.Imei == nil {
		return "", fmt.Errorf("Imei unavailable: %+v", activity)
	}

	if omni == nil {
		return "", fmt.Errorf("Omnicom data unavailable %+v", activity)
	}

	str = str + "//IMEI/" + activity.Imei.Value

	if omni.GetSpr() != nil {

		str = str + encodeSPR(omni.GetSpr())
		return str, nil
	}

	if omni.GetAr() != nil {

		str = str + encodeAR(omni.GetAr())
		return str, nil
	}

	if omni.GetGp() != nil {
		str = str + encodeGP(omni.GetGp())
		return str, nil
	}

	if omni.GetHpr() != nil {
		str = str + encodeHPR(omni.GetHpr())
		return str, nil
	}

	if omni.GetBm() != nil {
		str = str + encodeBM(omni.GetBm())
		return str, nil
	}

	if omni.GetAbm() != nil {
		str = str + encodeABM(omni.GetAbm())
		return str, nil
	}

	if omni.GetGa() != nil {
		str = str + encodeGFA(omni.GetGa())
		return str, nil
	}

	if omni.GetAa() != nil {
		str = str + encodeAA(omni.GetAa())
		return str, nil
	}

	if omni.GetAup() != nil {
		str = str + encodeAUP(omni.GetAup())
		return str, nil
	}

	if omni.GetBmstov() != nil {
		str = str + encodeBMSV(omni.GetBmstov())
		return str, nil
	}

	if omni.GetDg() != nil {
		str = str + encodeDGFS(omni.GetDg())
		return str, nil
	}

	if omni.GetGbmn() != nil {
		str = str + encodeGBMN(omni.GetGbmn())
		return str, nil
	}

	if omni.GetRmh() != nil {
		str = str + encodeRMH(omni.GetRmh())
		return str, nil
	}

	if omni.GetRsm() != nil {
		str = str + encodeRSM(omni.GetRsm())
		return str, nil
	}

	if omni.GetTma() != nil {
		str = str + encodeTMA(omni.GetTma())
		return str, nil
	}

	if omni.GetUaup() != nil {
		str = str + encodeUAUP(omni.GetUaup())
		return str, nil
	}

	if omni.GetUgpolygon() != nil {
		str = str + encodeUGFP(omni.GetUgpolygon())
		return str, nil
	}

	if omni.GetUgcircle() != nil {
		str = str + encodeUGFC(omni.GetUgcircle())
		return str, nil
	}

	if omni.GetUgp() != nil {
		str = str + encodeUGP(omni.GetUgp())
		return str, nil
	}

	if omni.GetUic() != nil {
		str = str + encodeUIC(omni.GetUic())
		return str, nil
	}
	return "", fmt.Errorf("Sentence type %v not implemented", reflect.TypeOf(omni))
}
