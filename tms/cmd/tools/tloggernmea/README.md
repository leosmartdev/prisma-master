# tloggernmea listens nmea messages from tgwad and log them

## Command line usage
- dblog string

    Write log & relevant objects to this EJDB database
- entryid int

    Entry ID for process
- filelog string

    Write log to this file
- fout string

    a file where will write data. stdout for stdout (default "stdout")
- host string

    address:port of tgwad (default "localhost:31228")
- httptest.serve string

    if non-empty, httptest.NewServer serves on this address and blocks
- log string

    Set the logging level (default "info")
- profile string

    Profile and listen on this port e.g. localhost:6060
- server_address string

- srcloc

    Find and write file:lineno to log? (default true)
- stdlog

    Write log to stderr?
- syslog

    Write log to syslog? (default true)
- trace value

    comma-separated list of tracers to enable
- version

    Print version then exit

Examples of running tloggernmea
```
    tloggernmea -stdlog -log debug : run tloggernmea with error info warn debug output into terminal
    tloggernmea -stdlog -trace mytracer : run tloggernmea with trace output into terminal
    tloggernmea -host 193.0.0.1:1089 -fout=mylogfile : will be connected to 193.0.0.1:1089 and write message to ./mylogfile
```
