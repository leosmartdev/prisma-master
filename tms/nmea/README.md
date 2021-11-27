# Introduction 

Libgonmea is a go library that provides functionalities for decoding NMEA streams from different maritime sensors. It is also able to encode NMEA stream from specific sentences type initialization. Currently the library is able to process [NMEA 0183][1] and [M1371][2] sentences.

# Parsing

To decode a NMEA 0183 or M1371 message libgonmea has a parser method called Parse:

func Parse(messages []string, index int) (SentenceI, error, int)

### Arguments:

- Messages: an array of strings containing NMEA 0183 and M1371 messages to be parsed.
- Index: integer pointing to the specific message to decode within the messages passed.

### Return Values

- SentenceI: Sentence interface containing the specific NMEA 0183 or M1371 sentence type for the message message parsed.
- Error: return an error in case the message is not supported or invalid.
- Int: should be ignored unless the parser is ingesting Multi messages Sentences then it returns the index of the last string parsed within Messages argument.

### Example

```go
package main

import (
    "fmt"
    "./gonmea"
)

func main() {
 
    var str = []string{"!AIVDO,1,1,,,102Cu;P01wVquDDRv417gw62000,0*22",
                       "!AIVDM,2,1,0,A,58wt8Ui`g??r21`7S=:22058<v05Htp000000015>8OA;0sk,0*7B",
                        "!AIVDM,2,2,0,A,eQ8823mDm3kP00000000000,2*5D",
                        "$RATTM,02,3.55,297.3,T,,,T,0.13,,N,,Q,,,M*02",
                        "$GPZDA,153415,10,12,2015,0,0*4B"}

     fmt.Println(len(str))
    for i:=0; i<len(str);i++ {
     m, err, j := nmea.Parse(str,i)
    if err != nil {
        fmt.Println(err)
    }
    if err == nil {fmt.Printf("%+v\n", m)}

    if j != 0 {
        i=j
    }
 
    fmt.Println()

}

}//end of main
```
### Output

> &{VDM_O:{Sentence:{SOS:! Talker:AI Format:VDO Fields:[1 1   102Cu;?P01wVquDDRv417gw62000 0] Checksum:22 Raw:!AIVDO,1,1,,,102Cu;?P01wVquDDRv417gw62000,0*22} Sentence_count:1 Sentence_index:1 Seq_msg_id:0 Channel: Encap_data:102Cu;?P01wVquDDRv417gw62000 Fill_bits:0} Message_id:1 Repeat_indicator:0 Mmsi:2424108 Navigational_status:15 Rate_of_turn:128 Speed_over_ground:1 Position_accuracy:true Longitude:265146282 Latitude:21544464 Course_over_ground:286 True_heading:511 Time_stamp:35 Special_manoeuvre:0 Spare:0 Raim_flag:true Comm_state:0}

> &{VDM_O:{Sentence:{SOS:! Talker:AI Format:VDM Fields:[2 1 0 A 58wt8Ui`g??r21`7S=:22058<v05Htp000000015>8OA;0sk 0] Checksum:7B Raw:!AIVDM,2,1,0,A,58wt8Ui`g??r21`7S=:22058<v05Htp000000015>8OA;0sk,0*7B&&!AIVDM,2,2,0,A,eQ8823mDm3kP00000000000,2*5D} Sentence_count:2 Sentence_index:1 Seq_msg_id:0 Channel:A Encap_data:58wt8Ui`g??r21`7S=:22058<v05Htp000000015>8OA;0skeQ8823mDm3kP00000000000 Fill_bits:0} Message_id:5 Repeat_indicator:0 Mmsi:603916439 Ais_version:0 Imo_number:439303422 Call_sign:  ZA83R Name:   ARCO AVON Ship_and_cargo_type:69 Dim_bow:113 Dim_stern:31 Dim_port:17 Dim_starboard:11 Position_device:0 Eta_month:3 Eta_day:23 Eta_hour:19 Eta_minute:45 Draught:132 Destination:  HOUSTON Data_terminal_avail:false Spare:0}

> &{Sentence:{SOS:$ Talker:RA Format:TTM Fields:[02 3.55 297.3 T   T 0.13  N  Q   M] Checksum:02 Raw:$RATTM,02,3.55,297.3,T,,,T,0.13,,N,,Q,,,M*02} Number:2 Distance:3.55 Bearing:297.3 Bearing_relative:T Speed:0 Course:0 Course_relative:T Cpa_distance:0.13 Cpa_time:0 Speed_distance_units:N Name: Status:Q Reference: Utc_time:0 Acquisition_type:M}

> &{Sentence:{SOS:$ Talker:GP Format:ZDA Fields:[153415 10 12 2015 0 0] Checksum:4B Raw:$GPZDA,153415,10,12,2015,0,0*4B} Utc_time:153415 Day:10 Month:12 Year:2015 Local_zone_hours:0 Local_zone_minutes:0}

# Encoding

To encode an NMEA 01883 or M1371 message libgonmea has an encode method for each sentence type.

func (s *<Sentence structure>) Encode() (string, error)

### Return Values:
-	String: raw NMEA or M1371 string
-	Error: returns error if the method fails to encode a given sentence either because the sentence is not supported or its initialization is not correct.

### Example

```go
package main

import (
	"fmt"
	"./gonmea"
)


func main(){

s := nmea.Sentence{}
s.SOS = "$"
s.Talker = "GP"
s.Format = "GLL"
	
GLL := nmea.GLL{s,3554.4456,"N",00528.9195,"W",135138,"A","V"}

str, err := GLL.Encode()
if err != nil {
fmt.Println(err)
}
fmt.Println(str)


sen := nmea.Sentence{}
sen.SOS = "!"
sen.Talker = "AI"
sen.Format = "VDM"

VDM := nmea.VDM_O{sen,1,1,0,"A","",0}

M1371 := nmea.M1371_10{VDM,10,0,265547250,0,2500912,0}

str, err = M1371.Encode()
if err != nil {
fmt.Println(err)
}
fmt.Println(str)



sen = nmea.Sentence{}
sen.SOS = "!"
sen.Talker = "AI"
sen.Format = "VDM"

VDM = nmea.VDM_O{sen,1,1,0,"B","",4}

M13711 := nmea.M1371_21{VDM,21,0,992429100,14,"TANGER MED 2",false,265115292,21528796,1,1,1,1,1,46,false,226,true,false,false,0,"MEHDI",1}

str, err = M13711.Encode()
if err != nil {
fmt.Println(err)
}
fmt.Println(str)

}
```
### Output

> $GPGLL,3554.4456,N,528.9195,W,135138,A,V*40

> !AIVDM,1,1,0,A,:3u?etP0V:C0,0*3B

> !AIVDM,1,1,0,B,E>jM4;7:0W3Ra@6RR@I00000000?kEJL:A0KP10888o>:3AB12@000000000@,4*41

# Sentence Generation

Libgonmea is a static library that was generated by libgonmea-gen.go using Sentence.json file.
Sentences.json was extracted initially from libcnmea, and contains definitions for all NMEA 0813 and M1371 sentences that libcnmea was supporting. Libgonmea-gen generates parsers and encoders for every specific sentence defined in Sentences.json which means that adding a NMEA sentence that is not supported by libgonmea is as simple as adding its definition to Sentences.json then running libgonmea-gen.go in the right directory where the code need to be generated.

### Example 

let's suppose that M1371 sentence of type 27 is not supported by our library. 

- We have then to add M1371_27 sentence definition as below to Sentences.Json.

> { "name": "M1371_27", "formatter":"","fieldCount":12,"fields": [{"name":"Message_id" ,"type":"uint8", "encbsize":6}, {"name":"Repeat_indicator" ,"type":"uint32", "encbsize":2}, {"name":"Mmsi" ,"type":"uint32", "encbsize":30}, {"name":"Position_accuracy" ,"type":"bool", "encbsize":1}, {"name":"Raim_flag" ,"type":"bool", "encbsize":1}, {"name":"Navigational_status" ,"type":"uint32", "encbsize":4}, {"name":"Longitude" ,"type":"int32", "encbsize":18}, {"name":"Latitude" ,"type":"int32", "encbsize":17}, {"name":"Speed_over_ground" ,"type":"uint32", "encbsize":6}, {"name":"Course_over_ground" ,"type":"uint32", "encbsize":9}, {"name":"Position_latency" ,"type":"bool", "encbsize":1}, {"name":"Spare" ,"type":"unknown", "encbsize":1}], "recordSize":37 }

- We have to build libgonmea-gen.go and copy it to gonmea directory 

> go build libgonmea-gen.go | mv libgonmea-gen gonmea/ 

- Run libgonmea-gen from gomea/ repository 

> ./libgonmea-gen   –file=/Path/to/Sentence.json

 


[1]:http://catb.org/gpsd/NMEA.html
[2]:http://catb.org/gpsd/AIVDM.html

