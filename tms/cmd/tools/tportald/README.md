# Intoduction
tportald provides Rest api to check multicast

## Command line usage
- addr string

    http listen and serve address (default ":9099")
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
- timeout int

    request timeout (default 5)
- trace value

    comma-separated list of tracers to enable
- version

    Print version then exit

Examples of running tportald
```
    tportald -stdlog -log debug : run tportald with error info warn debug output into terminal
    tportald -stdlog -trace mytracer : run tportald with trace output into terminal
    tportald -addr :8888 : listen http requests on 8888 port
```
