package mongo

import (
	"fmt"
	"time"
	"errors"

	"prisma/gogroup"
	api "prisma/tms/client_api"
	"prisma/tms/db"
	"prisma/tms/moc"
	"prisma/tms/rest"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"prisma/tms/log"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type MongoNotifyDb struct {
	mongo *MongoClient
	ctxt  gogroup.GoGroup
	misc  db.MiscDB
}

func NewNotifyDb(ctxt gogroup.GoGroup, client *MongoClient) db.NotifyDb {
	d := &MongoNotifyDb{
		mongo: client,
		misc:  NewMongoMiscData(ctxt, client),
		ctxt:  ctxt,
	}
	return d
}

var ErrNotFound = errors.New("not found")

func (d *MongoNotifyDb) GetById(id string) (*moc.Notice, error) {
	request := db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: "prisma.tms.moc.Notice",
			Obj: &db.GoObject{
				ID: id,
			},
		},
		Ctxt: d.ctxt,
	}
	results, err := d.misc.Get(request)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, ErrNotFound
	}
	// actually results should have 1 element
	if len(results) != 1 {
		log.Error("mongodb has several records for the same noticeId")
	}
	// get first element in result set, it does not matter which one if we have several notices
	notice, ok := results[0].Contents.Data.(*moc.Notice)
	if !ok {
		return nil, fmt.Errorf("bad notice data: %v", log.Spew(results[0].Contents.Data))
	}
	notice.DatabaseId = results[0].Contents.ID
	return notice, nil
}

func (d *MongoNotifyDb) GetByNoticeId(noticeId string) (*moc.Notice, error) {
	request := db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: "prisma.tms.moc.Notice",
			Obj: &db.GoObject{
				Data: &moc.Notice{
					NoticeId: noticeId,
				},
			},
		},
		Ctxt: d.ctxt,
	}
	results, err := d.misc.Get(request)
	if err != nil {
		return nil, err
	}
	if len(results) != 1 {
		return nil, mgo.ErrNotFound
	}
	notice, ok := results[0].Contents.Data.(*moc.Notice)
	if !ok {
		return nil, fmt.Errorf("bad notice data: %v", log.Spew(results[0].Contents.Data))
	}
	notice.DatabaseId = results[0].Contents.ID
	return notice, nil
}

func (d *MongoNotifyDb) GetPersistentStreamWithGroupContext(ctx gogroup.GoGroup, pipeline []bson.M) *db.NoticeStream {
	downstream := db.NewNoticeStream()
	upstream := d.misc.GetPersistentStream(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: "prisma.tms.moc.Notice",
		},
		Ctxt: ctx,
	}, bson.M{
		"me.action": bson.M{
			"$nin": []string{"CLEAR"},
		},
		"utime": bson.M{
			"$ne": "",
		},
	}, pipeline)
	go d.handleStream(ctx, upstream, downstream, true)
	return downstream
}

func (d *MongoNotifyDb) GetPersistentStream(pipeline []bson.M) *db.NoticeStream {
	downstream := db.NewNoticeStream()
	upstream := d.misc.GetPersistentStream(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: "prisma.tms.moc.Notice",
		},
		Ctxt: d.ctxt,
	}, bson.M{
		"me.action": bson.M{
			"$nin": []string{"CLEAR"},
		},
		"utime": bson.M{
			"$gt": time.Now().Add(-200000 * time.Hour),
		},
	}, pipeline)
	go d.handleStream(d.ctxt, upstream, downstream, true)
	return downstream
}

func (d *MongoNotifyDb) handleStream(ctx gogroup.GoGroup, upstream <-chan db.GoGetResponse, downstream *db.NoticeStream, persist bool) {
	for {
		select {
		case update, ok := <-upstream:
			if !ok {
				if !persist {
					log.Error("A channel was closed")
					return
				}
				log.Warn("Connection was lost, try to reconnect")
				upstream = d.misc.GetPersistentStream(db.GoMiscRequest{
					Req: &db.GoRequest{
						ObjectType: "prisma.tms.moc.Notice",
					},
					Ctxt: ctx,
				}, nil, nil)
				continue
			}
			if update.Status == api.Status_InitialLoadDone {
				downstream.InitialLoadDone <- true
				continue
			}
			notice := update.Contents.Data.(*moc.Notice)
			notice.DatabaseId = update.Contents.ID
			downstream.Updates <- notice
		case <-ctx.Done():
			return
		}
	}
}

func (d *MongoNotifyDb) GetStream() (*db.NoticeStream, error) {
	downstream := db.NewNoticeStream()
	upstream, err := d.misc.GetStream(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: "prisma.tms.moc.Notice",
		},
		Ctxt: d.ctxt,
	}, nil, nil)
	if err != nil {
		return nil, err
	}
	go d.handleStream(d.ctxt, upstream, downstream, false)
	return downstream, nil
}

func (d *MongoNotifyDb) Create(n *moc.Notice) error {
	req := db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: "prisma.tms.moc.Notice",
			Obj: &db.GoObject{
				ID:   n.DatabaseId,
				Data: n,
			},
		},
		Ctxt: d.ctxt,
	}
	response, err := d.misc.Upsert(req)
	if err != nil {
		return err
	}
	n.DatabaseId = response.Id
	return nil
}

func (d *MongoNotifyDb) Ack(id string, t *timestamp.Timestamp) error {
	db := d.mongo.DB()
	defer d.mongo.Release(db)

	found := false
	err := db.C("notices").Update(
		bson.M{
			"_id":       bson.ObjectIdHex(id),
			"me.action": moc.Notice_NEW.String(),
		},
		bson.M{
			"$set": bson.M{
				"me.ack_time.seconds": t.Seconds,
				"me.ack_time.nanos":   t.Nanos,
				"me.action":           moc.Notice_ACK.String(),
			},
		},
	)
	if err != nil && err != mgo.ErrNotFound {
		return err
	}
	if err == nil {
		found = true
	}

	err = db.C("notices").Update(
		bson.M{
			"_id":       bson.ObjectIdHex(id),
			"me.action": moc.Notice_ACK_WAIT.String(),
		},
		bson.M{
			"$set": bson.M{
				"etime":                   time.Now(),
				"me.ack_time.seconds":     t.Seconds,
				"me.ack_time.nanos":       t.Nanos,
				"me.cleared_time.seconds": t.Seconds,
				"me.cleared_time.nanos":   t.Nanos,
				"me.action":               moc.Notice_CLEAR.String(),
			},
		},
	)
	if err != nil && err != mgo.ErrNotFound {
		return err
	}
	if err == nil {
		found = true
	}

	if !found {
		return mgo.ErrNotFound
	}
	return nil
}

func (d *MongoNotifyDb) AckAll(t *timestamp.Timestamp) (int, error) {
	mdb := d.mongo.DB()
	defer d.mongo.Release(mdb)

	now, err := ptypes.Timestamp(t)
	if err != nil {
		return 0, err
	}
	iter := mdb.C("notices").Find(
		bson.M{
			"etime": db.MiscDataNotExpiredTime,
			"me.action": bson.M{
				"$in": []string{
					moc.Notice_NEW.String(), moc.Notice_ACK_WAIT.String(),
				},
			},
		},
	).Iter()
	bulk := mdb.C("notices").Bulk()
	result := bson.M{}
	changed := 0

	for iter.Next(&result) {
		if changed%500 == 0 {
			if changed > 0 {
				if _, err := bulk.Run(); err != nil {
					return 0, err
				}
			}
			bulk = mdb.C("notices").Bulk()
		}
		id := result["_id"]
		me := result["me"].(bson.M)
		prevAction := me["action"].(string)

		if prevAction == moc.Notice_NEW.String() {
			bulk.Update(
				bson.M{"_id": id},
				bson.M{
					"$set": bson.M{
						"me.action":               moc.Notice_ACK.String(),
						"me.updated_time.seconds": t.Seconds,
						"me.updated_time.nanos":   t.Nanos,
					},
				},
			)
		} else {
			bulk.Update(
				bson.M{"_id": id},
				bson.M{
					"$set": bson.M{
						"etime":                   now,
						"me.action":               moc.Notice_CLEAR.String(),
						"me.ack_time.seconds":     t.Seconds,
						"me.ack_time.nanos":       t.Nanos,
						"me.cleared_time.seconds": t.Seconds,
						"me.cleared_time.nanos":   t.Nanos,
					},
				},
			)
		}
		changed++
	}
	if err := iter.Close(); err != nil {
		return 0, err
	}
	if _, err := bulk.Run(); err != nil {
		return 0, err
	}
	return changed, nil
}

func (d *MongoNotifyDb) AckWait(id string, t *timestamp.Timestamp) error {
	db := d.mongo.DB()
	defer d.mongo.Release(db)

	err := db.C("notices").UpdateId(bson.ObjectIdHex(id),
		bson.M{
			"$set": bson.M{
				"me.updated_time.seconds": t.Seconds,
				"me.updated_time.nanos":   t.Nanos,
				"me.action":               moc.Notice_ACK_WAIT.String(),
			},
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func (d *MongoNotifyDb) Renew(id string, t *timestamp.Timestamp) error {
	db := d.mongo.DB()
	defer d.mongo.Release(db)

	err := db.C("notices").UpdateId(bson.ObjectIdHex(id),
		bson.M{
			"$set": bson.M{
				"me.updated_time.seconds": t.Seconds,
				"me.updated_time.nanos":   t.Nanos,
				"me.action":               moc.Notice_NEW.String(),
			},
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func (d *MongoNotifyDb) Clear(databaseID string, t *timestamp.Timestamp) error {
	db := d.mongo.DB()
	defer d.mongo.Release(db)

	err := db.C("notices").UpdateId(bson.ObjectIdHex(databaseID),
		bson.M{
			"$set": bson.M{
				"etime":                   time.Now(),
				"me.cleared_time.seconds": t.Seconds,
				"me.cleared_time.nanos":   t.Nanos,
				"me.action":               moc.Notice_CLEAR.String(),
			},
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func (d *MongoNotifyDb) GetActive() ([]*moc.Notice, error) {
	req := db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: "prisma.tms.moc.Notice",
		},
		Ctxt: d.ctxt,
	}
	results, err := d.misc.Get(req)
	if err != nil {
		return nil, err
	}
	ret := make([]*moc.Notice, 0, len(results))
	for _, r := range results {
		notice := r.Contents.Data.(*moc.Notice)
		notice.DatabaseId = r.Contents.ID
		ret = append(ret, notice)
	}
	return ret, nil
}

func (d *MongoNotifyDb) Startup() error {
	mdb := d.mongo.DB()
	defer d.mongo.Release(mdb)

	iter := mdb.C("notices").Find(
		bson.M{
			"me.etime":  db.MiscDataNotExpiredTime,
			"me.action": moc.Notice_NEW.String(),
		},
	).Iter()
	bulk := mdb.C("notices").Bulk()

	result := bson.M{}
	for iter.Next(&result) {
		id := result["_id"].(bson.ObjectId)
		bulk.Update(
			bson.M{"_id": id},
			bson.M{
				"$set": bson.M{
					"me.action": moc.Notice_ACK_WAIT.String(),
				},
			},
		)
	}

	_, err := bulk.Run()
	if err != nil {
		return err
	}
	return nil
}

func (d *MongoNotifyDb) UpdateTime(databaseID string, t *timestamp.Timestamp) error {
	db := d.mongo.DB()
	defer d.mongo.Release(db)

	return db.C("notices").UpdateId(bson.ObjectIdHex(databaseID),
		bson.M{
			"$set": bson.M{
				"me.updated_time.nanos":   t.Nanos,
				"me.updated_time.seconds": t.Seconds,
			},
		},
	)
}

func (d *MongoNotifyDb) TimeoutBySliceId(id []string) (int, error) {
	if id == nil || len(id) == 0 {
		return 0, nil
	}
	mdb := d.mongo.DB()
	defer d.mongo.Release(mdb)

	now := time.Now()
	pnow, err := ptypes.TimestampProto(now)
	if err != nil {
		return 0, err
	}
	changes, err := mdb.C("notices").UpdateAll(
		bson.M{
			"_id": bson.M{
				"$in": id,
			},
		},
		bson.M{
			"$set": bson.M{
				"etime":                   now,
				"me.action":               moc.Notice_CLEAR.String(),
				"me.cleared_time.seconds": pnow.Seconds,
				"me.cleared_time.nanos":   pnow.Nanos,
			},
		},
	)
	if err != nil {
		return 0, err
	}
	return changes.Updated, nil
}

func (d *MongoNotifyDb) Timeout(olderThan *timestamp.Timestamp) (int, error) {
	mdb := d.mongo.DB()
	defer d.mongo.Release(mdb)

	now := time.Now()
	pnow, err := ptypes.TimestampProto(now)
	if err != nil {
		return 0, err
	}
	changes, err := mdb.C("notices").UpdateAll(
		bson.M{
			"etime": db.MiscDataNotExpiredTime,
			"me.updated_time.seconds": bson.M{
				"$lt": olderThan.Seconds,
			},
			"me.action": moc.Notice_ACK.String(),
		},
		bson.M{
			"$set": bson.M{
				"etime":                   now,
				"me.action":               moc.Notice_CLEAR.String(),
				"me.cleared_time.seconds": pnow.Seconds,
				"me.cleared_time.nanos":   pnow.Nanos,
			},
		},
	)
	if err != nil {
		return 0, err
	}
	return changes.Updated, nil
}

func (d *MongoNotifyDb) GetHistory(query bson.M, pagination *rest.PaginationQuery) ([]bson.M, error) {
	mdb := d.mongo.DB()
	defer d.mongo.Release(mdb)

	pipe := []bson.M{
		bson.M{"$match": query},
		bson.M{"$sort": bson.M{pagination.Sort: -1}},
		bson.M{"$limit": pagination.Limit},
	}
	if pagination.AfterId != "" {
		v, err := time.Parse(time.RFC3339Nano, pagination.AfterId)
		if err != nil {
			return nil, err
		}
		query["ctime"] = bson.M{"$lt": v}
	}
	if pagination.BeforeId != "" {
		v, err := time.Parse(time.RFC3339Nano, pagination.BeforeId)
		if err != nil {
			return nil, err
		}
		query["ctime"] = bson.M{"$gt": v}
		pipe = []bson.M{
			bson.M{"$match": query},
			bson.M{"$sort": bson.M{pagination.Sort: 1}},
			bson.M{"$limit": pagination.Limit},
			bson.M{"$sort": bson.M{pagination.Sort: -1}},
		}
	}
	results := []bson.M{}
	err := mdb.C("notices").Pipe(pipe).All(&results)
	if err != nil {
		return nil, err
	}
	notices := []bson.M{}
	for i, result := range results {
		t, ok := result["ctime"].(time.Time)
		if !ok {
			return nil, fmt.Errorf("expecting time, got: %v", result["ctime"])
		}
		id := t.Format(time.RFC3339Nano)
		if i == 0 {
			pagination.BeforeId = id
		}
		if i == len(results)-1 {
			pagination.AfterId = id
		}
		notices = append(notices, result["me"].(bson.M))
	}
	return notices, nil
}
