// Package rule provides extra functions for rule engine.
package rule

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strings"

	"prisma/tms/util/rule/op"

	"github.com/globalsign/mgo/bson"
)

var (
	errNoField = errors.New("no field")
)

// Matches returns true if all the expressions evaluate to true using
// obj. If obj is not a bson.M, it assumes it is a struct and looks up
// field names via reflection
func Matches(exprs bson.M, obj interface{}) (bool, error) {
	return evalMap(op.And, exprs, obj)
}

func evalMap(oper string, exprs bson.M, obj interface{}) (bool, error) {
	first := true
	var result bool

	for k, v := range exprs {
		var value bool
		var err error
		if strings.HasPrefix(k, "$") {
			// if an operator here, the value must be a map
			bmap, ok := v.(bson.M)
			if !ok {
				return false, fmt.Errorf("expecting bson.M but got %v",
					reflect.TypeOf(v))
			}
			value, err = evalMap(k, bmap, obj)
		} else {
			value, err = evalField(k, v, obj)
		}
		if err != nil {
			return false, err
		}

		// Keep previous result to test with later results
		if first {
			result = value
			first = false
		} else {
			switch oper {
			case op.And:
				result = result && value
			case op.Or:
				result = result || value
			default:
				return false, fmt.Errorf("unexpected operator %v", oper)
			}
		}
		// Short circuit for ands
		if oper == op.And && !result {
			return false, nil
		}
	}
	return result, nil
}

func evalField(field string, matcher interface{}, obj interface{}) (bool, error) {
	value, err := lookup(field, obj)
	// Not having a field is okay
	if err == errNoField {
		return false, nil
	} else if err != nil {
		return false, err
	}

	switch m := matcher.(type) {
	case bson.M:
		exprs := m
		for op, operand := range exprs {
			result, err := evalOp(value, op, operand)
			if !result || err != nil {
				return false, err
			}
		}
		return true, nil
	}
	// simple equality if there is no operator
	return matcher == value, nil
}

func evalOp(value interface{}, oper string, operand interface{}) (bool, error) {
	switch oper {
	// for all of these comparsions, do them as floats
	case op.Gt, op.Gte, op.Lt, op.Lte:
		return evalFloat(value, oper, operand)
	case op.Eq:
		return value == operand, nil
	case op.RegEx:
		return evalRegEx(value, oper, operand)
	}
	return false, fmt.Errorf("unexpected operator %v", oper)
}

func evalFloat(a interface{}, oper string, b interface{}) (bool, error) {
	f1, err := toFloat(a)
	if err != nil {
		return false, err
	}
	f2, err := toFloat(b)
	if err != nil {
		return false, err
	}
	return evalFloatExpr(f1, oper, f2)
}

func evalFloatExpr(f1 float64, oper string, f2 float64) (bool, error) {
	switch oper {
	case op.Gt:
		return f1 > f2, nil
	case op.Gte:
		return f1 >= f2, nil
	case op.Lt:
		return f1 < f2, nil
	case op.Lte:
		return f1 <= f2, nil
	}
	return false, fmt.Errorf("unexpected operator %v", oper)
}

func toFloat(v interface{}) (float64, error) {
	switch f := v.(type) {
	case int:
		return float64(f), nil
	case int8:
		return float64(f), nil
	case int16:
		return float64(f), nil
	case int32:
		return float64(f), nil
	case int64:
		return float64(f), nil
	case uint:
		return float64(f), nil
	case uint8:
		return float64(f), nil
	case uint16:
		return float64(f), nil
	case uint32:
		return float64(f), nil
	case uint64:
		return float64(f), nil
	case float32:
		return float64(f), nil
	case float64:
		return f, nil
	default:
		return math.NaN(), fmt.Errorf("expecting a number, got %v", v)
	}
}

func evalRegEx(a interface{}, oper string, b interface{}) (bool, error) {
	str, ok := a.(string)
	if !ok {
		return false, fmt.Errorf("expected string, got %v", a)
	}
	def, ok := b.(bson.RegEx)
	if !ok {
		return false, fmt.Errorf("expected regex, but got: %v", def)
	}
	re, err := regexp.Compile(def.Pattern)
	if err != nil {
		return false, err
	}
	return re.MatchString(str), nil
}

func lookup(path string, obj interface{}) (interface{}, error) {
	switch target := obj.(type) {
	case bson.M:
		return lookupMap(path, target)
	}
	return lookupObj(path, obj)
}

func lookupMap(path string, obj bson.M) (interface{}, error) {
	names := strings.Split(path, ".")
	node := obj
	ok := true
	var value interface{}
	for i, name := range names {
		value, ok = node[name]
		if !ok {
			return nil, errNoField
		}
		if i < len(names)-1 {
			node, ok = value.(bson.M)
			if !ok {
				return nil, errNoField
			}
		}
	}
	return value, nil
}

func lookupObj(path string, obj interface{}) (interface{}, error) {
	names := strings.Split(path, ".")
	node := reflect.ValueOf(obj)
	var value reflect.Value
	for i, name := range names {
		value = node.FieldByName(name)
		if !value.IsValid() {
			return nil, errNoField
		}
		if i < len(names)-1 {
			node = value
		}
	}
	return value.Interface(), nil
}
