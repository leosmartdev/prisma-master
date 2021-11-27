# PRISMA Upgrade Guide

This guide will go over performing an upgrade from previous versions of PRISMA.

!!! warning
    Currently, the PRISMA upgrade process MAY RESULT IN DATA LOSS. There is no gurantee data will be
    protected.

## 1.7.3 -> 1.7.4

This release has a number of changes related to how the system interacts and sets up mongo. Because
of these changes, there are a number of things that must be done to ensure the upgrade runs
smoothly.

If you had performed the upgrade instructions for 1.7.1 -> 1.7.2, you will need to manually remove
the mongo.service file in `/etc/systemd/system/mongo.service`. In 1.7.4 we are no longer using a
custom mongo.service file, we are using the default provided by mongo
`/lib/systemd/system/mongod.service` and all configuration changes must now be made in the mongo
provided `/etc/mongod.conf`.

This means, before upgrading, you will need to update the mongod.conf file to setup replication,
ip addresses, database location, and any other customized settings from the old `mongo.service`
file.

### Steps

* Stop tms and mongo.
  ```
  sudo systemctl stop tms
  sudo systemctl stop mongo
  ```

* Remove any custom mongo.service files in `/etc`
  ```
  rm /etc/systemd/system/mongo.service
  ```

!!! caution
    There are two mongo service files, the only difference is a trailing d before the `.service.
    The TMS version is `mongo.service` with no `d` and this is the file we want to remove. Do not
    remove any service files named `mongod.service` as these are the mongo provided service files
    that will be used after the update.

* Update `/etc/mongod.conf` to configure mongo. For current options being passed to mongo see
  `/lib/systemd/system/mongo.service`. In general, the following are the added sections that will
  need to be updated if you are using a default installation with replicated servers:
  ```yaml
  storage:
    dbPath: /var/trident/db
  ...
  net:
    port: 27017
    bindIp: 127.0.0.1, <serverIP Address>
  ...
  replication:
    replSetName: rs0
  ```

!!! note
    It is advised that if you will be using mongo authentication when updating to 1.7.4 you first
    perform the upgrade and ensure mongo is working properly with the new configuration before
    adding the security and authentication sections, especially since you will need to add the new
    user credentials to the $external database before restarting mongo with the security turned on.

* Update ownership of database files. Since we are now using default mongo lifecycle, the user that
  will be owning the database folder is the `mongodb` user not the `prisma` user. So we will need
  to the chown the `/var/trident/db` directory to ensure mongo can still access the database.
  ```bash
  sudo chown -R mongodb:mongodb /var/trident/db
  ```

* Ensure there are no mongo lock files in tmp. If there are any, remove them.
  ```
  $ ls /tmp | grep mongo
  mongodb-27017.sock
  rm /tmp/mongo-27017.sock
  ```

At this point you can now install 1.7.4. 

## 1.7.2 -> 1.7.3

There are no special instructions for this upgrade.

## 1.7.1 -> 1.7.2

The are two things that need to be done for an upgrade from 1.7.1 to 1.7.2.

First, copy the mongo.service file into /etc so the 1.7.2 package does not overwrite replication
setup for mongo.

```
sudo cp /lib/systemd/system/mongo.service /etc/systemd/system/mongo.service
```

Also, remove the wkhtmltopdf that was installed in 1.7.1 so the older version that comes with
the installer package will correctly install the right wkpdftohtml.

```
sudo apt-get remove wkhtmltopdf
```

## 1.5.0 -> 1.7.2

First, verify you are upgrading from a 1.5 installation:

```bash
twebd --version
```

Also, check if the `tms-dev` package is installed, and if it is we need to remove it. If `tms-dev`
is installed the package installation will fail.

```
sudo apt list tms-dev
```

If installed, you will see `[installed]` in the output. To remove the package:

```
sudo apt-get remove -y tms-dev
```

### Update tmsd.conf

We need to open the `tmsd.conf` file and ensure the `tdatabased` line does not include `-mongo-data`
flag. If it does, the tms services will fail to start.

```
sudo vim /etc/trident/tmsd.conf
```

If the tdatabased line looks like this:

```
{tdatabased -mongo-data /var/trident/db }
```

Change it to remove thr `-mongo-data` flag:

```
{tdatabased}
```

### Install Server

Get the installer package onto the server and run the installer:

```
sudo sh PRISMA_Server_Install-1.7.1
```

!!! warning "If the installer fails"
    If the installer fails, run the following command then re-run the installer package as directed
    above.

    ```
    sudo chown -R prisma:prisma /var/trident/db
    ```

### Fix replication

If the setup includes replication, we will need to modify the `mongo.service` file to ensure the
replication name is set correctly. So open `/lib/systemd/system/mongo.service` and ensure the `--replSet` flag  on the ExecStart line does not say `replocal`.

```
ExecStart=/usr/bin/mongod --syslog --syslogFacility local0 --dbpath /var/trident/db --port 27017 --bind_ip_all --replSet replocal
```

The `replocal` should be changed to the name of your replica set, usually `rs0`.

Do this on all servers in the replica set.

!!! warning
    Before proceding, make sure to stop both servers fully, mongo and tms services.

    ```
    sudo systemctl stop mongo.service
    sudo systemctl stop tms.service
    ```

On the primary system, restart mongo:

```
sudo systemctl daemon-reload
sudo systemctl start mongo.service
```

Now, we need to change the ports for the replica set as they will be configured for the incorrect
port since the port changed in 1.7. You will only have to do this on the primary system, once the
replica set is configured correctly again, the primary should share the configuration with the
secondary. Open a connection to mongo using:

```
mongo
```

One in the mongo shell, use the following commands, and replacing the IP addresses with the correct
address for your system:

```
var config = rs.config()
config.members[0].host = "<ipaddress primary>:27017"
config.members[1].host = "<ipaddress secondary>:27017"
rs.reconfig(config, {force:true})
```

!!! note
    During this process, logging into mongo may show the SECONDARY server as PRIMARY. If this happens, log into the secondary (one marked PRIMARY), open mongo from the commnd line and type `rs.stepDown()`.

    This will tell the secondary to step down as the primary and the primary should then pick up again as the primary.

### Configuration

Last thing on the server we need to do is update the configuration in the database. With version 1.7.1,
the configuration sent to the client is stored in the `config` mongo collection, and needs to be injected
so it has the right IP addresses and other configuration information.

Take the `config.json` file for the system you are installing and run the following command to inject
the config into the system and restart tms service so the changes take effect:

```
mongo localhost/trident config.json
sudo systemctl restart tms.service
```

!!! warning
    The config.json is specific to every system, so insure you are installing the correct config
    for that IP address.

### Troubleshooting

If the admin account is locked and there is not another administrator account on the system, then
you will have to log into mongo and remove all the accounts and reinitialize.

```
mongo localhost/aaa
> db.users.remove({}):
> exit
mongo localhost/aaa /usr/share/tms-db/aaa.js
```

### Client Upgrade

For your installation, copy the new `production.json` to `~\.c2\production.json`.

Double click the installer to upgrade to the latest PRISMA Client.

## 1.5.0 -> 1.6

### Pre-Installation

Verify
 - `mongo --version`  MongoDB shell version v3.6.3
 - `twebd --version`  TMS twebd version: 1.5.0

### Server Installation

#### wkPDF Installation

Install
```bash
sudo apt-get update
sudo apt-get install xvfb libfontconfig wkhtmltopdf
```

Verify
 - `wkhtmltopdf --version` wkhtmltopdf 0.12.4

#### Redis Installation

Install
```bash
sudo apt-get update
sudo apt-get install redis-server
sudo systemctl restart redis-server.service
sudo systemctl enable redis-server.service
```

Verify
  - `redis-cli --version` redis-cli 4.0.9
  - `redis-cli ping` PONG

#### Consul Installation

Install
```bash
wget https://releases.hashicorp.com/consul/1.0.7/consul_1.0.7_linux_amd64.zip
unzip consul*.zip
sudo mv consul /usr/local/bin
rm consul*.zip
```

Verify
 - `consul --version` Consul v1.0.7

### TMS Installation

Stop
 - TMS `sudo service tms stop`
 - MongoDB `sudo service mongod stop`

Install
```bash
sudo dpkg -i tms_1.6.0-0.*_xenial_amd64.deb
sudo dpkg -i tms-db_1.6.0-0.*_xenial_amd64.deb
sudo dpkg -i tms-mcc_1.6.0-0.*_xenial_amd64.deb
```
[Installation Guide](installation.md#tms-install)

Verify
 - `tmsd --version` TMS tmsd version: 1.6.0

#### MongoDB Schema Update

1. Update MongoDB configuration (no tabs)
```yaml
replication:
  replSetName: replocal
```
2. Start MongoDB `sudo service mongod start`
2. Run schema
```bash
mongo mongodb://localhost:8201 /usr/share/tms-db/replication.js
mongo mongodb://localhost:8201/aaa /usr/share/tms-db/aaa.js
mongo mongodb://localhost:8201/trident /usr/share/tms-db/trident.js
```

#### Configuration Update
 - tmsd.conf
 - ~/.c2/production.json
