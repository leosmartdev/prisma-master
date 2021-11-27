package iridium

type MOLocationInformation struct {
	IEI byte

	Length uint16

	Latitude float64

	Longitude float64

	CEP uint32
}

func ParseMOlocationInformation(data []byte) MOLocationInformation {

	var location MOLocationInformation

	location.IEI = data[0]

	location.Length = uint16(data[2]) | uint16(data[1])<<8

	latlong := data[3:10]

	dir := int8(latlong[0])

	Dlat := float64(int8(latlong[1]))

	Minlat := float64(uint16(latlong[3])|uint16(latlong[2])<<8) / 1000 / 60

	location.Latitude = Dlat + Minlat

	Dlong := float64(int8(latlong[4]))

	Minlong := float64(uint16(latlong[6])|uint16(latlong[5])<<8) / 1000 / 60

	location.Longitude = Dlong + Minlong

	if dir == 1 {

		location.Longitude = -1 * location.Longitude

	} else {
		location.Latitude = -1 * location.Latitude
	}

	//location.Longitude = 0.0

	location.CEP = uint32(data[13]) | uint32(data[12])<<8 | uint32(data[11])<<16 | uint32(data[10])<<24

	return location
}
