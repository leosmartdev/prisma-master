package task

import (
	"prisma/tms/iridium"
)

// MT is a structure for storing data for a MT task
type MT struct {
	Header  iridium.MTHeader
	Payload iridium.MPayload
}
