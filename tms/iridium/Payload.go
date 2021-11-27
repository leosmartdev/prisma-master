// Package iridium provides functions to parse MT MO messages.
package iridium

import (
	"fmt"
	"prisma/tms/omnicom"
	"reflect"
	"errors"
)

type MPayload struct {
	IEI byte

	PayloadL uint16

	Omn omnicom.Omnicom
}

// ParserError it is used to include more information into the error message
type ParserError struct {
	msg string
}

func (p ParserError) Error() string {
	return "libomnicom error: " + p.msg
}

// ErrBadLengthPayload is used to issue error about wrong length of payload
var ErrBadLengthPayload = errors.New("data payload length does not match field Payload Length value")

func ParseMOPayload(data []byte) (MPayload, error) {

	var payload MPayload

	if len(data) == 0 {
		return payload, fmt.Errorf("empty payload")
	}

	payload.IEI = data[0]

	payload.PayloadL = uint16(data[2]) | uint16(data[1])<<8

	if payload.PayloadL != uint16(len(data)-3) {

		return payload, ErrBadLengthPayload
	}

	om, err := omnicom.Parse(data[3: payload.PayloadL+3])
	if err != nil {
		return payload, ParserError{err.Error()}
	}
	payload.Omn = om

	return payload, nil
}

func (mt MPayload) Encode() ([]byte, error) {

	var raw []byte

	raw = append(raw, mt.IEI)

	if mt.Omn == nil {
		return []byte{}, fmt.Errorf("MT omnicom payload is empty")
	}

	if reflect.TypeOf(mt.Omn).Elem().String() == "omnicom.RMH" {

		bytes, err := mt.Omn.(*omnicom.RMH).Encode()

		if err != nil {
			return []byte{}, err
		}

		mt.PayloadL = uint16(len(bytes))

		raw = append(raw, byte(mt.PayloadL>>8), byte(mt.PayloadL))

		raw = append(raw, bytes...)

	} else if reflect.TypeOf(mt.Omn).Elem().String() == "omnicom.UIC" {

		bytes, err := mt.Omn.(*omnicom.UIC).Encode()

		if err != nil {

			return []byte{}, err
		}

		mt.PayloadL = uint16(len(bytes))

		raw = append(raw, byte(mt.PayloadL>>8), byte(mt.PayloadL))

		raw = append(raw, bytes...)

	} else if reflect.TypeOf(mt.Omn).Elem().String() == "omnicom.RSM" {

		bytes, err := mt.Omn.(*omnicom.RSM).Encode()

		if err != nil {

			return []byte{}, err
		}

		mt.PayloadL = uint16(len(bytes))

		raw = append(raw, byte(mt.PayloadL>>8), byte(mt.PayloadL))

		raw = append(raw, bytes...)

	} else if reflect.TypeOf(mt.Omn).Elem().String() == "omnicom.UGP" {

		bytes, err := mt.Omn.(*omnicom.UGP).Encode()

		if err != nil {

			return []byte{}, err
		}

		mt.PayloadL = uint16(len(bytes))

		raw = append(raw, byte(mt.PayloadL>>8), byte(mt.PayloadL))

		raw = append(raw, bytes...)

	} else if reflect.TypeOf(mt.Omn).Elem().String() == "omnicom.UAUP" {

		bytes, err := mt.Omn.(*omnicom.UAUP).Encode()

		if err != nil {

			return []byte{}, err
		}

		mt.PayloadL = uint16(len(bytes))

		raw = append(raw, byte(mt.PayloadL>>8), byte(mt.PayloadL))

		raw = append(raw, bytes...)

	} else if reflect.TypeOf(mt.Omn).Elem().String() == "omnicom.UG_Polygon" {

		bytes, err := mt.Omn.(*omnicom.UG_Polygon).Encode()

		if err != nil {

			return []byte{}, err
		}

		mt.PayloadL = uint16(len(bytes))

		raw = append(raw, byte(mt.PayloadL>>8), byte(mt.PayloadL))

		raw = append(raw, bytes...)

	} else if reflect.TypeOf(mt.Omn).Elem().String() == "omnicom.UG_Circle" {

		bytes, err := mt.Omn.(*omnicom.UG_Circle).Encode()

		if err != nil {

			return []byte{}, err
		}

		mt.PayloadL = uint16(len(bytes))

		raw = append(raw, byte(mt.PayloadL>>8), byte(mt.PayloadL))

		raw = append(raw, bytes...)

	} else if reflect.TypeOf(mt.Omn).Elem().String() == "omnicom.DG" {

		bytes, err := mt.Omn.(*omnicom.DG).Encode()

		if err != nil {

			return []byte{}, err
		}

		mt.PayloadL = uint16(len(bytes))

		raw = append(raw, byte(mt.PayloadL>>8), byte(mt.PayloadL))

		raw = append(raw, bytes...)

	} else if reflect.TypeOf(mt.Omn).Elem().String() == "omnicom.SPR" {

		bytes, err := mt.Omn.(*omnicom.SPR).Encode()

		if err != nil {

			return []byte{}, err
		}

		mt.PayloadL = uint16(len(bytes))

		raw = append(raw, byte(mt.PayloadL>>8), byte(mt.PayloadL))

		raw = append(raw, bytes...)

	} else if reflect.TypeOf(mt.Omn).Elem().String() == "omnicom.AR" {

		bytes, err := mt.Omn.(*omnicom.AR).Encode()

		if err != nil {

			return []byte{}, err
		}

		mt.PayloadL = uint16(len(bytes))

		raw = append(raw, byte(mt.PayloadL>>8), byte(mt.PayloadL))

		raw = append(raw, bytes...)

	} else if reflect.TypeOf(mt.Omn).Elem().String() == "omnicom.AA" {

		bytes, err := mt.Omn.(*omnicom.AA).Encode()

		if err != nil {

			return []byte{}, err
		}

		mt.PayloadL = uint16(len(bytes))

		raw = append(raw, byte(mt.PayloadL>>8), byte(mt.PayloadL))

		raw = append(raw, bytes...)

	} else if reflect.TypeOf(mt.Omn).Elem().String() == "omnicom.TMA" {

		bytes, err := mt.Omn.(*omnicom.TMA).Encode()

		if err != nil {

			return []byte{}, err
		}

		mt.PayloadL = uint16(len(bytes))

		raw = append(raw, byte(mt.PayloadL>>8), byte(mt.PayloadL))

		raw = append(raw, bytes...)

	} else {
		return nil, fmt.Errorf("MT data payload %s sentence not yet implemented.\n", reflect.TypeOf(mt.Omn).Elem().String())
	}

	return raw, nil
}
