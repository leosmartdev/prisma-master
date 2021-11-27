package rule

import (
	"testing"

	google_protobuf1 "github.com/golang/protobuf/ptypes/wrappers"
	"github.com/stretchr/testify/assert"
)

var dataEqual = [][]interface{}{
	{int32(1), int32(1)},
	{float32(1), float32(1)},
	{float64(1), float64(1)},
	{int64(1), int64(1)},
	{"a", "a"},
	{&google_protobuf1.DoubleValue{Value: 1}, &google_protobuf1.DoubleValue{Value: 1}},
}
var dataGreater = [][]interface{}{
	{int32(1), int32(2)},
	{float32(1), float32(2)},
	{float64(1), float64(2)},
	{int64(1), int64(2)},
	{"a", "b"},
	{&google_protobuf1.DoubleValue{Value: 1}, &google_protobuf1.DoubleValue{Value: 2}},
}
var dataLesser = [][]interface{}{
	{int32(2), int32(1)},
	{float32(2), float32(1)},
	{float64(2), float64(1)},
	{int64(2), int64(1)},
	{"b", "a"},
	{&google_protobuf1.DoubleValue{Value: 2}, &google_protobuf1.DoubleValue{Value: 1}},
}
var dataError = [][]interface{}{
	{int32(2), ""},
	{float32(2), ""},
	{float64(2), ""},
	{int64(2), ""},
	{"b", 1},
	{&google_protobuf1.DoubleValue{Value: 2}, 1},
	{interface{}(1), interface{}(1)},
}

func TestCompGreaterValue(t *testing.T) {
	for _, data := range dataEqual {
		ret, err := compGreaterValue(data[0], data[1])
		assert.NoError(t, err)
		assert.False(t, ret)
	}
	for _, data := range dataLesser {
		ret, err := compGreaterValue(data[0], data[1])
		assert.NoError(t, err)
		assert.False(t, ret)
	}
	for _, data := range dataGreater {
		ret, err := compGreaterValue(data[0], data[1])
		assert.NoError(t, err)
		assert.True(t, ret)
	}
	for _, data := range dataError {
		_, err := compGreaterValue(data[0], data[1])
		assert.Error(t, err)
	}
}

func TestCompGreaterEqualValue(t *testing.T) {
	for _, data := range dataEqual {
		ret, err := compGreaterEqualValue(data[0], data[1])
		assert.NoError(t, err)
		assert.True(t, ret)
	}
	for _, data := range dataLesser {
		ret, err := compGreaterEqualValue(data[0], data[1])
		assert.NoError(t, err)
		assert.False(t, ret)
	}
	for _, data := range dataGreater {
		ret, err := compGreaterEqualValue(data[0], data[1])
		assert.NoError(t, err)
		assert.True(t, ret)
	}
	for _, data := range dataError {
		_, err := compGreaterEqualValue(data[0], data[1])
		assert.Error(t, err)
	}
}

func TestCompEqualValue(t *testing.T) {
	for _, data := range dataEqual {
		ret, err := compEqualValue(data[0], data[1])
		assert.NoError(t, err)
		assert.True(t, ret)
	}
	for _, data := range dataLesser {
		ret, err := compEqualValue(data[0], data[1])
		assert.NoError(t, err)
		assert.False(t, ret)
	}
	for _, data := range dataGreater {
		ret, err := compEqualValue(data[0], data[1])
		assert.NoError(t, err)
		assert.False(t, ret)
	}
	for _, data := range dataError {
		_, err := compEqualValue(data[0], data[1])
		assert.Error(t, err)
	}
}

func TestCompLesserValue(t *testing.T) {
	for _, data := range dataEqual {
		ret, err := compLesserValue(data[0], data[1])
		assert.NoError(t, err)
		assert.False(t, ret)
	}
	for _, data := range dataLesser {
		ret, err := compLesserValue(data[0], data[1])
		assert.NoError(t, err)
		assert.True(t, ret)
	}
	for _, data := range dataGreater {
		ret, err := compLesserValue(data[0], data[1])
		assert.NoError(t, err)
		assert.False(t, ret)
	}
	for _, data := range dataError {
		_, err := compLesserValue(data[0], data[1])
		assert.Error(t, err)
	}
}

func TestCompLesserEqualValue(t *testing.T) {
	for _, data := range dataEqual {
		ret, err := compLesserEqualValue(data[0], data[1])
		assert.NoError(t, err)
		assert.True(t, ret)
	}
	for _, data := range dataLesser {
		ret, err := compLesserEqualValue(data[0], data[1])
		assert.NoError(t, err)
		assert.True(t, ret)
	}
	for _, data := range dataGreater {
		ret, err := compLesserEqualValue(data[0], data[1])
		assert.NoError(t, err)
		assert.False(t, ret)
	}
	for _, data := range dataError {
		_, err := compLesserEqualValue(data[0], data[1])
		assert.Error(t, err)
	}
}
