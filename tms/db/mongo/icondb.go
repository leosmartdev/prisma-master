package mongo

import (
	"fmt"
	"prisma/tms/db"
	"prisma/tms/moc"
	"unsafe"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// ObjectTypes
const (
	IconObjectType      = "prisma.tms.moc.Icon"
	IconImageObjectType = "prisma.tms.moc.IconImage"
	ICON_NOTE_TYPE      = "Icon"
)

// MongoDB collections
const CollectionIcon = "icons"
const CollectionIconImage = "icon_images"

func NewIconDb(misc db.MiscDB) db.IconDB {
	client, ok := misc.(*MongoMiscClient)
	if !ok {
		return nil
	}
	return &MongoMiscClient{
		dbconn: client.dbconn,
		ctxt:   client.ctxt,
	}
}

func (iconDb *MongoMiscClient) UpsertIcon(icon *moc.Icon) error {
	_, err := iconDb.Upsert(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: IconObjectType,
			Obj: &db.GoObject{
				ID:   icon.Id,
				Data: icon,
			},
		},
		Ctxt: iconDb.ctxt,
		Time: &db.TimeKeeper{},
	})

	return err
}

func (iconDb *MongoMiscClient) FindAllIcons(mac_address string, withDeleted bool) ([]*moc.Icon, error) {
	var icons = make([]*moc.Icon, 0)

	req := db.GoRequest{
		ObjectType: IconObjectType,
	}
	sd, ti, err := iconDb.resolveTable(&req)
	if err != nil {
		return icons, err
	}
	raw := []bson.Raw{}
	c := Coder{TypeData: sd}

	var query map[string]interface{}
	if withDeleted == true {
		query = bson.M{
			"me.mac_address": mac_address,
		}
	} else {
		query = bson.M{
			"me.mac_address": mac_address,
			"me.deleted":     false,
		}
	}

	err = iconDb.dbconn.DB().C(ti.Name).Find(query).All(&raw)
	if err != mgo.ErrNotFound {
		for _, data := range raw {
			var obj DBMiscObject
			c.DecodeTo(data, unsafe.Pointer(&obj))
			icon, ok := obj.Obj.(*moc.Icon)
			if !ok {
				return icons, fmt.Errorf("Could not fetch Icon object")
			}
			icons = append(icons, icon)
		}
	}

	return icons, err
}

func (iconDb *MongoMiscClient) FindOneIcon(id string, withDeleted bool) (*moc.Icon, error) {
	var icon *moc.Icon

	iconData, err := iconDb.Get(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: IconObjectType,
			Obj: &db.GoObject{
				ID: id,
			},
		},
		Ctxt: iconDb.ctxt,
		Time: &db.TimeKeeper{},
	})

	if err == nil {
		icons := make([]*moc.Icon, 0)
		for _, iconDatum := range iconData {

			if mocIcon, ok := iconDatum.Contents.Data.(*moc.Icon); ok {
				if withDeleted == false && mocIcon.Deleted == true {
					continue
				}

				icons = append(icons, mocIcon)
			}
		}

		if len(icons) > 0 {
			icon = icons[0]
		} else {
			err = db.ErrorNotFound
		}
	}

	return icon, err
}

func (iconDb *MongoMiscClient) DeleteIcon(id string) error {
	query := bson.M{
		"_id": bson.ObjectIdHex(id),
	}

	update := bson.M{
		"$set": bson.M{
			"me.deleted": true,
		},
	}

	err := iconDb.dbconn.DB().C(CollectionIcon).Update(query, update)

	return err
}

func (iconDb *MongoMiscClient) UpsertIconImage(iconImage *moc.IconImage) error {
	_, err := iconDb.Upsert(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: IconImageObjectType,
			Obj: &db.GoObject{
				ID:   iconImage.Id,
				Data: iconImage,
			},
		},
		Ctxt: iconDb.ctxt,
		Time: &db.TimeKeeper{},
	})

	return err
}

func (iconDb *MongoMiscClient) FindAllIconImages(mac_address string) ([]*moc.IconImage, error) {
	var iconImages = make([]*moc.IconImage, 0)

	iconImageData, err := iconDb.Get(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: IconImageObjectType,
			Obj: &db.GoObject{
				Data: &moc.IconImage{
					MacAddress: mac_address,
				},
			},
		},
		Ctxt: iconDb.ctxt,
		Time: &db.TimeKeeper{},
	})

	if err == nil {
		for _, iconImageDatum := range iconImageData {
			if mocIconImage, ok := iconImageDatum.Contents.Data.(*moc.IconImage); ok {
				iconImages = append(iconImages, mocIconImage)
			}
		}
	}

	return iconImages, err
}
