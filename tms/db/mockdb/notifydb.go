package mockdb

import (
	"prisma/tms/db"
	"prisma/tms/moc"
	"prisma/tms/rest"

	"github.com/globalsign/mgo/bson"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/stretchr/testify/mock"
	"prisma/gogroup"
)

type NotifyDb struct {
	mock.Mock
}

func (d *NotifyDb) GetById(id string) (*moc.Notice, error) {
	args := d.Called(id)
	return args.Get(0).(*moc.Notice), args.Error(1)
}

func (d *NotifyDb) GetByNoticeId(id string) (*moc.Notice, error) {
	args := d.Called(id)
	return args.Get(0).(*moc.Notice), args.Error(1)
}

func (d *NotifyDb) Ack(id string, t *timestamp.Timestamp) error {
	args := d.Called(id, t)
	return args.Error(0)
}

func (d *NotifyDb) AckAll(t *timestamp.Timestamp) (int, error) {
	args := d.Called(t)
	return args.Int(0), args.Error(1)
}

func (d *NotifyDb) AckWait(id string, t *timestamp.Timestamp) error {
	args := d.Called(id, t)
	return args.Error(0)
}

func (d *NotifyDb) Clear(id string, t *timestamp.Timestamp) error {
	args := d.Called(id, t)
	return args.Error(0)
}

func (d *NotifyDb) Create(n *moc.Notice) error {
	args := d.Called(n)
	return args.Error(0)
}

func (d *NotifyDb) GetActive() ([]*moc.Notice, error) {
	args := d.Called()
	return args.Get(0).([]*moc.Notice), args.Error(1)
}

func (d *NotifyDb) GetPersistentStream(pipeline []bson.M) *db.NoticeStream {
	args := d.Called()
	return args.Get(0).(*db.NoticeStream)
}

func (d *NotifyDb) GetPersistentStreamWithGroupContext(ctx gogroup.GoGroup, pipeline []bson.M) *db.NoticeStream {
	args := d.Called()
	return args.Get(0).(*db.NoticeStream)
}

func (d *NotifyDb) GetStream() (*db.NoticeStream, error) {
	args := d.Called()
	return args.Get(0).(*db.NoticeStream), args.Error(1)
}

func (d *NotifyDb) Renew(id string, t *timestamp.Timestamp) error {
	args := d.Called(id, t)
	return args.Error(0)
}

func (d *NotifyDb) Startup() error {
	args := d.Called()
	return args.Error(0)
}

func (d *NotifyDb) TimeoutBySliceId(idSlice []string) (int, error) {
	args := d.Called(idSlice)
	return args.Int(0), args.Error(1)
}

func (d *NotifyDb) Timeout(olderThan *timestamp.Timestamp) (int, error) {
	args := d.Called(olderThan)
	return args.Int(0), args.Error(1)
}

func (d *NotifyDb) UpdateTime(id string, t *timestamp.Timestamp) error {
	args := d.Called(id, t)
	return args.Error(0)
}

func (d *NotifyDb) GetHistory(query bson.M, pagination *rest.PaginationQuery) ([]bson.M, error) {
	args := d.Called(query, pagination)
	return args.Get(0).([]bson.M), args.Error(1)
}

func NewNotifyDbStub() *NotifyDb {
	n := &NotifyDb{}
	n.On("AckWait", mock.AnythingOfType("string"), mock.AnythingOfType("*timestamp.Timestamp")).Return(nil)
	n.On("Clear", mock.AnythingOfType("string"), mock.AnythingOfType("*timestamp.Timestamp")).Return(nil)
	n.On("Create", mock.AnythingOfType("*moc.Notice")).Return(nil)
	n.On("Startup").Return(nil)
	n.On("Timeout", mock.AnythingOfType("*timestamp.Timestamp")).Return(0, nil)
	return n
}
