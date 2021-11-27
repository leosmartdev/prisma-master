package web

import (
	"errors"
	"net"
	"time"

	"prisma/gogroup"
	"prisma/tms/cmd/tools/tsimulator/object"
	"prisma/tms/cmd/tools/tsimulator/task"
	"prisma/tms/iridium"
	"prisma/tms/log"
)

const (
	// The time to read an mt message
	timeDeadLine = 4 * time.Second
	// bytes before data - IEI + length
	hmtStart = 3
	// Minimum size for an mt message to be processed
	minSizeMtMessage = 35
)

// DirectIPMT is a structure to handle MT messages
type DirectIPMT struct {
	control *object.Control
}

// NewDirectIPMT returns an instance of DirectIPMT that is linked to sea objects
func NewDirectIPMT(control *object.Control) *DirectIPMT {
	return &DirectIPMT{
		control: control,
	}
}

func (d *DirectIPMT) Listen(ctx gogroup.GoGroup, addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer l.Close()
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			c, err := l.Accept()
			if err != nil {
				return nil
			}
			go d.handleConn(c)
		}
	}
}

func (d *DirectIPMT) handleConn(c net.Conn) {
	defer c.Close()
	b := make([]byte, 1e6)
	if err := c.SetReadDeadline(time.Now().Add(timeDeadLine)); err != nil {
		log.Error(err.Error())
		return
	}
	if n, err := c.Read(b); err != nil {
		// what to do here?
		log.Error(err.Error())
		return
	} else {
		b = b[:n]
		log.Debug("This is the raw data received from tfleetd %+v", b)
	}
	resp, err := d.handleMT(b)
	if err != nil {
		log.Debug("Handle MT failed %v", err)
		data, err := resp.Encode()
		if err != nil {
			log.Error(err.Error())
			return
		}
		if _, err := c.Write(data); err != nil {
			log.Error(err.Error())
		}
		return
	}
	log.Debug("This is the MTC message %+v", resp)
	data, err := resp.Encode()
	if err != nil {
		log.Error(err.Error())
		return
	}
	if _, err := c.Write(data); err != nil {
		log.Error(err.Error())
	}
}

func getStatusByError(err error) iridium.SessionStatus {
	switch err.(type) {
	case iridium.ParserError:
		return iridium.StatusViolationProtocol
	}
	switch err {
	case iridium.ErrInvalidIMEI:
		return iridium.StatusInvalidIMEI
	case task.ErrTaskQueueFull:
		return iridium.StatusQueueFull
	case object.ErrObjectNotFound:
		return iridium.StatusUnknownIMEI
	}
	return iridium.StatusOK
}

func (d *DirectIPMT) handleMT(b []byte) (*iridium.MTResponse, error) {
	if len(b) < minSizeMtMessage {
		return &iridium.MTResponse{
			UniqueClientMessageID: "",
			MTMessageStatus:       getStatusByError(iridium.ParserError{}),
			IMEI:                  "",
		}, errors.New("Not enough size to be processed")
	}

	mtHeader, err := iridium.ParseMTHeader(b[hmtStart:])
	if err != nil {
		return &iridium.MTResponse{
			UniqueClientMessageID: "",
			MTMessageStatus:       getStatusByError(err),
			IMEI:                  "",
		}, err
	}
	log.Debug("MT Header %+v", mtHeader)

	obj, err := d.control.GetByIMEI(mtHeader.IMEI)
	if err != nil {
		return &iridium.MTResponse{
			UniqueClientMessageID: mtHeader.UniqueClientMessageID,
			MTMessageStatus:       getStatusByError(err),
			IMEI:                  mtHeader.IMEI,
		}, err
	}

	// IEI whole length + IEI header + header length
	mp, err := iridium.ParseMOPayload(b[hmtStart+hmtStart+mtHeader.MTHL:])
	if err != nil {
		return &iridium.MTResponse{
			UniqueClientMessageID: mtHeader.UniqueClientMessageID,
			MTMessageStatus:       getStatusByError(err),
			IMEI:                  mtHeader.IMEI,
		}, err
	}

	l, err := obj.AddTask(&task.MT{
		Header:  *mtHeader,
		Payload: mp,
	})
	if err != nil {
		return &iridium.MTResponse{
			UniqueClientMessageID: mtHeader.UniqueClientMessageID,
			MTMessageStatus:       getStatusByError(err),
			IMEI:                  mtHeader.IMEI,
		}, err
	}

	return &iridium.MTResponse{
		UniqueClientMessageID: mtHeader.UniqueClientMessageID,
		MTMessageStatus:       iridium.SessionStatus(l),
		IMEI:                  mtHeader.IMEI,
		AutoIDReference:       uint32(l),
	}, nil
}
