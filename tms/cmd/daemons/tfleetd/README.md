# Introduction
tfleetd handles beacons, MO and MT messages.

It can work with different sources like ingenu, iridium.

## Command line usage
- client-message-id string

    unique client message id should be 4 chars (default "C201")
- dblog string

    Write log & relevant objects to this EJDB database
- entryid int

    Entry ID for process
- filelog string

    Write log to this file
- gss string

    IPaddress:port the iridium gateway (default "12.47.179.12:10800")
- host string

    address:port of tgwad (default "localhost:31228")
- httptest.serve string

    if non-empty, httptest.NewServer serves on this address and blocks
- iridium-client string

    IPaddress:port where Mobile Originated messages are received (default "127.0.0.1:7777")
- log string

    Set the logging level (default "info")
- net string

    iridium || ingenu (default "iridium")
- omnicom true || false (default "true")

    Flag should be set to true when tfleetd is going to listen on omnicom messages
- password string

    <password> ingenu account password (default "0r011a_McMurd0%")
- profile string

    Profile and listen on this port e.g. localhost:6060
- server_address string

- solar true || false

    Set to true when tfleetd is expecting Omnicom Solar data
- srcloc

    Find and write file:lineno to log? (default true)
- stdlog

    Write log to stderr?
- syslog

    Write log to syslog? (default true)
- trace value

    comma-separated list of tracers to enable
- url string

    url ingenu url to fetch data (default "https://glds.ingenu.com/data/v1/receive/")
- url-ingenu string

    url ingenu url (default "https://glds.ingenu.com")
- username string

    username ingenu account username (default "orolia@orolia.com")
- version

    Print version then exit
- vms true || false

    tfleetd will expects Omnicom VMS data

Examples of running tfleetd
```
    tfleetd -stdlog -log debug : run tfleetd with error info warn debug output into terminal
    tfleetd -stdlog -trace mytracer : run tfleetd with trace output into terminal
    tfleetd -gss :10800 -vms=true : run tfleetd with gss source of tsimulator by default and accept vms data
```

### Useful documentation:
[https://oroliagroup.sharepoint.com/:w:/r/sites/mcmurdogroup/mcmurdord/pandore/_layouts/15/Doc.aspx?sourcedoc=%7B66C9BDAF-5E11-4BD2-AEBF-144A5E208AC4%7D&file=DRD15050B13%20-%20Iridium%20and%203G%20messages%20specification.docx&action=default&mobileredirect=true]
