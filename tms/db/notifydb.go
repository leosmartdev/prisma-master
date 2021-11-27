package db

import (
	"prisma/tms/moc"
	"prisma/tms/rest"

	"github.com/globalsign/mgo/bson"

	"github.com/golang/protobuf/ptypes/timestamp"
	"prisma/gogroup"
)

type NotifyDb interface {
	Ack(string, *timestamp.Timestamp) error
	AckAll(*timestamp.Timestamp) (int, error)
	AckWait(string, *timestamp.Timestamp) error
	Clear(string, *timestamp.Timestamp) error
	Create(*moc.Notice) error
	GetActive() ([]*moc.Notice, error)
	GetHistory(bson.M, *rest.PaginationQuery) ([]bson.M, error)
	// GetById returns a notice by database id
	GetById(id string) (*moc.Notice, error)
	// GetByNoticeId returns a notice by notice id
	GetByNoticeId(id string) (*moc.Notice, error)
	GetStream() (*NoticeStream, error)
	// GetPersistentStreamWithGroupContext returns stream
	// with specific context to be canceled
	GetPersistentStreamWithGroupContext(ctx gogroup.GoGroup, pipeline []bson.M) *NoticeStream
	// GetPersistentStream returns stream to watch for notices
	// pipeline is used to filter data for the watcher
	GetPersistentStream(pipeline []bson.M) *NoticeStream
	Startup() error
	Renew(string, *timestamp.Timestamp) error
	// TimeoutBySliceId updates notices by setting the expired state
	TimeoutBySliceId(id []string) (int, error)
	Timeout(*timestamp.Timestamp) (int, error)
	UpdateTime(string, *timestamp.Timestamp) error
}

type NoticeStream struct {
	Updates         chan *moc.Notice
	InitialLoadDone chan bool
}

func NewNoticeStream() *NoticeStream {
	return &NoticeStream{
		Updates:         make(chan *moc.Notice),
		InitialLoadDone: make(chan bool, 2),
	}
}
