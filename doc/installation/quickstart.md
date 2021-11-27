# Quickstart

To get started with PRISMA C2 on a new system quickly, you can follow the following procedures.

## Server Installation

Download the [PRISMA Installer][PRISMA_INSTALLER] on an Ubuntu 16.04 installation.

```bash
sudo sh PRISMA_Server_Install-<version>
```

## Configure Server

### tmsd.conf

Once installed, create a `tmsd.conf` file:

```bash
sudo vi /etc/trident/tmsd.conf
```

Add the following to that config file:

```
{tgwad}
{tmccd}
{tdatabased}
{twebd}
{tanalyzed}
{tauthd}
```

### mongod.conf and replication

Edit the mongod.conf file and create the replication.js file to setup mongo.

See the [mongo configuration tutorial](../configuration/mongodb-conf.md) for instructions on setting up mongo.

## Start Server

```
sudo systemctl stop tms.service
sudo systemctl start tms.service
```

## Install Client

Download the [Windows client installer.][PRISMA_INSTALLER] and using PowerShell create a `.c2` directory.

```bash
cd ~\
mkdir .c2
```

Create a file in the .c2 directory called `production.json` and add the following to that file ensuring that you add the correct IP for your installed server:

```json
{
    "configurations": [{
        "default": true,
        "name": "PRISMA RCC",
        "url": "http://<IPADDR>:8081/api/v2/config.json"
    }]
}
```

Double click the windows installer to install and run the application.

!!! note "If you see an error after the installer runs"
    If you don't create the configuration file in `~/.c2` first, the installer when completed will show an error that the configuration could not be loaded. This means that the installation did succeed, but after install it tried to run the client and couldn't find the configuration to reach the server. Just create the `~/.c2/production.json` and start the client from the start menu.

[PRISMA_INSTALLER]: https://documentation.mcmurdo.io/releases/stable/
