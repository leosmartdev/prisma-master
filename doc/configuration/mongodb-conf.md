# How To: Mongod Configuration File

You can configure mongod instances at startup using a configuration file. Information about the configuration file format and usage can be found in mongodb's documentation website. See [Configuration File Options](https://docs.mongodb.com/manual/reference/configuration-options/).

To configure mongo there are two parts, updating the `/etc/mongod.conf` sections as described below, and injecting the replication.js as also described below. 

After configuring be sure to restart mongo and then verify you can connect using the command line client. 

```
sudo systemctl restart mongod
```

To verify using the client `mongo <ipaddress>` should show a mongo prompt for each databse server with `PRIMARY>` or `SECONDARY>` as the prompt. If the primary and secondary are switched, just type `rs.stepDown()` in the mongo prompt labeled `PRIMARY>` that should be secondary server. 

## Sections

The following example is a mongodb configuration used in our developement environment. 

!!! info
    MongoDB configuration file uses YAML format which does not support tab characters: use spaces instead.

```yaml
storage:
  dbPath: /var/trident/db/

replication:
  replSetName: replocal

net:
  bindIpAll: true
  port: 27017

systemLog:
  destination: syslog
```

!!! info
    This example should not be used as a general standard, mongod should be configured based on specific project requirements.

Your config may have more options and comments, but the important lines to ensure PRISMA applications work correctly are the following sections. Be sure the entire file is configured before restarting mongo `sudo systemctl restart mongod`. 

### storage

We must specify the storage location for the database, which for PRISMA is `/var/trident/db`. So ensure the `dbPath` option is set to this: 

```yaml
storage:
  dbPath: /var/trident/db/
```

### net

Next, the net section is how we tell mongo what to listen to connections on, and also, if we are using SSL certifications, how to set those up. 

There are a few ways to accomplish this, the easiest way is to tell mongo to listen for connections to all ip addresses. 

```yaml
net:
  bindIpAll: true
  port: 27017
```

The second way is to white list IP Addresses mongo will run on, 

```yaml
net:
  bindIp: 127.0.0.1,10.140.100.101
  port: 27017
```

This config for instance, would listen to all connections trying to connect to 127.0.0.1 and 10.140.100.101. This is the ip address of the server mongo is running on, NOT the server the connection is coming from. So the IPs here would match a connection request using the mongo CLI like the following: `mongo mongodb://10.140.100.101/trident`.

Important, you must  use `bindIp` OR `bindIpAll`. Using both will cause mongo to error on start. 

### replication

PRISMA requires replication, so you must always have this section in the mongod.conf. It is usually commented out by default. 

For all production environments, the replication set name will be rs0. 

```yaml
replication:
  replSetName: rs0
```

This must always be paired with a `replication.js` that sets up the replication configuration. This is not enabled by default, so you will need to create `replication.js` file in `/etc/trident/db` (vv1.7.7 or later) or `/usr/share/tms-db/` (1.7.6 and earlier) that has the following entries (replace `<ipaddress rcc1X>` with the IPs of the database primary server for index 0 and secondary server for index 1. 

```js
var config = {  
  _id: "rs0", 
  members: [
    { _id: 0, host: "<ipaddress rcc1a>" },
    { _id: 1, host: "<ipaddress rcc1b>" }
  ]
};

rs.initiate(config);
```

On the next restart of tmsd, it will load this file into mongo. You can also just save it in the home directory and run `mongo localhost ~/replication.js` to manually insert the file. 

## Authenticated Mongo and SSL Certificates

For additional configuration to setup mongo to use SSL and Authorization, see the [SSl Page on configuring mongodb with SSL.](../installation/ssl/#configure-the-backend-to-use-ssl-with-mongodb)

## Toubleshooting

1. Running `mongo` says `connection refused`. 

    In general, this means mongo isn't started or didn't start correctly. First run `sudo systemctl status mongod` and see the output. If you see 2/INVALIDARGUMENT then the config file format is invalid. Check that the properties match the names above, replication set name is not in quotes, and that there are no tab characters in the file, only spaces. 

# How to: Change Default MongoDB Port

!!! info
    The following [link](https://docs.mongodb.com/manual/reference/default-mongodb-port/) lists the default TCP ports used by MongoDB

Changing default port in which MongoDB listens is a common practice to hide a server from spiders, crawlers and automated attacks. 

1. Open /etc/mongod.conf with your favorite code editor and search for the following lines:
```yaml
# network interfaces
net:
  port: 27017
```
Now change the port number to any other of your choice, save and reload mongod: 
```bat
orolia@prisma:~$ sudo systemctl restart mongod
orolia@prisma:~$ sudo systemctl status mongod
● mongod.service - High-performance, schema-free document-oriented database
   Loaded: loaded (/lib/systemd/system/mongod.service; enabled; vendor preset: enabled)
   Active: active (running) since Mon 2019-06-24 16:58:03 UTC; 2s ago
     Docs: https://docs.mongodb.org/manual
 Main PID: 2836 (mongod)
    Tasks: 14
   Memory: 136.0M
      CPU: 1.659s
   CGroup: /system.slice/mongod.service
           └─2836 /usr/bin/mongod --config /etc/mongod.conf

Jun 24 16:58:03 prisma systemd[1]: Started High-performance, schema-free document-oriented databas
```
Repeat the same for every mongod.conf file in the replication cluster.

!!! info 
    Take into account that the chosen port needs to be equal or greater than 1000 and must not be taken by any other service running in the same host. You can check if a certain port is already in use with: nc -l port_number.
    for this tutorial we assume port 30000 is available. 

2. Next, to reconfigure an existing replica set, overwriting the existing replica set configuration is needed. To run the procedure, you must connect to the primary of the replica set. Connect into mongod using mongo and the new configured port:
```bat
mongo mongodb://server-ip:30000
```
3. In order to verify the existing replica configuration use:
```js
printjson(rs.config())
```
The result should look like: 
```json
{
   "_id" : "rs0",
   "version" : 1,
   "protocolVersion" : NumberLong(1),
   "members" : [
      {
         "_id" : 0,
         "host" : "mongodb0.example.net:27017",
         "arbiterOnly" : false,
         "buildIndexes" : true,
         "hidden" : false,
         "priority" : 1,
         "tags" : {

         },
         "slaveDelay" : NumberLong(0),
         "votes" : 1
      },
      {
         "_id" : 1,
         "host" : "mongodb1.example.net:27017",
         "arbiterOnly" : false,
         "buildIndexes" : true,
         "hidden" : false,
         "priority" : 1,
         "tags" : {

         },
         "slaveDelay" : NumberLong(0),
         "votes" : 1
      }
   ],
   "settings" : {
      "chainingAllowed" : true,
      "heartbeatIntervalMillis" : 2000,
      "heartbeatTimeoutSecs" : 10,
      "electionTimeoutMillis" : 10000,
      "catchUpTimeoutMillis" : 2000,
      "getLastErrorModes" : {

      },
      "getLastErrorDefaults" : {
         "w" : 1,
         "wtimeout" : 0
      },
      "replicaSetId" : ObjectId("58858acc1f5609ed986b641b")
   }
}
```

The following sequence of operations updates the members[n].host of the first and second member. The operations are issued through a mongo shell connected to the primary only.

```js
cfg = rs.conf();
cfg.members[1].host = "mongodb0.example.net:30000";
cgf.members[2].host = "mongodb1.example.net:30000";
rs.reconfig(cfg,{force: true});
```
!!! info
    The configuration changes can be checked using: rs.status()