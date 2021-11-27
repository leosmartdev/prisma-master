package db

import "prisma/tms/moc"

type IconDB interface {
	UpsertIcon(icon *moc.Icon) error
	FindAllIcons(mac_address string, withDeleted bool) ([]*moc.Icon, error)
	FindOneIcon(id string, withDeleted bool) (*moc.Icon, error)
	DeleteIcon(id string) error
	UpsertIconImage(iconImage *moc.IconImage) error
	FindAllIconImages(mac_address string) ([]*moc.IconImage, error)
}
