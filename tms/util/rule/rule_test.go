package rule

import (
	"testing"

	"github.com/globalsign/mgo/bson"
)

type Bar struct {
	Baz int
}

type Foo struct {
	Bar Bar
}

type Debug struct {
	Foo Foo
}

func TestMatchPass(t *testing.T) {
	val := bson.M{"foo": 2}
	pred := bson.M{"foo": 2}
	yes, err := Matches(pred, val)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !yes {
		t.Fatalf("expecting a match")
	}
}

func TestMatchFail(t *testing.T) {
	val := bson.M{"foo": 4}
	pred := bson.M{"foo": 2}
	yes, err := Matches(pred, val)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if yes {
		t.Fatalf("not expecting a match")
	}
}

func TestPathPass(t *testing.T) {
	val := bson.M{"foo": bson.M{"bar": bson.M{"baz": 1}}}
	pred := bson.M{"foo.bar.baz": 1}
	yes, err := Matches(pred, val)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !yes {
		t.Fatalf("expecting a match")
	}
}

func TestPathFail(t *testing.T) {
	val := bson.M{"foo": bson.M{"bar": bson.M{"baz": 1}}}
	pred := bson.M{"foo.xxx.baz": 1}
	yes, err := Matches(pred, val)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if yes {
		t.Fatalf("not expecting a match")
	}
}

func TestReflectPathPass(t *testing.T) {
	val := Debug{Foo{Bar{Baz: 1}}}
	pred := bson.M{"Foo.Bar.Baz": 1}
	yes, err := Matches(pred, val)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !yes {
		t.Fatalf("expecting a match")
	}
}

func TestGreaterThanPass(t *testing.T) {
	val := bson.M{"foo": 20}
	pred := bson.M{"foo": bson.M{"$gt": 10}}
	yes, err := Matches(pred, val)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !yes {
		t.Fatalf("expecting a match")
	}
}

func TestGreaterThanFail(t *testing.T) {
	val := bson.M{"foo": 10}
	pred := bson.M{"foo": bson.M{"$gt": 20}}
	yes, err := Matches(pred, val)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if yes {
		t.Fatalf("not expecting a match")
	}
}

func TestGreaterThanEqualPass(t *testing.T) {
	val := bson.M{"foo": 20}
	pred := bson.M{"foo": bson.M{"$gte": 10}}
	yes, err := Matches(pred, val)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !yes {
		t.Fatalf("expecting a match")
	}

	val = bson.M{"foo": 20}
	pred = bson.M{"foo": bson.M{"$gte": 20}}
	yes, err = Matches(pred, val)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !yes {
		t.Fatalf("expecting a match")
	}
}

func TestGreaterThanEqualFail(t *testing.T) {
	val := bson.M{"foo": 10}
	pred := bson.M{"foo": bson.M{"$gte": 20}}
	yes, err := Matches(pred, val)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if yes {
		t.Fatalf("not expecting a match")
	}
}

func TestLessThanPass(t *testing.T) {
	val := bson.M{"foo": 5}
	pred := bson.M{"foo": bson.M{"$lt": 10}}
	yes, err := Matches(pred, val)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !yes {
		t.Fatalf("expecting a match")
	}
}

func TestLessThanFail(t *testing.T) {
	val := bson.M{"foo": 10}
	pred := bson.M{"foo": bson.M{"$lt": 5}}
	yes, err := Matches(pred, val)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if yes {
		t.Fatalf("not expecting a match")
	}
}

func TestLessThanEqualPass(t *testing.T) {
	val := bson.M{"foo": 5}
	pred := bson.M{"foo": bson.M{"$lte": 10}}
	yes, err := Matches(pred, val)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !yes {
		t.Fatalf("expecting a match")
	}

	val = bson.M{"foo": 20}
	pred = bson.M{"foo": bson.M{"$lte": 20}}
	yes, err = Matches(pred, val)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !yes {
		t.Fatalf("expecting a match")
	}
}

func TestLessThanEqualFail(t *testing.T) {
	val := bson.M{"foo": 20}
	pred := bson.M{"foo": bson.M{"$lte": 10}}
	yes, err := Matches(pred, val)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if yes {
		t.Fatalf("not expecting a match")
	}
}

func TestEqualsPass(t *testing.T) {
	val := bson.M{"foo": 10}
	pred := bson.M{"foo": bson.M{"$eq": 10}}
	yes, err := Matches(pred, val)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !yes {
		t.Fatalf("expecting a match")
	}
}

func TestEqualsFail(t *testing.T) {
	val := bson.M{"foo": 40}
	pred := bson.M{"foo": bson.M{"$eq": 10}}
	yes, err := Matches(pred, val)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if yes {
		t.Fatalf("not expecting a match")
	}
}

func TestAndPass(t *testing.T) {
	val := bson.M{"foo": "bar", "one": 1}
	pred := bson.M{"$and": bson.M{"foo": "bar", "one": 1}}
	yes, err := Matches(pred, val)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !yes {
		t.Fatalf("expecting a match")
	}
}

func TestAndFail(t *testing.T) {
	val := bson.M{"foo": "bar", "one": 1}
	pred := bson.M{"$and": bson.M{"foo": "bar", "one": 2}}
	yes, err := Matches(pred, val)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if yes {
		t.Fatalf("not expecting a match")
	}
}

func TestOrPass(t *testing.T) {
	val := bson.M{"foo": "bar", "one": 1}
	pred := bson.M{"$or": bson.M{"foo": "bar", "one": 2}}
	yes, err := Matches(pred, val)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !yes {
		t.Fatalf("expecting a match")
	}
}

func TestOrFail(t *testing.T) {
	val := bson.M{"foo": "bar", "one": 1}
	pred := bson.M{"$or": bson.M{"foo": "baz", "one": 2}}
	yes, err := Matches(pred, val)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if yes {
		t.Fatalf("not expecting a match")
	}
}

func TestRegExPass(t *testing.T) {
	val := bson.M{"foo": "bombardier"}
	pred := bson.M{"foo": bson.M{"$regex": bson.RegEx{
		Pattern: ".*bar.*",
	}}}
	yes, err := Matches(pred, val)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !yes {
		t.Fatalf("expecting a match")
	}
}

func TestRegExFail(t *testing.T) {
	val := bson.M{"foo": "bomber"}
	pred := bson.M{"foo": bson.M{"$regex": bson.RegEx{
		Pattern: ".*bar.*",
	}}}
	yes, err := Matches(pred, val)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if yes {
		t.Fatalf("not expecting a match")
	}
}
