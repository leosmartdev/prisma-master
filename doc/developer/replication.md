# PRISMA C2 Replication

For most installations, PRISMA is deployed on two servers, a primary and a backup server. In some installations, like Singapore, the databases are also on separate servers as well, so you have a total of 4 servers, but logically, the primary database and application are treated as a single pair, so all of the following methods and setups will apply.

## Normal Operations

For normal operations, the PRIMARY server (or primary application server if databases are separate) is running the tms application daemons. It is also running the PRIMARY mongo instance. The SECONDARY server is running the SECONDARY mongo instance and the tms application is stopped. 

## Failover process

If a failover is to occur, mongo will also failover automatically. But you can force a failover by logging into the primary mongo instance and using the command `rs.stepDown()` which will immediately cause the PRIMARY to step down to SECONDARY and SECONDARY to become primary. When mongo switches however, you will have to manually update the configuration in the database for the client to know which application server to connect to. So for installations, log into the secondary server and run 

```
mongo <secondary-mongo-host>/trident /path/to/secondary_config.js
```

Usually we create a secondary config file in the home directory of the secondary server. 

The application is currently a manual failover operation. You will need to log into the secondary application server and start tms, `sudo systemctl start tms` and if the primary server is still running then use `sudo systemctl stop tms` on the primary. You should run the config injection script above before starting the new application server. 

Now, to finish the failover, shutdown the client and use the PRISMA Secondary link on the desktop to open the application and connect to the secondary server. 

## Setup Replication

Set the replication set name to `rs0` in `/etc/mongod.conf`

```
replication:
  replSetName: "rs0"
```

Then, in mongo, initiate the config. 

```
$ mongo
> var config = {  
  _id: "rs0", 
  members: [
    { _id: 0, host: "<IPADDRESS PRIMARY DB>" },
    { _id: 1, host: "<IPADDRESS SECONDARY DB>" }
  ]
};
> rs.initiate(config);
```
