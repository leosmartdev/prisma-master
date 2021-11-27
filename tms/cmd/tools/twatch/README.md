# Introduction

twatch watches for tmsg messages

## Command line usage
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

Examples of running twatch
```
    twatch -stdlog -log debug : run twatch with error info warn debug output into terminal
    twatch -stdlog -trace mytracer : run twatch with trace output into terminal
```
