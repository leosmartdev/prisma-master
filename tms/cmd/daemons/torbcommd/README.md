# Introduction
A daemon to request AIS data from Machine to Machine Data Streaming pushed by
ORBCOMM server. It is based on a TCP/IP SSL connection.

## Authentication

To request connection with ORBCOMM server, we need to connect the port assigned
by ORBCOMM with specified username and password.

The default ip address and port in torbcommd assigned by ORBCOMM is 
63.146.183.130:9022.
It can be configured by adding argument when running torbcommd:
```
torbcommd -address=ip_address:port
```
The default username in torbcommd assigned by ORBCOMM is McMurdo_Lss.
It can be configured by adding argument when running torbcommd:
```
torbcommd -username=USERNAME
```
The default password in torbcommd assigned by ORBCOMM is JnHpv5BT5heY.
It can be configured by adding argument when running torbcommd:
```
torbcommd -password=PASSWORD
```