# Introduction
tmccrd is a real xml mcc data replay service

## Usage
Use: tmccrd -help for listing flags that can be used with the service

## Replay
tmccrd needs to point at a source directory where the raw data to be replayed resides, and destination directory which should be the ftp servcer directory that tmccd is watching.

###Example: 
```shell
tmccrd -stdlog -log debug -src-dir /srv/capture/ -dst-dir /srv/ftp/test/
```
This command will scan /srv/capture/ directory and move the data to /srv/ftp/test directory

## Capture
In order to capture real time data coming into tms, use the -ftp-capture and -capture-dir flags in tmccd.

###Example:
```shell
tmccd -protocol ftp -ftp-dir /srv/ftp -ftp-capture -capture-dir /srv/capture/ -stdlog -log debug
```