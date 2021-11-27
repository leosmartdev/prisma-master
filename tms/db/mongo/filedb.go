package mongo

import (
	"bytes"
	"context"
	"io"

	"prisma/tms/moc"

	"github.com/globalsign/mgo/bson"
)

const (
	CollectionFile = "fs"
)

type FileDb struct{}

func (d *FileDb) Create(ctx context.Context, f *moc.File) error {
	session, err := getSession(ctx)
	defer session.Close()
	if err != nil {
		return err
	}
	db := session.DB(DATABASE)
	dest, err := db.GridFS(CollectionFile).Create(f.Metadata.Name)
	defer dest.Close()
	dest.SetMeta(bson.M{"id": f.Metadata.Id})
	dest.SetContentType(f.Metadata.Type)
	_, err = io.Copy(dest, bytes.NewReader(f.Data))
	id, ok := dest.Id().(bson.ObjectId)
	if ok {
		f.Metadata.Id = id.Hex()
	}
	return err
}
