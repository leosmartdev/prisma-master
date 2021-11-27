package geo

import "testing"

func TestArcmin2DecimalLatitude(t *testing.T) {
	latitude, error := Arcmin2Decimal(111.038, "N")
	if error != nil {
		t.Errorf("error=%v", error)
	} else {
		t.Logf("latitude=%v", latitude)
	}
}

func TestArcmin2DecimalLongitude(t *testing.T) {
	longitude, error := Arcmin2Decimal(10401.394, "E")
	if error != nil {
		t.Errorf("error=%v", error)
	} else {
		t.Logf("longitude=%v", longitude)
	}
}
