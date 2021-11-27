package object

import "errors"

const (
	NavigationStatusAISSart        = 14
	NavigationStatusAISSartTesting = 15
	MessageTypePosition            = 1
	MessageTypeSRB                 = 5
)

// ErrMessageNotImplement is used to point out that a feature is not implemented or is not maintained
var ErrMessageNotImplement = errors.New("the device doesn't send this message")

// Informer is used to issue important messages for object on sea
type Informer interface {
	// Return a message about a position of this sea object
	// see: http://catb.org/gpsd/AIVDM.html#_types_1_2_and_3_position_report_class_a
	// Also in the future we can implement other states of a sea object, like "a sea object is a anchor"
	// it should send every 5 minutes with type 5 identification
	GetMessagePosition() ([]byte, error)
	// Return a static information about a sea object
	GetMessageStaticInformation() ([]byte, error)
}
