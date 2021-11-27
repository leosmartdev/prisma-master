package main

import (
	"net"
	"prisma/tms/cmd/tools/tsimulator/object"
	"prisma/tms/log"
	"prisma/tms/tmsg"
	"fmt"
)

// send data to application by ID
func sendDataToAppID(data []byte, to uint32) error {
	switch to {
	case tmsg.APP_ID_TFLEETD:
		conn, err := net.DialTCP("tcp", nil, tFleetdTCPAddr)
		if err != nil {
			return err
		}
		defer conn.Close()
		_, err = conn.Write(data)
		return err
	}
	return fmt.Errorf("The application %d is not found", to)
}

// Send a message to tmccd. It is used by sarat beacons
func sendDataToTmccd(obj object.ObjectCommunicator) (err error) {
	data, err := obj.GetPositionAlertingMessage()
	switch err {
	case object.ErrMessageNotImplement, object.ErrHolden:
		err = nil
	case nil:
		conn, errDial := net.DialTCP("tcp", nil, tmccdTCPAddr)
		if errDial != nil {
			return
		}
		defer conn.Close()
		_, errDial = conn.Write(data)
		if errDial != nil {
			return
		}
	default:
		log.Error("Got error in sendDataToTmccd: %v", err)
	}
	return
}

// It sends information about an object
func sendStaticInformation(conn net.Conn, obj object.Informer) (err error) {
	data, err := obj.GetMessageStaticInformation()
	switch err {
	case object.ErrMessageNotImplement:
		err = nil
	case nil:
		data = append(data, []byte("\r\n")...)
		if _, err = conn.Write(data); err != nil {
			log.Error("Error a connection: %v", err)
		}
	default:
		log.Error("Got error in sendStaticInformation: %v", err)
	}
	return
}

// It sends information about a current position of an object
func sendPositionInformation(conn net.Conn, obj object.Informer) (err error) {
	data, err := obj.GetMessagePosition()
	switch err {
	case object.ErrMessageNotImplement, object.ErrHolden:
		err = nil
	case nil:
		data = append(data, []byte("\r\n")...)
		if _, err = conn.Write(data); err != nil {
			log.Error("Error a connection: %v", err)
		}
	default:
		log.Error("Got error in sendStaticInformation: %v", err)
	}
	return
}

// It sends information about a current position and common data
func sendTrackedTargetMessage(conn net.Conn, obj object.ObjectCommunicator, latitude, longitude float64) (err error) {
	data, err := obj.GetTrackedTargetMessage(latitude, longitude)
	switch err {
	case object.ErrMessageNotImplement:
		err = nil
	case nil:
		data = append(data, []byte("\r\n")...)
		if _, err = conn.Write(data); err != nil {
			log.Error("Error a connection: %v", err)
		}
	default:
		log.Error("Got error in sendStaticInformation: %v", err)
	}
	return
}

// It sends information for starting alerting to tfleetd
func sendDataStartAlerting(obj object.ObjectCommunicator) (err error) {
	data, err := obj.GetStartAlertingMessage()
	switch err {
	case object.ErrMessageNotImplement:
		err = nil
	case nil:
		if data == nil {
			return
		}
		conn, errDial := net.DialTCP("tcp", nil, tFleetdTCPAddr)
		if errDial != nil {
			return
		}
		defer conn.Close()
		_, errDial = conn.Write(data)
		if errDial != nil {
			return
		}
	default:
		log.Error("Got error in sendDataStartAlerting: %v", err)
	}
	return
}

// It sends information for stopping alerting to tfleetd
func sendDataStopAlerting(obj object.ObjectCommunicator) (err error) {
	data, err := obj.GetStopAlertingMessage()
	switch err {
	case object.ErrMessageNotImplement:
		err = nil
	case nil:
		if data == nil {
			return
		}
		conn, errDial := net.DialTCP("tcp", nil, tFleetdTCPAddr)
		if errDial != nil {
			return
		}
		defer conn.Close()
		_, errDial = conn.Write(data)
		if errDial != nil {
			return
		}
	default:
		log.Error("Got error in sendDataStopAlerting: %v", err)
	}
	return
}

// It sends common information about an object to tfleetd. It is used by omnicom beacons
func sendDataToTFleetd(obj object.ObjectCommunicator) (err error) {
	data, err := obj.GetDataForIridiumNetwork()
	switch err {
	case object.ErrMessageNotImplement, object.ErrTooEarly, object.ErrHolden:
		err = nil
	case nil:
		conn, errDial := net.DialTCP("tcp", nil, tFleetdTCPAddr)
		if errDial != nil {
			return
		}
		defer conn.Close()
		_, errDial = conn.Write(data)
		if errDial != nil {
			return
		}
	default:
		log.Error("Got error in sendDataToTFleetd: %v", err)
	}
	return
}
