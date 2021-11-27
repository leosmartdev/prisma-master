package main

import (
	"testing"
)

func TestSplitMtData(t *testing.T) {
	bts := make([]byte, 2000)
	for i := 0; i < 2000; i++ {
		bts[i] = byte(i % 256)
	}
	pts, _, err := SplitMtData(bts)
	if err != nil {
		t.Error(err)
	}
	if len(pts) != 8 {
		t.Errorf("Number of packets should be 8")
	}
	for _, pt := range pts {
		if len(pt) > 270 {
			t.Errorf("No part can have lenght more than 270 for MT data")
		}
	}
}
