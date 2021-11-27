# Introduction
tspiderd is a deamon to get information of spider tracks.

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
- password string

    password (default "orolia")
- profile string

    Profile and listen on this port e.g. localhost:6060
- server_address string

- srcloc

    Find and write file:lineno to log? (default true)
- stdlog

    Write log to stderr?
- syslog

    Write log to syslog? (default true)
- system string

    System ID (default "Orolia")
- trace value

    comma-separated list of tracers to enable
- url string

    aff feed url (default "https://go.spidertracks.com/api/aff/feed")
- username string

    username (default "yonas.williams@orolia.com")
- version

    Print version then exit

Examples of running tspiderd
```
    tspiderd -stdlog -log debug : run twebd with error info warn debug output into terminal
    tspiderd -stdlog -trace mytracer : run twebd with trace output into terminal
    tspiderd -url http://localhost:8089/v1/aff/feed : run tspiderd where the source of spider tracks is another source
        (This address is simulator by default)
```


## Steps to run tspiderd

In order to run the spider tracks daemon, first create a vm using the vagrant up
command. Once in vagrant run a tmsd —start & in order to start tgwad,
tdatabased, and other daemons. Once tgwad and tdatabased are running you can now
run the spider tracks daemon by typing tspiderd into the command line, using any
flags available for running daemons. The standard flags will be the Orolia 
account, which will return the spider tracks heartbeat data which is a public 
test feed.  

Once the spider tracks daemon begins running it
will wait 30 seconds before requesting data from the spider tracks server and 
will then request data from the spider tracks server once every 30 seconds.
