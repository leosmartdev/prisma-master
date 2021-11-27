package moc

import (
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"io"
)

type JSPBUnmarshaler interface {
	Unmarshal(r io.Reader, pb proto.Message) error
}

func (t *GitLabEventMergeRequest) Unmarshal(r io.Reader, pb proto.Message) error {
	jspb := &jsonpb.Unmarshaler{
		AllowUnknownFields: true,
	}
	return jspb.Unmarshal(r, pb)
}
