# Introduction
twebd is used to manage resources of tms system.

## Command Line usage

In common cases twebd can be run by command
```
twebd
```

It has configuration that can be optioned passing parameters into command line

- certificate string

    The path to the certificate file (default "/etc/trident/certificate.pem")

- dblog string

    Write log & relevant objects to this EJDB database
- debug string

    address:port to expose variables to, visit at /debug/vars (default ":9090")
- decode-threads int

    Maximum number of threads to use for bson decoding, per client request, per database stream (default 8)
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
- key string

    The path to the key file (default "/etc/trident/key.pem")
- listen string

    Address:port to listen upon (default ":8080")
- listen-config string

    Address: port for configuration (default ":8081")
- log string

    Set the logging level (default "info")
- mongo-db-name string

    Database name (default "trident")
- mongo-url string

    MongoDB URL (default "mongodb://:8201")
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


Examples of running twebd
```
    twebd -stdlog -log debug : run twebd with error info warn debug output into terminal
    twebd -stdlog -trace mytracer : run twebd with trace output into terminal
```

## Maintainable resources

It includes:
- device
- fleet
- geofence
- incident
- multicast
- notice
- registry
- rule
- sarmap
- site
- track
- vessel
- zone

## Rest API

It can be viewed at https://<twebd-addr>/api/v2/apidocs.json. By default is https://localhost:8080/api/v2/apidocs.json

Each endpoints has security checking to allow an access for an endpoint. To see description [public/README.md]

## WebSocket

The daemon sends information using websocket. The information includes:
- updates for objects(expiration, pings, updates)
- updates for current session(expiration, updates)
- updates for transmissions
- updates for the client viewport
