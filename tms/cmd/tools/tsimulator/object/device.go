package object

import (
	"prisma/tms/omnicom"
	"time"
	"prisma/tms/cmd/tools/tsimulator/task"
)

// DeviceCommunicator provides features for sea objects
type DeviceCommunicator interface {
	ObjectCommunicator
	// Update information about object
	UpdateInformation(object Object)
}

// NullDevice doesn't implement nothing
type NullDevice struct {
}

// NewNullDevice returns NullDevice
func NewNullDevice() *NullDevice {
	return &NullDevice{}
}

func encodeDegree(degree float64) float64 {
	minutes := degree * 60
	return minutes * 10000
}

func eventTime() omnicom.Date_Event {

	var date omnicom.Date_Event

	date.Year = uint32(time.Now().Year() - 2000)
	date.Month = uint32(time.Now().Month())
	date.Day = uint32(time.Now().Day())
	date.Minute = (uint32(time.Now().Hour()) * 60) + uint32(time.Now().Minute())

	return date
}

func (*NullDevice) GetMessageStaticInformation() ([]byte, error) {
	return nil, ErrMessageNotImplement
}

func (*NullDevice) GetMessagePosition() ([]byte, error) {
	return nil, ErrMessageNotImplement
}

func (*NullDevice) GetTrackedTargetMessage(latitude, longitude float64) ([]byte, error) {
	return nil, ErrMessageNotImplement
}

func (*NullDevice) GetDataForIridiumNetwork() ([]byte, error) {
	return nil, ErrMessageNotImplement
}

func (*NullDevice) GetStartAlertingMessage() ([]byte, error) {
	return nil, ErrMessageNotImplement
}

func (*NullDevice) GetStopAlertingMessage() ([]byte, error) {
	return nil, ErrMessageNotImplement
}

func (*NullDevice) GetPositionAlertingMessage() ([]byte, error) {
	return nil, ErrMessageNotImplement
}

func (*NullDevice) StartAlerting(uint) error {
	return ErrMessageNotImplement
}

func (*NullDevice) StopAlerting(uint) error {
	return ErrMessageNotImplement
}

func (*NullDevice) UpdateInformation(Object) {
}

func (*NullDevice) AddTask(interface{}) (len int, err error) {
	return 0, ErrMessageNotImplement
}

func (*NullDevice) GetChannelHandledTask() <-chan task.Result {
	return nil
}
