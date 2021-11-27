# Introduction
tmccd handles messages of sarsat beacons.

## Command line usage

- address string

    -address = IPaddress:Port The address and port to request for mcc messages (default ":9999")
- capture-dir string

    captured file are copied to this directory (default "/srv/capture")
- dblog string

    Write log & relevant objects to this EJDB database
- entryid int

    Entry ID for process
- filelog string

    Write log to this file
- ftp-capture

    capture raw ftp data coming from mccs
- ftp-dir string

    watch this ftp directory for new messages (default "/srv/ftp")
- host string

    address:port of tgwad (default "localhost:31228")
- httptest.serve string

    if non-empty, httptest.NewServer serves on this address and blocks
- log string

    Set the logging level (default "info")
- profile string

    Profile and listen on this port e.g. localhost:6060
- protocol string

    -protocol = Transport protocol (eg: TCP,FTP,AMHS...) (default "ftp")
- server_address string

- sit185-template string

    file path to the sit template (default "/etc/trident/sit185-template.json")
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


Examples of running tgwad
```
    tmccd -stdlog -log debug : run tmccd with error info warn debug output into terminal
    tmccd -stdlog -trace mytracer : run tmccd with trace output into terminal
    tmccd -gss :10800 -vms=true : run tmccd with gss source of tsimulator by default and accept vms data
```
