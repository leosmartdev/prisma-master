# Introduction
tping sends tmsg message to check live of a daemon.

## Command line usage
- app string

    Application to ping, default: tgwad (default "tgwad")
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
- profile string

    Profile and listen on this port e.g. localhost:6060
- server_address string

- site string

    Site to ping, default: local
- srcloc

    Find and write file:lineno to log? (default true)
- stdlog

    Write log to stderr?
- syslog

    Write log to syslog? (default true)
- timeout string

    Amount of time to wait for response (default "5s")
- trace value

    comma-separated list of tracers to enable
- version

    Print version then exit

Examples of running tping
```
    tping -stdlog -log debug : run tping with error info warn debug output into terminal
    tping -stdlog -trace mytracer : run tping with trace output into terminal
    tping -app tfleetd : ping tfleetd daemon
```
