# MCC Integration

The tms-mcc package installs vsftpd as the FTP server used to receive
messages from MCCs. Each MCC shall be assigned a separate username and
password. The FTP password file is found at /etc/trident/vsftpd.passwd
and is managed using the Apache htpasswd utility. Received message are
ingest by [tmccd](../../linked-documents/tms/cmd/daemons/tmccd/README.md) which is
also installed by tms-mcc package.

## FTP

Examples on how to set up an FTP account useing htpasswd:

```
sudo htpasswd -d /etc/trident/vsftpd.passwd alice
```

Make a directory for each FTP user in /srv/ftp. The directory should be
owned by vsftpd with group mcm. Example:

```
sudo install -m 775 -o vsftpd -g mcm -d /srv/ftp/alice
```
## [tmccd](../../linked-documents/tms/cmd/daemons/tmccd/README.md)

In a development VM or a fresh install, a default vsfptd.password is provided
that includes user 'test' with password 'test.

When starting [tmccd](../../linked-documents/tms/cmd/daemons/tmccd/README.md), use the following flags:

```
tmccd --protocol ftp --ftp-dir /srv/ftp --sit185-template /etc/trident/sit185-template.json
```

### Data formats

Prisma RCCs are able to ingest data coming from MCCs XML or SIT185 over tcp or ftp.
If tmccd is presented with an unknown data format or fails, it will use a default
parser that display the raw MCC input to the end user.
SIT185 a plain test date format; concequentely, tmccd parser need a regex template
to be able to extract data from a SIT185 text file.

#### xml

MCC xml data format is a feature only available in Orolia's MCC. See example below:

```xml
<?xml version="1.0" ?>
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
</topMessage>
```


#### sit185

SIT185 is a plain text data format; concequentely, tmccd parser need a regex template
to be able to extract data from a SIT185 text file.

Example of a SIT185.txt would look like below:

```yaml
1. DISTRESS COSPAS-SARSAT POSITION CONFLICT ALERT
2. MSG NO 02698 AUMCC REF C1ADE28809C0185
3. DETECTED AT 06 APR 07 1440 UTC BY SARSAT S11
4. DETECTION FREQUENCY 406.0246 MHz
5. COUNTRY OF BEACON REGISTRATION 525/ INDONESIA
6. USER CLASS SERIAL USER-LOCATION - ELT
AIRCRAFT 24-BIT ADDRESS 8A2027
7. EMERGENCY CODE NIL
8. POSITIONS
CONFIRMED - NIL
DOPPLER A - 07 00.1 S 098 42.2 E PROB 50 PERCENT
DOPPLER B - 05 42.1 S 107 20.2 E PROB 50 PERCENT
DOA/ALTITUDE - NIL
ENCODED - NIL
UPDATE TIME WITHIN 4 HOURS OF DETECTION TIME
9. ENCODED POSITION PROVIDED BY INTERNAL DEVICE
10. NEXT PASS/EXPECTED DATA TIMES CONFIRMED - NIL
DOPPLER A - 06 APR 07 1805 UTC AULUTW ALBANY LUT AUSTRALIA
DOPPLER B - 06 APR 07 1956 UTC AULUTW ALBANY LUT AUSTRALIA
DOA - NIL
ENCODED - NIL
11. HEX ID C1ADE28809C0185 HOMING SIGNAL 121.5 MHZ
12. ACTIVATION TYPE NIL
13. BEACON NUMBER ON AIRCRAFT OR VESSEL 00
14. OTHER ENCODED INFORMATION CSTA CERTIFICATE NO 0097
BEACON MODEL - TECHTEST, UK 503-1
AIRCRAFT 24-BIT ADDRESS ASSIGNED TO INDONESIA
15. OPERATIONAL INFORMATION
RELIABILITY OF DOPPLER POSITION DATA - SUSPECT
LUT ID INLUT1 BANGALORE, INDIA
16. REMARKS
THIS POSITION 51 KILOMETRES FROM PREVIOUS ALERT
END OF MESSAGE
```
Example of Sit185-template.json:

```json
    {
	"msg_num": "2\\..+?MSG.*?NO.+?(?P<msg_num>\\d+).*",
	"date": "3\\..+?DETECTED AT.+?(?P<date>.+UTC)",
	"date_format": "%D %M %Y %H%MN %TZ",
	"hex_id": "11\\..+?HEX.+ID\\s*(?P<hex_id>\\S*)\\s",
	"doppler_a": "(?m)8\\.(?s).*DOPPLER\\s*A\\s*[-|–]\\s*(?P<doppler_a>NIL|UNKNOWN|(?P<doppler_a_lat_degree>\\d+).*?(?P<doppler_a_lat_min>\\d+\\.\\d+).*?(?P<doppler_a_lat_cardinal_point>[N|S]).*?(?P<doppler_a_lon_degree>\\d+).*?(?P<doppler_a_lon_min>\\d+\\.\\d+).*?(?P<doppler_a_lon_cardinal_point>[E|W]).*?(?P<doppler_a_prob>\\d+|$).*?).*9\\.",
	"doppler_b": "(?m)8\\.(?s).*DOPPLER\\s*B\\s*[–|-]\\s*(?P<doppler_b>NIL|UNKNOWN|(?P<doppler_b_lat_degree>\\d+).*?(?P<doppler_b_lat_min>\\d+\\.\\d+).*?(?P<doppler_b_lat_cardinal_point>[N|S]).*?(?P<doppler_b_lon_degree>\\d+).*?(?P<doppler_b_lon_min>\\d+\\.\\d+).*?(?P<doppler_b_lon_cardinal_point>[E|W]).*?(?P<doppler_b_prob>\\d+|$).*?).*9\\.",
	"encoded":	"8\\.(?s).*ENCODED\\s*?[-|–]\\s*(?P<encoded>NIL|UNKNOWN|(?P<encoded_lat_degree>.*?\\d+)(?P<encoded_lat_min>.*?\\d+\\.\\d+)(?P<encoded_lat_cardinal_point>.*?[N|S]).*?(?P<encoded_lon_degree>\\d+)(?P<encoded_lon_min>.*?\\d+\\.\\d+)(?P<encoded_lon_cardinal_point>.*?[E|W]).*?\n).*9\\.",
	"confirmed": "8\\.(?s).*CONFIRMED\\s*?[–|-]\\s*(?P<confirmed>NIL|UNKNOWN|(?P<confirmed_lat_degree>.*?\\d+)(?P<confirmed_lat_min>.*?\\d+\\.\\d+)(?P<confirmed_lat_cardinal_point>.*?[N|S])(?P<confirmed_lon_degree>.*?\\d+)(?P<confirmed_lon_min>.*?\\d+\\.\\d+)(?P<confirmed_lon_cardinal_point>.*?[E|W]).*?\n).*9\\.",
	"doa": 	"8\\.(?s).*DOA\\s*[–|-]\\s*(?P<doa>((?P<doa_lat_degree>.*?\\d+)(?P<doa_lat_min>.*?\\d+\\.\\d+)(?P<doa_lat_cardinal_point>.*?[N|S])(?P<doa_lon_degree>.*?\\d+)(?P<doa_lon_min>.*?\\d+\\.\\d+)(?P<doa_lon_cardinal_point>.*?[E|W]))(.*?ALTITUDE\\s*(?P<doa_elevation>.*?\\d+)|)).*9\\."
}
```
!!! info
    If you need to test a regex against a SIT185 input, online regex testers can be used. [regex101](https://regex101.com) is a good reference, also make sure you check the golang flavor to get a valid evaluation.

## Annex: golang regex syntax

The syntax of the regular expressions accepted is the same general syntax used by Perl, Python, and other languages. More precisely, it is the syntax accepted by RE2 except for \C. For an overview of the syntax, see below:

```C
/*Syntax

The regular expression syntax understood by this package when parsing with
the Perl flag is as follows. Parts of the syntax can be disabled by passing
alternate flags to Parse.

Single characters:

    .              any character, possibly including newline (flag s=true)
    [xyz]          character class
    [^xyz]         negated character class
    \d             Perl character class
    \D             negated Perl character class
    [[:alpha:]]    ASCII character class
    [[:^alpha:]]   negated ASCII character class
    \pN            Unicode character class (one-letter name)
    \p{Greek}      Unicode character class
    \PN            negated Unicode character class (one-letter name)
    \P{Greek}      negated Unicode character class

Composites:

    xy             x followed by y
    x|y            x or y (prefer x)

Repetitions:

    x*             zero or more x, prefer more
    x+             one or more x, prefer more
    x?             zero or one x, prefer one
    x{n,m}         n or n+1 or ... or m x, prefer more
    x{n,}          n or more x, prefer more
    x{n}           exactly n x
    x*?            zero or more x, prefer fewer
    x+?            one or more x, prefer fewer
    x??            zero or one x, prefer zero
    x{n,m}?        n or n+1 or ... or m x, prefer fewer
    x{n,}?         n or more x, prefer fewer
    x{n}?          exactly n x

Implementation restriction: The counting forms x{n,m}, x{n,}, and x{n}
reject forms that create a minimum or maximum repetition count above 1000.
Unlimited repetitions are not subject to this restriction.

Grouping:

    (re)           numbered capturing group (submatch)
    (?P<name>re)   named & numbered capturing group (submatch)
    (?:re)         non-capturing group
    (?flags)       set flags within current group; non-capturing
    (?flags:re)    set flags during re; non-capturing

    Flag syntax is xyz (set) or -xyz (clear) or xy-z (set xy, clear z). The flags are:

    i              case-insensitive (default false)
    m              multi-line mode: ^ and $ match begin/end line in addition to begin/end text (default false)
    s              let . match \n (default false)
    U              ungreedy: swap meaning of x* and x*?, x+ and x+?, etc (default false)

Empty strings:

    ^              at beginning of text or line (flag m=true)
    $              at end of text (like \z not Perl's \Z) or line (flag m=true)
    \A             at beginning of text
    \b             at ASCII word boundary (\w on one side and \W, \A, or \z on the other)
    \B             not at ASCII word boundary
    \z             at end of text

Escape sequences:

    \a             bell (== \007)
    \f             form feed (== \014)
    \t             horizontal tab (== \011)
    \n             newline (== \012)
    \r             carriage return (== \015)
    \v             vertical tab character (== \013)
    \*             literal *, for any punctuation character *
    \123           octal character code (up to three digits)
    \x7F           hex character code (exactly two digits)
    \x{10FFFF}     hex character code
    \Q...\E        literal text ... even if ... has punctuation

Character class elements:

    x              single character
    A-Z            character range (inclusive)
    \d             Perl character class
    [:foo:]        ASCII character class foo
    \p{Foo}        Unicode character class Foo
    \pF            Unicode character class F (one-letter name)

Named character classes as character class elements:

    [\d]           digits (== \d)
    [^\d]          not digits (== \D)
    [\D]           not digits (== \D)
    [^\D]          not not digits (== \d)
    [[:name:]]     named ASCII class inside character class (== [:name:])
    [^[:name:]]    named ASCII class inside negated character class (== [:^name:])
    [\p{Name}]     named Unicode property inside character class (== \p{Name})
    [^\p{Name}]    named Unicode property inside negated character class (== \P{Name})

Perl character classes (all ASCII-only):

    \d             digits (== [0-9])
    \D             not digits (== [^0-9])
    \s             whitespace (== [\t\n\f\r ])
    \S             not whitespace (== [^\t\n\f\r ])
    \w             word characters (== [0-9A-Za-z_])
    \W             not word characters (== [^0-9A-Za-z_])

ASCII character classes:

    [[:alnum:]]    alphanumeric (== [0-9A-Za-z])
    [[:alpha:]]    alphabetic (== [A-Za-z])
    [[:ascii:]]    ASCII (== [\x00-\x7F])
    [[:blank:]]    blank (== [\t ])
    [[:cntrl:]]    control (== [\x00-\x1F\x7F])
    [[:digit:]]    digits (== [0-9])
    [[:graph:]]    graphical (== [!-~] == [A-Za-z0-9!"#$%&'()*+,\-./:;<=>?@[\\\]^_`{|}~])
    [[:lower:]]    lower case (== [a-z])
    [[:print:]]    printable (== [ -~] == [ [:graph:]])
    [[:punct:]]    punctuation (== [!-/:-@[-`{-~])
    [[:space:]]    whitespace (== [\t\n\v\f\r ])
    [[:upper:]]    upper case (== [A-Z])
    [[:word:]]     word characters (== [0-9A-Za-z_])
    [[:xdigit:]]   hex digit (== [0-9A-Fa-f])

func IsWordChar(r rune) bool
type EmptyOp uint8
    const EmptyBeginLine EmptyOp = 1 << iota ...
    func EmptyOpContext(r1, r2 rune) EmptyOp
type Error struct{ ... }
type ErrorCode string
    const ErrInternalError ErrorCode = "regexp/syntax: internal error" ...
type Flags uint16
    const FoldCase Flags = 1 << iota ...
type Inst struct{ ... }
type InstOp uint8
    const InstAlt InstOp = iota ...
type Op uint8
    const OpNoMatch Op = 1 + iota ...
type Prog struct{ ... }
    func Compile(re *Regexp) (*Prog, error)
type Regexp struct{ ... }
    func Parse(s string, flags Flags) (*Regexp, error)
*/
```

The regexp implementation provided by the golang "regexp" standard library package is guaranteed to run in time linear in the size of the input. (This is a property not guaranteed by most open source implementations of regular expressions.) For more information about this property, read any book about automata theory or check out [Regular Expression Matching Can Be Simple And Fast](https://swtch.com/~rsc/regexp/regexp1.html) paper.

## References:

- [Annex E - COSPAS-SARSAT STANDARD FOR THE TRANSMISSION OF MESSAGES VIA FTP](A.002_FTP.pdf)
- [Golang regular expression syntax accepted by RE2](https://github.com/google/re2/wiki/Syntax)
