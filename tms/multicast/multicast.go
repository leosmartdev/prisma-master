// Package multicast contains a fabric method to get different implementations for different multicast messages.
package multicast

import (
	"fmt"

	"prisma/tms/db"
	"prisma/tms/moc"
	"prisma/tms/multicast/messages"
	"prisma/tms/omnicom"
	"prisma/tms/tmsg"
)

// Multicast is an interface for working with beacon messages
type Multicast interface {
	// It returns structure for sending to tgwad, also it returns an id for watching the status of the message
	GetMessage(misc db.MiscDB) (*omnicom.Omni, string, string, error)
}

// Parsing determines and returns a message instance for working
func Parsing(cmd *moc.DeviceConfiguration) (Multicast, error) {

	msg, err := tmsg.Unpack(cmd.Configuration)
	if err != nil {
		return nil, err
	}
	switch cmd.Configuration.TypeUrl {
	case "type.googleapis.com/prisma.tms.omnicom.OmnicomConfiguration":
		config, ok := msg.(*omnicom.OmnicomConfiguration)
		if !ok {
			return nil, fmt.Errorf("Could not reflect OmnicomConfiguration payload")
		}
		return messages.NewOmnicomConfigMulticast(config), nil

	default:
		return nil, fmt.Errorf("Type not handled in multicast %+v", cmd.Configuration.TypeUrl)
	}

	return nil, nil

}
