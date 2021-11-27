package tmsg

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	pb "github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
)

func PackFrom(msg pb.Message) (*any.Any, error) {
	typeName := pb.MessageName(msg)
	if typeName == "" {
		return nil, errors.New(fmt.Sprintf("Message type not known: %v", reflect.TypeOf(msg)))
	}

	value, err := pb.Marshal(msg)
	if err != nil {
		return nil, err
	}

	return &any.Any{
		TypeUrl: "type.googleapis.com/" + typeName,
		Value:   value,
	}, nil
}

func Unpack(any *any.Any) (pb.Message, error) {
	if any == nil {
		return nil, nil
	}
	typeURL := any.TypeUrl
	bytes := any.Value
	split := strings.Split(typeURL, "/")
	typeName := split[len(split)-1]

	ty := pb.MessageType(typeName)
	if ty == nil {
		return nil, errors.New(fmt.Sprintf("Message type not known: %s", typeURL))
	}

	ty = ty.Elem()

	instance := reflect.New(ty).Interface()
	msg := instance.(pb.Message)
	err := pb.Unmarshal(bytes, msg)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func UnpackTo(any *any.Any, msg pb.Message) {
	typeURL := any.TypeUrl
	bytes := any.Value
	split := strings.Split(typeURL, "/")
	typeName := split[len(split)-1]

	ty := pb.MessageType(typeName)
	if ty == nil {
		panic(fmt.Sprintf("Message type not known: %s", typeURL))
	}

	if reflect.TypeOf(msg) != ty {
		panic(fmt.Sprintf("Message is wrong type: %v vs %v", reflect.TypeOf(msg), ty))
	}

	err := pb.Unmarshal(bytes, msg)
	if err != nil {
		panic(err)
	}
}
