package public

import (
	"testing"

	"prisma/tms/moc"
	"prisma/tms/rest"
)

func TestVesselCreateDeviceIdFull(t *testing.T) {
	var devices []*moc.Device
	devices = append(devices, &moc.Device{
		Id:       "abc",
		Type:     "",
		DeviceId: "",
		Networks: nil,
	})
	vessel := &moc.Vessel{
		Name:      "myvessel",
		Devices:   devices,
		Crew:      nil,
		Fleet: nil,
	}
	errs := rest.SanitizeValidate(vessel, SchemaVesselCreate)
	t.Log(errs)
	if len(errs) > 0 {
		t.Error(errs)
	}
}

func TestVesselCreateDeviceIdEmpty(t *testing.T) {
	var devices []*moc.Device
	devices = append(devices, &moc.Device{
		Type:     "",
		DeviceId: "",
		Networks: nil,
	})
	vessel := &moc.Vessel{
		Devices:   devices,
		Crew:      nil,
		Fleet: nil,
	}
	errs := rest.SanitizeValidate(vessel, SchemaVesselUpdate)
	t.Log(errs)
	if len(errs) == 1 {
		t.Error("expected: error")
	}
}
