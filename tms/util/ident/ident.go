// Package ident provides functions to concat information and get HEX of that.
package ident

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

type IdentBuilder struct {
	buf bytes.Buffer
}

func newBuilder() *IdentBuilder {
	return &IdentBuilder{}
}

func (i *IdentBuilder) With(field string, value interface{}) *IdentBuilder {
	i.buf.WriteString(fmt.Sprintf("(%v:%v)", field, value))
	return i
}

func (i *IdentBuilder) Hash() string {
	sum := md5.Sum(i.buf.Bytes())
	return hex.EncodeToString(sum[:])
}

func (i *IdentBuilder) String() string {
	return i.buf.String()
}

func With(field string, value interface{}) *IdentBuilder {
	return newBuilder().With(field, value)
}
