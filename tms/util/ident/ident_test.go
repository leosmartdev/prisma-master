package ident

import "testing"

func TestBuilder(t *testing.T) {
	want := "(foo:1)(bar:2)"
	got := With("foo", 1).With("bar", 2).String()
	if want != got {
		t.Fatalf("wanted %v ; got %v", want, got)
	}
}

func TestBuilderBytes(t *testing.T) {
	want := "(foo:12345678)(bar:ABCDEFGH)"
	value1 := "12345678"
	value2 := "ABCDEFGH"
	got := With("foo", value1).With("bar", value2).String()
	if want != got {
		t.Fatalf("wanted %v ; got %v", want, got)
	}
	wantb := "(foo:[49 50 51 52 53 54 55 56])(bar:[65 66 67 68 69 70 71 72])"
	gotb := With("foo", []byte(value1)).With("bar", []byte(value2)).String()
	if wantb != gotb {
		t.Fatalf("wanted %v ; got %v", wantb, gotb)
	}
}

func TestHashMD5(t *testing.T) {
	want := "22286048b79e718ee6c863e282219a38"
	got := With("foo", 1).With("bar", 2).Hash()
	if want != got {
		t.Fatalf("wanted %v ; got %v", want, got)
	}

}
