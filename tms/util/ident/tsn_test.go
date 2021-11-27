package ident

import "testing"

func TestTSN(t *testing.T) {
	seconds = 0
	Clock = func() int64 {
		return 10
	}
	s, c := TSN()
	if s != 10 {
		t.Fatalf("expecting 10 ; got %v", s)
	}
	if c != 1 {
		t.Fatalf("expecting 1 ; got %v", c)
	}
}

func TestTSNCounterIncrement(t *testing.T) {
	seconds = 0
	Clock = func() int64 {
		return 10
	}
	s, c := TSN()
	s, c = TSN()
	if s != 10 {
		t.Fatalf("expecting 10 ; got %v", s)
	}
	if c != 2 {
		t.Fatalf("expecting 2 ; got %v", c)
	}
}

func TestTSNTimeIncrement(t *testing.T) {
	seconds = 0
	Clock = func() int64 {
		return 10
	}
	s, c := TSN()
	Clock = func() int64 {
		return 11
	}
	s, c = TSN()
	if s != 11 {
		t.Fatalf("expecting 11 ; got %v", s)
	}
	if c != 1 {
		t.Fatalf("expecting 1 ; got %v", c)
	}
}
