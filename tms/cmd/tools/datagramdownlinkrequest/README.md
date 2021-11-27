### datagramdownlinkrequest sends RSM with DatagramDownlinkRequest using for testing

## Command line usage
- config string

    Location of configuration file (default "/etc/trident/tmsd.conf")
- control string

    Location of TMSD control socket (default "/tmp/tmsd.sock")
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
- kill

    Not try making subprocesses exit gracefully by sending sigint the second time when tmsd -stop, kill them instead
- kt int

    Wait time to kill the subprocesses when tmsd -stop (default 60)
- lli int

    Long time interval to try launching not started or crashed sub-processes in the second phase (default 10)
- llt int

    Times to try launching not started or crashed sub-processes in the second phase (default 30)
- log string

    Set the logging level (default "info")
- pid string

    Location of file for tmsd to place subprocess PIDs (default "/tmp/tmsd.pid")
- profile string

    Profile and listen on this port e.g. localhost:6060
- ri int

    Interval to send tms info to front end (default 300)
- se int

    Wait time to try the second time to make subprocesses exit gracefully when tmsd -stop (default 5)
- server_address string

- si int

    Time interval to print status of subprocesses when tmsd -stop (default 5)
- sli int

    Short time interval to try launching not started or crashed sub-processes in the first phase (default 1)
- slt int

    Times to try launching not started or crashed sub-processes in the first phase (default 30)
- srcloc

    Find and write file:lineno to log? (default true)
- stdlog

    Write log to stderr?
- syslog

    Write log to syslog? (default true)
- trace value

    comma-separated list of tracers to enable
- users string

    A list of default users to switch to launch tmsd instead if a root try to launch it, will be tried in order (default "prisma")
- version

    Print version then exit

Examples of running datagramdownlinkrequest
```
    datagramdownlinkrequest -stdlog -log debug : run datagramdownlinkrequest with error info warn debug output into terminal
    datagramdownlinkrequest -stdlog -trace mytracer : run datagramdownlinkrequest with trace output into terminal
```

