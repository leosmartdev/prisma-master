# Introduction
tdatabased is used to record track information into mongodb.

tdatabased listens tgwad messages to get tracks and record them into mongoDB.

## Command line usage

In common cases tdatabased can be run by command
```
tdatabased
```

  -dblog string
    	Write log & relevant objects to this EJDB database
  -decode-threads int
    	Maximum number of threads to use for bson decoding, per client request, per database stream (default 8)
  -entryid int
    	Entry ID for process
  -fast-timers
    	use fast timers for testing
  -filelog string
    	Write log to this file
  -host string
    	address:port of tgwad (default "localhost:31228")
  -httptest.serve string
    	if non-empty, httptest.NewServer serves on this address and blocks
  -insthreads int
    	Number of goroutines to spawn for database inserts (default 10)
  -log string
    	Set the logging level (default "info")
  -mongo-db-name string
    	Database name (default "trident")
  -mongo-url string
    	MongoDB URL (default "mongodb://:27017")
  -profile string
    	Profile and listen on this port e.g. localhost:6060
  -schemas string
    	provide files to up schemas. Separate them by ','. Example: /path/to/file.ext,/path1/to1/file1.ext
  -server_address string
    	
  -srcloc
    	Find and write file:lineno to log? (default true)
  -sslCAFile string
    	Certificate Authority file for SSL (default "/etc/trident/mongoCA.crt")
  -sslCertFile string
    	X509 public key (default "/etc/trident/mongo.crt")
  -sslKeyfile string
    	X509 private key (default "/etc/trident/mongo.key")
  -stdlog
    	Write log to stderr?
  -syslog
    	Write log to syslog? (default true)
  -trace value
    	comma-separated list of tracers to enable
  -version
    	Print version then exit