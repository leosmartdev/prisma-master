package routing

import (
	jsonpb "github.com/golang/protobuf/jsonpb"
)

func (r *Route) MarshalJSON() ([]byte, error) {
	m := jsonpb.Marshaler{
		Indent: "  ",
	}
	s, err := m.MarshalToString(r)
	if err != nil {
		return nil, err
	}
	return []byte(s), err
}
func (r *Route) UnmarshalJSON(b []byte) error {
	return jsonpb.UnmarshalString(string(b), r)
}
