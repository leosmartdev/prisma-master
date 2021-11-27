package omnicom

import (
	omni "prisma/tms/omnicom"
	"time"
)

func CreateOmnicomDate(messageTime time.Time) omni.Date {
	var date omni.Date

	if messageTime.IsZero() {
		messageTime = time.Now().UTC()
	}
	date.Year = uint32(messageTime.Year() - 2000)
	date.Month = uint32(messageTime.Month())
	date.Day = uint32(messageTime.Day())
	date.Minute = uint32((uint32(messageTime.Hour()) * 60) + uint32(messageTime.Minute()))

	return date
}
