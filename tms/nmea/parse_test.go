package nmea

import (
	"fmt"
	"testing"
)


func TestParse(t *testing.T) {

	/*  var str = []string {"!AIVDM,1,1,,B,E>jM4;7:0W3Ra@6RR@I00000000?kEJL:A0KP10888o>:0,4*42",
	                        "!AIVDO,1,1,,,102CuBkP00OLuClC>nL00?wF0000,0*11",
	                        "!AIVDM,1,1,,A,13oh8v02j?wUkDDDRg0k52a428Cr,0*30",
	                        "!AIVDM,1,1,,A,13cV5b002IwV9FNDS28BuBQ200S@,0*35",
	                        "!AIVDO,1,1,,,102Cu;?P01wVquDDRv417gw62000,0*22",
	                        "!AIVDM,1,1,,B,A028n?ktedb<02H0a7l`2OqBtm@3w3ok:P;ue?E<,0*0C",
	                        "!AIVDM,1,1,,B,14Vmsj001vwW2`TDU;`jf2>v00SS,0*5F",
	                        "!AIVDM,2,1,0,A,58wt8Ui`g??r21`7S=:22058<v05Htp000000015>8OA;0sk,0*7B",
	                        "!AIVDM,2,2,0,A,eQ8823mDm3kP00000000000,2*5D",
	                        "!AIVDM,1,1,,A,13F<4J04BlwWSEtDUvtVA55400S>,0*15",
	                        "!AIVDM,2,1,9,B,61c2;qLPH1m@wsm6ARhp<ji6ATHd<C8f=Bhk>34k;S8i=3To,0*2C",
	                        "!AIVDM,2,2,9,B,Djhi=3Di<2pp=34k>4D,2*03","$GPZDA,135137,10,12,2015,0,0*4E",
	                        "$GPTTM,01,4.54,354.0,T,,,T,4.46,,N,,Q,,,M*0A",
	                        "$GPTTM,02,3.56,297.1,T,,,T,1.88,,N,,Q,,,M*04",
	                        "$RATTM,01,4.54,353.9,T,8.90,264.8,T,4.54,-0.5,N,,Q,,,M*3C",
	                        "$RATTM,02,3.55,297.3,T,,,T,0.13,,N,,Q,,,M*02",
	                        "$GPRMC,135137,A,3554.4456,N,00528.9195,W,0.0,256.1,101215,1,W*62",
	                        "$GPRMB,A,0.00,R,START ,001   ,3554.443,N,00528.919,W,0.0,164.3,-0.0,A*7B",
	                        "$GPRMB,A,0.66,L,003,004,4917.24,N,12309.57,W,001.3,052.5,000.5,V*20",
	                        "$GPGLL,3554.4456,N,00528.9195,W,135138,A*3A",
	                        "$GPGGA,135138,3554.4456,N,00528.9195,W,1,8,1.8,23,M,48,M,,*5E",
	                        "$GPVTG,,T,258.9,M,0.0,N,0.0,K*66","!AIVDM,1,1,,B,342O:cA1h2wWKD0DcGKaLIAb2Dw:,0*64",
	                        "!AIVDM,2,1,8,A,53WaP2000000<p7C;KIH:0luE=<622222222220j1@53340Ht0000000,0*71",
	                        "!AIVDM,2,2,8,A,000000000000000,2*2C",
	                        "!AIVDM,1,1,,B,13Fw280000wW8uLDc?HcQ09b0D3S,0*20",
	                        "!AIVDM,1,1,,B,4028ipQuw5;QlwVLddDW0aG000S:,0*63",
	                        "!AIVDM,2,1,0,A,58wt8Ui`g??r21`7S=:22058<v05Htp000000015>8OA;0sk,0*7B",
	                        "!AIVDM,2,2,0,A,eQ8823mDm3kP00000000000,2*5D",
	                        "$GPZDA,153415,10,12,2015,0,0*4B",
	                        "$GPRMC,153415,A,3555.2532,N,00524.1251,W,0.0,236.0,101215,1,W*6B",
	                        "$GPRMB,A,0.01,L,START ,001   ,3555.257,N,00524.130,W,0.0,310.7,0.0,V*56",
	                        "$GPGLL,3555.2532,N,00524.1252,W,153416,A*3B",
	                        "$GPGGA,153416,3555.2532,N,00524.1252,W,1,8,1.1,5,M,48,M,,*62",
	                        "$GETTM,4,5.122,14.00,T,0.000,269.90,T,0.000,0.000,N,,Q,,114526.20,M*19",
	                        "$GETTM,2,7.180,348.20,T,26.500,82.50,T,0.000,0.000,N,,Q,,114528.76,M*2A",
	                        "$GETTM,3,5.429,345.40,T,2.000,345.40,T,0.000,0.000,N,,Q,,114528.76,M*2E",
	                        "$GETTM,4,5.122,14.00,T,0.000,269.90,T,0.000,0.000,N,,Q,,114528.93,M*1F",
	                        "$RATTM,01,4.55,354.8,T,,,T,4.38,,N,,Q,,,M*0E",
	                        "$RATTM,01,4.55,354.8,T,,,T,4.38,,N,,Q,,,M*0E",
	                        "!AIVDM,1,1,,B,E>jM4;7:0W3Ra@6RR@I00000000?kEJL:A0KP10888o>:,4*72",
				"!AIVDM,1,1,0,A,:3u?etP0V:C0,0*3B",
	                        "!AIVDM,1,1,,B,E>jM4;7:0W3Ra@2V6RT24hI0000?kEJL:A0KP10888o>:00,4*3E",
	                        "!AIVDM,1,1,0,B,E>jM4;7:0W3Ra@6RR@I00000000?kEJL:A0KP10888o>:00,4*42","!AIVDM,1,1,0,B,E>jM4;7:0W3Ra@6RR@I00000000?kEJL:A0KP10888o>:;IJ9:@0000000000,4*39",
				"$GPRMB,A,0.66,L,003,004,4917.24,N,12309.57,W,001.3,052.5,000.5,V*20",                       			   "!AIVDM,1,1,0,B,E>jM4;7:0W3Ra@6RR@I00000000?kEJL:A0KP10888o>:3AB12@000000000@,4*41"}
	*/

	s, err := Parse("$GPGLL,3554.4456,N,528.9195,W,,,*5A")

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%+v\n\n\n", s)
	}

	s, err = Parse("!AIVDO,1,1,,,102CuBkP00OLuClC>nL00?wF0000,0*11")

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%+v\n\n\n", s)
	}

	s, err, _ = ParseArray([]string{"!AIVDM,2,1,8,A,53WaP2000000<p7C;KIH:0luE=<622222222220j1@53340Ht0000000,0*71", "!AIVDM,2,2,8,A,000000000000000,2*2C"})

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%+v\n", s)
	}

	s, err = Parse("!AIVDM,1,1,0,B,E>jM4;7:0W3Ra@6RR@I00000000?kEJL:A0KP10888o>:3AB12@000000000@,4*41")

	if err != nil {
		fmt.Println(err)
		fmt.Print("\n\n\n")
	} else {
		fmt.Printf("%+v\n\n\n", s)
	}

	s, err = Parse("$RATTM,01,4.54,353.9,T,8.90,264.8,T,4.54,-0.5,N,,Q,,,M*3C")

	if err != nil {
		fmt.Println(err)
		fmt.Print("\n\n\n")
	} else {
		fmt.Printf("%+v\n\n\n", s)
	}

	s, err = Parse("!AIVDM,1,1,0,B,E>jM4;7:0W3Ra@6RR@I00000000?kEJL:A")

	if err != nil {
		fmt.Println(err)
		fmt.Print("\n\n\n")
	} else {
		fmt.Printf("%+v\n\n\n", s)
	}

	s, err = Parse("!AIVDM,2,1,8,A,53WaP2000000<p7C;KIH:0luE=<622222222220j1@53340Ht0000000,0*71")

	if err != nil {
		fmt.Println(err)
		fmt.Print("\n\n\n")
	} else {
		fmt.Printf("%+v\n\n\n", s)
	}

	s, err = Parse("!AIVDM,2,2,8,A,000000000000000,2*2C")

	if err != nil {
		fmt.Println(err)
		fmt.Print("\n\n\n")
	} else {
		fmt.Printf("%+v\n\n\n", s)
	}

	s, err, _ = ParseArray([]string{"!AIVDM,2,1,0,B,C8u:8C@t7@TnGCKfm6Po`e6N`:Va0L2J;06HV50JV?SjBPL3,0*28", "!AIVDM,2,2,0,B,11RP,0*17"})

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%+v\n\n\n", s)
	}

	s, err = Parse("$RATTM,01,4.54,353.9,T,8.90,264.8,T,4.54,-0.5,N")

	if err != nil {
		fmt.Println(err)
		fmt.Print("\n\n\n")
	} else {
		fmt.Printf("%+v\n\n\n", s)
	}

	s, err = Parse("!BSVDM,1,1,,B,1005KowP?w<tSF0l4Q@>4?wv0TST,0*00")

	if err != nil {
		fmt.Println(err)
		fmt.Print("\n\n\n")
	} else {
		fmt.Printf("%+v\n\n\n", s)
	}

	s, err, _ = ParseArray([]string{"!BSVDM,2,1,9,A,5005Koh00000lV22220l5@60Td4r22222222221S1p@3400Ht6H888888888,0*04", "!BSVDM,2,2,9,A,88888888888,2*3C"})

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%+v\n\n\n", s)
	}

	s, err = Parse("$GPRMC,123519,A,4807.038,N,01131.000,E,022.4,084.4,230394,003.1,W*6A")
	if err != nil {
		fmt.Printf("RMC error: %+v\n", s)
	} else {
		fmt.Printf("\nRMC: %+v", s)
	}

	s, err = Parse("!BSVDM,1,1,,A,17ldoM8P?w<tSF0l4Q@>4?v00d2F,0*44")
	if err != nil {
		fmt.Printf("%+v\n", s)
	} else {
		fmt.Printf("%+v\n", s)
	}

	fmt.Println("I am here")
	fmt.Println()

	s, err = Parse("!AIVDM,1,1,0,A,4951805511150804849844848484848484848484858104484848484848,*2E")
	if err != nil {
		fmt.Printf("%+v\n", s)
	} else {
		fmt.Printf("%+v\n", s)
	}
}
