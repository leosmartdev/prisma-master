package sit185

import "testing"
import "regexp"

func TestSit185_Subtest(t *testing.T) {

	fn := []string{"msg_num", "date", "hex_id", "doppler_a", "encoded", "confirmed", "doa", "doppler_b"}
	regexpstr := []string{
		"2\\..+?MSG.*?NO.+?(?P<msg_num>\\d+).*",
		"3\\..+?DETECTED AT.+?(?P<date>.+UTC)",
		"11\\..+?HEX.+ID\\s*(?P<hex_id>\\S*)\\s",
		"(?m)8\\.(?s).*DOPPLER\\s*A\\s*[-|–]\\s*(?P<doppler_a>NIL|UNKNOWN|(?P<doppler_a_lat_degree>\\d+).*?(?P<doppler_a_lat_min>\\d+\\.\\d+).*?(?P<doppler_a_lat_cardinal_point>[N|S]).*?(?P<doppler_a_lon_degree>\\d+).*?(?P<doppler_a_lon_min>\\d+\\.\\d+).*?(?P<doppler_a_lon_cardinal_point>[E|W]).*?(?P<doppler_a_prob>\\d+|$).*?).*9\\.",
		"8\\.(?s).*ENCODED\\s*?[-|–]\\s*(?P<encoded>NIL|UNKNOWN|(?P<encoded_lat_degree>.*?\\d+)(?P<encoded_lat_min>.*?\\d+\\.\\d+)(?P<encoded_lat_cardinal_point>.*?[N|S]).*?(?P<encoded_lon_degree>\\d+)(?P<encoded_lon_min>.*?\\d+\\.\\d+)(?P<encoded_lon_cardinal_point>.*?[E|W]).*?\n).*9\\.",
		"8\\.(?s).*CONFIRMED\\s*?[–|-]\\s*(?P<confirmed>NIL|UNKNOWN|(?P<confirmed_lat_degree>.*?\\d+)(?P<confirmed_lat_min>.*?\\d+\\.\\d+)(?P<confirmed_lat_cardinal_point>.*?[N|S])(?P<confirmed_lon_degree>.*?\\d+)(?P<confirmed_lon_min>.*?\\d+\\.\\d+)(?P<confirmed_lon_cardinal_point>.*?[E|W]).*?\n).*9\\.",
		"8\\.(?s).*DOA\\s*[–|-]\\s*(?P<doa>((?P<doa_lat_degree>.*?\\d+)(?P<doa_lat_min>.*?\\d+\\.\\d+)(?P<doa_lat_cardinal_point>.*?[N|S])(?P<doa_lon_degree>.*?\\d+)(?P<doa_lon_min>.*?\\d+\\.\\d+)(?P<doa_lon_cardinal_point>.*?[E|W]))(.*?ALTITUDE\\s*(?P<doa_elevation>.*?\\d+)|)).*9\\.",
		"(?m)8\\.(?s).*DOPPLER\\s*B\\s*[–|-]\\s*(?P<doppler_b>NIL|UNKNOWN|(?P<doppler_b_lat_degree>\\d+).*?(?P<doppler_b_lat_min>\\d+\\.\\d+).*?(?P<doppler_b_lat_cardinal_point>[N|S]).*?(?P<doppler_b_lon_degree>\\d+).*?(?P<doppler_b_lon_min>\\d+\\.\\d+).*?(?P<doppler_b_lon_cardinal_point>[E|W]).*?(?P<doppler_b_prob>\\d+|$).*?).*9\\.",
	}

	tt := []struct {
		name       string
		msg        string
		regexpstrs []string
		fieldname  []string
	}{
		{"DISTRESS COSPAS-SARSAT ALERT", sit185DCSA, regexpstr, fn},
		{"SAMPLE 406 MHz UNRESOLVED DOPPLER POSITION MATCH", UnresolvedDoppPos, regexpstr, fn},
		{"SAMPLE 406 MHz INITIAL ENCODED POSITION ALERT", InitEncPosAlert, regexpstr, fn},
		{"SAMPLE 406 MHz INITIAL ALERT WITH NO LOCATION", InitAltNoLoc, regexpstr, fn},
		{"SAMPLE 406 MHz POSITION CONFIRMATION ALERT", PosConfAlt, regexpstr, fn},
		{"SAMPLE 406 MHz POSITION CONFIRMATION ALERT (SGB, LOCATION - PLB)", PosConfAltPLB, regexpstr, fn},
		{"SAMPLE 406 MHz DOA POSITION CONFIRMATION ALERT", DoaPosConfAlt, regexpstr, fn},
		{"SAMPLE 406 MHz NOCR ENCODED POSITION ALERT", NocrEncPosAlt, regexpstr, fn},
		{"SAMPLE 406 MHz INITIAL DOPPLER POSITION ALERT", InitDopPosAlt, regexpstr, fn},
		{"SAMPLE 406 MHz INITIAL DOA POSITION ALERT", InitDoaPosAlt, regexpstr, fn},
		{"SAMPLE 406 MHz INITIAL ALERT", InitAlt, regexpstr, fn},
		{"SAMPLE 406 MHz ALERT WITH UNRELIABLE BEACON MESSAGE", AltUnrBeaMssg, regexpstr, fn},
		{"SAMPLE 406 MHz CONFIRMED UPDATE POSITION ALERT", ConfUpdPosAlt, regexpstr, fn},
		{"SAMPLE 406 MHz POSITION ALERT", PosAlt, regexpstr, fn},
		{"SAMPLE 406 MHz DOPPLER POSITION CONFLICT ALERT", DopPosConfAlt, regexpstr, fn},
		{"SAMPLE 406 MHz DOPPLER INITIAL ALERT", DopInitAlt, regexpstr, fn},
		{"SAMPLE 406 MHz DOPPLER CONFIRMED ALERT", DopConfAlt, regexpstr, fn},
		{"Malaysia 406 MHz CONFIRMED", MalaysiaSit185CONFIRMED, regexpstr, fn},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			var rexps []RegularExpPattern
			for key, regexpstr := range tc.regexpstrs {
				rexp, err := regexp.Compile(regexpstr)
				if err != nil {
					t.Errorf("%s failed because of: %v", tc.name, err)
				}

				rexps = append(rexps, RegularExpPattern{tc.fieldname[key], rexp})

			}

			sit, err := Parse(tc.msg, rexps)
			if err != nil {
				t.Errorf("%s failed because of: %v", tc.name, err)
			}

			t.Logf("HexId: %+v", sit.Fields["hex_id"])
			t.Logf("Confirmed: %+v", sit.Fields["confirmed"])
			t.Logf("Doa: %+v", sit.Fields["doa"])
			t.Logf("Doppler A:  %+v", sit.Fields["doppler_a"])
			t.Logf("Doppler B: %+v", sit.Fields["doppler_b"])
			t.Logf("Date:  %+v", sit.Fields["date"])
			t.Logf("Encoded: %+v", sit.Fields["encoded"])
			t.Logf("All Fields: %+v", sit.Fields)
		})
	}
}
