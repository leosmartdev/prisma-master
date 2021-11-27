# Introduction
tnoid is a daemon that receives data from AIS, Radars.

tnoid connects to a server and receives nmea messages.
It has generated messages and they can be seen in tms/nmea folder

## Command line usage
- address string

    -address = IPaddress:Port The address and port to request for nmea messages (default "127.0.0.1:2001")
- dblog string

    Write log & relevant objects to this EJDB database
- entryid int

    Entry ID for process
- filelog string

    Write log to this file
- host string

    address:port of tgwad (default "localhost:31228")
- httptest.serve string

    if non-empty, httptest.NewServer serves on this address and blocks
- log string

    Set the logging level (default "info")
- name string

    Name of this app. (default "tnoid0")
- profile string

    Profile and listen on this port e.g. localhost:6060
- radar_latitude float

    radar latitude
- radar_latitude_count int

    count of radar latitudes (default 1)
- radar_longitude float

    radar longitude
- radar_longitude_count int

    count of radar longitudes (default 1)
- ri int

    Interval to send tms info to front end (default 300)
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


Examples of running tnoid
```
    tnoid -stdlog -log debug : run tnafexportd with error info warn debug output into terminal
    tnoid -stdlog -trace mytracer : run tnafexportd with trace output into terminal
    tnoid -radar_longitude=104.3 -radar_latitude 0.3 : run tnoid and relative position of received radars will be computed of this arguments
```
