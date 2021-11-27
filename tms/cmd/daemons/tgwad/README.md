# Introduction
tgwad is a broadcast daemon.

To use it in your code see TSIClient, also take a look at tmsg.GClient.

## Command line usage
- ca string

    File which contains the certificate authority
- cert string

    File which contains our certificate
- db string

    Path to tgwad database for delayed remote delivery (default "/tmp/tgwad.db")
- dblog string

    Write log & relevant objects to this EJDB database
- debug-port string

    Debugging port (default ":8083")
- entryid int

    Entry ID for process
- fast-timers

    use fast timers for testing
- filelog string

    Write log to this file
- host string

    address:port of tgwad (default "localhost:31228")
- httptest.serve string

    if non-empty, httptest.NewServer serves on this address and blocks
- insthreads int

    Number of goroutines to spawn for database inserts (default 10)
- log string

    Set the logging level (default "info")
- name string

    Name of the local site (default "local")
- num uint

    Number of the local site (default 1)
- port string

    TCP interface:port to listen on (default ":31228")
- priv string

    File which contains our private key
-profile string

    Profile and listen on this port e.g. localhost:6060
- secure-port string

    Port for accepting TLS connections
- server_address string

- site value

    Specify a site in this format: <name>,<num>[,gw][,<route>]*. E.g. hq,5,tcp:10.5.0.5:31228
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
    tgwad -stdlog -log debug : run tgwad with error info warn debug output into terminal
    tgwad -stdlog -trace mytracer : run tgwad with trace output into terminal
    tgwad -gss :10800 -vms=true : run tgwad with gss source of tsimulator by default and accept vms data
```


TGWAD can open a debug port and expose some interesting URLs. Here's a few:
http://localhost:8083/router/listeners
http://localhost:8083/router/history
http://localhost:8083/remote/<site name>/queue
http://localhost:8083/remote/<site name>/status

