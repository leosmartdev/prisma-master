// Package rule provides rule engine to issue actions for specific conditions - rules.
package rule

import (
	"reflect"
	"strings"
	"errors"
	"regexp"
)

func getValueFieldByNameICase(ref reflect.Value, field string) reflect.Value {
	field = strings.ToLower(field)
	ref = reflect.Indirect(ref)
	for i := 0; i < ref.NumField(); i++ {
		if strings.ToLower(ref.Type().Field(i).Name) == field {
			return ref.Field(i)
		}
	}
	return reflect.Value{}
}

func getValueBySpecificField(goal interface{}, original interface{},
	field string) (interface{}, interface{}, error) {

	var (
		reflectValueGoal,
		reflectValueOriginal reflect.Value
		i int
	)
	reflectValueGoal = reflect.ValueOf(goal)
	reflectValueOriginal = reflect.ValueOf(original)
	// what is about deeps?
	if strs := strings.Split(field, "."); len(strs) > 1 {
		for i = 0; i < len(strs)-1; i++ {
			reflectValueGoal = getValueFieldByNameICase(reflectValueGoal, strs[i])
			reflectValueOriginal = getValueFieldByNameICase(reflectValueOriginal, strs[i])
			if !reflectValueGoal.IsValid() || !reflectValueOriginal.IsValid() {
				return nil, nil, errors.New("Value is invalid")
			}
			reflectValueGoal = reflectValueGoal.Elem()
			reflectValueOriginal = reflectValueOriginal.Elem()
			if !reflectValueGoal.IsValid() || !reflectValueOriginal.IsValid() {
				return nil, nil, errors.New("Value is invalid")
			}
		}
		field = strs[i]
	}
	reflectValueGoal = getValueFieldByNameICase(reflectValueGoal, field)
	reflectValueOriginal = getValueFieldByNameICase(reflectValueOriginal, field)
	if !reflectValueGoal.IsValid() || !reflectValueOriginal.IsValid() {
		return nil, nil, errors.New("Value is invalid")
	}
	return reflectValueGoal.Interface(), reflectValueOriginal.Interface(), nil
}

func compRegExValue(val, valOrigin interface{}) (bool, error) {
	valRegEx, ok := val.(string)
	valOriginStr, okOrig := valOrigin.(string)
	if !ok || !okOrig {
		return false, errors.New("bad types")
	}
	re, err := regexp.Compile(valRegEx)
	if err != nil {
		return false, err
	}
	return re.MatchString(valOriginStr), nil
}
