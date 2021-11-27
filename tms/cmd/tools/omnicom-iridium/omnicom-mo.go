// omnicom-mo sends prepared MO messages to specific server.
package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

var (
	Address = flag.String("address", "127.0.0.1:7777", "<IPaddress:port> where to send the raw data")
)

func main() {
	flag.Parse()

	tcpAddr, err := net.ResolveTCPAddr("tcp4", *Address)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(0)
	}

	MOHeaderWithLocationInformation := []byte{
		01, 00, 45,
		01,       // IEI Header 1
		00, 0x1C, // MO Header length 28
		0, 0x12, 0xD6, 0x87, // CDR Reference Auto ID 1234567
		0x33, 0x30, 0x30, 0x30, // IMEI "3000..
		0x33, 0x34, 0x30, 0x31, // 3401
		0x30, 0x31, 0x32, 0x33, // 0123
		0x34, 0x35, 0x30, // 450"
		0,          // Session Status,
		0xD4, 0x31, // MOMSN 54321
		0x30, 0x39, // MTMSN 12345
		0x43, 0xB5, 0x39, 0xE1,
		0x03,       // IEI Location Information
		0x00, 0x0B, // Length of data starting from next byte
		0x01,                   // Bit0: 0 - East, 1 - West      Bit1: 0 - North, 1 - South
		0x2F,                   // Latitude degrees. [0-90]
		0xBF,                   // Latitude. Thousandths of a minute. Most Significant byte
		0xD9,                   // Latitude. Thousandths of a minute. Least Significant byte
		0x03,                   // Longitude degrees [0-180]
		0x63,                   // Longitude degrees. Thousandths of a minute. Most Significant byte
		0x21,                   // Longitude degrees. Thousandths of a minute. Least Significant byte
		0x00, 0x00, 0x00, 0x03} // CEP Radius [1-2000]

	MOHeaderWithDataPayload := []byte{1, 0, 95, 1, 0, 28, 153, 86, 64, 67, 51,
		48, 48, 50, 51, 52, 48, 49, 48, 48, 51,
		48, 52, 53, 48, 0, 60, 141, 0, 0, 87, 242,
		100, 93, 3, 0, 11, 1, 47, 185, 60, 3, 110,
		224, 0, 0, 0, 3, 2, 0, 47, 1, 158, 113, 0,
		1, 2, 0, 1, 33, 67, 96, 205, 119, 209, 84,
		29, 196, 91, 5, 1, 20, 0, 0, 0, 0, 0, 0, 8,
		133, 13, 136, 181, 223, 69, 80, 119, 16, 18, 172,
		4, 80, 0, 0, 0, 0, 0, 185}

	MOHeaderWithLocationInformation2 := []byte{1, 0, 63, 1, 0, 28, 153, 165, 140, 64, 51, 48,
		48, 50, 51, 52, 48, 49, 48, 48, 51, 48, 52, 53,
		48, 2, 61, 25, 0, 0, 87, 243, 112, 61, 3, 0, 11,
		1, 47, 189, 12, 2, 207, 193, 0, 0, 0, 107, 2, 0,
		15, 6, 33, 68, 67, 205, 119, 171, 84, 29, 128,
		4, 183, 1, 0, 74}

	MOHistoryMessage := []byte{1, 0, 95, 1, 0, 28, 153, 165, 205, 27,
		51, 48, 48, 50, 51, 52, 48, 49, 48, 48,
		51, 48, 52, 53, 48, 0, 61, 27, 0, 0, 87,
		243, 113, 28, 3, 0, 11, 1, 47, 191, 217,
		3, 99, 33, 0, 0, 0, 2, 2, 0, 47, 1, 3, 225,
		0, 2, 2, 0, 1, 33, 68, 66, 141, 119, 171, 212,
		29, 116, 11, 85, 1, 17, 128, 0, 0, 0, 0, 0, 8,
		133, 17, 15, 53, 222, 175, 80, 117, 208, 18, 220,
		4, 70, 0, 0, 0, 0, 0, 73}

	MOGeofenceconfirmation := []byte{1, 0, 62, 1, 0, 28, 153, 176, 250, 91, 51, 48,
		48, 50, 51, 52, 48, 49, 48, 48, 51, 48, 52, 53,
		48, 0, 61, 45, 1, 119, 87, 243, 152, 68, 3, 0,
		11, 1, 47, 191, 217, 3, 99, 33, 0, 0, 0, 3, 2, 0,
		14, 4, 136, 2, 20, 69, 134, 215, 122, 173, 65, 215, 64, 192, 79}

	MOSinglepositionreport := []byte{1, 0, 63, 1, 0, 28, 153, 177, 102, 48, 51, 48,
		48, 50, 51, 52, 48, 49, 48, 48, 51, 48, 52, 53,
		48, 0, 61, 47, 0, 0, 87, 243, 153, 165, 3, 0, 11,
		1, 47, 185, 156, 3, 98, 98, 0, 0, 0, 9, 2, 0, 15,
		6, 33, 68, 89, 173, 119, 171, 84, 29, 120, 12,
		195, 1, 0, 180}

	MOAlertReport := []byte{1, 0, 67, 1, 0, 28, 153, 178, 119, 172, 51, 48,
		48, 50, 51, 52, 48, 49, 48, 48, 51, 48, 52, 53,
		48, 0, 61, 50, 0, 0, 87, 243, 156, 105, 3, 0, 11,
		1, 47, 191, 217, 3, 99, 33, 0, 0, 0, 3, 2, 0, 19,
		2, 90, 50, 20, 69, 174, 215, 122, 173, 65, 215, 136,
		81, 22, 210, 4, 64, 0, 90}

	MOGlobalParam := []byte{1, 0, 137, 1, 0, 28, 153, 179, 112, 89, 51, 48, 48, 50, 51,
		52, 48, 49, 48, 48, 51, 48, 52, 53, 48, 0, 61, 55, 0, 0, 87, 243,
		159, 90, 3, 0, 11, 1, 47, 185, 60, 3, 110, 224, 0, 0, 0, 5, 2,
		0, 89, 3, 0, 0, 49, 12, 33, 68, 92, 237, 119, 171, 84, 29, 124,
		4, 0, 128, 1, 104, 242, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 4, 1,
		0, 8, 1, 2, 5, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 51, 53, 49, 53, 55, 57, 48, 53, 55, 50, 56, 49, 52,
		48, 50, 51, 48, 48, 50, 51, 52, 48, 49, 48, 48, 51, 48, 52, 53,
		48, 19}

	MOGeoDelete := []byte{1, 0, 62, 1, 0, 28, 154, 22, 108, 77, 51, 48, 48, 50, 51,
		52, 48, 49, 48, 48, 51, 48, 52, 53, 48, 0, 61, 196, 1, 127,
		87, 244, 226, 189, 3, 0, 11, 1, 47, 197, 131, 3, 99, 33, 0,
		0, 0, 6, 2, 0, 14, 4, 249, 18, 20, 85, 72, 215, 122, 173, 65,
		215, 128, 128, 112}

	MTConfirmation := []byte{1, 0, 28, 68, 0, 25, 116, 101, 115, 116,
		51, 48, 48, 50, 51, 52, 48, 49, 48, 48, 51, 48,
		52, 53, 49, 0, 0, 0, 0, 255, 254}

	bytes := [][]byte{MTConfirmation, MOGeoDelete, MOGlobalParam, MOAlertReport,
		MOSinglepositionreport, MOGeofenceconfirmation, MOHeaderWithLocationInformation2,
		MOHeaderWithDataPayload, MOHeaderWithLocationInformation, MOHistoryMessage}

	for {
		for _, message := range bytes {
			fmt.Println(0)
			conn, err := net.DialTCP("tcp", nil, tcpAddr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
				os.Exit(1)
			}

			conn.Write(message) // don't care about return value
			conn.Close()
			time.Sleep(2 * time.Second)

		}

	} //end of for loop
}
