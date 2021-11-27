package tmsg

import (
	. "prisma/tms"
)

func MessageType(msg *TsiMessage) string {
	return msg.Type()
}
