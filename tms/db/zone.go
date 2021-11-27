package db

import "prisma/tms/moc"

// ZoneDB is a composition of MiscDB and extra behavior specific to Zone
type ZoneDB interface {
	MiscDB
	GetOne(omnID uint32) (*moc.Zone, error)
}
