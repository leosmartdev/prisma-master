package tms

import (
	"errors"
)

var (
	UnsupportedFeature = errors.New("Unsupported feature")
	UnknownOption      = errors.New("Unknown option encountered")
)
