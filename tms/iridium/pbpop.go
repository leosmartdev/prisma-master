package iridium

import "prisma/tms/omnicom"

func PopulateMOProtobuf(MOH MOHeader, MOP MPayload) (*Iridium, error) {

	sen := &Iridium{}

	pbMO := MobileOriginatedHeader{}

	pbMO.MO_IEI = append(pbMO.MO_IEI, MOH.MO_IEI)

	pbMO.MOHL = uint32(MOH.MOHL)

	pbMO.CDR = MOH.CDR

	pbMO.IMEI = MOH.IMEI

	pbMO.SessStatus = MOH.SessStatus

	pbMO.MOMSN = uint32(MOH.MOMSN)

	pbMO.MTMSN = uint32(MOH.MTMSN)

	pbMO.TimeOfSession = MOH.TimeOfSession

	sen.Moh = &pbMO

	pbPayload := Payload{}

	pbPayload.IEI = append(pbPayload.IEI, MOP.IEI)

	pbPayload.PayloadL = uint32(MOP.PayloadL)

	Omnicom, err := omnicom.PopulateProtobuf(MOP.Omn)

	if err != nil {
		return nil, err
	}

	pbPayload.Omnicom = Omnicom

	sen.Payload = &pbPayload

	return sen, nil

}

func PopulateLocationProtobuf(MOH MOHeader, Location MOLocationInformation) (*Iridium, error) {

	sen := &Iridium{}

	pbMO := MobileOriginatedHeader{}

	pbMO.MO_IEI = append(pbMO.MO_IEI, MOH.MO_IEI)

	pbMO.MOHL = uint32(MOH.MOHL)

	pbMO.CDR = MOH.CDR

	pbMO.IMEI = MOH.IMEI

	pbMO.SessStatus = MOH.SessStatus

	pbMO.MOMSN = uint32(MOH.MOMSN)

	pbMO.MTMSN = uint32(MOH.MTMSN)

	pbMO.TimeOfSession = MOH.TimeOfSession

	sen.Moh = &pbMO

	pbLocation := MobileOriginatedLocationInformation{}

	pbLocation.IEI = append(pbLocation.IEI, Location.IEI)

	pbLocation.Length = uint32(Location.Length)

	pbLocation.Latitude = Location.Latitude

	pbLocation.Longitude = Location.Longitude

	pbLocation.CEP = Location.CEP

	sen.Moli = &pbLocation

	return sen, nil
}

func PopulateMobileTerminatedConfirmationProtobuf(MTC MTConfirmation) (*Iridium, error) {

	sen := &Iridium{}

	pbMTC := MessageTerminatedConfirmation{}

	pbMTC.IEI = append(pbMTC.IEI, MTC.IEI)

	pbMTC.MTCL = uint32(MTC.MTCL)

	pbMTC.UCM_IF = MTC.UCM_IF

	pbMTC.IMEI = MTC.IMEI

	pbMTC.Auto_IDR = MTC.Auto_IDR

	pbMTC.MT_Message_Status = int32(MTC.MT_Message_Status)

	sen.Mtc = &pbMTC

	return sen, nil
}

func PopulateProtobufToMobileTerminated(MT *Iridium) (MTHeader, MPayload, error) {

	var MTH MTHeader

	MTH.IEI = MT.Mth.MO_IEI[0]

	MTH.MTHL = uint16(MT.Mth.MTHL)

	MTH.UniqueClientMessageID = MT.Mth.UniqueClientMessageID

	MTH.IMEI = MT.Mth.IMEI

	MTH.MTflag = uint16(MT.Mth.MTflag)

	var MTP MPayload

	MTP.IEI = MT.Payload.IEI[0]

	MTP.PayloadL = uint16(MT.Payload.PayloadL)

	var err error

	MTP.Omn, err = omnicom.PopulateOmnicom(MT.Payload.Omnicom)

	if err != nil {

		return MTHeader{}, MPayload{}, err
	}

	return MTH, MTP, nil

}

func PopulateMTProtobuf(MTH MTHeader, MOP MPayload) (*Iridium, error) {

	sen := &Iridium{}

	pbMT := MobileTerminatedHeader{}
	pbMT.MO_IEI = append(pbMT.MO_IEI, MTH.IEI)
	pbMT.MTHL = uint32(MTH.MTHL)
	pbMT.UniqueClientMessageID = MTH.UniqueClientMessageID
	pbMT.IMEI = MTH.IMEI
	pbMT.MTflag = uint32(MTH.MTflag)

	sen.Mth = &pbMT

	pbPayload := Payload{}
	pbPayload.IEI = append(pbPayload.IEI, MOP.IEI)
	pbPayload.PayloadL = uint32(MOP.PayloadL)

	Omnicom, err := omnicom.PopulateProtobuf(MOP.Omn)
	if err != nil {
		return nil, err
	}
	pbPayload.Omnicom = Omnicom

	sen.Payload = &pbPayload

	return sen, nil
}
