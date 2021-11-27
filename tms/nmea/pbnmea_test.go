package nmea

import (
	"fmt"
	"testing"
)

func TestPbNmea(t *testing.T) {

	s, _ := Parse("$GPGLL,3554.4456,N,528.9195,W,,,*5A")

	sen, _ := PopulateProtobuf(s)

	fmt.Printf("%+v\n", sen)

	s, _ = Parse("$GPZDA,135137,10,12,2015,0,0*4E")

	sen, _ = PopulateProtobuf(s)

	fmt.Printf("%+v\n", sen)

	s, _ = Parse("!AIVDO,1,1,,,102CuBkP00OLuClC>nL00?wF0000,0*11")

	sen, _ = PopulateProtobuf(s)

	fmt.Printf("%+v\n", sen)

	s, _, _ = ParseArray([]string{"!AIVDM,2,1,8,A,53WaP2000000<p7C;KIH:0luE=<622222222220j1@53340Ht0000000,0*71", "!AIVDM,2,2,8,A,000000000000000,2*2C"})

	sen, _ = PopulateProtobuf(s)

	fmt.Printf("%+v\n", sen)

	s, _ = Parse("!AIVDM,1,1,,,E>jM4;7:0W3Ra@6RR@I00000000?kEJL:A0KP10888o>:3AB12@000000000@,*07")

	sen, _ = PopulateProtobuf(s)

	fmt.Printf("%+v\n", sen)

	s, _ = Parse("$GPRMC,123519,A,4807.038,N,01131.000,E,022.4,084.4,230394,003.1,W*6A")

	sen, _ = PopulateProtobuf(s)

	fmt.Printf("%+v\n", sen)

	s, _ = Parse("!BSVDM,1,1,,A,17ldoM8P?w<tSF0l4Q@>4?v00d2F,0*44")

	sen, _ = PopulateProtobuf(s)

	fmt.Printf("%+v\n", sen)

}
