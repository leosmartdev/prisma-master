package public

import (
	"testing"
	"prisma/tms/moc"
	"github.com/stretchr/testify/assert"
)

func TestDuplicatingDevice(t *testing.T) {
	vessel := new(moc.Vessel)
	assert.Equal(t, "", duplicatingDevice(vessel))
	vessel.Devices = append(vessel.Devices, &moc.Device{Id: "test"})
	vessel.Devices = append(vessel.Devices, &moc.Device{Id: "test_another"})
	assert.Equal(t, "", duplicatingDevice(vessel))
	vessel.Devices = append(vessel.Devices, &moc.Device{Id: "test"})
	assert.Equal(t, "test", duplicatingDevice(vessel))
}
