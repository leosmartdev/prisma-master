package validate

import "testing"

func TestMMSIOk(t *testing.T) {
	arr := []string{"820512345", "203123456", "077512345", "003091234", "009991234", "111654123", "997741234", "984561234", "970100000", "972334444", "974557777"}
	for _, v := range arr {
		errs := MMSI(v)
		if len(errs) > 0 {
			t.Error(errs)
		}
	}
}

func TestMMSIWrongFormat(t *testing.T) {
	arr := []string{"799512345", "803123456", "077612345", "009091234", "009981234", "112654123", "99776123a", "96888asdf", "971a1zxcv", "97z889999", "9]4114444"}
	for _, v := range arr {
		errs := MMSI(v)
		if len(errs) == 0 {
			t.Error("wrong")
		}

	}
}
