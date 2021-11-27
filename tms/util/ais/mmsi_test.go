package ais

import "testing"

func TestFormatMMSI(t *testing.T) {
	want := "000000123"
	got := FormatMMSI(123)
	if want != got {
		t.Fatalf("want %v ; got %v", want, got)
	}
}
