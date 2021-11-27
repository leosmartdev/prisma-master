# Introduction
rsm is a cli dev tool that allows us to request specific message from an Omnicom beacon for testing. if expects back a Mobile Terminated Confirmation (MTC) message that reflects the results of the transmission between the server and the iridium gateway queue. 

The list of possible messages to request is below: 

 - 0x00: send an alert report: message 'Alert report". Response with “Alert Report(0x02)”
 - 0x01: send the last position recorded with the message "History position report(0x01)".
 - 0x02: make a new position acquisition and send it with the message "History position report(0x01)".
 - 0x03: send the global parameters setting.Response with Global parameters(0x03)
 - 0x04: send the parameters URL. Response with API url parameters(0x08)
 - 0x10: test the 3G. The Dome send a “Single position report(0x06) and a History report(0x01) in a 3g message.

The tool aslo accepts an MT flag parameter, this flag can enable the tool to free the MT queue in case its full. 

 - 0x0001: Delete all MT payloads in the IMEI’s MT queue
 - 0x0002: Send ring alert with no associated MT payload (normal ring alert rules apply)
 - 0x0000: No flag

## Command line usage

### Parameters
- gss string
    IPaddress:port the iridium gateway (default "12.47.179.12:10800")
- imei string
    !5 digit imei (default "300234010031990")
- msg_to_ask int
    0x00 || 0x01 || 0x02 || 0x03 || 0x04. read documentation for more info (default 3)
- mt_flag
    0x0001 || 0x0002

### Examples of how to run the cli: 
 
 
```bash
rsm -gss 12.47.179.12:10800 -imei 300234010031990 -msg_to_ask  0x02 -mt_flag 0x0002
```
