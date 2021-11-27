package lib

const testXML = `
/00003 00000/5257/18 029 0923 /147/0001/01 
<?xml version="1.0" ?> 
<topMessage>     
	<header dest="0001" orig="5257" number="00002" date="2018-01-29T09:13:10Z" />     
	<message>         
		<resolvedAlertMessage>             
		<header>                 
			<siteId>1878</siteId>                 
			<beacon>9A22BE29630F010</beacon>             
		</header>             
		<composite>                 
			<location latitude="34.8776" longitude="33.4054" />                 
			<duration>PT62M</duration>             
		</composite>         
		</resolvedAlertMessage>     
	</message> 
</topMessage>
/LASSIT 
/ENDMSG`

const ResolvedAlertMessageFromIndonesia = `/00003 00000/5257/18 029 0923
 
/147/0001/01
 
<?xml version="1.0" ?>
 
<topMessage>
 
    <header dest="0001" orig="5257" number="00002" date="2018-01-29T09:13:10Z" />
 
    <message>
 
        <resolvedAlertMessage>
 
            <header>
 
                <siteId>1878</siteId>
 
                <beacon>9A22BE29630F010</beacon>
 
            </header>
 
            <composite>
 
                <location latitude="34.8776" longitude="33.4054" />
 
                <duration>PT62M</duration>
 
            </composite>
 
        </resolvedAlertMessage>
 
    </message>
 
</topMessage>
 
/LASSIT
 
/ENDMSG`

const FreeformMessageSample = `<?xml version="1.0" ?>
<topMessage>
    <header dest="2770" orig="2570" number="14063" date="2014-10-14T13:46:17Z" />
    <message>
        <freeformMessage>
            <subject>Test Message</subject>
            <body>This is the body of the test message</body>
        </freeformMessage>
    </message>
</topMessage>`

// with the same beaconID generate positions for 1st message that gets send is unlocated, 2nd is located and the final is confirmed.
//After generate a new beaconID and new positions and do again step 1

// Confirmed Alert message has the fixed final position of a SARSAT beacon encapsulated in <composite>
// <elemental> are positions that are not sure, but were used to calculate the composite.
// we want to make sure that the <elemental> are not very far from <composite> for coherence.
const ConfirmedAlertMessageSample = `<?xml version="1.0" ?>
<topMessage>
    <header dest="2770" orig="2570" number="14063" date="2014-10-14T13:46:17Z" />
    <message>
        <resolvedAlertMessage>
            <header>
                <siteId>27534</siteId>
                <beacon>2DD42ED73F81FE0</beacon>
			</header>
            <composite> 
                <location latitude="-3.6581" longitude="110.2116" />
                <duration>PT1155M</duration>
            </composite>
            <elemental>
                <side>B</side>
                <location latitude="-3.6578" longitude="110.2108" />
                <satellite>13</satellite>
                <orbitNumber>10753</orbitNumber>
                <tca>2014-10-14T13:40:19.954Z</tca>
            </elemental>
            <elemental>
                <side>A</side>
                <location latitude="-3.6584" longitude="110.2119" />
                <satellite>13</satellite>
                <orbitNumber>10753</orbitNumber>
                <tca>2014-10-14T02:56:55.316Z</tca>
            </elemental>
            <elemental>
                <side>A</side>
                <location latitude="-3.6583" longitude="110.2113" />
                <satellite>13</satellite>
                <orbitNumber>10753</orbitNumber>
                <tca>2014-10-14T01:17:21.963Z</tca>
            </elemental>
        </resolvedAlertMessage>
    </message>
</topMessage>`

// This message has not position, and it will generate an Alert in the system we just need to update
// the date and message number and send it to tmccd.
const UnlocatedAlertMessageSample = `<?xml version="1.0" ?>
<topMessage>
    <header dest="2770" orig="2570" number="13866" date="2014-10-09T21:12:58Z" />
    <message>
        <unlocatedAlertMessage>
            <header>
                <siteId>27258</siteId>
                <beacon>ADCC404C8400315</beacon>
            </header>
            <tca>2014-10-09T13:01:07.029Z</tca>
            <satellite>7</satellite>
            <orbitNumber>20203</orbitNumber>
        </unlocatedAlertMessage>
    </message>
</topMessage>
	`

// Located alert message that can have multiple positions that are not sure. We will show all of the,
// in the map as a multipoint target. change dates, and message number + positions.
const LocatedAlertMessageSample = `<?xml version="1.0" ?>
<topMessage>
    <header dest="2770" orig="2570" number="13859" date="2014-10-09T19:46:28Z" />
    <message>
        <incidentAlertMessage>
            <header>
                <siteId>27206</siteId>
                <beacon>ADC667150EFE241</beacon>
            </header>
            <elemental>
                <satellite>10</satellite>
                <orbitNumber>48374</orbitNumber>
                <tca>2014-10-09T13:10:26.848Z</tca>
                <dopplerA probability="59">
                    <location latitude="-2.8900" longitude="110.2116" />
                </dopplerA>
                <dopplerB>
                    <location latitude="-2.8906" longitude="110.2111" />
                </dopplerB>
			</elemental>
        </incidentAlertMessage>
    </message>
</topMessage>`

const LocatedAlertMessageMeoElementSample = `<?xml version="1.0" ?>
<topMessage>
    <header dest="43AR" orig="4310" number="00023" date="2015-03-03T19:54:24Z" />
    <message>
        <incidentAlertMessage>
            <header>
                <siteId>128</siteId>
                <beacon>A608130D34D34D1</beacon>
            </header>
            <meoElemental>
                <satellite>312</satellite>
                <tca>2015-03-02T16:20:39.895Z</tca>
                <doa>
                    <location latitude="1.8900" longitude="107.2116" />
                </doa>
            </meoElemental>
        </incidentAlertMessage>
    </message>
</topMessage>`

const UnlocatedAlertMessageWithHeadersAndFooters = `/00001 00000/5030/17 299 1532
/122/503A/012/01
<?xml version="1.0" ?>
<topMessage>
    <header dest="503A" orig="5030" number="00001" date="2017-10-26T15:32:45Z" />
    <message>
        <unlocatedAlertMessage>
            <header>
                <siteId>20</siteId>
                <beacon>3EF43F8ABF81FE0</beacon>
            </header>
            <tca>2016-07-24T06:44:24.640Z</tca>
            <satellite>12</satellite>
            <orbitNumber>0</orbitNumber>
        </unlocatedAlertMessage>
    </message>
</topMessage>
/LASSIT
/ENDMSG`

//sit185 sample message
const Sit185DCSA = `
/00042 00000/5030/17 304 1918
1.  DISTRESS COSPAS-SARSAT ALERT
2.  MSG NO 00042  AUMCC REF NO 00890
3.  DETECTED AT 28 JUL 2016 2338 UTC BY GOES 15
4.  DETECTION FREQUENCY  406.0399 MHZ
5.  COUNTRY OF BEACON REGISTRATION  512/NEWZEALAND
6.  USER CLASS  STANDARD LOCATION - PLB
    SERIAL NO: 4350
    IDENTIFICATION  239/4350
7.  EMERGENCY CODE  NIL
8.  POSITIONS
        CONFIRMED - 36 36.20S  174 43.00E
        DOPPLER A - NIL
        DOPPLER B - NIL
        DOA       - 05 10.2 S 178 01.2 E EXPECTED ACCURACY 03 NMS
                    ALTITUDE 45 METRES
        ENCODED   - NIL
        TIME OF UPDATE UNKNOWN
9.  ENCODED POSITION PROVIDED BY INTERNAL DEVICE
10. NEXT PASS TIMES
        CONFIRMED - UNKNOWN
        DOPPLER A - UNKNOWN
        DOPPLER B - UNKNOWN
        DOA       - UNKNOWN
        ENCODED   - UNKNOWN
11. HEX ID  400E77A1FCFFBFF    HOMING SIGNAL  121.5
12. ACTIVATION TYPE  UNKNOWN
13. BEACON NUMBER ON AIRCRAFT OR VESSEL NO  NIL
14. OTHER ENCODED INFORMATION  
    A.  ENCODED POSITION UNCERTAINTY PLUS-MINUS 2 SECONDS OF
        LATITUDE AND LONGITUDE
15. OPERATIONAL INFORMATION  NIL
16. REMARKS  NIL
END OF MESSAGE
`

// UknownMccMessageFormat ...
const UknownMccMessageFormat = `
My name is El Mehdi Rahoui, and I need help me please
I am drawning somewhere around here: @47.7901492,-3.5284062
Thank you
`

// ChileXML failed to parse xml
const ChileXML = `/00063 00000/7250/19 008 1248

/145/725C/01

<?xml version="1.0" ?>

<topMessage>

    <header dest="725C" orig="7250" number="00063" date="2019-01-08T12:48:13Z" />

    <message>

        <incidentAlertMessage>

            <header>

                <siteId>6784</siteId>

                <beacon>9C7F4B013595551</beacon>

            </header>

            <meoElemental>

                <satellite>312</satellite>

                <tca>2019-01-08T12:47:32.946Z</tca>

                <doa>

                    <location latitude="-33.2523" longitude="-70.6555" />

                </doa>

            </meoElemental>

        </incidentAlertMessage>

    </message>

</topMessage>

/LASSIT

/ENDMSG`
