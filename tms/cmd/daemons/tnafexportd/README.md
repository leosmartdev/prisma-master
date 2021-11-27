# Introduction
tnafexportd is a daemon listening on tsi messages with omnicom data and exports it in a format called NAF. That format is required by the Moroccan ministry of fisheries so we can integrate our data into their existing system.

## Command line usage

- address string

    -address <ip address> address to forward naf messages to (default ":9977")
- conn string

    conn type <tcp || sftp> (default "tcp")
- dblog string

    Write log & relevant objects to this EJDB database
- dir string

    sftp directory
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
- password string

    sftp password
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
- username string

    sftp username
- version

    Print version then exit

Examples of running tnafexportd
```
    tnafexportd -stdlog -log debug : run tnafexportd with error info warn debug output into terminal
    tnafexportd -stdlog -trace mytracer : run tnafexportd with trace output into terminal
```
