package db

import "prisma/tms/sar"

// Interface to a database which contains tracks
type SarsatDB interface {
	//GetMessageStream() (<-chan sar.SarsatMessage, error)
	InsertMessage(*sar.SarsatMessage) error
	InsertAlert(*sar.SarsatAlert, *sar.SarsatMessage) error
}
