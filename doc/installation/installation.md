# PRISMA Installation Guide

This guide provides detailed documentation on installing the Server portions of the PRISMA RCC and PRISMA Fleet systems. This will cover online and offline installations of fresh systems.

For instructions on installing the Windows PRISMA client, see [PRISMA Client Installation Guide](./client-installation.md).

For instructions on upgrading from a previous release, see [Upgrade documentation.](./upgrade.md)


## Introduction

This documentation covers the installation of the PRISMA server and client software for use in an RCC deployment. This guide will only cover a single-server install at this time. It will be updated in the future with instructions for a multi-server install (which will entail thorough details on how to configure a application server and database setup).

Even though the system is branded as "PRISMA RCC", in this document we will be going over both the backend server-side which is referred to as TMS as well as PRISMA which is the front end client. The name "TMS" (was originally derived from our legacy software "Trident Maritime System") and "C2" (stands for "Command and Control" but is not used as Client name any more. Client is simply `PRISMA Client`).

## Requirements

Before starting, be sure that you have the following:
 - PRISMA Client installation EXE (Windows Installations)
 - PRISMA_Server_Install-X.X.X installation package
 - Ubuntu 16.04 installation media
 - Server for the backend TMS software
 - Windows workstation(s) for the PRISMA client software
## RCC Overview


Messages from the MCC are sent to the RCC using the FTP protocol. A standard FTP server, vsftpd, places incoming files under the /srv/ftp directory. The tmccd (Trident Mission Command Center Daemon) linux process will be listening in for any updates that come in. When there is a new incoming message, tmccd will parse that particular message into a TMS track object. Which then passes that track object to the tgwad (Trident Gateway Access Daemon) linux process and deletes the incoming message file from the FTP directory. If tmccd is unable to parse a message, an error will be written to the tmsd.log and the message will not be removed.

The tanalyzed (Trident Analytics Engine Daemon) linux process is designed to subscribe to the tgwad linux process. It's basically looking at each track object that enters the system and if certain conditions are met, tanalyzed generates an alert which prompts the user to take some action. In the case of an RCC single-server install, an alert is generated any time a new beacon is detected. The tdatabased (Trident Database Daemon) linux process also subscribes to tgwad and inserts all new track(s) into our database. The database instance is managed by the tdatabased linux process. When the tdatabased linux process starts, it starts the database and when tdatabased linux process stops, it stops the database.

The PRISMA client interacts with the backend database via twebd (Trident Webserver Daemon) linux process. The client must first authenticate with the tauthd (Trident Authentcation Daemon) linux process, if successful then a token will be passed to the client which is then used to communicate with twebd. The tauthd linux process subscribes to any changes to tracks, alerts, incidents etc., etc. and sends those changes to the PRISMA client though a web socket. Requests from the PRISMA client are served using a standard REST interface and data that needs to be persistent is stored in the tdatabased.

## Server Installation

### Machine Settings

 - Configure storage arrays to use RAID 1+0 when there are at least four disks, otherwise use RAID 1.
 - Ensure the machine is configured to boot in UEFI mode.

### Ubuntu 16.04 Installation

Install Ubuntu 16.04 normally but with the following considerations:
  - Server should be configured manually with a static IP address or with a dedicated IP address through DHCP.
  - Standalone systems that are not on the Internet do not require DNS. Leave the address for the DNS server blank.
  - The recommend hostnames are as follows:
    - prisma: single-server server
  - The recommended username for the administrator should be "orolia". We don't want people using "prisma" (due to the fact that user will be reserved for the TMS during runtime).
  - Do not encrypt the home directory
  - See the section below on partitioning the disks
  - If you are asked to force UEFI boot mode, please select "Yes"
  - Please make sure to select "No Automatic Updates"
  - At the package selection screen, make sure to select "OpenSSH Server"


#### Partitioning

When installing from a USB stick, only two devices should appear when selecting the target for the OS installation. Be sure to select the correct device–the USB stick is usually listed first with the server RAID array listed second. If the RAID is listed second, the device should be /dev/sdb.

Application servers can be partitioned by selecting "Guided - Use entire disk and setup LVM".
single-server servers or dedicated database servers should be partitioned manually. Create the following standard partitions:

| Size      | Use As                  |   Mount              |
| --------- |:-----------------------:|:--------------------:|
| 512 MB	  | ext2	                  | /boot                |
| 512 MB	  | EFI                     | System Partition     |
| Remaining	| Physical Volume for LVM |                      |

Now select "Configure the Logical Volume Manager". Create a single volume group with the same name as the machine's hostname. Select the physical volume that was created for LVM. It should be the third partition on the RAID array and it should have a large amount of space. This is usually device `/dev/sdb3`.

Create three logical volumes, exit back to the partioning menu, and assign mount points, filesystem types, and options for the newly created logical volumes:

| Name  | Size      | Use As |   Mount      | Options  |
| ----- |:---------:|:------:|:------------:|:--------:|
| swap	| 32GB      | swap   |
| root	| 32GB      | ext4	 | /
| db	  | Remaining | XFS	   | /var/trident	| noatime

Before leaving the partioning menu, double check and verify the configuration. Do not forget to set the noatime option for the `/var/trident partition`. When the OS is installed and running you can double check your partitioning by evaluating the output of running:

```
df -aTh
```


### TMS Installation

Once you are logged in as the orolia user, you are ready to run the PRISMA installer package. The installer package should be named `PRISMA_Server_Install-X.X.X` where `X.X.X` is the version you are installing.

Run the installer with elevated permissions (required):

```
sudo sh PRISMA_Server_Install-1.8.0
```

!!! note
    This will install `tms` `tms-db` and `tms-mcc`. At this time if you do not need the `mcc` or `db` packages installed on this server, then you may remove them when the installation is complete using:

    ```
    sudo apt remove tms-mcc tms-db
    ```

    In future versions, the installer should ask up front which applications to install to avoid this issue.

!!! warning "If mongo db installation fails"
    The mongo db installation can fail due to a GPG check failure. If this happens, just run the following command then retry the install:

    ```
    sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv 2930ADAE8CAF5059EE73BB4B58712A2291FA4AD5
    ```

    This should only happen if the server was already configured to get mongo directly from the mongo Ubuntu repository, and should not be an issue for installs using the PRISMA Installer package.

### TLS Certificates

The PRISMA client application communicates with TMS using a secure connection with TLS. The twebd and tauthd linux processes on the application server will only work with a valid vertifcate. This single-server install will install a default certificate for "localhost" (the path for the certificate is):

```
/etc/trident/certificate.pem
```

The default certificate comes with a key, which is located:

```
/etc/trident/key.pem
```

For production installations, please obtain a signed SSL certerificate then install the certificate and keys as described in [SSL documentation.](./ssl.md). For development systems, you may run `tselfsign` to create a self signed certificate. Documentation for this is also in the [SSL documentation.](./ssl.md)

### TMS Configuration

Single-server installs have a default configuration, which lives in a configuration file referred to as:

```
/etc/trident/tmsd.conf
```

Let's create the configuration file for the PRISMA RCC single-server install. Pleae run the following command below:

```
sudo vi tmsd.conf
```

Press the "i" key on the keyboard and enter the following content:

```
{tgwad}
{tmccd}
{tdatabased}
{twebd}
{tanalyzed}
{tauthd}
```

This configuration file will get a basic system up and running, but for more complex systems more configuration will be required.

### Configure Mongo

Edit the mongod.conf file and create the replication.js file to setup mongo.

See the [mongo configuration tutorial](../configuration/mongodb-conf.md) for instructions on setting up mongo.

### Insert Client Configuration

We need to insert the client configuration into the database so the PRISMA client will get the correct information to talk to the backend along with setting default map coordinates, site information, etc...

!!! todo
    Document how to inject a corrected config into the system, and provide a script to easily do this. Including, how to set lat/lng/zoom of map, policy descriptions, brand name.

###  Validate that the FTP service is configured properly

For `tms-mcc` installations, you will need to ensure the ftp drop from the MCC is configured and working correctly so beacon alerts are processed by the TMS server.

On the application server, validate that the FTP service is running:
```
ftp localhost
```

Login with user "test" and password "test". Exit with "quit".

Create an FTP user which the MCC will use to send messages. The commands listed below use a username of "mcc". Change this to the desired username of your choice.
```
sudo htpasswd -d /etc/trident/vsftpd.passwd mcc
```

When entering a password be aware that it is neither stored nor transmitted securly and can be discovered. Do not use a password that is important and in-use elsewhere.

Create a directory where the FTP files will be deposited. The directory name should match the name of the MCC FTP user.
```
sudo install -o vsftpd -g prisma -m 2775 -d /srv/ftp/mcc
```

Verify a file can be uploaded with the following:
```
echo "hello world" > hello
ftp localhost
---login---
put hello
quit
cat /srv/ftp/mcc/hello
sudo rm /srv/ftp/mcc/hello
```

Remove the test user with:
```
sudo htpasswd -D /etc/trident/vsftpd.passwd test
```

#### Validate access to FTP from outside

Verify that FTP port is opened on network settings (either port 22 for sftp or port 21 for ftp). For servers in AWS, navigate to Network & Security > Network Interfaces and find the server in question. Under the details tab/security groups, there will be a link to view inbound rules, verify that the appropriate port is enabled for the security group being used.

Test FTP access from outside using the credentials created in the steps above.

### Common Installation Configurations

In addition to a single server installation as described above, PRISMA can be configured to work with replication servers and DB servers on separate systems. The following section will describe to configuration, anatomy, and other details of these additional setups.

#### Multi Server RCC Installation

!!! todo
    Document the anatomy of a server install with a replication server on site, eg similar to CHILE, and why it's different than a single server installation.

#### Multi Server RCC Installation with DB on Separate System

!!! todo
    Document the anatomy of a server install with a replication server on site and database on a completely separate server, eg similar to Singapore, and why its different.

## Initial Start

Start TMS with the following:
```
sudo systemctl start tms.service
```

Perform the steps under "Verifying TMS services" in the "Troubleshooting" section.

## Feature Configuration

### Incident Transfer

To enable communication between multiple installations:

1. Configure tgwad with the IP and ports.  This creates the communication channel.

```
tgwad --num 1 --name Iquique -site Puerto,2,tcp:10.106.143.12:31229 -site Talcahuano,4,tcp:10.206.243.62:31228
```

2.  Configure `/site` endpoint with information about each site.  This provides the end-user with information about a site, mainly the location so it shows on the map.

!!!TODO
    provide script to upload JSON like tconfig and tpolicy example site.json

After completing the above, you will see the sites on the map and a color indicator will show the connection status, green=ok, yellow=bad

# Operations

## RCC End-to-End Test

Launch the PRISMA client and login to PRISMA RCC. Please validate that the lat/long coordinates entered in the config file is is where your map view routes to after you login to the windows client.

Login to the applcation server and upload a sample message to the FTP server:
```
cd /usr/share/tms-mcc
ftp localhost
--login--
put sample.xml
ren sample.xml sample.txt
quit
```

A beacon should now appear on the map, keep in mind this is a simulation and is not a real beacon.

## Replay Test

Note: This should only be done on an empty database that can be deleted after use.

  * Install the tms-dev debian package.
  * Download the mcc-xml-capture.tar.gz data file and install with the following:
    ```
    tar xf mcc-xml-capture.tar.gz
    sudo cp -r capture /srv/capture
    sudo chown -R prisma:prisma /srv/capture
    ```
  * Add the following to /etc/trident/tmsd/conf
    ```
    {tmccrd
      --src-dir /srv/capture
      --dst-dir /srv/ftp/mcc
    }
    ```
  * Restart TMS:
    ```
    sudo systemctl restart tms.service
    ```

Data does not stream quickly, but after a few minutes, you should see some becaons from the replay data.

## Starting and Stopping TMS Services

Start TMS with the following:
```
sudo systemctl start tms.service
```

Stop with:
```
sudo systemctl stop tms.service
```

Restart with:
```
sudo systemctl restart tms.service
```

## Clearing the database

When testing the system before it is operational, it may be necessary to clean out the database to remove any test data. On the database machine, execute the following:

```
sudo systemctl stop tms.service
sudo systemctl stop mongo.service
sudo rm -rf /var/trident/db/*
sudo systemctl start mongo.service
mongo mongodb://localhost:8201/aaa /usr/share/tms-db/aaa.js
mongo mongodb://localhost:8201/trident /usr/share/tms-db/trident.js
sudo systemctl start tms.service
```


This should not be done on any database that has real data that must be preserved.

# Troubleshooting

## TMS Log Files

A log file of important TMS events can be found at `/var/log/tmsd.log`. To watch in real-time as items are logged, run the following:
```
tail -f /var/log/tmsd.log
```

To search the log for messages from a specific daemon:
```
grep twebd /var/log/tmsd.log
```

This log file only contains the log entries for the current day. The previous day will be in tmsd.log.1 and six more days can be found in tmsd.log.2.gz, tmsd.log.3.gz, etc. Gzipped logs can be viewed with:
```
zcat /var/log/tmsd.log.*.gz
```

and searched with:
```
zcat /var/log/tmsd.log.*.gz | grep twebd
```

In normal operations, the log file should be fairly quiet. If a lot of entries are being quickly written to the log regulary, this usually indicates a problem with the system.

## Verifying TMS Services

Check the status of the TMS services with `tmsd --info`

The output should look similar to the following:

```
pid  name       status   started  the last successful start  command line args
2739 tgwad      running  1 times  2017-10-31 14:42:25 UTC     --num=1 --name=rcc
2729 tmccd      running  1 times  2017-10-31 14:42:25 UTC
2719 tdatabased running  1 times  2017-10-31 14:42:25 UTC
2744 twebd      running  1 times  2017-10-31 14:42:25 UTC
2717 tanalyzed  running  1 times  2017-10-31 14:42:25 UTC
2714 tauthd     running  1 times  2017-10-31 14:42:25 UTC
```

Verify that all processes in the tmsd.conf file appear in this list and that every status shows "running". If the process has started more than once, that indicates that it crashed and was restarted. Check the TMS log files for the error.

On the application server, verify twebd is working correctly with:
```
curl --insecure https://localhost:8080/api/v2/apidocs.json
```

An extensive JSON document should be returned without error.

On the database machine, connect to the mongo instance with:
```
mongo localhost:8201/trident
show collections
quit()
```

There should be one startup warning, "Access control is not enabled for this database".  Any other warnings or errors should be resolved.

## PRISMA Client Debugging

If the PRISMA client is misbehaving or is otherwise not working, check the contents of the web console. Start the application from the command line with:
```
"C:\Program Files\PRISMA-win-x64\PRISMA" -- --devtools
```

The application should now start with the development tools attached to the right side of the application window. The console tab should show any errors generated by the applciation. Note that on startup, the client always tries to connect to tauthd to see if it has a valid session and will fail. This is normal. Also check the network tab to monitor network activity.

NOTE: The exe is probably not in Program Files. If you right click on Primsa in the Start menu and
select show in Explorer that will take you to a shortcut where you can right click again and select
get info to see the actual location of the EXE file. Open Power Shell and `cd` to that directory to
then run the command above.

## Track counts

From the database server, query the database to see how many tracks have been ingested into the system with:
```
mongo localhost:8201/trident
db.tracks.count()
quit()
```

As new data comes in, this number should go up.

## Incoming MCC messages

Messages that have been properly handled should be cleared out from the incoming FTP directory. Look for invalid messages that may not have been ingested with:
```
ls -R /srv/ftp/*
```

## Permissions

If the log file indicates that services cannot open files because permission is denied, reset permissions with the following commands:
```
sudo chown prisma:prisma /tmp/tgwad* /tmp/tmsd*
```

On database servers, also run the following commands:
```
sudo chown -R prisma:prisma /var/trident/db
```


## Common Problems

### "Service Unavailable" is displayed when starting PRISMA.

This appears before the login screen.
The client is unable to connect to twebd on the applcation server to get the configuration.

  - Run PRISMA as described in "PRISMA Client Debugging" section. Are there any errors printed to the Windows command line window?
  - Is the PRISMA `~\.c2\production.json` file configured with the correct IP address of the application server?
    ```
    {
      "configurationRemote": true,
      "configuration": "http://<IpAddress>:8081/api/v2/config.json"
    }
    ```
  - Can the workstation can ping the application server?
  - Is the twebd service running on the application server?
  - Can the URL http://X:8081/api/v2/config.json (where X is the IP address of the application server) can be visited in a web browser?
  - Are there any errors in the TMS log?

### "Server unreachable. Contact adminstrator" is displayed when starting PRISMA.

This appears on the login screen. The client is unable to connect to tauthd on the application server.

  - Run PRISMA as described in "PRISMA Client Debugging" section. Are there any errors in the web console?
  - Are the TLS certifcates have been properly installed?
  - Does visiting https://X:8080/api/v2/apidocs.json (where X is the IP address of the application server) should return a JSON document with the browser indicating that a secure connection?
  - Is the tauthd service is running on the application server?
  - Are there any errors in the TMS log?

### No items appear on the map.

  - Run PRISMA as described in "PRISMA Client Debugging" section. Are there any errors in the web console?
  - Are there any errors in the TMS log?
  - Are all TMS services running?
  - Is the FTP server running? Can a file be updated to the FTP server?
  - Are there any unprocessed messages in the FTP directory?
  - Are tracks being addded to the database?
  - Are the servers set to the UTC timezone?

### Beacons appear on the map but no alerts are generated.

The tanalyzed daemon is not working properly.

  - Is the tanalyzed daemon running?
  - Are there any errors in the TMS log?

##  Beacons disappear after fifteen minutes.

The tanalyzed daemon is not working properly.

  - Is the tanalyzed daemon running?
  - Are there any errors in the TMS log?

## Changes for Next Version
  - Multi-server install: Application Server/Database install. This will have detailed setup of an Application Server & Database install.
  - A section that goes over the detailed steps needed for an upgrade with no data loss for our customers.
  - Recover administrator password: There is no way to recover the administrators password, once the admin user is locked you have to wipe the users collection & add the admin user in mongo again).
  - Update this wiki for config file: (how to address PRISMA Fleet config & PRISMA RCC config).
